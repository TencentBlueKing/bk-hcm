## 1. TCloud CAM AttachUserPolicy Adaptor

- [x] 1.1 在 `pkg/adaptor/types/account/tcloud.go` 新增 `TCloudAttachUserPolicyOption` 类型定义（`TargetUin uint64`、`PolicyId uint64`）及 `Validate` 方法
- [x] 1.2 在 `pkg/adaptor/tcloud/cam_policy.go` 实现 `AttachUserPolicy` 方法，封装 `cam.NewAttachUserPolicyRequest` 调用
- [x] 1.3 在 `pkg/adaptor/tcloud/interface.go` 的 `TCloud` 接口中新增 `AttachUserPolicy` 方法声明
- [x] 1.4 更新 `pkg/adaptor/mock/tcloud/tcloud_mock.go`，为 `AttachUserPolicy` 生成 mock 方法

## 2. hc-service 批量绑定权限策略接口

- [x] 2.1 在 `pkg/api/hc-service/sub-account/tcloud.go` 新增 `TCloudAttachUserPoliciesReq`（`AccountID string`、`TargetUin uint64`、`PolicyIds []uint64`）及 `Validate` 方法
- [x] 2.2 在 `pkg/api/hc-service/sub-account/tcloud.go` 新增 `TCloudAttachUserPoliciesResult`（`SuccessCount uint64`、`FailedPolicyIds []uint64`）
- [x] 2.3 在 `cmd/hc-service/service/account/secret.go` 新增 `TCloudAttachUserPolicies` handler，获取 TCloud adaptor 并循环调用 `AttachUserPolicy`
- [x] 2.4 在 `cmd/hc-service/service/account/service.go` 路由注册中新增 `POST /vendors/tcloud/sub_accounts/attach_user_policies`

## 3. hc-service Client 封装

- [x] 3.1 在 `pkg/client/hc-service/tcloud/account.go` 新增 `AttachUserPolicies(kt, req) (*TCloudAttachUserPoliciesResult, error)` client 方法，调用 hc-service 新接口

## 4. 创建流程实现

- [x] 4.1 在 `cmd/cloud-server/service/application/handlers/sub-account/create-sub-account/deliver.go` 实现 `attachPermissionToCloud` 方法：
  - 检查 `PermissionTemplateIDs` 是否为空
  - 查询权限模版获取 `cloud_id`（云上策略 ID）
  - 调用 hc-service `AttachUserPolicies` 批量绑定
  - 错误处理：记录日志但不阻塞流程
- [x] 4.2 在 `deliverForTCloud` 方法的 `saveLocalSubAccount` 成功后、`sendSubAccountMail` 前调用 `attachPermissionToCloud`
- [x] 4.3 在 `cmd/cloud-server/service/application/handlers/sub-account/create-sub-account/create_itsm_ticket.go` 的 `RenderItsmForm` 方法中增加权限模版名称渲染：
  - 查询权限模版获取名称列表
  - 渲染为 `绑定权限模版: 模版1,模版2,...` 格式

## 5. 更新流程实现

- [x] 5.1 在 `pkg/api/cloud-server/application/update_sub_account.go` 的 `SubAccountUpdateReq` 结构体中新增 `PermissionTemplateIDs` 字段（`[]string`，`omitempty`）
- [x] 5.2 在 `cmd/cloud-server/service/application/handlers/sub-account/update-sub-account/check.go` 实现 `checkPermissionTemplate` 方法：
  - 检查 `PermissionTemplateIDs` 是否为 nil
  - 查询权限模版并校验数量匹配
  - 校验每个模版的 `policy_library_id` 不为空
  - 校验每个模版的 `account_id` 与二级账号 ID 匹配
- [x] 5.3 在 `CheckReq` 方法中调用 `checkPermissionTemplate`
- [x] 5.4 在 `cmd/cloud-server/service/application/handlers/sub-account/update-sub-account/deliver.go` 实现 `updatePermissionTemplateOnCloud` 方法：
  - 检查 `PermissionTemplateIDs` 是否为 nil（跳过）或空数组（记录日志）
  - 查询权限模版获取 `cloud_id`（云上策略 ID）
  - 调用 hc-service `AttachUserPolicies` 批量绑定
  - 错误处理：记录日志但不阻塞流程
- [x] 5.5 在 `deliverForTCloud` 方法中调用 `updatePermissionTemplateOnCloud`
- [x] 5.6 在 `cmd/cloud-server/service/application/handlers/sub-account/update-sub-account/create_itsm_ticket.go` 的 `RenderItsmForm` 方法中增加权限模版名称渲染：
  - 处理 nil（不渲染）、空数组（清空）、有值（查询名称列表）三种情况
  - 渲染为 `修改权限模版: ...` 格式
