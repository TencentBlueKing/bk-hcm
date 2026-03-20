## Why

cloud-server 层的 permission policy library 目前仅实现了 Create 和 List 接口，缺少 Update 接口。API 文档（`PATCH /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}`）已编写，data-service 层的 BatchUpdate 链路（handler → DAO → DB）也已就绪，但 cloud-server 层没有对应的路由、handler 和 API Model，导致前端无法调用更新操作。同时，现有的 cloud-server svc struct 未集成审计组件，Update 操作将缺少审计追踪。

## What Changes

- 在 `pkg/api/cloud-server/permission_policy_library.go` 新增 `PermissionPolicyLibraryUpdateReq` 请求模型（name、policy_document、bk_biz_ids、memo 均可选）
- 在 `cmd/cloud-server/service/permission-policy-library/` 新增 `update.go`，实现 `UpdatePermissionPolicyLibrary` handler：路径参数解析（vendor + id）、请求解码校验、IAM 鉴权、审计记录、单条→批量请求适配、调用 DS TCloud client
- 修改 `cmd/cloud-server/service/permission-policy-library/service.go`：注册 PATCH 路由、svc struct 添加 audit 字段并初始化

## Capabilities

### New Capabilities
- `cloud-server-permission-policy-library-update`: cloud-server 层的权限策略库 Update 接口，包括路由注册、handler 实现、API Model 定义、IAM 鉴权和审计记录

### Modified Capabilities

## Impact

- `pkg/api/cloud-server/permission_policy_library.go`：新增 UpdateReq struct
- `cmd/cloud-server/service/permission-policy-library/service.go`：新增路由、audit 字段
- `cmd/cloud-server/service/permission-policy-library/update.go`：新文件
- `cmd/cloud-server/service/permission-policy-library/create.go`：需适配 svc struct 新增的 audit 字段（无逻辑变化）
- 依赖已有的 `audit.Interface`（`cmd/cloud-server/logics/audit/`）和 `PermissionPolicyLibraryAuditResType`
- 依赖已有的 DS TCloud client `BatchUpdate` 方法
