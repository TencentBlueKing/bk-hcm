## Why

平台已支持通过 ITSM 审批流创建三级账号，但尚缺对应的删除能力。运维人员需要一个安全、可审计的方式来移除不再使用的三级账号，包括清理云上用户、本地 sub_account 记录以及 account 表中的登记账号条目。缺少此功能会导致过期账号累积，增加安全风险和管理负担。

## What Changes

- 在 `cmd/cloud-server/service/application/handlers/sub-account/delete-sub-account/` 下新增删除三级账号审批流 handler 包，遵循 create-sub-account 的代码模式。
- 实现审批前校验：验证三级账号的密钥是否已删除（因密钥管理功能尚未实现，先以 TODO 占位）。
- 实现审批通过后的交付逻辑（先实现腾讯云）：
  - 调用腾讯云 CAM `DeleteUser` API 删除云上子用户。
  - 通过 data-service `BatchDelete` 删除本地 sub_account 记录。
  - 通过三级账号的 `cloud_id` 在 account 表中匹配并删除对应的登记账号记录。
- 在 hc-service 新增 API 端点（`POST /vendors/tcloud/sub_accounts/delete`）及对应的 TCloud adaptor 方法 `DeleteUser`。
- 在 hc-service TCloud account client 中新增 `DeleteSubAccount` 方法。
- 在 cloud-server web-service 层注册 `delete_sub_account` API 路由。
- 通过 `init()` 自注册模式注册 `SubAccountActionDelete` handler。
- 抽取创建和删除流程中可复用的公共结构体。

## Capabilities

### New Capabilities
- `delete-sub-account`：端到端的三级账号删除工作流，包括请求校验、ITSM 审批单创建、审批通过后的资源清理（云上 + 本地 DB）。

### Modified Capabilities

## Impact

- **代码**：新增包 `cmd/cloud-server/service/application/handlers/sub-account/delete-sub-account/`（5 个文件，遵循 create 模式）；在 `pkg/adaptor/tcloud/account.go` 新增 adaptor 方法；新增 hc-service 端点和 client 方法；新增路由注册。
- **API**：新增 web-server 端点 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/delete_sub_account`；新增 hc-service 端点 `POST /vendors/tcloud/sub_accounts/delete`。
- **依赖**：腾讯云 CAM SDK（`cam.NewDeleteUserRequest` / `DeleteUserWithContext`）。
- **扩展性**：handler 架构使用 vendor-switch 分发，后续添加 AWS/Azure/GCP/HuaWei 删除支持仅需在 `Deliver()` 的 switch 中添加新 case 及对应的 adaptor/client 方法。
