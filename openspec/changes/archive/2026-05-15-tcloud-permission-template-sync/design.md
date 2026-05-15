## Context

系统已具备权限模板（`permission_template`）的完整 CRUD 能力（data-service 层），以及 TCloud CAM 策略的创建/更新 adaptor 封装（`CreatePolicy`、`UpdatePolicy`）。但目前缺少将云上 CAM 策略自动同步到本地 `permission_template` 表的定时同步机制。

当前代码库中：
- `pkg/adaptor/tcloud/cam_policy.go` 已有 `CreatePolicy`、`UpdatePolicy`，但无 `ListPolicies`（分页拉取策略列表）和 `GetPolicyDetail`（获取策略详情含 `PolicyDocument`）
- res-sync tcloud `client` 已有 `SubAccount`、`Account` 等同步方法，`Interface` 定义在 `cmd/hc-service/logics/res-sync/tcloud/client.go`
- hc-service sync 层已有 `SyncSubAccount` handler，参考路由 `POST /vendors/tcloud/sub_accounts/sync`
- cloud-server sync 层 `SyncAllResource` 已有 `syncFuncMap` 和 `getSyncOrder`，可直接扩展
- data-service client 已有 `TCloud.PermissionTemplate.BatchCreate`、`BatchUpdate`、`ListPermissionTemplateExt` 和 `Global.PermissionTemplate.BatchDelete`
- `PermissionTemplateCloudResType` 枚举常量尚未定义

同步逻辑的核心是：拉取云上该二级账号下的所有策略列表（含预设策略和自定义策略，通过 `GetPolicyList` 的 `Scope` 和 `PolicyType` 参数控制），与本地 `permission_template` 表（`account_id = 该账号`）做三路对比（新增/更新/删除）。

## Goals / Non-Goals

**Goals:**
- 实现 TCloud CAM `ListPolicies`（分页拉取策略列表）和 `GetPolicyDetail`（获取策略详情）adaptor 封装
- 实现 res-sync tcloud 层 `PermissionTemplate` 同步方法，含三路对比逻辑
- 在 hc-service sync 层暴露 `POST /vendors/tcloud/permission_templates/sync` 接口
- 在 cloud-server `SyncAllResource` 中注册权限模板同步，纳入全量同步流程
- 新增 `PermissionTemplateCloudResType` 枚举常量

**Non-Goals:**
- 不修改 `permission_template` 的 CRUD 接口
- 不实现 CAM 策略的创建/更新/删除（仅同步，不写回云上）
- 不实现其他云厂商的权限模板同步

## Decisions

### Decision 1：ListPolicies 分两步：先 List 获取 PolicyID，再 GetPolicyDetail 获取 PolicyDocument

**选择**：腾讯云 `GetPolicyList`（[文档](https://cloud.tencent.com/document/product/598/34574)）返回策略列表但不含 `PolicyDocument`，支持通过 `PolicyType` 参数过滤策略类型（`"Custom"` 自定义策略、`"QCS"` 预设策略，不传则返回所有类型）；`GetPolicy`（[文档](https://cloud.tencent.com/document/product/598/34570)）可获取单个策略的完整详情含 `PolicyDocument`。因此同步流程为：先分页拉取所有策略 ID 和基础信息（不限 `PolicyType`，同步全部类型），再逐个（或批量）获取 `PolicyDocument`。

**理由**：`GetPolicyList` 不返回 `PolicyDocument`，必须通过 `GetPolicy` 单独获取。`policy_hash` 基于 `PolicyDocument` 计算，是判断是否需要更新的依据。

**替代方案**：仅用 `GetPolicyList` 对比名称/描述。否决：无法对比策略内容变化，会漏掉策略文档变更。

### Decision 2：同步逻辑放在 res-sync tcloud 层，参考 SubAccount 实现

**选择**：在 `cmd/hc-service/logics/res-sync/tcloud/permission_template.go` 实现 `PermissionTemplate` 方法，使用 `common.Diff` 做三路对比，复用现有的 `listPermissionTemplateFromDB`（通过 `dbCli.TCloud.PermissionTemplate.ListPermissionTemplateExt`）和 `dbCli.Global.PermissionTemplate.BatchDelete`。

**理由**：与 `SubAccount` 同步逻辑结构一致，复用 `common.Diff` 工具函数，代码风格统一。

**替代方案**：直接在 hc-service handler 中实现同步逻辑。否决：违反分层架构，逻辑应在 res-sync 层。

### Decision 3：hc-service sync handler 使用 TCloudGlobalSyncReq（仅含 AccountID）

**选择**：权限模板是账号级别的资源（不分 Region），使用 `TCloudGlobalSyncReq`（含 `AccountID`），与 `SyncSubAccount` 保持一致。

**理由**：CAM 策略不区分 Region，无需传入 Region 参数。

### Decision 4：cloud-server SyncPermissionTemplate 参考 SyncSubAccount 实现

**选择**：在 `cmd/cloud-server/service/sync/tcloud/permission_template.go` 实现 `SyncPermissionTemplate` 函数，调用 `cliSet.HCService().TCloud.Account.SyncPermissionTemplate`（或新增专用 client 方法），并在 `syncFuncMap` 中注册。

**理由**：与现有 `SyncSubAccount` 实现完全对称，复用 `SyncDetail` 状态管理。

### Decision 5：对比逻辑使用 policy_hash 判断是否需要更新

**选择**：云上拉取到 `PolicyDocument` 后计算 SHA256 hash，与本地 `policy_hash` 对比。若不同则更新 `policy_document` 和 `policy_hash`（以及 `name`、`memo`）。

**理由**：`policy_hash` 已在 data-service 层自动维护，是判断内容变化的最高效方式，避免全量字符串对比。

## Risks / Trade-offs

- **GetPolicy 调用量大**：若账号下有大量策略（含预设和自定义），需逐个调用 `GetPolicy` 获取详情，可能触发 API 限频。缓解：分批调用，每批间隔适当时间；利用 `policy_hash` 对比，仅对 hash 变化的策略调用 `GetPolicy`（但首次同步仍需全量拉取）。
- **policy_library_id 等字段不覆盖**：同步时不修改 `policy_library_id`、`policy_library_version`、`policy_library_sync_time`，这些字段由"应用策略库"流程管理，同步逻辑不干预。

## Migration Plan

- 无数据迁移需求，新增代码不影响现有数据
- 部署后，定时同步任务会在下次触发时自动同步权限模板数据

## Open Questions

- 无
