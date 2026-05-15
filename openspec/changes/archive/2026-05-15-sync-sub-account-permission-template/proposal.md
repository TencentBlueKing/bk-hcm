## Why

当前系统在同步三级账号（子账号）时，缺少同步子账号绑定的权限模版信息的能力。运营人员无法在本地看到每个子账号绑定了哪些权限策略，也无法基于此做权限管理和审计。需要在子账号同步流程中新增同步每个子账号权限模版信息的操作。

## What Changes

- 在 `sub_account` 表新增 `permission_template_ids` 字段（JSON 数组），存储子账号绑定的本地权限模版 ID 列表
- 新增 TCloud CAM `ListAttachedUserAllPolicies` adaptor 方法，获取子用户绑定的策略列表（含限流重试机制）
- 在 res-sync tcloud 层新增 `SubAccountPermissionTemplate` 同步方法，同步每个子账号绑定的权限模版
- 在 `SyncSubAccount` 服务层调用链中增加 `SubAccountPermissionTemplate` 同步步骤
- 修改 sub_account 表的 DAO、API Model、Service 层以支持 `permission_template_ids` 字段的读写

## Capabilities

### New Capabilities
- `tcloud-cam-list-attached-user-policies`: TCloud CAM ListAttachedUserAllPolicies API adaptor 封装，支持分页拉取子用户绑定的策略列表，含限流重试机制
- `sync-sub-account-permission-template`: 子账号权限模版同步能力，查询子账号绑定的云上策略，匹配本地 permission_template 记录，更新 sub_account 表的 permission_template_ids 字段

### Modified Capabilities
- `sub-account-crud`: sub_account 表新增 `permission_template_ids` 字段，需扩展 SQL DDL、Table 定义、DAO、API Model、Service Handler

## Impact

- **新增代码**：
  - SQL DDL 文件（新增 `permission_template_ids` 字段）
  - `pkg/adaptor/types/account/` 新增 `TCloudListAttachedUserAllPoliciesOption`、`TCloudAttachedPolicy` 类型定义
  - `pkg/adaptor/tcloud/` 新增 `ListAttachedUserAllPolicies` 方法
  - `cmd/hc-service/logics/res-sync/tcloud/` 新增 `SubAccountPermissionTemplate` 方法
- **修改代码**：
  - `pkg/dal/table/cloud/sub_account.go` 新增 `PermissionTemplateIDs` 字段
  - `pkg/api/core/cloud/sub-account/` 新增 `permission_template_ids` 字段
  - `cmd/hc-service/service/sync/tcloud/sub_account.go` 在同步流程中增加 `SubAccountPermissionTemplate` 调用
  - `pkg/adaptor/tcloud/interface.go` 接口扩展
- **API 变更**：sub_account 相关接口的响应体增加 `permission_template_ids` 字段
- **依赖**：腾讯云 CAM SDK（`cam:ListAttachedUserAllPolicies`），已在项目中引入
