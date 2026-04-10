## Why

"应用权限策略库（更新）"接口当前接收 `account_ids`（二级账号 ID 列表），但更新场景的语义是对**已应用的权限模版**进行同步更新——权限模版 ID 才是精准标识符，通过它可以直接定位到账号和云策略，无需再做关联查询。

## What Changes

- `BizApplyPermissionPolicyLibraryUpdateReq`：请求参数 `account_ids` → `permission_template_ids`
- `ApplyPermissionPolicyLibraryReq`（resource 级）：请求参数 `account_ids` → `permission_template_ids`（仅 update 接口）
- 新增 `ApplyTemplateResult` 响应结构（`permission_template_id` + `status` + `reason`），替换 update 路径上的 `ApplyAccountResult`
- 新增 `ApplyPermissionPolicyLibraryUpdateResult` 响应结构
- 拆分 `ApplyPermPolicyLibContent`（当前 create/update 共用）为独立结构体：
  - `ApplyPermPolicyLibBaseContent`（仅含 Action/Vendor/BkBizID，用于 handler dispatch）
  - `ApplyPermPolicyLibCreateContent`（原 Content，含 AccountID）
  - `ApplyPermPolicyLibUpdateContent`（含 PermissionTemplateID，替换 AccountID）
- Handler 工厂签名改为 sub-account 模式（factory 接收 base + rawContent，内部二次反序列化）
- `apply-update` handler 的 CheckReq、Deliver、ITSM 渲染全部改为基于 PermissionTemplateID 操作
- applier 新增 `ApplyUpdateByTemplateIDs` 方法，删除旧 `ApplyUpdate` 方法
- 两份接口文档同步更新

## Capabilities

### New Capabilities

- `apply-perm-lib-update-by-template`: 应用权限策略库（更新）接口改为以 permission_template_id 数组作为入参及响应标识符，涵盖 Biz 申请单流程和 Resource 直接执行路径

### Modified Capabilities

- `cloud-server-permission-policy-library-update`: apply update 接口的请求/响应字段变更（account_ids → permission_template_ids，ApplyAccountResult → ApplyTemplateResult）

## Impact

- **API 破坏性变更**：`PUT /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}/apply` 和 `POST .../apply_permission_policy_library_update` 的请求/响应字段均变更（**BREAKING**）
- **影响文件**：
  - `pkg/api/cloud-server/application/permission_policy_library.go`
  - `pkg/api/cloud-server/permission_policy_library.go`
  - `cmd/cloud-server/service/application/handlers/permission-policy-library/base.go`
  - `cmd/cloud-server/service/application/handlers/permission-policy-library/apply-create/init.go`
  - `cmd/cloud-server/service/application/handlers/permission-policy-library/apply-update/`（全部 4 个文件）
  - `cmd/cloud-server/service/permission-policy-library/applier.go`
  - `cmd/cloud-server/service/application/create.go`
  - `cmd/cloud-server/service/permission-policy-library/apply.go`
  - `docs/api-docs/web-server/docs/biz/permission-policy-library/apply_permission_policy_library_update.md`
  - `docs/api-docs/web-server/docs/resource/permission-policy-library/apply_permission_policy_library_update.md`
