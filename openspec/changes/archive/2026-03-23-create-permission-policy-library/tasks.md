## 1. API 模型定义

- [x] 1.1 在 `pkg/api/cloud-server/permission_policy_library.go` 中新增 `PermissionPolicyLibraryCreateReq` 结构体（含 Name、PolicyDocument、BkBizIDs、Memo 字段及 validate tag）和 `Validate()` 方法
- [x] 1.2 在 `pkg/api/cloud-server/permission_policy_library.go` 中新增 `PermissionPolicyLibraryCreateResult` 结构体（含 ID 字段）

## 2. Handler 实现

- [x] 2.1 新建 `cmd/cloud-server/service/permission-policy-library/create.go`，实现 `CreatePermissionPolicyLibrary` Handler：vendor 校验、请求解码、参数校验、IAM 鉴权（AuthorizeWithPerm）、构造 DS BatchCreate 请求、调用 TCloud client、返回单个 ID

## 3. 路由注册

- [x] 3.1 在 `cmd/cloud-server/service/permission-policy-library/service.go` 的 `InitService` 中新增 Create 路由：`h.Add("CreatePermissionPolicyLibrary", http.MethodPost, "/vendors/{vendor}/permission_policy_libraries/create", svc.CreatePermissionPolicyLibrary)`
