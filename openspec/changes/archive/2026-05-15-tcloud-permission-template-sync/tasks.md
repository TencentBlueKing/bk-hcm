## 1. TCloud CAM Adaptor 封装

- [x] 1.1 在 `pkg/adaptor/types/account/tcloud_cam_list_policy.go` 新增 `TCloudListPoliciesOption`（`Region string`、`Page uint64`、`Rp uint64`）、`TCloudPolicyItem`（`PolicyID uint64`、`PolicyName string`、`Description string`、`PolicyType string`、`CreateTime string`）类型定义
- [x] 1.2 在 `pkg/adaptor/types/account/tcloud_cam_list_policy.go` 新增 `TCloudGetPolicyDetailOption`（`Region string`、`PolicyID uint64`）、`TCloudPolicyDetail`（`PolicyID uint64`、`PolicyName string`、`PolicyDocument string`、`Description string`、`PolicyType string`、`CreateTime string`）类型定义，并为 `TCloudGetPolicyDetailOption` 添加 `Validate` 方法
- [x] 1.3 在 `pkg/adaptor/tcloud/cam_list_policy.go` 实现 `ListPolicies` 方法，调用 `cam.NewListPoliciesRequest()`，设置 `Scope="Local"`，支持分页，返回策略列表和总数
- [x] 1.4 在 `pkg/adaptor/tcloud/cam_list_policy.go` 实现 `GetPolicyDetail` 方法，调用 `cam.NewGetPolicyRequest()`，返回含 `PolicyDocument` 的策略详情
- [x] 1.5 在 `pkg/adaptor/tcloud/interface.go` 的 `TCloud` 接口中新增 `ListPolicies` 和 `GetPolicyDetail` 方法声明

## 2. res-sync tcloud PermissionTemplate 同步逻辑

- [x] 2.1 新建 `cmd/hc-service/logics/res-sync/tcloud/permission_template.go`，定义 `SyncPermissionTemplateOption`（`AccountID string`）及 `Validate` 方法
- [x] 2.2 实现 `listPermissionTemplateFromCloud` 方法：分页调用 `cli.cloudCli.ListPolicies` 拉取所有自定义策略（`PolicyType == "Custom"`），再逐个调用 `cli.cloudCli.GetPolicyDetail` 获取 `PolicyDocument`，组装为带 hash 的策略列表
- [x] 2.3 实现 `listPermissionTemplateFromDB` 方法：通过 `cli.dbCli.TCloud.PermissionTemplate.ListPermissionTemplateExt` 分页拉取本地 `account_id = opt.AccountID` 的所有记录
- [x] 2.4 实现 `createPermissionTemplate` 方法：将云上新增策略批量调用 `cli.dbCli.TCloud.PermissionTemplate.BatchCreate` 插入，`policy_library_id`、`policy_library_version`、`policy_library_sync_time` 为 nil
- [x] 2.5 实现 `updatePermissionTemplate` 方法：对 policy_hash 有变化的记录批量调用 `cli.dbCli.TCloud.PermissionTemplate.BatchUpdate` 更新 `name`、`policy_document`、`memo`
- [x] 2.6 实现 `deletePermissionTemplate` 方法：调用 `cli.dbCli.Global.PermissionTemplate.BatchDelete` 删除云上已不存在的本地记录
- [x] 2.7 实现 `PermissionTemplate` 主方法：调用上述辅助方法，执行三路对比逻辑（新增/更新/删除）
- [x] 2.8 在 `cmd/hc-service/logics/res-sync/tcloud/client.go` 的 `Interface` 中新增 `PermissionTemplate(kt *kit.Kit, opt *SyncPermissionTemplateOption) (*SyncResult, error)` 方法声明

## 3. hc-service sync handler 及路由注册

- [x] 3.1 新建 `cmd/hc-service/service/sync/tcloud/permission_template.go`，实现 `SyncPermissionTemplate` handler：解析 `TCloudGlobalSyncReq`，获取 sync client，调用 `syncCli.PermissionTemplate`
- [x] 3.2 在 `cmd/hc-service/service/sync/tcloud/service.go` 中注册路由 `h.Add("SyncPermissionTemplate", http.MethodPost, "/permission_templates/sync", v.SyncPermissionTemplate)`

## 4. hc-service client 封装

- [x] 4.1 在 `pkg/client/hc-service/tcloud/` 中找到 Account client 文件（或新建），新增 `SyncPermissionTemplate(kt *kit.Kit, req *sync.TCloudGlobalSyncReq) error` 方法，调用 `POST /permission_templates/sync`

## 5. cloud-server sync 层

- [x] 5.1 新建 `cmd/cloud-server/service/sync/tcloud/permission_template.go`，实现 `SyncPermissionTemplate` 函数（签名与 `ResSyncFunc` 兼容），参考 `SyncSubAccount` 实现，调用 `cliSet.HCService().TCloud.Account.SyncPermissionTemplate`
- [x] 5.2 在 `pkg/criteria/enumor/cloud_resource_type.go` 中新增 `PermissionTemplateCloudResType CloudResourceType = "permission_template"` 常量
- [x] 5.3 在 `cmd/cloud-server/service/sync/tcloud/sync_all_resource.go` 的 `syncFuncMap` 中注册 `enumor.PermissionTemplateCloudResType: SyncPermissionTemplate`，并在 `getSyncOrder` 中将其添加到 `SubAccountCloudResType` 之后
