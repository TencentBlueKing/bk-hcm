## Why

权限策略库（permission_policy_library）的 data-service 层 CRUD 已完成，但 cloud-server 层尚未暴露任何 handler，导致 web-server 无法通过标准 API 提供权限策略库列表查询能力。本次需要在 cloud-server 实现 List 接口，打通从前端到数据层的查询链路。

## What Changes

- 在 `cmd/cloud-server/service/` 下新增 `permission-policy-library` 模块，注册 List 路由
- 新增 cloud-server 层响应结构体 `PermissionPolicyLibraryResult`，在 `BasePermissionPolicyLibrary` 基础上扩展 `associated_account_count` 字段（暂不赋值，留 TODO）
- 在 `cmd/cloud-server/service/service.go` 中注册新模块

## Capabilities

### New Capabilities

- `cloud-server-permission-policy-library-list`: cloud-server 层权限策略库列表查询能力，包含 vendor 过滤、IAM 鉴权（ListResourceAuthRes）、count/数据双模式支持

### Modified Capabilities

（无）

## Impact

- `pkg/api/cloud-server/` — 新增 cloud-server 层响应类型
- `cmd/cloud-server/service/permission-policy-library/` — 新增 handler 模块（service.go + list.go）
- `cmd/cloud-server/service/service.go` — 新增 import 和 InitService 调用
