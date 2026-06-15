## 1. TCloud Adaptor 和 hc-service 层

- [x] 1.1 在 `pkg/adaptor/tcloud/account.go` 新增 `DeleteUser` 方法，封装 CAM `DeleteUser` API 调用
- [x] 1.2 在 `pkg/api/hc-service/sub-account/` 新增 `delete.go`，定义 `DeleteSubAccountReq` 请求结构体
- [x] 1.3 在 `cmd/hc-service/service/account/sub_account.go` 新增 `TCloudDeleteSubAccount` handler 方法
- [x] 1.4 在 `cmd/hc-service/service/account/service.go` 注册 `TCloudDeleteSubAccount` 路由
- [x] 1.5 在 `pkg/client/hc-service/tcloud/account.go` 新增 `DeleteSubAccount` client 方法

## 2. delete-sub-account handler 包核心文件

- [x] 2.1 创建 `cmd/cloud-server/service/application/handlers/sub-account/delete-sub-account/init.go`：定义 `ApplicationOfDeleteSubAccount` 结构体，`init()` 注册 `SubAccountActionDelete` handler，实现 `newHandlerFromContent` 工厂方法
- [x] 2.2 创建 `cmd/cloud-server/service/application/handlers/sub-account/delete-sub-account/check.go`：实现 `CheckReq` 方法，校验三级账号存在性、二级账号有效性，以及 TODO 占位密钥校验
- [x] 2.3 创建 `cmd/cloud-server/service/application/handlers/sub-account/delete-sub-account/prepare.go`：定义 `deleteSubAccountContent` 内容结构体，实现 `GenerateApplicationContent` 和 `GetItsmApprover`
- [x] 2.4 创建 `cmd/cloud-server/service/application/handlers/sub-account/delete-sub-account/create_itsm_ticket.go`：实现 `RenderItsmTitle` 和 `RenderItsmForm`
- [x] 2.5 创建 `cmd/cloud-server/service/application/handlers/sub-account/delete-sub-account/deliver.go`：实现 `Deliver` 方法，按顺序执行云上删除 → sub_account 表删除 → account 表登记记录删除

## 3. cloud-server 路由和入口

- [x] 3.1 在 `pkg/api/cloud-server/application/` 新增删除三级账号请求结构体 `SubAccountDeleteReq`（包含 IDs 列表）
- [x] 3.2 在 `cmd/cloud-server/service/application/create.go` 新增 `CreateBizForDeleteSubAccount` 方法（按二级账号分组创建审批单）
- [x] 3.3 在 `cmd/cloud-server/service/application/init.go` 的 `bizService` 中注册 `delete_sub_account` 路由

## 4. 确保 init 引入和编译验证

- [x] 4.1 确认 `delete-sub-account` 包被正确引入（通过 create.go import），保证 `init()` 被执行
- [x] 4.2 编译验证，确保所有接口方法已实现且无编译错误
