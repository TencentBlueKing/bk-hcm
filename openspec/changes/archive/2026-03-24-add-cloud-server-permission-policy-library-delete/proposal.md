## Why

cloud-server 层的权限策略库（permission_policy_library）已完成 Create / List / Update 接口，但缺少 Delete 接口。用户无法通过 API 删除不再需要的策略库记录，需要补齐 CRUD 闭环。

## What Changes

- 在 cloud-server 层新增 `DELETE /vendors/{vendor}/permission_policy_libraries/{id}` 接口
- 删除前查询记录并校验路径参数 vendor 与记录实际 vendor 一致
- 删除前预留云权限模板关联检查位（当前跳过，留 TODO）
- 删除前记录审计日志
- 通过 Global Client 调用 data-service 层已有的 BatchDelete 完成物理删除

## Capabilities

### New Capabilities
- `cloud-server-permission-policy-library-delete`: 权限策略库删除接口，包含参数校验、鉴权、vendor 匹配校验、审计、物理删除

### Modified Capabilities

（无）

## Impact

- `cmd/cloud-server/service/permission-policy-library/service.go`：注册 Delete 路由
- `cmd/cloud-server/service/permission-policy-library/delete.go`：新增 Delete handler
- 底层 data-service / DAO / Client SDK 已就绪，无需变更
