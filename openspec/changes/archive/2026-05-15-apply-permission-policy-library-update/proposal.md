## Why

系统已实现权限策略库"应用（创建）"接口，将策略库的内容首次同步到二级账号（创建 CAM 策略 + 本地模板）。当策略库内容发生版本升级后，需要一个配套的"应用（更新）"接口，将最新策略内容同步更新到已应用的二级账号——覆盖云上 CAM 策略内容并刷新本地模板的版本信息。

## What Changes

- 新增 TCloud CAM `UpdatePolicy` adaptor 封装（类型定义 + 方法实现）：`TCloudUpdatePolicyOption` 含 `Region string`、`PolicyID uint64`（required）、`PolicyDocument *string`、`Description *string`
- 新增 hc-service 层 CAM 策略更新接口（`PATCH /vendors/tcloud/permission_templates/cam/update_policy`），供 cloud-server 调用：`UpdateCAMPolicyReq` 含 `AccountID`、`PolicyID`（required）、`PolicyDocument *string`、`Description *string`（至少提供一个）
- 新增 hc-service client 方法（`UpdateCAMPolicy`）
- 扩展 cloud-server 层 `applier.go` 公共逻辑，新增 `GetAccountTemplate`、`TCloudUpdateCAMPolicy`（同时更新 PolicyDocument 和 Description）、`TCloudUpdateLocalTemplate`（更新 PolicyDocument、PolicyLibraryVersion、PolicyLibrarySyncTime、Memo）、`ApplyUpdate` 系列方法
- 新增 cloud-server 层 `ApplyPermissionPolicyLibraryUpdate` handler（`PUT /vendors/{vendor}/permission_policy_libraries/{id}/apply`）
- 在路由注册中新增 `PUT` 路由

## Capabilities

### New Capabilities
- `apply-permission-policy-library-update`: cloud-server 层"应用权限策略库（更新）"接口，含 TCloud CAM UpdatePolicy adaptor、hc-service 更新接口及 cloud-server applier 扩展

### Modified Capabilities
- `permission-template-crud`: applier 新增 `GetAccountTemplate`（查询单个账号的已应用模板记录），`BatchUpdate` 路径已有但通过 applier 封装新调用链路

## Impact

- **新增代码**：`pkg/adaptor/types/account/` 新增 UpdatePolicy 类型、`pkg/adaptor/tcloud/cam_policy.go` 新增 `UpdatePolicy` 方法、`cmd/hc-service/service/permission-template/cam_policy.go` 新增 handler、`pkg/api/hc-service/permission-template/` 新增 Update 请求模型、`pkg/client/hc-service/tcloud/permission_template.go` 新增 client 方法、`cmd/cloud-server/service/permission-policy-library/applier.go` 新增方法、`cmd/cloud-server/service/permission-policy-library/apply.go` 新增 handler
- **修改代码**：`pkg/adaptor/tcloud/interface.go` 新增 `UpdatePolicy` 接口声明、`cmd/hc-service/service/permission-template/service.go` 新增路由、`cmd/cloud-server/service/permission-policy-library/service.go` 新增 PUT 路由
- **受影响系统**：hc-service（新增 API）、cloud-server（新增路由）
