## ADDED Requirements

### Requirement: List permission policy library via cloud-server

系统 SHALL 在 cloud-server 层暴露 `POST /vendors/{vendor}/permission_policy_libraries/list` 接口，支持按 vendor 过滤、IAM 鉴权后查询权限策略库列表。handler 实现 SHALL 将核心逻辑抽取为接受 `authHandler handler.ListAuthResHandler` 参数的私有方法，公开入口方法传入具体的 authHandler，以便后续扩展不同鉴权模式（如 biz 域使用 `ListBizAuthRes`）。

#### Scenario: 查询数据列表

- **WHEN** 用户发送 `POST /api/v1/cloud/vendors/tcloud/permission_policy_libraries/list` 请求，body 包含 `filter`（rules 非空）和 `page`（count=false, start=0, limit=20）
- **THEN** 系统 SHALL 返回 `{ "code": 0, "data": { "count": 0, "details": [...] } }`，details 中每条记录包含 `id`, `name`, `policy_document`, `policy_hash`, `version`, `bk_biz_ids`, `memo`, `vendor`, `associated_account_count`, `creator`, `reviser`, `created_at`, `updated_at` 字段

#### Scenario: 查询总数

- **WHEN** 用户发送请求 body 中 `page.count=true`
- **THEN** 系统 SHALL 返回 `{ "code": 0, "data": { "count": <N>, "details": null } }`，count 为满足 filter 条件的总记录数

#### Scenario: 无权限用户查询

- **WHEN** 用户不具备 `sys.CloudVendorConfig` IAM 权限
- **THEN** 系统 SHALL 返回空列表 `{ "code": 0, "data": { "count": 0, "details": [] } }`

#### Scenario: vendor 路径参数校验

- **WHEN** 用户传入不合法的 vendor 值（非 tcloud 等有效枚举值）
- **THEN** 系统 SHALL 返回参数校验错误

### Requirement: vendor 作为过滤条件注入

系统 SHALL 将 URL 路径中的 `{vendor}` 参数作为 `vendor` 字段等值过滤条件，与用户传入的 filter 合并后发送给 data-service。

#### Scenario: vendor 过滤生效

- **WHEN** 用户请求 `/vendors/tcloud/permission_policy_libraries/list`
- **THEN** 最终发给 data-service 的 filter 中 SHALL 包含 `vendor == "tcloud"` 条件

### Requirement: associated_account_count 字段占位

响应结构体 SHALL 包含 `associated_account_count` 字段（int 类型），当前固定返回 0，待后续补充实际计算逻辑。

#### Scenario: 字段存在但值为 0

- **WHEN** 查询返回数据列表
- **THEN** 每条记录中 `associated_account_count` SHALL 为 0
