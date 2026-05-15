## ADDED Requirements

### Requirement: 查询策略库下业务范围内的权限模版列表
系统 SHALL 提供接口，按策略库 ID 返回所有已应用且账号当前仍在策略库业务范围（bk_biz_ids）内的权限模版，全量返回不分页。

#### Scenario: 成功返回范围内的权限模版
- **WHEN** 请求 `GET /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}/permission_templates`
- **THEN** 系统返回该策略库下所有 `policy_library_id = {id}` 且对应账号的 `bk_biz_id` 在策略库 `bk_biz_ids` 范围内的 permission_template 记录

#### Scenario: 账号业务范围外的模版不返回
- **WHEN** 策略库的 `bk_biz_ids` 更新后移除了某个业务，该业务下账号的历史模版依然存在于数据库
- **THEN** 接口不返回这些业务范围外的模版

#### Scenario: 策略库不存在时返回错误
- **WHEN** 路径参数 `{id}` 对应的策略库不存在
- **THEN** 系统返回对应错误

#### Scenario: 无权限时返回权限拒绝错误
- **WHEN** 请求者无 `PermissionPolicyLibrary.Find` 权限
- **THEN** 系统返回 `PermissionDenied` 错误

#### Scenario: 暂不支持的云厂商返回错误
- **WHEN** 路径参数 `{vendor}` 不是 `tcloud`
- **THEN** 系统返回 unsupported vendor 错误

### Requirement: 返回数据包含扩展字段
系统 SHALL 在返回的权限模版中包含云厂商差异扩展字段（extension）。

#### Scenario: TCloud 模版包含 cloud_type 扩展字段
- **WHEN** vendor 为 `tcloud` 时查询权限模版列表
- **THEN** 每条模版的 `extension.cloud_type` 字段正确返回（1-自定义策略，2-预设策略）
