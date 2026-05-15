## ADDED Requirements

### Requirement: 查询权限模板关联的三级账号ID列表

系统 SHALL 提供 `GET /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/permission_templates/{id}/sub_account_ids` 接口，查询业务下云权限模板关联的三级账号ID列表。全量返回，不分页。

#### Scenario: 正常查询返回关联三级账号ID列表
- **WHEN** 发送 GET 请求，`bk_biz_id=1`，`vendor=tcloud`，`id=perm_tmpl_001`
- **THEN** 系统查询 `sub_account` 表中 `permission_template_ids` 包含 `perm_tmpl_001` 且 `bk_biz_ids` 包含 `1` 的三级账号，返回其 `cloud_id` 列表
- **AND** 响应为 `{ "code": 0, "message": "", "data": { "sub_account_ids": ["00000001", "00000002"] } }`

#### Scenario: 无关联三级账号
- **WHEN** 发送 GET 请求，指定权限模板无关联三级账号
- **THEN** 响应为 `{ "code": 0, "message": "", "data": { "sub_account_ids": [] } }`

#### Scenario: 权限模板不存在
- **WHEN** 发送 GET 请求，指定权限模板 ID 不存在
- **THEN** 系统正常查询 sub_account 表，返回空列表（不校验权限模板是否存在）

#### Scenario: 不支持的 vendor
- **WHEN** 发送 GET 请求，vendor 不是 tcloud
- **THEN** 系统返回 `unsupported vendor` 错误

#### Scenario: 无业务访问权限
- **WHEN** 发送 GET 请求，用户无指定业务的访问权限
- **THEN** 系统返回权限不足错误

### Requirement: sub_account DAO 支持按 permission_template_ids 过滤

系统 SHALL 在 `sub_account` DAO 的 `List` 方法 `columnTypes` 中注册 `permission_template_ids` 字段，使其支持 `json_overlaps` 操作符过滤。`permission_template_ids` 类型 SHALL 为 `enumor.String`（与 `types.StringArray` 元素类型一致）。

#### Scenario: json_overlaps 过滤 permission_template_ids
- **WHEN** 使用 `filter.Expression` 包含 `RuleJsonOverlaps("permission_template_ids", []string{"tmpl_001"})` 查询
- **THEN** DAO 生成 SQL `JSON_OVERLAPS(permission_template_ids, JSON_ARRAY(:placeholder))` 条件，返回 `permission_template_ids` 数组中包含 `"tmpl_001"` 的记录
