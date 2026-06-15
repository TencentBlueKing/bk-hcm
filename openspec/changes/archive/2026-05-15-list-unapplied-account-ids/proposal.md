## Why

权限策略库"应用（创建/更新）"接口需要前端提供目标账号ID列表，但目前系统缺乏一个接口来查询"哪些账号还没有应用过此策略库"，导致前端无法自动筛选候选账号，用户需要手动比对。

## What Changes

- 新增 cloud-server 层 `GET /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}/unapplied_account_ids` 接口：根据策略库的业务范围（BkBizIDs）匹配二级账号，排除已在权限模板表中关联此策略库的账号，全量返回未应用账号ID列表
- 新增 `PolicyLibraryApplier` 中的辅助方法：`ListUnappliedAccountIDs`、`listAllInScopeAccountIDs`、`listAllAppliedAccountIDs`
- 新增 cloud-server 响应模型 `PermissionPolicyLibraryAccountIDsResult`

## Capabilities

### New Capabilities
- `list-unapplied-account-ids`: cloud-server 层查询未应用策略库的二级账号ID列表接口，含 applier 辅助方法和响应模型

### Modified Capabilities

## Impact

- **新增代码**：`pkg/api/cloud-server/permission_policy_library.go` 新增 `PermissionPolicyLibraryAccountIDsResult`、`cmd/cloud-server/service/permission-policy-library/` 新增 handler 文件和 applier 方法
- **修改代码**：`cmd/cloud-server/service/permission-policy-library/service.go` 新增 GET 路由
- **受影响系统**：cloud-server（新增只读查询接口）
