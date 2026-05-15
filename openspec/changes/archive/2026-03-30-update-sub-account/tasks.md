## 1. API 请求/响应结构体

- [x] 1.1 新建 `pkg/api/cloud-server/application/update_sub_account.go`，定义 `SubAccountUpdateReq` 和 `SubAccountUpdateItem` 结构体（可选字段使用 `*string` 指针类型），含 Validate 方法
- [x] 1.2 新建 `pkg/api/hc-service/sub-account/update.go`，定义 hc-service 更新子用户的请求 `UpdateSubAccountReq` 结构体（可选字段使用 `*string` 指针类型），含 Validate 方法

## 2. TCloud Adaptor 层

- [x] 2.1 在 `pkg/adaptor/types/account/tcloud.go` 新增 `UpdateUserOption` 结构体（Name 必填，其余可选字段使用 `*string` 指针类型），含 Validate 方法
- [x] 2.2 在 `pkg/adaptor/tcloud/account.go` 新增 `UpdateUser(kt *kit.Kit, opt *UpdateUserOption) error` 方法，封装 `cam.NewUpdateUserRequest()` 调用，仅将非 nil 字段设置到请求参数
- [x] 2.3 在 `pkg/adaptor/tcloud/interface.go` 的 TCloud 接口中新增 `UpdateUser` 方法签名

## 3. hc-service 更新子用户 API

- [x] 3.1 在 `cmd/hc-service/service/account/service.go` 注册路由 `POST /vendors/tcloud/sub_accounts/update`
- [x] 3.2 在 `cmd/hc-service/service/account/sub_account.go` 实现 `TCloudUpdateSubAccount` handler：根据 account_id 获取凭证 → 构建 TCloud adaptor → 调用 UpdateUser（仅传入非 nil 字段）→ 返回结果
- [x] 3.3 在 `pkg/client/hc-service/tcloud/account.go` 新增 `UpdateSubAccount` client 方法

## 4. cloud-server Handler 实现

- [x] 4.1 新建 `cmd/cloud-server/service/application/handlers/sub-account/update-sub-account/init.go`，定义 `ApplicationOfUpdateSubAccount` 结构体、构造函数和 `init()` 注册 `SubAccountActionUpdate`
- [x] 4.2 新建 `handlers/sub-account/update-sub-account/check.go`，实现 `CheckReq()`：校验参数、通过 data-service 验证三级账号 ID 存在、获取所属 account_id、验证二级账号存在
- [x] 4.3 新建 `handlers/sub-account/update-sub-account/prepare.go`，实现 `GenerateApplicationContent()` 序列化审批内容（含 `updateSubAccountContent` 结构体）
- [x] 4.4 新建 `handlers/sub-account/update-sub-account/create_itsm_ticket.go`，实现 `RenderItsmTitle()` 和 `RenderItsmForm()`
- [x] 4.5 新建 `handlers/sub-account/update-sub-account/deliver.go`，实现 `Deliver()`：vendor switch → hc-service UpdateSubAccount（仅传非 nil 字段）→ data-service BatchUpdate（仅传非 nil 字段）

## 5. cloud-server 入口串联

- [x] 5.1 在 `cmd/cloud-server/service/application/create.go` 实现 `CreateBizForUpdateSubAccount`：鉴权 → 解析请求 → 遍历 sub_accounts 逐个创建 handler 并调用 `create()`

## 6. 验证与清理

- [x] 6.1 确认编译通过，无 lint 错误
