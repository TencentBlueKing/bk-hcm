## 1. Cloud-Server 响应类型

- [x] 1.1 在 `pkg/api/cloud-server/` 新增 `permission_policy_library.go`，定义 `PermissionPolicyLibraryResult`（内嵌 `BasePermissionPolicyLibrary` + `AssociatedAccountCount int`）和 `PermissionPolicyLibraryListResult`（Count + Details）

## 2. Cloud-Server Handler 模块

- [x] 2.1 在 `cmd/cloud-server/service/permission-policy-library/` 新增 `service.go`，定义 `svc` 结构体（client + authorizer）、`InitService(c)` 函数，注册路由 `POST /vendors/{vendor}/permission_policy_libraries/list` → `ListPermissionPolicyLibrary`
- [x] 2.2 在 `cmd/cloud-server/service/permission-policy-library/` 新增 `list.go`，实现 `ListPermissionPolicyLibrary` handler：提取 vendor → 解码/校验 ListReq → ListResourceAuthRes 鉴权 → tools.And 合并 vendor filter → 调用 DS SDK → 包装响应（associated_account_count=0 + TODO）

## 3. 服务注册

- [x] 3.1 在 `cmd/cloud-server/service/service.go` 中 import `permissionpolicylibrary` 包，并在 `apiSet()` 中调用 `permissionpolicylibrary.InitService(c)`
