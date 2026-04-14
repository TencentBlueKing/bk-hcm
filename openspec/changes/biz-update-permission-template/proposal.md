## Why

业务侧已有创建云权限模板的能力，但缺少更新能力。当自定义权限模板（`policy_library_id` 为空、TCloud `cloud_type=1`）需要绑定到新的权限策略库时，用户目前无法通过审批流进行变更，只能手动操作或重建模板。

## What Changes

- 新增业务侧更新云权限模板接口：`POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/update_permission_template`
- 接口通过 ITSM 审批流执行，审批通过后：用新策略库的策略内容更新云端 CAM Policy，同时更新本地权限模板记录（含 `policy_library_id` 关联、`policy_document`、`memo` 等）
- 在 `applier.go` 中新增 `ApplyUpdateWithTmplInfo` 方法，专门处理「自定义模板绑定策略库并更新」的场景
- 新增 `UpdateTmplBaseInfo` 结构体，将 memo 等可覆盖字段封装传入 `updateTCloudCAMPolicy` 和 `updateTCloudLocalTemplate`，支持 memo 和 `policy_library_id` 写入

## Capabilities

### New Capabilities

- `biz-update-permission-template`: 业务侧更新云权限模板申请单接口，包含 CheckReq 校验（自定义模板判断）、ITSM 渲染、Deliver 执行（云端+本地更新）

### Modified Capabilities

- `permission-template-crud`: `updateTCloudCAMPolicy` 和 `updateTCloudLocalTemplate` 改用 `UpdateTmplBaseInfo` 结构体传参，支持 memo 覆盖；`updateTCloudLocalTemplate` 始终写入 `policy_library_id` 字段

## Impact

- **新增文件**：`handlers/permission-template/update/` 下 `init.go`、`check.go`、`create_itsm_ticket.go`、`deliver.go`
- **修改文件**：
  - `cmd/cloud-server/service/permission-policy-library/applier.go`：新增 `UpdateTmplBaseInfo` 结构体、`ApplyUpdateWithTmplInfo` 方法；将 `updateTCloudCAMPolicy` 和 `updateTCloudLocalTemplate` 改用 `UpdateTmplBaseInfo` 传参；`ApplyUpdate` 使用 `UpdateTmplBaseInfo{Memo: library.Memo}` 保持向后兼容；`CheckPermTmplUpdatability` 扩展为同时支持自定义模板和已绑同一策略库的模板
  - `pkg/criteria/enumor/permission_template.go`：新增 `PermTemplateActionUpdate`
  - `pkg/api/cloud-server/application/permission_template.go`：新增 `BizUpdatePermissionTemplateReq`
  - `cmd/cloud-server/service/application/create.go`：新增 handler 函数
  - `cmd/cloud-server/service/application/init.go`：注册路由
  - `cmd/cloud-server/service/application/approve.go`：blank import 注册
