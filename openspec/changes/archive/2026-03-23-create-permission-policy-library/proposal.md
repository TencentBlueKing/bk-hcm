## Why

cloud-server 层目前仅实现了权限策略库的 List 接口，缺少 Create 接口。data-service 层的批量创建已完备，但前端无法通过 web-server API 创建权限策略库。需要在 cloud-server 层新增 Create 路由，完成从 API 网关到数据服务的完整链路。

## What Changes

- 在 cloud-server 层新增 `POST /vendors/{vendor}/permission_policy_libraries/create` 路由和 Handler
- 新增 cloud-server 层的 Create 请求/响应 API 模型（单个创建语义，区别于 DS 层的批量语义）
- Handler 内完成：vendor 校验、请求解码校验、IAM 鉴权（纯资源类型+动作，不依赖 accountID）、单个→批量适配、调用 DS TCloud client、返回单个 ID

## Capabilities

### New Capabilities
- `cloud-server-permission-policy-library-create`: cloud-server 层创建权限策略库接口，包含路由注册、鉴权、请求适配和响应转换

### Modified Capabilities

## Impact

- `cmd/cloud-server/service/permission-policy-library/service.go`：新增 Create 路由注册
- `cmd/cloud-server/service/permission-policy-library/create.go`：新增文件，Create Handler 实现
- `pkg/api/cloud-server/permission_policy_library.go`：新增 CS 层 Create 请求/响应模型
