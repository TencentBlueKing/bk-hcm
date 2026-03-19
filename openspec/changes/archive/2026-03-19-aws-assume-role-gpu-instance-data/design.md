## Context

HCM 当前通过 `cloud.account` 存储的 AK/SK 直接调用 AWS API 管理资源。但 AWS 场景下，`cloud.account` 存储的是一级根账号（Management Account）中 IAM User 的凭证，二级资源账号（Member Account）没有在 HCM 中录入独立 AK/SK。访问成员账号资源必须通过 STS AssumeRole 获取临时凭证。

下游 GPU 资源分析平台自行维护资源账号 CloudID 列表，通过 HCM 透传接口拉取实例数据。调用模式为逐 region 请求。

## Goals / Non-Goals

**Goals:**
- 新增 AssumeRole 能力，支持通过根账号凭证访问成员账号资源
- 提供 GPU 实例数据透传接口（实例类型含 GPU 字段 + 实例列表），含 cloud-server 资源视角入口和蓝鲸 API 网关开放接口
- 编写 GPU 接口文档
- 非破坏性：不改动现有 `Aws()` / `AwsRoot()` 方法，所有改造通过新增代码实现

**Non-Goals:**
- 不实现预留实例接口
- 不做多 region 聚合（下游平台逐 region 调用）
- 不引入 Redis 等外部组件

## Decisions

### 1. AssumeRole 链路放在哪里

**决策**：新增 `AwsWithAssumeRole` 方法到 `CloudAdaptorClient`，不改动现有 `Aws()` 和 `AwsRoot()`。

**理由**：改造现有方法影响 30+ 个已有接口，风险过高。新增方法遵循开闭原则，只有新接口走 AssumeRole 路径。

**否决方案**：改造 `Aws()` / `AwsRoot()` 方法 —— 影响面太大。

### 2. Role Chain（角色链式调用）

**决策**：由下游平台通过 `role_chain`（角色名数组）参数透传，HCM 按数组顺序执行 N 步 AssumeRole 链式调用。中间角色在管理账号中 AssumeRole，最终角色在目标成员账号中 AssumeRole。每步用目标账号 CloudID + `site` 自动拼接完整 ARN。

**理由**：AWS 运维架构中常用 Role Chaining 模式增强安全控制（如先 Assume 一个 Caller 角色获取受限权限，再 Assume 最终 ReadOnly 角色访问资源）。使用数组而非单一 `role_name` 可灵活支持 1 步到 N 步的 AssumeRole 场景。AWS 官方文档确认 Role Chaining 是受支持的模式。

**否决方案**：
- 硬编码角色名 —— 不灵活，换角色名需改代码。
- 单一 `role_name` 参数 —— 无法支持多步 AssumeRole，不满足 AWS 增强安全架构需求。

### 2.1 ExternalId 支持

**决策**：所有 AssumeRole 接口新增可选参数 `external_id`，由下游平台透传。HCM 仅在 Role Chain 的**最后一步**（即 Assume 目标成员账号中的角色）时将 ExternalId 传入 STS AssumeRole 调用。中间步骤不传。

**理由**：AWS 运维部署的 ReadOnlyRole（StackSets 批量推送到各子账号）的 Trust Policy 配置了 `sts:ExternalId` 条件验证，不带正确 ExternalId 的 AssumeRole 调用会被 AccessDenied。CallerRole（管理账号中的中枢角色）的 Trust Policy 不要求 ExternalId。因此 ExternalId 只需在最后一步使用。

**设计要点**：
- `external_id` 为可选字段（`omitempty`），不传时不影响无 ExternalId 要求的部署场景
- ExternalId 不影响缓存 key：同一 roleArn 的临时凭证不因 ExternalId 不同而不同，ExternalId 仅是授权校验参数

### 3. 凭证缓存方案

**决策**：进程内内存缓存（`map` + `sync.Mutex`），缓存 key 为 `cloudAccountID + ":" + roleArn`（由 `AwsWithAssumeRole` 内部构建后传入缓存模块），过期前 10 分钟提前刷新，含降级策略。Role Chain 场景下每一步 AssumeRole 独立缓存。

**理由**：STS 调用有 ~100-200ms 延迟且有频率限制。临时凭证无需跨 Pod 共享，进程内缓存足够。降级策略在 STS 短暂不可用时仍能服务。Role Chain 各步独立缓存可最大化复用中间凭证。

**否决方案**：
- Redis 缓存 —— 引入外部组件依赖，不值得。
- 不缓存 —— STS 延迟和频率限制不可忽视。

### 4. 数据源体系

**决策**：使用 Account-Set 体系（`root_account` + `main_account`），不使用 Cloud Account 体系（`cloud.account` + `cloud.sub_account`）。参考 GCP GPU monitoring 的 `GcpRoot` + `MainAccount` 模式。

