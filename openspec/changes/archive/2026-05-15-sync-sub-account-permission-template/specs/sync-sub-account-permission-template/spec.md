## ADDED Requirements

### Requirement: 同步子账号权限模板绑定关系

系统 SHALL 提供 `SubAccountPermissionTemplate` 同步方法，同步指定账号下所有子账号绑定的权限模板信息到本地 `sub_account` 表的 `permission_template_ids` 字段。

#### Scenario: 成功同步子账号权限模板
- **WHEN** 调用 `SubAccountPermissionTemplate` 方法，传入有效的 `AccountID`
- **THEN** 系统查询该账号下所有子账号，逐个调用 `ListAttachedUserAllPolicies` 获取绑定的策略列表，匹配本地 `permission_template` 记录，更新 `sub_account` 表的 `permission_template_ids` 字段

#### Scenario: 云上策略不存在于本地权限模板表
- **WHEN** 同步过程中发现云上策略的 `policy_id` 在本地 `permission_template` 表中不存在（通过 `cloud_id` 匹配）
- **THEN** 系统记录错误日志（包含策略 ID 和子账号信息），跳过该策略，继续处理其他策略

#### Scenario: 子账号未绑定任何策略
- **WHEN** 同步某个子账号时，该子账号在云上未绑定任何策略
- **THEN** 系统将该子账号的 `permission_template_ids` 设置为空数组（`[]`）

### Requirement: 同步失败返回错误

系统 SHALL 在权限模板同步失败时返回错误，中断同步流程，让上层调用方感知并处理。

#### Scenario: Adaptor 调用失败
- **WHEN** 调用 `ListAttachedUserAllPolicies` 方法失败（如网络错误、API 错误）
- **THEN** 系统记录错误日志，返回错误，中断同步流程

#### Scenario: 数据库更新失败
- **WHEN** 更新 `sub_account` 表的 `permission_template_ids` 字段失败
- **THEN** 系统记录错误日志，返回错误，中断同步流程

### Requirement: 同步结果日志记录

系统 SHALL 记录同步过程中的关键信息，包括成功数量、失败数量、跳过的策略等，便于运维排查问题。

#### Scenario: 同步完成记录汇总日志
- **WHEN** `SubAccountPermissionTemplate` 方法执行完成
- **THEN** 系统记录汇总日志，包括：处理的子账号总数、成功更新数量、失败数量、跳过的策略数量（如果大于 0）
