## ADDED Requirements

### Requirement: 权限模板表 DDL

系统 SHALL 提供 `permission_template` 表的 SQL DDL，包含以下字段：id, cloud_id, name, account_id, policy_library_id, policy_library_version, policy_library_sync_time, policy_document, policy_hash, memo, extension, vendor, tenant_id, creator, reviser, created_at, updated_at。DDL 文件 SHALL 包含 `id_generator` 插入和 `hcm_version` 视图更新。

#### Scenario: 表创建成功
- **WHEN** 执行 DDL 脚本
- **THEN** 创建 `permission_template` 表，主键为 `id`，`tenant_id` 默认值为 `'default'`，`created_at`/`updated_at` 自动维护

### Requirement: 批量创建权限模板

系统 SHALL 提供 `POST /vendors/{vendor}/permission_templates/create` 接口，支持按 vendor 批量创建权限模板。请求体和 extension 字段 SHALL 采用 Go 泛型模式（类似 CVM 的 `batchCreateCvm[T Extension]`），通过 `PermissionTemplateBatchCreateReq[T PermissionTemplateExtension]` 泛型结构体和 `batchCreatePermissionTemplate[T](cts, svc, vendor)` 泛型函数实现 vendor 分发。Core API 层 SHALL 定义 `PermissionTemplateExtension` 接口约束（当前仅含 `TCloudPermissionTemplateExtension`）。请求体 SHALL 包含 `permission_templates` 数组（必填字段：cloud_id, name, account_id, policy_document, extension；可选字段：policy_library_id, policy_library_version, policy_library_sync_time, memo）。系统 SHALL 自动计算 `policy_hash`（SHA256）。数量限制 SHALL 遵守 `constant.BatchOperationMaxLimit`。

#### Scenario: TCloud vendor 创建成功
- **WHEN** 发送 POST 请求，vendor=tcloud，包含合法的 permission_templates 数组
- **THEN** 系统创建记录，自动生成 ID，计算 policy_hash，写入 Create 审计，返回 `BatchCreateResult{IDs}`

#### Scenario: 不支持的 vendor
- **WHEN** 发送 POST 请求，vendor 不是 tcloud
- **THEN** 系统返回 `unsupported vendor` 错误

#### Scenario: extension 缺失
- **WHEN** 发送 POST 请求，某条记录未提供 extension
- **THEN** 系统返回参数校验错误

#### Scenario: 超过批量限制
- **WHEN** 请求中 permission_templates 数量超过 BatchOperationMaxLimit
- **THEN** 系统返回参数校验错误

### Requirement: 批量更新权限模板

系统 SHALL 提供 `PATCH /vendors/{vendor}/permission_templates/batch/update` 接口，支持按 vendor 批量更新权限模板。Update 请求体 SHALL 同样采用泛型模式 `PermissionTemplateBatchUpdateReq[T PermissionTemplateExtension]`，通过 `batchUpdatePermissionTemplate[T](cts, svc)` 泛型函数实现 vendor 分发。可更新字段：name, policy_document, memo, extension, policy_library_id, policy_library_version, policy_library_sync_time。当 `policy_document` 非空时 SHALL 自动重算 `policy_hash`。不可更新字段：id（仅作标识）, cloud_id, account_id, vendor, creator, created_at, updated_at, tenant_id。

#### Scenario: 更新含 policy_document 变更
- **WHEN** 发送 PATCH 请求，某条记录包含新的 policy_document
- **THEN** 系统自动重算 policy_hash 并更新

#### Scenario: 仅更新 memo
- **WHEN** 发送 PATCH 请求，某条记录仅包含 memo 字段
- **THEN** 系统仅更新 memo 和 reviser，其他字段不变

#### Scenario: 置空可选字段
- **WHEN** 发送 PATCH 请求，将 memo 设为空字符串或 policy_library_id 设为空
- **THEN** 系统 SHALL 允许将这些 blanked fields 置空

### Requirement: 批量删除权限模板

系统 SHALL 提供 `DELETE /permission_templates/batch` 接口（无 vendor 路径参数），通过 `filter.Expression` 指定删除条件。

#### Scenario: 按 filter 删除
- **WHEN** 发送 DELETE 请求，filter 指定 id in [...]
- **THEN** 系统物理删除匹配记录

#### Scenario: filter 为空
- **WHEN** 发送 DELETE 请求，filter 为 nil
- **THEN** 系统返回 InvalidParameter 错误

### Requirement: 列表查询权限模板（不含 extension）

系统 SHALL 提供 `POST /permission_templates/list` 接口（无 vendor 路径参数），支持 filter + page 分页查询，支持 count 模式。返回 `BasePermissionTemplate`，不包含 extension 字段。

#### Scenario: 分页查询
- **WHEN** 发送 POST 请求，指定 filter 和 page（非 count 模式）
- **THEN** 系统返回 `PermissionTemplateListResult{Details: []BasePermissionTemplate}` 包含匹配记录（不含 extension）

#### Scenario: Count 模式
- **WHEN** 发送 POST 请求，page.count=true
- **THEN** 系统返回 `PermissionTemplateListResult{Count}` 仅包含数量

### Requirement: 列表查询权限模板（含 extension）