**理由**：
- 与 GCP GPU monitoring 保持一致的架构模式：`root_account` 提供凭证，`main_account` 提供目标账号标识（GCP 为 CloudProjectID，AWS 为 CloudID/AWS Account ID）
- `root_account` 表中的凭证与 `cloud.account` 实际对应同一个 AWS 管理账号，但 `root_account` + `main_account` 是 HCM 中用于账单和监控场景的标准体系
- 入参模式与 GCP monitoring 对齐：`root_account_id` + `main_account_id`，语义清晰

**否决方案**：使用 `cloud.account` + `cloud.sub_account` 体系 —— 虽然 `sub_account` 通过 Organizations API 同步可覆盖全量成员账号，但需要额外补完同步链路，且与 GCP monitoring 的模式不一致。

### 5. BaseSecret 新增 CloudSessionToken

**决策**：在 `BaseSecret` 结构体新增 `CloudSessionToken` 字段，`newClientSet` 从硬编码 `""` 改为读取该字段。

**理由**：其他厂商该字段为零值空字符串，传入 `NewStaticCredentials` 等效当前硬编码 `""`，行为不变。

### 6. GPU 数据接口入参设计

**决策**：下游平台传 `root_account_id`（根账号 HCM 内部 ID）+ `main_account_id`（二级账号 HCM 内部 ID）+ `role_chain`（角色名数组）+ `region` + 可选的 `external_id`。hc-service 用 `root_account_id` 调 `AwsRoot()` 获取根账号 AK/SK，用 `main_account_id` 查 `main_account` 表获取目标成员账号的 CloudID（AWS Account ID）。

**理由**：与 GCP GPU monitoring 的入参模式完全对齐（`root_account_id` + `main_account_id` + 业务参数），语义清晰，下游平台调用方式统一。

**否决方案**：下游传 `cloud_id` + `role_chain` + `region`，HCM 内部反查 `sub_account` 表 —— 需要额外补完 sub_account 同步链路，且与 GCP monitoring 模式不一致。

### 7. GPU 接口分层：cloud-server 入口 + hc-service 实现

**决策**：GPU 数据接口由 cloud-server 提供资源视角入口（含鉴权），内部通过微服务调用 hc-service。hc-service handler 通过 `AwsWithAssumeRole` 获取 adaptor client 后，调用已有 adaptor 方法完成数据查询：

- **实例类型查询**：handler → `client.ListInstanceType(kt, opt)` → adaptor 内部调用 `ec2.DescribeInstanceTypes`
- **实例列表查询**：handler → `client.ListCvm(kt, opt)` → adaptor 内部调用 `ec2.DescribeInstances`

同时在蓝鲸 API 网关 YAML 中注册为开放接口。

**透传策略**：
- **实例列表**（`instances/list`）采用完全透传——handler 直接返回 AWS SDK 原始 `ec2.Instance` 对象，不做字段映射和裁剪，下游可获取 AWS DescribeInstances 的全量数据。JSON 字段名为 PascalCase（AWS SDK Go v1 无 json tag），client 层用 `json.RawMessage` 接收。
- **实例类型**（`instance_types/list`）沿用 HCM 既有模式，映射到 `AwsInstanceTypeResp`（含 GPU 扩展字段）。

**理由**：遵循 HCM 四层架构规范和 adaptor 封装模式：

1. **cloud-server** 负责鉴权和路由
2. **hc-service handler** 负责请求解码、调用 adaptor、响应返回
3. **pkg/adaptor/aws/** 负责 SDK 交互，封装分页、类型转换等细节

实例列表接口定位为透传代理，下游平台需要的字段可能随业务发展变化，直接返回 AWS 原始数据避免了频繁修改 HCM 响应结构。EC2 Instance 类型规格目录（`DescribeInstanceTypes`）与账号无关，同一 Region 下所有账号返回相同数据，下游可通过 HCM 原有的非 AssumeRole 接口获取，无需每个资源账号都单独查询。

## Risks / Trade-offs

| 风险 | 缓解措施 |
| --- | --- |
| BaseSecret 新增字段影响其他厂商 | 零值空字符串等效 `""`，已验证无影响 |
| 临时凭证分页拉取中过期 | 提前 10 分钟刷新缓冲；单次 API 秒级完成 |
| STS 在 opt-in Region 未启用 | 使用全局 STS endpoint 或激活目标 Region |
| 管理账号 IAM User 缺少 STS 权限 | 前置依赖：AWS 运维确认并补充权限 |
| Role Chaining 会话时长限制 | AWS 限制链式 AssumeRole 最大会话时长为 1 小时（无法延长），单次 API 秒级完成，影响可控 |
