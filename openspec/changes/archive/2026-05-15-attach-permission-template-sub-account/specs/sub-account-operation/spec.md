## MODIFIED Requirements

### Requirement: 创建流程 Deliver 交付逻辑
`ApplicationOfCreateSubAccount.Deliver()` SHALL：
1. 按 vendor switch-case 分发（本期仅 tcloud，其他返回 not supported 错误）
2. 调用 hc-service `CreateSubAccount` 创建云上子用户
3. 调用 data-service `SubAccount.BatchCreate` 写入本地 sub_account 表
4. **调用 `attachPermissionToCloud` 方法绑定权限模版到云上子用户**
5. 调用 `SendMail` 将 SecretId 和 SecretKey 发送到 `receive_email`
6. 邮件发送失败 SHALL 记录错误日志但不影响 deliver 状态
7. 返回 `Completed` 状态和包含 sub_account ID 的 detail

#### Scenario: TCloud 正常交付
- **WHEN** vendor 为 tcloud，hc-service 和 data-service 调用均成功，权限模版绑定成功或失败
- **THEN** 返回 `Completed` 状态，密钥邮件发送到 receive_email，权限绑定结果记录到日志

#### Scenario: 云上创建失败
- **WHEN** hc-service 调用 AddUser 返回错误
- **THEN** 返回 `DeliverError` 状态和错误信息

#### Scenario: 本地写入失败
- **WHEN** hc-service 成功但 data-service BatchCreate 失败
- **THEN** 返回 `DeliverError` 状态，日志记录云上已创建但本地写入失败

#### Scenario: 权限绑定失败
- **WHEN** 云上创建和本地写入均成功，但 `attachPermissionToCloud` 调用失败
- **THEN** deliver 仍返回 `Completed` 状态，错误记录到日志

#### Scenario: 邮件发送失败
- **WHEN** 云上创建、本地写入、权限绑定均成功，但 SendMail 失败
- **THEN** deliver 仍返回 `Completed` 状态，错误记录到日志

#### Scenario: 不支持的 vendor
- **WHEN** vendor 为 aws（本期未实现）
- **THEN** 返回 `DeliverError` 和 not supported 错误

### Requirement: 更新流程 CheckReq 校验逻辑
`ApplicationOfUpdateSubAccount.CheckReq()` SHALL：
1. 调用 `req.Validate()` 校验请求参数
2. 调用 `CheckSubAccountExists` 校验三级账号存在
3. 校验二级账号存在且业务 ID 匹配
4. **调用 `checkPermissionTemplate` 方法校验权限模版**

#### Scenario: 权限模版校验成功
- **WHEN** 传入有效的 `PermissionTemplateIDs`，所有校验通过
- **THEN** 返回 nil

#### Scenario: 权限模版 ID 为空
- **WHEN** `PermissionTemplateIDs` 为 nil
- **THEN** 跳过权限模版校验，返回 nil

#### Scenario: 权限模版数量不匹配
- **WHEN** 查询到的权限模版数量与 `PermissionTemplateIDs` 数量不一致
- **THEN** 返回错误 "permission templates count mismatch"

#### Scenario: 权限模版 policy_library_id 为空
- **WHEN** 某个权限模版的 `policy_library_id` 为空
- **THEN** 返回错误 "permission template(id=xxx) has empty policy_library_id"

#### Scenario: 权限模版 account_id 不匹配
- **WHEN** 某个权限模版的 `account_id` 与三级账号所属二级账号 ID 不一致
- **THEN** 返回错误 "permission template(id=xxx) account_id does not match"

### Requirement: 更新流程 Deliver 交付逻辑
`ApplicationOfUpdateSubAccount.Deliver()` SHALL：
1. 按 vendor switch-case 分发（本期仅 tcloud，其他返回 not supported 错误）
2. 调用 hc-service `UpdateSubAccount` 更新云上子用户
3. 调用 data-service `SubAccount.BatchUpdate` 更新本地 sub_account 表
4. **调用 `updatePermissionTemplateOnCloud` 方法更新权限模版**
5. 返回 `Completed` 状态和包含 sub_account ID 的 detail

#### Scenario: TCloud 正常交付
- **WHEN** vendor 为 tcloud，hc-service 和 data-service 调用均成功，权限模版更新成功或失败
- **THEN** 返回 `Completed` 状态，权限更新结果记录到日志

