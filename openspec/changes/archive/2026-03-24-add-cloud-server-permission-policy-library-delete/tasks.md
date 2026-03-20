## 1. 路由注册

- [x] 1.1 在 `cmd/cloud-server/service/permission-policy-library/service.go` 的 `InitService` 中新增 `h.Add("DeletePermissionPolicyLibrary", http.MethodDelete, "/vendors/{vendor}/permission_policy_libraries/{id}", svc.DeletePermissionPolicyLibrary)`

## 2. Delete Handler 实现

- [x] 2.1 新建 `cmd/cloud-server/service/permission-policy-library/delete.go`，实现 `DeletePermissionPolicyLibrary` handler：解析路径参数 vendor 和 id、校验 vendor 枚举合法性、校验 id 非空
- [x] 2.2 在 handler 中完成 IAM 鉴权：`authorizer.AuthorizeWithPerm` 使用 `ResourceAttribute{Type: PermissionPolicyLibrary, Action: Delete}`
- [x] 2.3 在 handler 中查询记录并校验 vendor：通过 `Global.PermissionPolicyLibrary.ListPermissionPolicyLibrary` 查询 `filter: id = {id}`，校验记录存在 + vendor 匹配
- [x] 2.4 预留云权限模板关联检查 TODO 注释
- [x] 2.5 在 handler 中记录删除审计：`svc.audit.ResDeleteAudit(kt, PermissionPolicyLibraryAuditResType, []string{id})`
- [x] 2.6 构造 `PermissionPolicyLibraryBatchDeleteReq`（filter: id = {id}），调用 `Global.PermissionPolicyLibrary.BatchDelete` 完成物理删除
