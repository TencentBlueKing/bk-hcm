## ADDED Requirements

### Requirement: SubAccount 表支持存储权限模板 ID 列表

系统 SHALL 在 `sub_account` 表新增 `permission_template_ids` 字段（JSON 数组类型），用于存储子账号绑定的本地权限模板 ID 列表。

#### Scenario: 查询子账号返回权限模板 ID 列表
- **WHEN** 查询子账号详情或列表时
- **THEN** 系统返回 `permission_template_ids` 字段，包含该子账号绑定的所有权限模板 ID

#### Scenario: 更新子账号的权限模板 ID 列表
- **WHEN** 同步流程更新子账号的权限模板绑定关系时
- **THEN** 系统更新 `permission_template_ids` 字段，覆盖原有值

#### Scenario: 新创建的子账号权限模板 ID 列表为空
- **WHEN** 新创建子账号时
- **THEN** `permission_template_ids` 字段默认为空数组（`[]`）或 null

### Requirement: API 支持 permission_template_ids 字段

系统 SHALL 在子账号相关的 API 请求/响应中支持 `permission_template_ids` 字段。

#### Scenario: 创建子账号 API 不支持指定权限模板
- **WHEN** 调用创建子账号 API 时传入 `permission_template_ids` 字段
- **THEN** 系统忽略该字段，权限模板绑定通过同步流程维护

#### Scenario: 查询子账号 API 返回权限模板 ID 列表
- **WHEN** 调用查询子账号详情或列表 API 时
- **THEN** 系统返回 `permission_template_ids` 字段

#### Scenario: 更新子账号 API 不支持直接修改权限模板
- **WHEN** 调用更新子账号 API 时传入 `permission_template_ids` 字段
- **THEN** 系统忽略该字段，权限模板绑定通过同步流程维护
