## 1. API 响应模型

- [x] 1.1 在 `pkg/api/cloud-server/permission_policy_library.go` 新增 `PermissionPolicyLibraryAccountIDsResult` 结构体（含 `AccountIDs []string` JSON tag: `account_ids`）

## 2. Applier 辅助方法

- [x] 2.1 在 `cmd/cloud-server/service/permission-policy-library/applier.go` 新增私有方法 `listAllAppliedAccountIDs(kt, libraryID)` — 分页扫描 permission_template 表，返回去重后的 account_id 列表
- [x] 2.2 在 `applier.go` 新增私有方法 `listAllInScopeAccountIDs(kt, vendor, bizIDs)` — 分页扫描账号表（vendor + bk_biz_id 过滤），返回去重后的 account_id 列表
- [x] 2.3 在 `applier.go` 新增公开方法 `ListUnappliedAccountIDs(kt, vendor, libraryID)` — 调用上述两个方法并通过 `slice.NotIn` 计算差集

## 3. Handler

- [x] 3.1 在 `cmd/cloud-server/service/permission-policy-library/list.go` 追加 `ListPermissionPolicyLibraryUnappliedAccountIDs` handler（vendor 校验 + id 校验 + Find 鉴权 + 调用 applier）

## 4. 路由注册

- [x] 4.1 在 `cmd/cloud-server/service/permission-policy-library/service.go` 注册 `GET /vendors/{vendor}/permission_policy_libraries/{id}/unapplied_account_ids` 路由
