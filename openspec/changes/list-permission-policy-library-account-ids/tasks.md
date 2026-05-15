## 1. 路由注册

- [x] 1.1 在 `cmd/cloud-server/service/permission-policy-library/service.go` 注册 Resource 接口路由：`GET /vendors/{vendor}/permission_policy_libraries/{id}/account_ids`
- [x] 1.2 在 `cmd/cloud-server/service/permission-policy-library/service.go` 注册 Biz 接口路由：`GET /bizs/{bk_biz_id}/vendors/{vendor}/permission_policy_libraries/{id}/account_ids`

## 2. Resource 接口实现

- [x] 2.1 在 `cmd/cloud-server/service/permission-policy-library/list.go` 新增 `ListPermissionPolicyLibraryAccountIDs` Handler
- [x] 2.2 解析 vendor、id 路径参数并做合法性校验
- [x] 2.3 鉴权：`meta.PermissionPolicyLibrary + meta.Find + ResourceID=id`
- [x] 2.4 调用 `applier.ListAllAppliedAccountIDs(kt, id)` 获取全量已应用账号 ID
- [x] 2.5 返回 `PermissionPolicyLibraryAccountIDsResult{AccountIDs: accountIDs}`

## 3. Biz 接口实现

- [x] 3.1 在 `cmd/cloud-server/service/permission-policy-library/list.go` 新增 `ListBizPermissionPolicyLibraryAccountIDs` Handler
- [x] 3.2 解析 bk_biz_id（int64）、vendor、id 路径参数并做合法性校验
- [x] 3.3 鉴权：`meta.Biz + meta.Access + BizID=bk_biz_id`
- [x] 3.4 调用 `applier.GetPolicyLibraryDetail(kt, id)` 获取策略库详情，校验 bk_biz_id 在 library.BkBizIDs 中，否则返回 400
- [x] 3.5 调用 `applier.ListAllAppliedAccountIDs(kt, id)` 获取全量已应用账号 ID
- [x] 3.6 批量查询账号表，过滤出 `bk_biz_id` == 路径参数的账号 ID
- [x] 3.7 返回过滤后的 `PermissionPolicyLibraryAccountIDsResult{AccountIDs: accountIDs}`

## 4. 辅助方法（可选重构）

- [x] 4.1 将 `applier.listAllAppliedAccountIDs` 由私有方法改为公开（或在 Handler 内联等效逻辑）

## 5. 验证

- [x] 5.1 确认代码编译通过（`go build ./cmd/cloud-server/...`）
- [x] 5.2 对照接口文档验证路由、响应结构符合预期
