## Why

权限策略库支持将策略应用到多个二级账号，但目前缺少一个接口来查询某个策略库下已应用的权限模版列表，导致前端无法构建"可更新账号"的选择界面，Apply Update 操作缺乏前置数据支持。

## What Changes

- 新增 `GET /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}/permission_templates` 接口
- 返回指定策略库下所有**当前业务范围内**已应用的权限模版（全量，不分页）
- 在 `PolicyLibraryApplier` 上新增 `ListTemplatesInScope` 公有方法，封装业务范围过滤逻辑

## Capabilities

### New Capabilities

- `list-permission-policy-library-permission-templates`: 查询策略库下已应用且当前在业务范围内的权限模版列表

### Modified Capabilities

（无现有 spec 层需求变更）

## Impact

- `cmd/cloud-server/service/permission-policy-library/applier.go`：新增 `ListTemplatesInScope` 方法
- `cmd/cloud-server/service/permission-policy-library/list.go`：新增 handler
- `cmd/cloud-server/service/permission-policy-library/service.go`：注册新路由
- `pkg/api/cloud-server/permission_policy_library.go`：新增响应类型
