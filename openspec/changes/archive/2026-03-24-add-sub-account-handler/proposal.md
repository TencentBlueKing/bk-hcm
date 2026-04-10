## Why

HCM 平台已有二级账号(account)的创建审批流，但缺少三级账号(sub_account)的创建能力。业务侧需要通过审批流在云上创建子用户(如腾讯云 CAM 子用户)，获取密钥后通知申请人，并将账号信息持久化到本地 sub_account 表。当前 `CreateBizForAddSubAccount` 路由已注册但实现为空(stub)，需要完整实现整条链路。

## What Changes

- 新增一个应用类型枚举 `OperateSubAccount = "operate_sub_account"`，统一覆盖子账号新增/更新/删除操作，具体操作类型通过 application content 中的 `action` 字段区分，后续扩展无需新增审批流或枚举值
- 新增 `SubAccountAddReq` API 请求/响应结构体，定义三级账号创建入参（支持批量提交，每批上限100）
- 新增 `handlers/sub_account/` 包，实现 `ApplicationHandler` 接口：参数校验、ITSM 表单渲染、审批通过后的资源交付
- 在 TCloud adaptor 层新增 `AddUser` 方法，封装腾讯云 CAM `AddUser` API 调用
- 在 hc-service 层新增创建子用户 API，供 cloud-server handler 在交付阶段调用
- 在 hc-service client 层新增 `CreateSubAccount` 方法
- 交付阶段调用 hc-service 创建云上子用户，成功后通过 data-service 写入 sub_account 表，并通过 CMSI 发送密钥邮件到 `receive_email`
- 完善 `CreateBizForAddSubAccount` 入口函数，串联认证鉴权、handler 创建和审批流
- 在 approve.go 中补充 `OperateSubAccount` 类型的 handler 分发逻辑，根据 content 中的 `action` 字段分发到对应 handler（后续 update/delete 在同一 case 分支内扩展）

## Capabilities

### New Capabilities
- `sub-account-operation`: 三级账号操作审批流（本期实现创建），统一使用 `operate_sub_account` 审批类型。包括 cloud-server handler、hc-service 云上创建、data-service 本地持久化、密钥邮件发送。审批流设计兼容后续更新/删除操作扩展

### Modified Capabilities

## Impact

- **API**: 新增 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/add_sub_account`
- **枚举**: `pkg/criteria/enumor/application.go` 新增 `OperateSubAccount` 一个类型
- **审批流**: `approval_process` 表新增一条 `operate_sub_account` 记录
- **cloud-server**: `cmd/cloud-server/service/application/` 下 create.go、approve.go 需要修改
- **handlers**: 新增 `cmd/cloud-server/service/application/handlers/sub_account/` 包
- **hc-service**: 新增子用户创建 API 路由和实现（`cmd/hc-service/service/account/`）
- **adaptor**: `pkg/adaptor/tcloud/account.go` 新增 `AddUser` 方法
- **client**: `pkg/client/hc-service/tcloud/account.go` 新增 `CreateSubAccount` 方法
- **API 定义**: `pkg/api/cloud-server/application/` 新增 `add_sub_account.go`；`pkg/api/hc-service/` 新增子用户创建请求/响应类型
- **依赖**: 腾讯云 SDK `cam.AddUser` API；蓝鲸 CMSI 邮件服务；ITSM 审批服务
