## Why

资源运营管理员在创建三级账号时，需要同时为其绑定预设的权限模版（permission_template），以简化权限分配流程。同时，在更新三级账号时，也需要支持权限模版的更新。当前系统已支持在创建请求中传入 `permission_template_ids`，并在 `sub_account` 表中存储该关联关系，但缺少：
1. 实际调用腾讯云 CAM API 将策略绑定到子用户的流程
2. 更新流程中对 `permission_template_ids` 的支持和校验
3. 工单渲染中展示权限模版名称的功能

需要实现：
- `attachPermissionToCloud` 方法，查询权限模版获取云上策略 ID，并通过 hc-service 调用腾讯云 `AttachUserPolicy` API 完成云上权限绑定
- 更新流程的 `checkPermissionTemplate` 校验逻辑
- 工单渲染中查询并展示权限模版名称

## What Changes

### 创建流程变更
- 新增 TCloud CAM `AttachUserPolicy` adaptor 方法（`pkg/adaptor/tcloud/cam_policy.go`），封装腾讯云 CAM 策略绑定 API，支持限流失败重试
- 新增 `TCloudAttachUserPolicyOption` 类型定义（`pkg/adaptor/types/account/`）
- 新增 hc-service 层批量绑定权限策略接口（`POST /vendors/tcloud/sub_accounts/attach_user_policies`），供 cloud-server 调用
- 新增 hc-service client 方法（`pkg/client/hc-service/tcloud/account.go`）
- 新增 cloud-server `attachPermissionToCloud` 方法（`cmd/cloud-server/service/application/handlers/sub-account/create-sub-account/deliver.go`），在三级账号创建成功后调用
- 新增创建流程工单渲染中权限模版名称展示

### 更新流程变更
- 新增 `SubAccountUpdateReq.PermissionTemplateIDs` 字段，支持更新权限模版
- 新增更新流程 `checkPermissionTemplate` 校验方法，校验规则与创建流程一致
- 新增更新流程工单渲染中权限模版名称展示
- 新增更新流程 `updatePermissionTemplateOnCloud` 方法，实现权限模版差异更新

## Capabilities

### New Capabilities
- `tcloud-cam-attach-user-policy`: TCloud CAM AttachUserPolicy API adaptor 封装，支持批量绑定策略到子用户，包含限流重试机制

### Modified Capabilities
- `sub-account-operation`: 扩展三级账号创建（add）和更新（update）流程：
  - 创建流程：在 `Deliver` 阶段增加权限模版绑定步骤，工单渲染展示权限模版名称
  - 更新流程：支持 `permission_template_ids` 字段，增加校验和差异更新逻辑，工单渲染展示权限模版名称

## Impact

- 影响文件：
  - `pkg/adaptor/types/account/tcloud.go` - 新增 `TCloudAttachUserPolicyOption` 类型
  - `pkg/adaptor/tcloud/cam_policy.go` - 新增 `AttachUserPolicy` 方法
  - `pkg/adaptor/tcloud/interface.go` - 新增接口方法声明
  - `pkg/api/hc-service/sub-account/` - 新增批量绑定请求/响应结构体
  - `cmd/hc-service/service/sub-account/` - 新增批量绑定 handler
  - `pkg/client/hc-service/tcloud/account.go` - 新增 client 方法
  - `pkg/api/cloud-server/application/update_sub_account.go` - 新增 `PermissionTemplateIDs` 字段
  - `cmd/cloud-server/service/application/handlers/sub-account/create-sub-account/deliver.go` - 实现 `attachPermissionToCloud`
  - `cmd/cloud-server/service/application/handlers/sub-account/create-sub-account/create_itsm_ticket.go` - 渲染权限模版名称
  - `cmd/cloud-server/service/application/handlers/sub-account/update-sub-account/check.go` - 新增 `checkPermissionTemplate` 方法
  - `cmd/cloud-server/service/application/handlers/sub-account/update-sub-account/deliver.go` - 实现 `updatePermissionTemplateOnCloud` 方法
  - `cmd/cloud-server/service/application/handlers/sub-account/update-sub-account/create_itsm_ticket.go` - 渲染权限模版名称
- API 变更：新增 hc-service 内部接口，SubAccountUpdateReq 结构体新增字段
- 依赖：腾讯云 CAM `AttachUserPolicy` API（https://cloud.tencent.com/document/product/598/34579）