#### Scenario: 权限模版未变更
- **WHEN** `PermissionTemplateIDs` 为 nil
- **THEN** 跳过权限模版更新，返回 `Completed` 状态

#### Scenario: 权限模版更新失败
- **WHEN** 云上更新和本地更新均成功，但 `updatePermissionTemplateOnCloud` 调用失败
- **THEN** deliver 仍返回 `Completed` 状态，错误记录到日志

## ADDED Requirements

### Requirement: 创建流程 attachPermissionToCloud 方法

系统 SHALL 在 `cmd/cloud-server/service/application/handlers/sub-account/create-sub-account/deliver.go` 实现 `attachPermissionToCloud` 方法。

方法逻辑：
1. 若 `a.req.PermissionTemplateIDs` 为空，直接返回 nil
2. 通过 `a.Client.DataService().Global.PermissionTemplate.ListPermissionTemplate` 查询权限模版，获取 `cloud_id`（云上策略 ID）
3. 从 `createResult` 中获取子用户 UIN
4. 调用 `a.Client.HCService().TCloud.Account.AttachUserPolicies` 批量绑定策略
5. 绑定失败时记录错误日志但不返回错误（不阻塞流程）

#### Scenario: 无权限模版
- **WHEN** `a.req.PermissionTemplateIDs` 为空
- **THEN** 方法直接返回 nil，不执行任何操作

#### Scenario: 成功绑定权限模版
- **WHEN** 传入有效的 `PermissionTemplateIDs`，权限模版查询成功，hc-service 调用成功
- **THEN** 策略成功绑定到云上子用户，返回 nil

#### Scenario: 权限模版查询失败
- **WHEN** 权限模版查询返回错误
- **THEN** 记录错误日志，返回 nil（不阻塞流程）

#### Scenario: 部分策略绑定失败
- **WHEN** hc-service 返回部分失败的策略 ID
- **THEN** 记录警告日志，返回 nil（不阻塞流程）

#### Scenario: 全部策略绑定失败
- **WHEN** hc-service 返回错误
- **THEN** 记录错误日志，返回 nil（不阻塞流程）

### Requirement: 更新流程 checkPermissionTemplate 方法

系统 SHALL 在 `cmd/cloud-server/service/application/handlers/sub-account/update-sub-account/check.go` 实现 `checkPermissionTemplate` 方法。

方法逻辑：
1. 若 `a.req.PermissionTemplateIDs` 为 nil，直接返回 nil
2. 通过 `a.Client.DataService().Global.PermissionTemplate.ListPermissionTemplate` 查询权限模版
3. 校验数量是否匹配：`len(result.Details) == len(a.req.PermissionTemplateIDs)`
4. 校验每个模版的 `policy_library_id` 不为空
5. 校验每个模版的 `account_id` 与三级账号所属二级账号 ID 一致

#### Scenario: 无权限模版变更
- **WHEN** `a.req.PermissionTemplateIDs` 为 nil
- **THEN** 方法直接返回 nil，不执行任何校验

#### Scenario: 权限模版校验成功
- **WHEN** 传入有效的 `PermissionTemplateIDs`，所有校验通过
- **THEN** 返回 nil

#### Scenario: 权限模版数量不匹配
- **WHEN** 查询到的权限模版数量与请求 ID 数量不一致
- **THEN** 返回错误

#### Scenario: policy_library_id 为空
- **WHEN** 某个权限模版的 `policy_library_id` 为空
- **THEN** 返回错误

#### Scenario: account_id 不匹配
- **WHEN** 某个权限模版的 `account_id` 与二级账号 ID 不一致
- **THEN** 返回错误

### Requirement: 更新流程 updatePermissionTemplateOnCloud 方法

系统 SHALL 在 `cmd/cloud-server/service/application/handlers/sub-account/update-sub-account/deliver.go` 实现 `updatePermissionTemplateOnCloud` 方法。

方法逻辑：
1. 若 `a.req.PermissionTemplateIDs` 为 nil，直接返回 nil（表示不更新权限模版）
2. 若 `a.req.PermissionTemplateIDs` 为空数组，记录日志提示暂不支持清空权限，返回 nil
3. 查询权限模版获取 `cloud_id`（云上策略 ID）
4. 获取三级账号的 `cloud_id`（子用户 UIN）
5. 调用 `a.Client.HCService().TCloud.Account.AttachUserPolicies` 批量绑定策略
6. 绑定失败时记录错误日志但不返回错误（不阻塞流程）

