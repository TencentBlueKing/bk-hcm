## ADDED Requirements

### Requirement: TCloud CAM ListPolicies adaptor 方法

系统 SHALL 在 `pkg/adaptor/tcloud/cam_list_policy.go` 中新增 `ListPolicies` 方法，封装腾讯云 CAM `GetPolicyList` API 调用，支持分页拉取指定账号下的所有策略（`Scope="Local"`，`PolicyType` 可选，不传则返回所有类型，包括预设策略 `"QCS"` 和自定义策略 `"Custom"`）。

`TCloudListPoliciesOption` 类型定义在 `pkg/adaptor/types/account/tcloud_cam_list_policy.go`：
- 字段：`Region`（string，可选，默认 `constant.TCloudDefaultRegion`）、`Page`（uint64，从 1 开始）、`Rp`（uint64，每页数量，最大 200）

`TCloudPolicyItem` 类型定义在同文件：
- 字段：`PolicyID`（uint64）、`PolicyName`（string）、`Description`（string）、`PolicyType`（string，`"Custom"` 或 `"QCS"`）、`CreateTime`（string）

方法签名：`ListPolicies(kt *kit.Kit, opt *account.TCloudListPoliciesOption) ([]account.TCloudPolicyItem, uint64, error)`，返回策略列表、总数和错误。

`TCloud` interface SHALL 新增 `ListPolicies` 方法声明。

#### Scenario: 成功拉取策略列表

- **WHEN** 传入有效的 `Page` 和 `Rp`
- **THEN** 系统 SHALL 调用 `cam.NewGetPolicyListRequest()`，设置 `Scope="Local"`，不限制 `PolicyType`，返回所有类型策略列表和总数

#### Scenario: opt 为 nil

- **WHEN** `opt` 为 nil
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: 云 API 返回错误

- **WHEN** 腾讯云 CAM API 返回错误
- **THEN** 系统 SHALL 透传云 API 错误信息

### Requirement: TCloud CAM GetPolicyDetail adaptor 方法

系统 SHALL 在 `pkg/adaptor/tcloud/cam_list_policy.go` 中新增 `GetPolicyDetail` 方法，封装腾讯云 CAM `GetPolicy` API 调用，获取单个策略的完整详情（含 `PolicyDocument`）。

`TCloudGetPolicyDetailOption` 类型定义在 `pkg/adaptor/types/account/tcloud_cam_list_policy.go`：
- 字段：`Region`（string，可选）、`PolicyID`（uint64，required）

`TCloudPolicyDetail` 类型定义在同文件：
- 字段：`PolicyID`（uint64）、`PolicyName`（string）、`PolicyDocument`（string）、`Description`（string）、`PolicyType`（string）、`CreateTime`（string）

方法签名：`GetPolicyDetail(kt *kit.Kit, opt *account.TCloudGetPolicyDetailOption) (*account.TCloudPolicyDetail, error)`

`TCloud` interface SHALL 新增 `GetPolicyDetail` 方法声明。

#### Scenario: 成功获取策略详情

- **WHEN** 传入有效的 `PolicyID`
- **THEN** 系统 SHALL 调用 `cam.NewGetPolicyRequest()`，返回含 `PolicyDocument` 的策略详情

#### Scenario: PolicyID 为 0

- **WHEN** `PolicyID` 为 0
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

#### Scenario: 策略不存在

- **WHEN** 传入不存在的 `PolicyID`
- **THEN** 系统 SHALL 透传云 API 错误信息

### Requirement: res-sync tcloud PermissionTemplate 同步方法

系统 SHALL 在 `cmd/hc-service/logics/res-sync/tcloud/permission_template.go` 中实现 `PermissionTemplate` 同步方法，执行三路对比逻辑。

`SyncPermissionTemplateOption` 结构体字段：`AccountID`（string，required）。

同步流程：
1. 调用 `listPermissionTemplateFromCloud`：分页调用 `ListPolicies` 拉取所有策略（含预设策略和自定义策略），再逐个调用 `GetPolicyDetail` 获取 `PolicyDocument`，组装为 `[]TCloudPolicyWithDetail`
2. 调用 `listPermissionTemplateFromDB`：通过 `dbCli.TCloud.PermissionTemplate.ListPermissionTemplateExt` 拉取本地 `account_id = opt.AccountID` 的所有记录
3. 三路对比：
   - 云上有、本地无 → 调用 `dbCli.TCloud.PermissionTemplate.BatchCreate` 插入（`policy_library_id` 等字段为空）
   - 云上有、本地有 → 对比 `policy_hash`，若变化则调用 `dbCli.TCloud.PermissionTemplate.BatchUpdate` 更新 `name`、`policy_document`、`memo`
   - 云上无、本地有 → 调用 `dbCli.Global.PermissionTemplate.BatchDelete` 删除

`Interface` SHALL 新增 `PermissionTemplate(kt *kit.Kit, opt *SyncPermissionTemplateOption) (*SyncResult, error)` 方法声明。

#### Scenario: 云上有、本地无

