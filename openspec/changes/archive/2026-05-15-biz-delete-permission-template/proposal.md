## Why

当前系统支持通过申请单流程创建和更新云权限模板，但缺少对应的删除操作。用户在不需要某个自定义权限模板时，无法走审批流程进行删除，需要补齐该能力。

## What Changes

- 新增 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/delete_permission_template` 接口，创建"删除云权限模板"的 ITSM 申请单
- 审批通过后，删除云上 CAM Policy 并删除本地权限模板记录
- 校验约束：仅允许删除自定义策略（非预设策略），且关联三级账号数为 0
- 新增 `PermTemplateActionDelete` 枚举值
- 新增 hc-service `DeleteCAMPolicy` 接口（TCloud）及对应 adaptor 方法
- 新增 application handler `ApplicationOfDeletePermTemplate`，不嵌入 `PolicyLibraryApplier`，deliver 逻辑内联

## Capabilities

### New Capabilities

- `biz-delete-permission-template`: 业务侧通过 ITSM 申请单删除云权限模板的完整审批交付流程，包含参数校验、ITSM 表单渲染、审批通过后的云资源删除和本地记录清理

### Modified Capabilities

（无现有 spec 级别的需求变更）

## Impact

- `pkg/criteria/enumor/permission_template.go` — 新增 `PermTemplateActionDelete`
- `pkg/api/cloud-server/application/permission_template.go` — 新增 `BizDeletePermissionTemplateReq`
- `pkg/adaptor/types/account/tcloud_policy.go` — 新增 `TCloudDeletePolicyOption`
- `pkg/adaptor/tcloud/cam_policy.go` — 新增 `DeletePolicy` 方法
- `pkg/api/hc-service/permission-template/permission_template.go` — 新增 `DeleteCAMPolicyReq`
- `cmd/hc-service/service/permission-template/cam_policy.go` — 新增 `TCloudDeleteCAMPolicy` handler
- `cmd/hc-service/service/permission-template/service.go` — 注册路由
- `pkg/client/hc-service/tcloud/permission_template.go` — 新增 `DeleteCAMPolicy` client 方法
- `cmd/cloud-server/service/application/handlers/permission-template/delete/` — 新增 4 个文件（init/check/deliver/create_itsm_ticket）
- `cmd/cloud-server/service/application/create.go` — 新增 `CreateBizForDeletePermissionTemplate`
- `cmd/cloud-server/service/application/init.go` — 注册路由
