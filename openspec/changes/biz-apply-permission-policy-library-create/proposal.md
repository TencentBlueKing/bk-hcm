## Why

业务运营人员需要通过审批流程（ITSM）为指定二级账号批量申请应用权限策略库（创建），
审批通过后系统自动在云上创建 CAM 策略并绑定到对应账号。现有的非业务接口直接执行，
缺少审批环节，不满足企业内控要求。

## What Changes

- 新增业务接口：`POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/apply_permission_policy_library_create`
- 接口接收 `policy_library_id` 和 `account_ids` 列表，为每个账号创建一个 ITSM 审批单，返回所有审批单 ID 数组
- 新增 `ApplicationType` 枚举值：`apply_permission_policy_library`（使用 Action 字段区分 create/update，枚举类型 `PermPolicyLibAction`）
- 遵循 三级账号申请单模式：base 基类 + Registry 分发 + 批量循环创建
- 审批通过后交付逻辑：复用 `PolicyLibraryApplier.ApplyCreate` 方法执行云上 CAM 策略创建

## Capabilities

### New Capabilities

- `biz-apply-permission-policy-library`: 业务侧通过 ITSM 审批流程应用权限策略库（创建）的能力，
  包含请求结构、Handler 实现、路由注册、审批交付逻辑

### Modified Capabilities

（无 spec 级别的行为变更）

## Impact

**新增代码**：
- `pkg/criteria/enumor/application.go` — 新增 `ApplyPermissionPolicyLibrary` ApplicationType
- `pkg/criteria/enumor/permission_policy_library.go` — 新增 `PermPolicyLibAction` 枚举（`apply_create` / `apply_update`）
- `pkg/api/cloud-server/application/permission_policy_library.go` — 新增 `BizApplyPermissionPolicyLibraryCreateReq` 和 `ApplyPermPolicyLibContent` 结构体
- `cmd/cloud-server/service/application/handlers/permission-policy-library/` — base + apply-create handler（init/check/create_itsm_ticket/deliver），其中 init.go 包含 `BuildContent` 辅助函数
- `cmd/cloud-server/service/application/create.go` — 新增 `CreateBizForApplyPermissionPolicyLibraryCreate` 和 `createBizForApplyPermPolicyLibCreate` 方法（返回 `*core.BatchCreateResult`）
- `cmd/cloud-server/service/application/approve.go` — 新增 `ApplyPermissionPolicyLibrary` case 分发
- `cmd/cloud-server/service/application/init.go` — 注册 biz 路由

**依赖系统**：
- ITSM 审批服务（approval_process 需配置 `apply_permission_policy_library` 类型）
- 现有 `PolicyLibraryApplier` 公共方法（`ApplyCreate` 统一方法，内部封装 TCloudCreateCAMPolicy / TCloudCreateLocalTemplate / RecordApplyAudit）
- 鉴权：`meta.PermissionPolicyLibrary` + `meta.Apply`
