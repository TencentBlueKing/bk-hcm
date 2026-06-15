## ADDED Requirements

### Requirement: Biz 路径 apply_update 接口以 permission_template_ids 为入参

系统 SHALL 在 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/apply_permission_policy_library_update` 接口中，将请求参数从 `account_ids` 替换为 `permission_template_ids`（string array，required，min=1，max=100）。

每个 `permission_template_id` 对应一张 ITSM 审批单。`policy_library_id` 字段保持不变。

#### Scenario: 正常提交 permission_template_ids

- **WHEN** 用户提交 `{ "policy_library_id": "xxx", "permission_template_ids": ["t1", "t2"] }`
- **THEN** 系统 SHALL 为每个 permission_template_id 创建一张 ITSM 审批单，返回 `{ "data": { "ids": ["app1", "app2"] } }`

#### Scenario: permission_template_ids 为空数组

- **WHEN** 用户提交 `{ "policy_library_id": "xxx", "permission_template_ids": [] }`
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: permission_template_ids 超过 100 条

- **WHEN** `permission_template_ids` 数组长度大于 100
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

### Requirement: Biz 路径 CheckReq 基于 PermissionTemplateID 校验

系统 SHALL 在创建 ITSM 审批单前，对 `permission_template_id` 执行以下校验：
1. PermissionTemplateID 非空
2. 调用 `GetPermTmplAccountIDs` 通过模版 ID 查询关联账号，查询失败或结果不等于 1 则校验失败
3. 权限模版关联账号的 `bk_biz_id` 在策略库的 `bk_biz_ids` 范围内（通过 `CheckAccountsBizInScope`）

系统 SHALL NOT 额外调用 `CheckAccountApplied`（模版存在即代表已应用）。

#### Scenario: PermissionTemplateID 不存在或查询结果异常

- **WHEN** 传入的 `permission_template_id` 在 DB 中不存在，或 `GetPermTmplAccountIDs` 返回的账号数量不等于 1
- **THEN** 系统 SHALL 返回校验失败错误

#### Scenario: 模版关联账号不在 biz scope 内

- **WHEN** 模版关联账号的 `bk_biz_id` 不在策略库的 `bk_biz_ids` 列表中
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: 校验通过

- **WHEN** 模版存在（账号数量为 1）且账号 biz 在 scope 内
- **THEN** 系统 SHALL 继续创建 ITSM 审批单

### Requirement: Biz 路径审批通过后 Deliver 按 TemplateID 执行更新

系统 SHALL 在 ITSM 审批通过后的 `Deliver` 方法中，使用存储于审批单 content 的 `permission_template_id`，调用 `applier.ApplyUpdate`（以 `[]string{permissionTemplateID}` 为入参）执行云策略和本地模版的更新，并返回 `ApplyTemplateResult`。

#### Scenario: Deliver 成功

- **WHEN** ITSM 审批通过，content 中 `permission_template_id` 对应的模版存在
- **THEN** 系统 SHALL 完成云策略更新 + 本地模版更新，Deliver 返回 `Completed` 状态

#### Scenario: Deliver 失败

- **WHEN** 云策略更新失败
- **THEN** 系统 SHALL 返回 `DeliverError` 状态，detail map 中包含 `permission_template_id` 和错误原因

### Requirement: Resource 路径 apply update 接口以 permission_template_ids 为入参

系统 SHALL 在 `PUT /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}/apply`（更新路径）中，将请求参数从 `account_ids` 替换为 `permission_template_ids`（string array，required，min=1，max=100）。

#### Scenario: 正常提交 permission_template_ids

- **WHEN** 用户提交 `{ "permission_template_ids": ["t1", "t2", "t3"] }`
- **THEN** 系统 SHALL 对每个模版执行云策略更新，返回每个模版的执行结果列表

#### Scenario: permission_template_ids 为空

- **WHEN** `permission_template_ids` 为空数组
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

### Requirement: update 路径响应使用 ApplyTemplateResult

系统 SHALL 在 update 执行路径（Resource API 直接执行 和 Biz Deliver）中，响应结果使用新结构 `ApplyTemplateResult`：
- `permission_template_id`（string）：被操作的模版 ID
- `status`（string）：`success` 或 `failed`
- `reason`（string，可选）：失败时的原因

`ApplyPermissionPolicyLibraryUpdateResult` 包含 `results []ApplyTemplateResult`。

Create 路径继续使用原 `ApplyAccountResult`（含 `account_id`），不受影响。

#### Scenario: 全部成功

- **WHEN** 所有模版更新均成功
- **THEN** 响应 `results` 中每条记录 SHALL 包含对应 `permission_template_id` 和 `status: "success"`

#### Scenario: 部分失败

- **WHEN** 某个模版更新失败（如云策略更新异常）
- **THEN** 该条记录 SHALL 返回 `status: "failed"` 和 `reason`，其他成功条目不受影响

### Requirement: applier 改造 ApplyUpdate 及相关方法

系统 SHALL 对 `PolicyLibraryApplier` 进行以下改造：
- **改造 `ApplyUpdate(kt, vendor, libraryID, templateIDs)`**：入参从 `accountIDs []string` 改为 `templateIDs []string`，返回类型从 `*ApplyPermissionPolicyLibraryResult` 改为 `*ApplyPermissionPolicyLibraryUpdateResult`。内部先调用 `GetPermTmplAccountIDs` 获取账号列表做 biz scope 校验，再按 templateIDs 遍历执行更新。
- **新增 `GetPermTmplAccountIDs(kt, templateIDs)`**：批量查询权限模版，提取并去重 `AccountID` 列表返回。
- **新增 `GetTCloudTemplateByID(kt, templateID)`**：按 ID 查询 TCloud 权限模版（带 extension），不存在时返回 **error**（非 nil）。
- **`tcloudApplyUpdateForTemplate(kt, library, templateID)`**：调用 `GetTCloudTemplateByID` 查模版 → 使用模版内的 `CloudID` 和 `AccountID` → 调用 `TCloudUpdateCAMPolicy` + `TCloudUpdateLocalTemplate`，返回 `ApplyTemplateResult`。
- **审计方法拆分**：将原 `RecordApplyAudit` 拆分为 `RecordApplyCreateAudit(kt, libraryID, accountID)` 和 `RecordApplyUpdateAudit(kt, libraryID, permTmplID)`，分别用于 create 和 update 路径。

#### Scenario: GetTCloudTemplateByID 模版不存在

- **WHEN** templateID 在 DB 中不存在
- **THEN** `GetTCloudTemplateByID` SHALL 返回 error，`tcloudApplyUpdateForTemplate` SHALL 在结果中标记 failed

#### Scenario: ApplyUpdate 全部成功

- **WHEN** 所有 templateID 均存在且 cloud 更新成功
- **THEN** `ApplyUpdate` SHALL 返回所有结果均为 success 的 `ApplyPermissionPolicyLibraryUpdateResult`

### Requirement: 审批单 content 拆分为 create/update 独立结构体

系统 SHALL 将当前的 `ApplyPermPolicyLibContent` 拆分为：
- `ApplyPermPolicyLibBaseContent`：`Action`、`Vendor`、`BkBizID`，用于 handler dispatch
- `ApplyPermPolicyLibCreateContent`：embed base + `PolicyLibraryID` + `AccountID`
- `ApplyPermPolicyLibUpdateContent`：embed base + `PolicyLibraryID` + `PermissionTemplateID`

`ActionHandlerFactory` 函数签名 SHALL 调整为 `func(opt *HandlerOption, base *ApplyPermPolicyLibBaseContent, content string) (ApplicationHandler, error)`。

`NewHandlerFromApplication` SHALL 先将 rawContent 反序列化为 `ApplyPermPolicyLibBaseContent` 获取 `Action`，再将 rawContent 原样传给对应工厂函数，工厂内部二次反序列化为 action-specific struct。

`ApplicationBasePermissionPolicyLibrary` 中的 `Content` 字段 SHALL 改为 `Base *ApplyPermPolicyLibBaseContent`；各 action-specific handler 自行持有强类型的 `Content` 字段。

#### Scenario: update 审批单 content 反序列化

- **WHEN** 从 DB 读取的 content 为 update action 的 JSON（含 `permission_template_id`）
- **THEN** `NewHandlerFromApplication` SHALL 正确解析并构建 `ApplicationOfApplyPermPolicyLibUpdate` handler

#### Scenario: create 审批单 content 反序列化

- **WHEN** 从 DB 读取的 content 为 create action 的 JSON（含 `account_id`）
- **THEN** `NewHandlerFromApplication` SHALL 正确解析并构建 `ApplicationOfApplyPermPolicyLibCreate` handler

### Requirement: ITSM 表单展示 permission_template_id 相关信息

系统 SHALL 在 `RenderItsmTitle` 和 `RenderItsmForm` 中展示权限模版 ID 及通过 `GetPermTmplAccountIDs` 查询到的账号信息，替代原来直接使用 `account_id`。

`RenderItsmForm` 中展示的字段顺序 SHALL 为：业务、云厂商、云账号、权限模版ID、权限策略库、策略库ID、策略内容。

#### Scenario: ITSM 标题包含模版信息

- **WHEN** 渲染 update 类型的 ITSM 工单标题
- **THEN** 标题 SHALL 为 `申请应用权限策略库({library.Name})到权限模版({permission_template_id})`
