## Context

系统已实现权限策略库"应用（创建）"接口（`POST /vendors/{vendor}/permission_policy_libraries/{id}/apply`），将策略库首次下发到指定二级账号：在云上创建 CAM 策略、在本地创建 `permission_template` 记录。公共 applier 逻辑封装在 `applier.go` 中，设计时已预留了"应用（更新）"的复用入口。

现在需要实现配套的更新接口（`PUT` 同 URL），将策略库最新版本的 `policy_document` 覆盖更新到已应用账号的 CAM 策略和本地模板。

- TCloud adaptor 已有 `CreatePolicy`，无 `UpdatePolicy`
- hc-service `PermissionTemplateClient` 已有 `CreateCAMPolicy`，无 `UpdateCAMPolicy`
- data-service `PermissionTemplateClient` 已有 `BatchCreate`、`BatchUpdate`、`ListPermissionTemplateExt`
- `applier.go` 中的 `CheckAccountApplied` 只返回 `bool`，更新流程需要取出已有模板的 `CloudID`

## Goals / Non-Goals

**Goals:**
- 实现 Resource 层 `PUT /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}/apply` 接口
- 封装 TCloud CAM `UpdatePolicy` adaptor 能力
- 在 hc-service 暴露 CAM 策略更新接口
- 在 applier 中新增可复用的 Update 系列方法
- 复用现有的 `GetPolicyLibraryDetail`、`CheckAccountsBizInScope`、`RecordApplyAudit`

**Non-Goals:**
- 不修改 Create 接口任何逻辑
- 不实现 Biz 层审批接口
- 不做并发/异步任务优化

## Decisions

### Decision 1：新增 `GetAccountTemplate` 而非修改 `CheckAccountApplied`

**选择**：在 applier 中新增 `GetAccountTemplate(kt, libraryID, accountID)` 方法，返回完整的 `PermissionTemplateExt` 记录（nil 表示未应用）。`CheckAccountApplied` 保持不变，Create 流程继续使用它。

**理由**：Update 流程既需要判断"是否已应用"，又需要取出 `CloudID`（用于调 TCloud UpdatePolicy 时传 `PolicyID`）。若修改 `CheckAccountApplied` 签名，会破坏 Create 流程。新增独立方法职责更清晰。

**替代方案**：复用 `CheckAccountApplied` 再单独 List 一次。否决：两次 DB 查询浪费，不如一次返回所需数据。

### Decision 2：`TCloudUpdateCAMPolicy` 只返回 `error`，同时支持更新 Description

**选择**：TCloud `UpdatePolicy` API 成功时无新的 PolicyID 返回，hc-service handler 返回 `nil data`，client 方法签名为 `(kt, req) → error`。`TCloudUpdatePolicyOption` 和 `UpdateCAMPolicyReq` 中的 `PolicyDocument` 与 `Description` 均为可选指针（`*string`），Validate 要求至少提供其中一个。applier 层调用时同时传入 `PolicyDocument` 和 `Description`（取 `library.Memo`）。

**理由**：与 `CreatePolicy` 不同，`UpdatePolicy` 只是覆盖已有策略内容，无需返回资源标识符。使用指针类型可以精确表达"不更新该字段"的语义，保持与 TCloud CAM SDK 接口行为一致。

### Decision 3：本地模板更新字段来自策略库当前状态

**选择**：更新本地模板时，`PolicyLibraryVersion` 取 `library.Version`，`PolicyLibrarySyncTime` 取 `time.Now().UTC()`，`PolicyDocument` 取 `library.PolicyDocument`，`Memo` 取 `library.Memo`。与 Create 流程一致。

**理由**：应用更新的语义是"将账号的模板同步到策略库当前版本"，因此以策略库的当前版本号和同步时间戳为准。`Memo` 也一并同步，保持模板描述与策略库一致。

### Decision 4：Update 流程中账号校验逻辑与 Create 互反

**选择**：Update 流程逐账号执行时，若 `GetAccountTemplate` 返回 nil（未应用），直接返回 `failed: "该二级账号未应用此权限策略库"`；若非 nil，取出 `CloudID` 转 `uint64` 后调 UpdatePolicy。

**理由**：Create 检查"已应用则报错"，Update 检查"未应用则报错"，两者完全互反，清晰表达各自的前置条件。

## Risks / Trade-offs

- **[风险] CloudID 转 uint64 失败**：`permission_template.cloud_id` 为 string，需解析为 uint64 传给 TCloud UpdatePolicy。若历史数据中存在非数字 cloud_id，解析会失败。→ 缓解：解析失败直接返回该账号的 `failed` 结果，不影响其他账号。
- **[风险] 云上策略已被手动删除**：账号本地有模板记录，但云上 CAM 策略已被删除。UpdatePolicy 会返回云 API 错误。→ 缓解：云 API 错误会被捕获并作为该账号的 `failed` 原因返回，与 Create 模式一致。
- **[权衡] 同步执行耗时**：100 个账号逐个调用 CAM API 可能耗时较长。→ 与 Create 接口相同的设计预期，前端已知同步特性。