#### Scenario: 无权限模版变更
- **WHEN** `a.req.PermissionTemplateIDs` 为 nil
- **THEN** 方法直接返回 nil，不执行任何操作

#### Scenario: 清空权限模版
- **WHEN** `a.req.PermissionTemplateIDs` 为空数组
- **THEN** 记录警告日志提示暂不支持清空权限，返回 nil

#### Scenario: 成功更新权限模版
- **WHEN** 传入有效的 `PermissionTemplateIDs`，权限模版查询成功，hc-service 调用成功
- **THEN** 策略成功绑定到云上子用户，返回 nil

#### Scenario: 权限模版查询失败
- **WHEN** 权限模版查询返回错误
- **THEN** 记录错误日志，返回 nil（不阻塞流程）

#### Scenario: 策略绑定失败
- **WHEN** hc-service 返回错误
- **THEN** 记录错误日志，返回 nil（不阻塞流程）

### Requirement: 创建流程工单渲染权限模版名称

系统 SHALL 在 `cmd/cloud-server/service/application/handlers/sub-account/create-sub-account/create_itsm_ticket.go` 的 `RenderItsmForm` 方法中增加权限模版名称渲染。

渲染逻辑：
1. 若 `a.req.PermissionTemplateIDs` 为空，跳过该项
2. 调用 `a.Client.DataService().Global.PermissionTemplate.ListPermissionTemplate` 查询权限模版名称
3. 将名称列表以逗号分隔渲染：`绑定权限模版: 模版1,模版2,...`

#### Scenario: 无权限模版
- **WHEN** `a.req.PermissionTemplateIDs` 为空
- **THEN** 工单不渲染权限模版项

#### Scenario: 渲染权限模版成功
- **WHEN** `a.req.PermissionTemplateIDs` 有值，查询成功
- **THEN** 工单展示权限模版名称列表

#### Scenario: 查询权限模版失败
- **WHEN** 查询权限模版失败
- **THEN** 记录错误日志，工单不渲染权限模版项（不阻塞工单创建）

### Requirement: 更新流程工单渲染权限模版名称

系统 SHALL 在 `cmd/cloud-server/service/application/handlers/sub-account/update-sub-account/create_itsm_ticket.go` 的 `RenderItsmForm` 方法中增加权限模版名称渲染。

渲染逻辑：
1. 若 `a.req.PermissionTemplateIDs` 为 nil，跳过该项（表示未变更）
2. 若为空数组，渲染 `修改权限模版: 清空`
3. 若有值，查询权限模版名称，渲染 `修改权限模版: 模版1,模版2,...`

#### Scenario: 无权限模版变更
- **WHEN** `a.req.PermissionTemplateIDs` 为 nil
- **THEN** 工单不渲染权限模版项

#### Scenario: 清空权限模版
- **WHEN** `a.req.PermissionTemplateIDs` 为空数组
- **THEN** 工单渲染 `修改权限模版: 清空`

#### Scenario: 渲染权限模版成功
- **WHEN** `a.req.PermissionTemplateIDs` 有值，查询成功
- **THEN** 工单展示权限模版名称列表

#### Scenario: 查询权限模版失败
- **WHEN** 查询权限模版失败
- **THEN** 记录错误日志，工单渲染 `修改权限模版: [查询失败]`

### Requirement: SubAccountUpdateReq 扩展

系统 SHALL 在 `pkg/api/cloud-server/application/update_sub_account.go` 的 `SubAccountUpdateReq` 结构体中新增 `PermissionTemplateIDs` 字段。

字段定义：
- `PermissionTemplateIDs`: `[]string`，可选，使用 `omitempty` 标签，`nil` 表示不更新权限模版，空数组表示清空权限模版

#### Scenario: 更新请求包含权限模版
- **WHEN** 请求中 `permission_template_ids` 有值
- **THEN** 字段被正确解析

#### Scenario: 更新请求不包含权限模版
- **WHEN** 请求中未包含 `permission_template_ids` 字段
- **THEN** 字段值为 nil
