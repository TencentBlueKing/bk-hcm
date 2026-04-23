## MODIFIED Requirements

### Requirement: 批量更新权限模板

系统 SHALL 提供 `PATCH /vendors/{vendor}/permission_templates/batch/update` 接口，支持按 vendor 批量更新权限模板。Update 请求体 SHALL 同样采用泛型模式 `PermissionTemplateBatchUpdateReq[T PermissionTemplateExtension]`，通过 `batchUpdatePermissionTemplate[T](cts, svc)` 泛型函数实现 vendor 分发。可更新字段：name, policy_document, memo, extension, policy_library_id, policy_library_version, policy_library_sync_time。当 `policy_document` 非空时 SHALL 自动重算 `policy_hash`。不可更新字段：id（仅作标识）, cloud_id, account_id, vendor, creator, created_at, updated_at, tenant_id。

`PolicyLibraryApplier` SHALL 新增 `UpdateTmplBaseInfo` 结构体（含 `Memo *string`），作为 update 场景可覆盖字段的载体。

`PolicyLibraryApplier.updateTCloudCAMPolicy` 方法 SHALL 接受 `tmplInfo UpdateTmplBaseInfo` 参数：
- 使用 `tmplInfo.Memo` 作为 CAM Policy 的 description（调用方负责设置默认值）

`PolicyLibraryApplier.updateTCloudLocalTemplate` 方法 SHALL 接受 `tmplInfo UpdateTmplBaseInfo` 参数：
- 使用 `tmplInfo.Memo` 更新 memo 字段（调用方负责设置默认值）
- 该方法 SHALL 始终写入 `policy_library_id` 字段（值来自 `library.ID`），以支持自定义模板绑定策略库

现有 `ApplyUpdate` 调用 `applyTCloudUpdate` 时传 `UpdateTmplBaseInfo{Memo: library.Memo}`，行为不变（向后兼容）。

系统 SHALL 在 `PolicyLibraryApplier` 中新增以下方法：
- `ApplyUpdateWithTmplInfo(kt, vendor, libraryID string, templateIDs []string, tmplInfo UpdateTmplBaseInfo)`: 接受显式 templateIDs 和 tmplInfo，调用 `CheckPermTmplUpdatability` 校验后按 template 逐一执行云端和本地更新

`PolicyLibraryApplier.CheckPermTmplUpdatability` SHALL 允许以下任一条件的模板进行更新：
1. `policy_library_id IS NULL` AND `extension.cloud_type == TCloudCustomPolicy(1)`：自定义模板，允许绑定任意新策略库
2. `policy_library_id != ""` AND `policy_library_id == targetPolicyLibraryID`：已绑定同一策略库，允许重新应用（幂等）

#### Scenario: 更新含 policy_document 变更

- **WHEN** 发送 PATCH 请求，某条记录包含新的 policy_document
- **THEN** 系统自动重算 policy_hash 并更新

#### Scenario: 仅更新 memo

- **WHEN** 发送 PATCH 请求，某条记录仅包含 memo 字段
- **THEN** 系统仅更新 memo 和 reviser，其他字段不变

#### Scenario: 置空可选字段

- **WHEN** 发送 PATCH 请求，将 memo 设为空字符串或 policy_library_id 设为空
- **THEN** 系统 SHALL 允许将这些 blanked fields 置空

#### Scenario: updateTCloudLocalTemplate 使用 tmplInfo.Memo

- **WHEN** 调用 `updateTCloudLocalTemplate` 时 `tmplInfo.Memo` 为非 nil
- **THEN** 本地模板记录的 memo 字段更新为 `tmplInfo.Memo` 值，同时写入 `policy_library_id`

#### Scenario: ApplyUpdate 向后兼容

- **WHEN** 调用原有 `ApplyUpdate`（不传 tmplInfo）
- **THEN** 内部以 `UpdateTmplBaseInfo{Memo: library.Memo}` 调用 `applyTCloudUpdate`，本地模板 memo 更新为策略库 memo，行为与修改前一致

#### Scenario: ApplyUpdateWithTmplInfo 成功

- **WHEN** 调用 `ApplyUpdateWithTmplInfo(kt, tcloud, libraryID, []string{templateID}, tmplInfo)`
- **THEN** 云端 CAM Policy 已更新，本地模板 `policy_library_id`、`policy_document`、`memo` 均已写入

#### Scenario: CheckPermTmplUpdatability - 自定义模板允许更新

- **WHEN** 模板 `policy_library_id=nil` 且 `cloud_type=TCloudCustomPolicy`
- **THEN** `CheckPermTmplUpdatability` 返回 nil，允许更新

#### Scenario: CheckPermTmplUpdatability - 已绑同一策略库允许幂等更新

- **WHEN** 模板 `policy_library_id` 非空且等于目标 `policyLibraryID`
- **THEN** `CheckPermTmplUpdatability` 返回 nil，允许更新

#### Scenario: CheckPermTmplUpdatability - 已绑不同策略库拒绝

- **WHEN** 模板 `policy_library_id` 非空且不等于目标 `policyLibraryID`
- **THEN** `CheckPermTmplUpdatability` 返回 `InvalidParameter` 错误