系统 SHALL 提供 `POST /vendors/{vendor}/permission_templates/list` 接口（带 vendor 路径参数），支持 filter + page 分页查询，支持 count 模式。返回 `PermissionTemplate[T]`（embed `BasePermissionTemplate` + `Extension *T`），通过泛型函数 `convPermissionTemplateExtListResult[T]` 实现 vendor 分发和 extension 反序列化。DS API 层 SHALL 定义 `PermissionTemplateExtListReq`、`PermissionTemplateExtListResult[T]`、`PermissionTemplateExtListResp[T]` 泛型结构体。

#### Scenario: TCloud vendor 分页查询
- **WHEN** 发送 POST 请求，vendor=tcloud，指定 filter 和 page（非 count 模式）
- **THEN** 系统返回 `PermissionTemplateExtListResult[TCloudPermissionTemplateExtension]{Details}` 包含含 extension 的匹配记录

#### Scenario: Count 模式
- **WHEN** 发送 POST 请求，vendor=tcloud，page.count=true
- **THEN** 系统返回 `PermissionTemplateExtListResult{Count}` 仅包含数量

#### Scenario: 不支持的 vendor
- **WHEN** 发送 POST 请求，vendor 不是 tcloud
- **THEN** 系统返回 `unsupported vendor` 错误

### Requirement: TCloud 扩展字段定义

Core API 层 SHALL 定义 `PermissionTemplateExtension` 接口约束（union type，当前仅含 `TCloudPermissionTemplateExtension`）和 `TCloudPermissionTemplateExtension` 结构体，包含 `CloudType enumor.TCloudPolicyType`（枚举类型，`TCloudCustomPolicy=1` 自定义策略，`TCloudPresetPolicy=2` 预设策略）。`TCloudPolicyType` 定义在 `pkg/criteria/enumor/permission_template.go`，提供 `Validate()` 方法。`BasePermissionTemplate` 不包含 Extension 字段，仅用于无 vendor 的 List 接口。`PermissionTemplate[T PermissionTemplateExtension]` 嵌入 `BasePermissionTemplate` 并携带 `Extension *T`，用于带 vendor 的 ListExt 接口。extension 在 Table 层使用 `types.JsonField`，在 Service handler 中通过 `json.MarshalToString`/`json.UnmarshalFromString` 在类型化结构体和 JSON 字符串之间转换。

#### Scenario: Create 时 extension 序列化
- **WHEN** Create handler 接收到 TCloudPermissionTemplateExtension
- **THEN** 将其序列化为 JSON 字符串存入 `types.JsonField`

#### Scenario: ListExt 时 extension 反序列化
- **WHEN** ListExt handler 从数据库读取 extension JsonField
- **THEN** 通过泛型函数 `convPermissionTemplateExtListResult[T]` 将其反序列化为 `*T` 填入 `PermissionTemplate[T]`

#### Scenario: List 时不返回 extension
- **WHEN** List handler（无 vendor）从数据库读取记录
- **THEN** 返回 `BasePermissionTemplate`，不包含 extension 字段

### Requirement: 审计记录

系统 SHALL 为 Create/Update/Delete 操作记录审计。Create 审计在 DAO `BatchCreateWithTx` 内写入，使用 `enumor.PermissionTemplateAuditResType`。Update 和 Delete 审计通过 DS audit 模块的 `permissionTemplateUpdateAuditBuild` 和 `permissionTemplateDeleteAuditBuild` 构建。

#### Scenario: Create 审计
- **WHEN** 批量创建权限模板成功
- **THEN** DAO 层在同一事务内写入 Create 审计记录，包含 ResID, ResName, ResType, Action=Create, Vendor

#### Scenario: Update 审计
- **WHEN** 上层调用 DS audit 的 Update 审计接口
- **THEN** audit 模块查询已有记录，构建含变更字段的审计记录

#### Scenario: Delete 审计
- **WHEN** 上层调用 DS audit 的 Delete 审计接口
- **THEN** audit 模块查询已有记录，构建含完整数据快照的审计记录

### Requirement: Client 封装

系统 SHALL 提供 TCloud vendor client（`BatchCreate`, `BatchUpdate`, `ListPermissionTemplateExt`）和 Global client（`ListPermissionTemplate`, `BatchDelete`）。vendor client 的 base path 已包含 `/vendors/tcloud` 前缀，SubResourcef 中不再拼接 vendor 路径。

#### Scenario: TCloud client 调用 Create
- **WHEN** 调用 `tcloud.PermissionTemplate.BatchCreate(kt, req)`
- **THEN** 发送 POST 到 `/permission_templates/create`，返回 `BatchCreateResult`

#### Scenario: TCloud client 调用 ListExt
- **WHEN** 调用 `tcloud.PermissionTemplate.ListPermissionTemplateExt(kt, req)`
- **THEN** 发送 POST 到 `/permission_templates/list`，返回 `PermissionTemplateExtListResult[TCloudPermissionTemplateExtension]`

#### Scenario: Global client 调用 List
- **WHEN** 调用 `global.PermissionTemplate.ListPermissionTemplate(kt, req)`
- **THEN** 发送 POST 到 `/permission_templates/list`，返回 `PermissionTemplateListResult`（不含 extension）
