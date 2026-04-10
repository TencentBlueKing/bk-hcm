## 1. API Model 定义

- [x] 1.1 在 `pkg/api/cloud-server/permission_policy_library.go` 新增 `PermissionPolicyLibraryUpdateReq` struct（name omitempty max=128、policy_document omitempty、bk_biz_ids omitempty、memo *string omitempty max=255）及 `Validate()` 方法

## 2. Service 层路由与结构

- [x] 2.1 修改 `cmd/cloud-server/service/permission-policy-library/service.go`：svc struct 新增 `audit audit.Interface` 字段，InitService 中从 `c.Audit` 初始化
- [x] 2.2 在 service.go 的 handler 注册中新增 `PATCH /vendors/{vendor}/permission_policy_libraries/{id}` 路由，指向 `svc.UpdatePermissionPolicyLibrary`

## 3. Handler 实现

- [x] 3.1 新建 `cmd/cloud-server/service/permission-policy-library/update.go`，实现 `UpdatePermissionPolicyLibrary` handler：解析路径参数 vendor 和 id、解码请求体、校验
- [x] 3.2 在 handler 中完成 IAM 鉴权：`authorizer.AuthorizeWithPerm` 使用 `ResourceAttribute{Type: PermissionPolicyLibrary, Action: Update}`
- [x] 3.3 在 handler 中记录审计：`converter.StructToMap(req)` 获取变更字段 → `svc.audit.ResUpdateAudit(kt, PermissionPolicyLibraryAuditResType, id, updateFields)`
- [x] 3.4 在 handler 中完成单条→批量适配：构造 `PermissionPolicyLibraryBatchUpdateReq`，调用 `svc.client.DataService().TCloud.PermissionPolicyLibrary.BatchUpdate()`，按 vendor switch 分发（当前仅 tcloud）
