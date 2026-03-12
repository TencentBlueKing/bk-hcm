## Context

HCM 当前通过 `cloud.account` 存储的 AK/SK 直接调用 AWS API 管理资源。但 AWS 场景下，`cloud.account` 存储的是一级根账号（Management Account）中 IAM User 的凭证，二级资源账号（Member Account）没有在 HCM 中录入独立 AK/SK。访问成员账号资源必须通过 STS AssumeRole 获取临时凭证。

AWS 的 `sub_account` 与其他厂商不同：TCloud/HuaWei/GCP 的 sub_account 同步的是 IAM 子用户（三级），而 AWS 同步的是 Organizations 成员账号（二级）。这是因为 AWS 通过 `organizations.ListAccounts` API 获取成员列表，而非 IAM ListUsers。

下游 GPU 资源分析平台自行维护资源账号 CloudID 列表，通过 HCM 透传接口拉取实例数据。调用模式为逐 region 请求。

详细技术设计文档：[AWS AssumeRole 跨账号访问与 GPU 实例数据接口](../../docs/design/2026-03%20设计：AWS%20AssumeRole%20跨账号访问与%20GPU%20实例数据接口.md)

## Goals / Non-Goals

**Goals:**
- 新增 AssumeRole 能力，支持通过一级根账号凭证访问二级成员账号资源
- 补完 AWS sub_account 同步链路，使 sub_account 表有 AWS 数据
- 提供 GPU 实例数据透传接口（实例类型含 GPU 字段 + 实例列表），含 cloud-server 资源视角入口和蓝鲸 API 网关开放接口
- 编写 GPU 接口文档
- 非破坏性：不改动现有 `Aws()` / `AwsRoot()` 方法，所有改造通过新增代码实现

**Non-Goals:**
- 不实现预留实例接口
- 不做多 region 聚合（下游平台逐 region 调用）
- 不修改 sub_account 数据模型（Role ARN 由下游平台透传 role_name，HCM 自动拼接）
- 不引入 Redis 等外部组件

## Decisions

### 1. AssumeRole 链路放在哪里

**决策**：新增 `AwsWithAssumeRole` 方法到 `CloudAdaptorClient`，不改动现有 `Aws()` 和 `AwsRoot()`。

**理由**：改造现有方法影响 30+ 个已有接口，风险过高。新增方法遵循开闭原则，只有新接口走 AssumeRole 路径。

**否决方案**：改造 `Aws()` / `AwsRoot()` 方法 —— 影响面太大。

### 2. Role Chain（角色链式调用）

**决策**：由下游平台通过 `role_chain`（角色名数组）参数透传，HCM 按数组顺序执行 N 步 AssumeRole 链式调用。中间角色在管理账号中 AssumeRole，最终角色在目标成员账号（`cloud_id`）中 AssumeRole。每步用 `sub_account.CloudID`（或管理账号 CloudID）+ `site` 自动拼接完整 ARN。

**理由**：AWS 运维架构中常用 Role Chaining 模式增强安全控制（如先 Assume 一个 Caller 角色获取受限权限，再 Assume 最终 ReadOnly 角色访问资源）。使用数组而非单一 `role_name` 可灵活支持 1 步到 N 步的 AssumeRole 场景。AWS 官方文档确认 Role Chaining 是受支持的模式。

**否决方案**：
- 在 sub_account 扩展字段存储 CloudRoleArn 并手动录入 —— 运维成本高，角色名固定后无必要。
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

**决策**：使用 Cloud Account 体系（`cloud.account` + `cloud.sub_account`），不使用 Account-Set 体系（`root_account` + `main_account`）。

**理由**：`main_account` 仅包含需要出账核算的账号，无法覆盖全部成员账号。`sub_account` 通过 Organizations API 同步，覆盖全量。

### 5. BaseSecret 新增 CloudSessionToken

**决策**：在 `BaseSecret` 结构体新增 `CloudSessionToken` 字段，`newClientSet` 从硬编码 `""` 改为读取该字段。

**理由**：其他厂商该字段为零值空字符串，传入 `NewStaticCredentials` 等效当前硬编码 `""`，行为不变。

### 6. GPU 数据接口入参简化

**决策**：下游平台只需传 `cloud_id`（成员账号 AWS Account ID）+ `role_chain`（角色名数组）+ `region` + 可选的 `external_id`，HCM 内部自动用 CloudID 反查 sub_account 表找到对应的 `account_id`（根账号 AK/SK）。下游不需要感知 HCM 内部的 account_id 和 sub_account_id。

**理由**：AWS Account ID 全球唯一（AWS 官方文档确认 "uniquely identifies an AWS account"，且 AWS 确认不会出现 ID collision），用 CloudID 反查一定能得到唯一结果。简化下游调用步骤，从 3 步（查 sub_account → 查 region → 查数据）减少为 2 步（查 region → 查数据），甚至查 region 也可内聚。

**否决方案**：下游传 account_id + sub_account_id + role_name + region —— 下游需要先调一次 sub_account list 接口做 ID 转换，多一步调用且暴露了 HCM 内部概念。

### 7. GPU 接口分层：cloud-server 入口 + hc-service 实现

**决策**：GPU 数据接口由 cloud-server 提供资源视角入口（含 `ResOperateAuth` 鉴权），内部通过微服务调用 hc-service。hc-service handler 通过 `AwsWithAssumeRole` 获取 adaptor client 后，调用已有 adaptor 方法完成数据查询：

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

### 8. sub_account 同步频率

**决策**：复用 HCM 已有的定时同步机制（与其他厂商对齐），不单独设置 AWS 频率。

**理由**：sub_account 同步只需注册到 syncOrder，自动跟随 cloud-server 的定时任务周期。新增成员账号后，下一轮同步即可出现在表中。若实际使用中发现延迟不可接受，可后续调整。

## Risks / Trade-offs

| 风险 | 缓解措施 |
| --- | --- |
| BaseSecret 新增字段影响其他厂商 | 零值空字符串等效 `""`，已验证无影响 |
| 临时凭证分页拉取中过期 | 提前 10 分钟刷新缓冲；单次 API 秒级完成 |
| STS 在 opt-in Region 未启用 | 使用全局 STS endpoint 或激活目标 Region |
| 成员账号处于 SUSPENDED 状态 | sub_account 同步记录 Status，调用前可跳过 |
| 管理账号 IAM User 缺少 STS 权限 | 前置依赖：AWS 运维确认并补充权限 |
| AWS sub_account 同步链路未接通 | 前置依赖：补完 3 处代码（client + 入口 + syncOrder） |
| Role Chaining 会话时长限制 | AWS 限制链式 AssumeRole 最大会话时长为 1 小时（无法延长），单次 API 秒级完成，影响可控 |

## Open Questions

| # | 事项 | 说明 | 影响 |
| - | --- | --- | --- |
| D1 | cloud.account 与 root_account 的关系 | 两套体系中持有管理账号 AK/SK 的记录是否对应同一个 AWS 账号？推论：sub_account 同步需要 Organizations API 权限，说明 cloud.account 的 AK/SK 必然属于管理账号 | 如果不对应，需要额外关联逻辑 |
| D2 | sub_account 同步频率是否足够 | 补完同步链路后，新增子账号多久能出现在 sub_account 表中？ | 影响下游平台发现新账号的时效性 |
| E1 | 管理账号 IAM User 的 STS 权限 | cloud.account 的 AK/SK 对应的 IAM User 是否已有 sts:AssumeRole 权限？ | 所有 AssumeRole 调用的前提条件 |
