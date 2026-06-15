## Why

系统已具备权限模板（`permission_template`）的 CRUD 能力，但缺少将云上 CAM 策略自动同步到本地的定时同步机制。当二级账号下的 CAM 策略在云上发生变化（新增、修改、删除）时，本地 `permission_template` 表无法感知，导致本地数据与云上状态不一致。需要实现权限模板的定时同步逻辑，使本地数据与腾讯云 CAM 策略保持一致。

## What Changes

- 新增 TCloud CAM `ListPolicies` adaptor 封装（`pkg/adaptor/types/account/` 类型定义 + `pkg/adaptor/tcloud/` 方法实现），调用腾讯云 [GetPolicyList](https://cloud.tencent.com/document/product/598/34574) 接口分页拉取二级账号下所有策略
- 新增 TCloud CAM `GetPolicyDetail` adaptor 封装，调用腾讯云 [GetPolicy](https://cloud.tencent.com/document/product/598/34570) 接口获取策略详情（含 `PolicyDocument`）
- 在 `pkg/adaptor/tcloud/interface.go` 的 `TCloud` 接口中新增 `ListPolicies` 和 `GetPolicyDetail` 方法声明
- 新增 res-sync tcloud 层 `PermissionTemplate` 同步方法（`cmd/hc-service/logics/res-sync/tcloud/permission_template.go`），实现三路对比逻辑（云上有本地无→插入、云上有本地有→对比 hash 更新、云上无本地有→删除）
- 在 res-sync tcloud `Interface` 中新增 `PermissionTemplate` 方法声明
- 新增 hc-service sync 层 `SyncPermissionTemplate` handler（`cmd/hc-service/service/sync/tcloud/permission_template.go`），注册路由 `POST /vendors/tcloud/permission_templates/sync`
- 新增 hc-service client 方法 `SyncPermissionTemplate`（`pkg/client/hc-service/tcloud/`）
- 新增 cloud-server sync 层 `SyncPermissionTemplate` 函数（`cmd/cloud-server/service/sync/tcloud/permission_template.go`），参考 `SyncSubAccount` 实现
- 在 cloud-server `SyncAllResource` 的 `syncFuncMap` 和 `getSyncOrder` 中注册 `PermissionTemplateCloudResType`
- 新增 `PermissionTemplateCloudResType` 枚举常量（`pkg/criteria/enumor/cloud_resource_type.go`）

## Capabilities

### New Capabilities
- `tcloud-permission-template-sync`: 腾讯云权限模板定时同步能力，含 CAM ListPolicies/GetPolicyDetail adaptor 封装、res-sync 同步逻辑、hc-service sync handler 及 cloud-server 同步入口

### Modified Capabilities
- `permission-template-crud`: 同步逻辑复用现有的 `BatchCreate`、`BatchUpdate`、`BatchDelete` 接口，无需修改 CRUD 接口本身

## Impact

- **新增代码**：
  - `pkg/adaptor/types/account/tcloud_cam_list_policy.go`：`TCloudListPoliciesOption`、`TCloudPolicyItem`、`TCloudGetPolicyDetailOption`、`TCloudPolicyDetail` 类型定义
  - `pkg/adaptor/tcloud/cam_list_policy.go`：`ListPolicies`、`GetPolicyDetail` 方法实现
  - `cmd/hc-service/logics/res-sync/tcloud/permission_template.go`：`SyncPermissionTemplateOption`、`PermissionTemplate` 同步方法及辅助函数
  - `cmd/hc-service/service/sync/tcloud/permission_template.go`：`SyncPermissionTemplate` handler
  - `cmd/cloud-server/service/sync/tcloud/permission_template.go`：`SyncPermissionTemplate` 函数
- **修改代码**：
  - `pkg/adaptor/tcloud/interface.go`：新增 `ListPolicies`、`GetPolicyDetail` 接口声明
  - `cmd/hc-service/logics/res-sync/tcloud/client.go`：新增 `PermissionTemplate` 接口声明
  - `cmd/hc-service/service/sync/tcloud/service.go`：注册新路由
  - `pkg/client/hc-service/tcloud/`：新增 `SyncPermissionTemplate` client 方法
  - `cmd/cloud-server/service/sync/tcloud/sync_all_resource.go`：注册 `PermissionTemplateCloudResType`
  - `pkg/criteria/enumor/cloud_resource_type.go`：新增 `PermissionTemplateCloudResType` 常量
- **受影响系统**：hc-service（新增同步 API）、cloud-server（全量同步流程扩展）
- **依赖**：腾讯云 CAM SDK（`cam.GetPolicyList`、`cam.GetPolicy`），已在项目中引入
