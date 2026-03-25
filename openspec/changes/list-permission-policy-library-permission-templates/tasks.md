## 1. API 响应类型

- [x] 1.1 在 `pkg/api/cloud-server/permission_policy_library.go` 中新增 `PermissionPolicyLibraryPermTmplResult` 响应类型，`Details` 字段类型为 `any`（支持多云扩展）

## 2. Applier 核心逻辑

- [x] 2.1 在 `cmd/cloud-server/service/permission-policy-library/applier.go` 中新增公有方法 `ListTemplatesInScope(kt, vendor, libraryID)`
- [x] 2.2 实现：调用 `GetPolicyLibraryDetail` 获取策略库详情（含 bk_biz_ids）
- [x] 2.3 实现：调用 `listAllInScopeAccountIDs` 构建 in_scope_set
- [x] 2.4 实现：分页扫描 permission_template 表（`policy_library_id = {libraryID}`），内存过滤 AccountID ∈ in_scope_set
- [x] 2.5 实现：vendor 不为 tcloud 时返回 unsupported vendor 错误

## 3. Cloud-Server Handler

- [x] 3.1 在 `cmd/cloud-server/service/permission-policy-library/list.go` 中新增 `ListPermissionPolicyLibraryPermissionTemplates` handler
- [x] 3.2 实现：校验 vendor 路径参数
- [x] 3.3 实现：校验 id 路径参数
- [x] 3.4 实现：鉴权（`PermissionPolicyLibrary.Find + ResourceID = {id}`）
- [x] 3.5 实现：调用 `applier.ListTemplatesInScope` 并返回结果

## 4. 路由注册

- [x] 4.1 在 `cmd/cloud-server/service/permission-policy-library/service.go` 中注册新路由 `GET /vendors/{vendor}/permission_policy_libraries/{id}/permission_templates`