- **WHEN** 云上存在策略 A，本地 `permission_template` 表中无对应记录
- **THEN** 系统 SHALL 调用 `BatchCreate` 插入记录，`policy_library_id`、`policy_library_version`、`policy_library_sync_time` 为 nil

#### Scenario: 云上有、本地有，policy_hash 变化

- **WHEN** 云上策略 A 的 `PolicyDocument` 发生变化，本地 `policy_hash` 与云上不一致
- **THEN** 系统 SHALL 调用 `BatchUpdate` 更新 `policy_document`、`policy_hash`（由 data-service 自动重算）、`name`、`memo`

#### Scenario: 云上有、本地有，policy_hash 未变化

- **WHEN** 云上策略 A 的 `PolicyDocument` 未变化，本地 `policy_hash` 与云上一致
- **THEN** 系统 SHALL 跳过该记录，不执行任何更新

#### Scenario: 云上无、本地有

- **WHEN** 本地存在策略 B 的记录，但云上已删除
- **THEN** 系统 SHALL 调用 `BatchDelete` 删除本地记录

#### Scenario: 云上和本地均为空

- **WHEN** 云上无策略，本地也无记录
- **THEN** 系统 SHALL 直接返回，不执行任何操作

### Requirement: hc-service sync SyncPermissionTemplate handler

系统 SHALL 在 `cmd/hc-service/service/sync/tcloud/permission_template.go` 中新增 `SyncPermissionTemplate` handler，暴露 `POST /vendors/tcloud/permission_templates/sync` 接口。

handler 解析 `TCloudGlobalSyncReq`（含 `AccountID`），通过 `svc.syncCli.TCloud(cts.Kit, req.AccountID)` 获取 sync client，调用 `syncCli.PermissionTemplate(cts.Kit, &tcloud.SyncPermissionTemplateOption{AccountID: req.AccountID})`。

路由 SHALL 在 `cmd/hc-service/service/sync/tcloud/service.go` 中注册。

#### Scenario: 同步成功

- **WHEN** 发送 POST 请求，`account_id` 有效
- **THEN** 系统 SHALL 执行三路对比同步，返回 `{ "code": 0, "data": null }`

#### Scenario: account_id 缺失

- **WHEN** 请求中 `account_id` 为空
- **THEN** 系统 SHALL 返回 InvalidParameter 错误

### Requirement: hc-service client SyncPermissionTemplate 方法

系统 SHALL 在 `pkg/client/hc-service/tcloud/` 中新增 `SyncPermissionTemplate(kt *kit.Kit, req *sync.TCloudGlobalSyncReq) error` client 方法，调用 hc-service 的 `POST /vendors/tcloud/permission_templates/sync` 接口。

#### Scenario: cloud-server 通过 client 调用

- **WHEN** cloud-server 调用 `cliSet.HCService().TCloud.Account.SyncPermissionTemplate(kt, req)`
- **THEN** 请求 SHALL 正确路由到 hc-service 的 `POST /vendors/tcloud/permission_templates/sync`

### Requirement: cloud-server SyncPermissionTemplate 函数

系统 SHALL 在 `cmd/cloud-server/service/sync/tcloud/permission_template.go` 中新增 `SyncPermissionTemplate` 函数，签名与 `ResSyncFunc` 兼容（`func(kt *kit.Kit, cliSet *client.ClientSet, accountID string, regions []string, sd *detail.SyncDetail) error`）。

函数调用 `sd.ResSyncStatusSyncing`、`cliSet.HCService().TCloud.Account.SyncPermissionTemplate`、`sd.ResSyncStatusSuccess`，参考 `SyncSubAccount` 实现。

#### Scenario: 同步成功

- **WHEN** 调用 `SyncPermissionTemplate`，hc-service 同步成功
- **THEN** 函数 SHALL 调用 `sd.ResSyncStatusSuccess` 并返回 nil

#### Scenario: hc-service 返回错误

- **WHEN** hc-service 同步失败
- **THEN** 函数 SHALL 返回错误，不调用 `sd.ResSyncStatusSuccess`

### Requirement: PermissionTemplateCloudResType 枚举常量及 SyncAllResource 注册

系统 SHALL 在 `pkg/criteria/enumor/cloud_resource_type.go` 中新增 `PermissionTemplateCloudResType CloudResourceType = "permission_template"` 常量。

系统 SHALL 在 `cmd/cloud-server/service/sync/tcloud/sync_all_resource.go` 的 `syncFuncMap` 中注册 `enumor.PermissionTemplateCloudResType: SyncPermissionTemplate`，并在 `getSyncOrder` 中将其添加到 `SubAccountCloudResType` 之后。

#### Scenario: 全量同步包含权限模板

- **WHEN** 调用 `SyncAllResource`
- **THEN** 系统 SHALL 在同步子账号之后同步权限模板

#### Scenario: PermissionTemplateCloudResType 常量有效

- **WHEN** 使用 `enumor.PermissionTemplateCloudResType`
- **THEN** 其值 SHALL 为 `"permission_template"`
