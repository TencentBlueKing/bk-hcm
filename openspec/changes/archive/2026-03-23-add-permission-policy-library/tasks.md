## 1. 数据层（table + DAO）

- [x] 1.1 在 `pkg/dal/table/cloud/` 新增 `permission_policy_library.go`，定义 `PermissionPolicyLibraryTable` 结构体、Column 描述符、`InsertValidate`、`UpdateValidate`、`TableName`
- [x] 1.2 在 `pkg/dal/table/table.go` 注册表名常量 `PermissionPolicyLibraryTable` 并配置 `EnableTenant: true`
- [x] 1.3 在 `pkg/dal/dao/types/` 新增 `permission_policy_library.go`，定义 `ListPermissionPolicyLibraryDetails` 类型
- [x] 1.4 在 `pkg/dal/dao/cloud/permission-policy-library/` 新增 `permission_policy_library.go`，实现 `Interface`（Create、Update、List、DeleteWithTx）
- [x] 1.5 在 `pkg/dal/dao/dao.go` 的 `Set` 接口新增 `PermissionPolicyLibrary()` 方法，并在 `set` 实现中注册 DAO 实例

## 2. API 类型定义

- [x] 2.1 在 `pkg/api/core/cloud/` 新增 `permission_policy_library.go`，定义 `BasePermissionPolicyLibrary` 核心类型
- [x] 2.2 在 `pkg/api/data-service/cloud/` 新增 `permission_policy_library.go`，定义 Create/Update/BatchDelete/List/Get 的 Req 和 Resp 结构体，包含 `Validate()` 方法

## 3. DataService Handler

- [x] 3.1 在 `cmd/data-service/service/cloud/permission-policy-library/` 新增 `service.go`，注册路由：
  - `POST /vendors/{vendor}/permission_policy_libraries/create`
  - `PATCH /vendors/{vendor}/permission_policy_libraries/{id}`
  - `DELETE /permission_policy_libraries/batch`
  - `POST /permission_policy_libraries/list`
  - `GET /permission_policy_libraries/{id}`
- [x] 3.2 新增 `create.go`，实现 Create handler（从 URL 取 vendor，计算 SHA256 hash，version=1，使用 AutoTxn）
- [x] 3.3 新增 `update.go`，实现 Update handler（先 List 查当前记录，比较 hash，条件递增 version）
- [x] 3.4 新增 `delete.go`，实现 BatchDelete handler（物理删除，使用 DeleteWithTx）
- [x] 3.5 新增 `list.go`，实现 List handler（支持 count 模式和数据模式）
- [x] 3.6 新增 `get.go`，实现 Get handler（通过 id 精确查询，不存在返回 RecordNotFound）
- [x] 3.7 在 `cmd/data-service/service/service.go` 导入并调用 `permissionpolicylibrary.InitService(capability)`

## 4. SDK 客户端

- [x] 4.1 在 `pkg/client/data-service/tcloud/` 新增 `permission_policy_library.go`，为 `restClient` 添加 `CreatePermissionPolicyLibrary` 和 `UpdatePermissionPolicyLibrary` 方法
- [x] 4.2 在 `pkg/client/data-service/global/` 新增 `permission_policy_library.go`，实现 `PermissionPolicyLibraryClient` struct，包含 `List`、`Get`、`BatchDelete` 方法
- [x] 4.3 在 `pkg/client/data-service/global/client.go` 的 `Client` struct 新增 `PermissionPolicyLibrary *PermissionPolicyLibraryClient` 字段，并在 `NewClient` 中初始化
