## Context

系统已具备三级账号创建和更新流程（`sub-account-operation`），支持在创建请求中传入 `permission_template_ids` 并存储到 `sub_account` 表。当前系统还具备：
- TCloud adaptor（`pkg/adaptor/tcloud/`）已有 CAM Client 封装，实现了 `CreatePolicy`、`UpdatePolicy`、`ListAttachedUserAllPolicies` 等方法
- hc-service 通过 `CloudAdaptorClient.TCloud(kt, accountID)` 获取带密钥的 TCloud adaptor 实例
- `permission_template` 表存储了云上策略信息（`cloud_id` 字段存储云上策略 ID，`name` 字段存储模版名称）
- cloud-server 在 `deliver.go` 中有空的 `attachPermissionToCloud` 方法待实现

腾讯云 CAM `AttachUserPolicy` API 文档：https://cloud.tencent.com/document/product/598/34579

## Goals / Non-Goals

**Goals:**
- 实现 TCloud CAM `AttachUserPolicy` adaptor 方法，支持为子用户绑定策略
- 在 hc-service 暴露批量绑定权限策略接口，供 cloud-server 调用
- 实现创建流程 cloud-server `attachPermissionToCloud` 方法，在三级账号创建成功后调用
- 实现更新流程 `checkPermissionTemplate` 校验方法，校验规则与创建流程一致
- 实现更新流程 `updatePermissionTemplateOnCloud` 方法，支持权限模版差异更新
- 创建和更新流程的工单渲染中展示权限模版名称
- 支持限流失败重试机制

**Non-Goals:**
- 本次不实现其他云厂商的策略绑定
- 本次不实现权限解绑流程（DetachUserPolicy）
- 本次不修改 `permission_template` 表结构

## Decisions

### Decision 1: Adaptor 方法签名设计

**选择**：实现 `AttachUserPolicy(kt *kit.Kit, opt *TCloudAttachUserPolicyOption) error` 单策略绑定方法，在 handler 层循环调用实现批量绑定。

**理由**：
- 腾讯云 `AttachUserPolicy` API 一次只能绑定一个策略
- 单策略绑定方法签名简单，便于错误定位和重试
- 与现有 `CreatePolicy`、`UpdatePolicy` 等方法风格一致

**替代方案**：实现批量绑定方法内部循环。否决：错误处理复杂，部分成功部分失败时难以表达。

### Decision 2: 限流重试机制

**选择**：使用 adaptor 已有的 `SetRateLimitRetryWithRandomInterval` 限流重试机制，复用全局配置。

**理由**：
- CAM API 有频率限制（每秒 20 次请求）
- 已有限流重试机制成熟稳定
- 随机间隔避免惊群效应

### Decision 3: hc-service 批量绑定接口设计

**选择**：在 hc-service 新增 `POST /vendors/tcloud/sub_accounts/attach_user_policies` 接口，接收子用户 UIN 和策略 ID 列表。

请求结构体 `TCloudAttachUserPoliciesReq`：
- `account_id`: string，必填，用于获取凭证
- `target_uin`: uint64，必填，目标子用户 UIN
- `policy_ids`: []uint64，必填，云上策略 ID 列表

**理由**：
- 批量接口减少网络往返
- 在 hc-service 层处理批量逻辑，cloud-server 保持简洁
- 接口内部串行调用 adaptor，便于错误处理

### Decision 4: 创建流程 Deliver 阶段调用时机

**选择**：在 `saveLocalSubAccount` 成功后、`sendSubAccountMail` 前调用 `attachPermissionToCloud`。

**理由**：
- 三级账号必须在云上创建成功后才能绑定权限
- 本地记录写入成功后才有 `cloud_id` 可用于后续操作
- 权限绑定失败不应影响邮件发送（绑定失败记录日志，返回 Completed 状态）

### Decision 5: 更新流程校验设计

**选择**：在更新流程的 `CheckReq` 中新增 `checkPermissionTemplate` 方法，复用创建流程的校验逻辑。

校验规则：
1. 若 `permission_template_ids` 为空，直接返回 nil
2. 查询权限模版表，校验数量是否匹配
3. 校验每个模版的 `policy_library_id` 不为空
4. 校验每个模版的 `account_id` 与请求中的二级账号 ID 匹配（需从现有 sub_account 记录中获取）

**理由**：
- 与创建流程校验规则一致，降低理解成本
- 确保权限模版存在且有效
- 确保权限模版属于同一二级账号

### Decision 6: 更新流程权限模版更新策略

**选择**：在更新流程 Deliver 阶段实现 `updatePermissionTemplateOnCloud` 方法，进行差异更新。

更新逻辑：
1. 若 `permission_template_ids` 未变更（nil），跳过更新
2. 若 `permission_template_ids` 为空数组，清空所有权限（暂不实现，记录日志）
3. 若 `permission_template_ids` 有值，查询权限模版获取云上策略 ID，调用 hc-service 批量绑定

**理由**：
- 更新场景下可能只更新其他字段，不涉及权限模版
- 差异更新避免不必要的 API 调用
- 本次暂不实现权限解绑，后续可通过同步流程修复

### Decision 7: 工单渲染权限模版名称

**选择**：在创建和更新流程的 `RenderItsmForm` 方法中，查询权限模版并渲染名称列表。

渲染逻辑：
1. 若 `permission_template_ids` 为空，不渲染该项
2. 查询权限模版表，获取名称列表
3. 渲染为 `绑定权限模版: 模版1,模版2,...` 格式

**理由**：
- 工单需要展示完整的变更信息供审批人查看
- 权限模版名称比 ID 更直观
- 与其他字段渲染风格一致

## Risks / Trade-offs

- **Risk**: 策略绑定 API 调用失败导致三级账号无权限
  - **Mitigation**: 记录详细错误日志，返回 Completed 状态但不阻塞流程；`permission_template_ids` 已存储到 `sub_account` 表，后续可通过同步流程修复

- **Risk**: 批量绑定时部分策略成功部分失败
  - **Mitigation**: 记录成功/失败的策略 ID，日志中详细记录失败原因

- **Risk**: CAM API 限流导致绑定超时
  - **Mitigation**: 使用限流重试机制；考虑批量绑定时的并发控制（串行调用）

- **Risk**: 更新流程未实现权限解绑，可能导致历史权限残留
  - **Mitigation**: 本次仅实现绑定新权限，后续可通过同步流程对账修复；更新日志中记录提示信息
