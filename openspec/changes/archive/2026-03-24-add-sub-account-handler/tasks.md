## 1. 枚举与类型定义

- [x] 1.1 在 `pkg/criteria/enumor/application.go` 新增 `OperateSubAccount ApplicationType = "operate_sub_account"` 常量，更新 `Validate()` switch-case

## 2. API 请求/响应结构体

- [x] 2.1 新建 `pkg/api/cloud-server/application/add_sub_account.go`，定义 `SubAccountAddReq` 和 `SubAccountItem` 结构体，含 Validate 方法
- [x] 2.2 新建 `pkg/api/hc-service/sub-account/create.go`，定义 hc-service 创建子用户的请求 `CreateSubAccountReq` 和响应 `CreateSubAccountResp` 结构体

## 3. TCloud Adaptor 层

- [x] 3.1 在 `pkg/adaptor/types/account/tcloud.go` 新增 `AddUserOption` 和 `AddUserResult` 结构体
- [x] 3.2 在 `pkg/adaptor/tcloud/account.go` 新增 `AddUser(kt *kit.Kit, opt *AddUserOption) (*AddUserResult, error)` 方法，封装 `cam.NewAddUserRequest()` 调用

## 4. hc-service 创建子用户 API

- [x] 4.1 在 `cmd/hc-service/service/account/service.go` 注册路由 `POST /vendors/tcloud/sub_accounts/create`
- [x] 4.2 实现 handler：根据 account_id 获取凭证 → 构建 TCloud adaptor → 调用 AddUser → 返回结果
- [x] 4.3 在 `pkg/client/hc-service/tcloud/account.go` 新增 `CreateSubAccount` client 方法

## 5. cloud-server Handler 实现

- [x] 5.1 新建 `cmd/cloud-server/service/application/handlers/sub_account/init.go`，定义 `ApplicationOfAddSubAccount` 结构体和构造函数
- [x] 5.2 新建 `handlers/sub_account/check.go`，实现 `CheckReq()`：校验参数、验证 account_id 存在、验证 name 不重复
- [x] 5.3 新建 `handlers/sub_account/prepare.go`，实现 `PrepareReq()`/`PrepareReqFromContent()`/`GenerateApplicationContent()`/`GetItsmApprover()`/`GetBkBizIDs()`
- [x] 5.4 新建 `handlers/sub_account/create_itsm_ticket.go`，实现 `RenderItsmTitle()` 和 `RenderItsmForm()`
- [x] 5.5 新建 `handlers/sub_account/deliver.go`，实现 `Deliver()`：vendor switch → hc-service CreateSubAccount → data-service BatchCreate → SendMail

## 6. cloud-server 入口与审批串联

- [x] 6.1 在 `cmd/cloud-server/service/application/create.go` 实现 `CreateBizForAddSubAccount`：鉴权 → 解析请求 → 遍历 sub_accounts 逐个创建 handler 并调用 `create()`
- [x] 6.2 在 `cmd/cloud-server/service/application/approve.go` 的 `getHandlerByApplication` 中新增 `OperateSubAccount` case，根据 content 中的 `action` 字段分发到对应 Handler
- [x] 6.3 在 `createApplication` 中将 `OperateSubAccount` 加入需要记录 bkBizIDs 的类型判断

## 7. 验证与清理

- [x] 7.1 确认编译通过，无 lint 错误
- [x] 7.2 确认 `approval_process` 表需新增 `operate_sub_account` 类型记录（文档或 SQL 备注）

> **部署备注**: 需在 `approval_process` 表中插入一条记录：
> ```sql
> INSERT INTO approval_process (application_type, service_id, managers)
> VALUES ('operate_sub_account', <itsm_service_id>, '<manager_list>');
> ```
> 其中 `service_id` 和 `managers` 与具体 ITSM 审批流配置对齐。
