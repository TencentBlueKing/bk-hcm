## ADDED Requirements

### Requirement: ApplicationType 枚举扩展
系统 SHALL 在 `enumor.ApplicationType` 中新增 `OperateSubAccount`("operate_sub_account") 常量，并在 `Validate()` 中支持校验。

该类型统一覆盖子账号的新增/更新/删除操作，具体操作类型通过 application content 中的 `action` 字段（如 `"add"`、`"update"`、`"delete"`）区分。`approval_process` 表中 `"operate_sub_account"` 记录直接匹配 `string(OperateSubAccount)`，无需额外映射方法，`getApprovalProcessInfo` 现有逻辑无需修改。

#### Scenario: Validate 通过新类型
- **WHEN** 传入 `ApplicationType("operate_sub_account")`
- **THEN** `Validate()` 返回 nil

### Requirement: SubAccountAddReq 请求结构体
系统 SHALL 定义 `SubAccountAddReq` 结构体，包含以下字段：
- `action`: string，必填，操作类型（本期固定为 `"add"`，用于 content 中区分操作）
- `vendor`: `enumor.Vendor`，必填
- `sub_accounts`: `[]SubAccountItem` 数组，必填，长度 1~100

`SubAccountItem` 包含：
- `account_id`: string，必填，所属二级账号 ID
- `name`: string，必填，三级账号名称
- `receive_email`: string，必填，账号开通接收邮箱
- `email`: string，可选，三级账号邮箱
- `phone_num`: string，可选，手机号
- `country_code`: string，可选，手机区域代码
- `managers`: []string，可选，账号管理者
- `memo`: string，可选，备注

#### Scenario: 正常校验通过
- **WHEN** 提交包含 1 个 sub_account 的请求，vendor 为 tcloud，account_id/name/receive_email 均填写
- **THEN** `Validate()` 返回 nil

#### Scenario: sub_accounts 为空
- **WHEN** 提交 sub_accounts 为空数组
- **THEN** `Validate()` 返回参数错误

#### Scenario: sub_accounts 超过 100 条
- **WHEN** 提交 sub_accounts 长度为 101
- **THEN** `Validate()` 返回参数错误

### Requirement: CreateBizForAddSubAccount 入口
系统 SHALL 实现 `CreateBizForAddSubAccount` 方法，处理 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/add_sub_account` 请求。

系统 SHALL 校验 bk_biz_id 路径参数有效性并进行业务访问权限鉴权。

系统 SHALL 对 `sub_accounts` 数组中的每个元素独立创建一张 ITSM 审批单和 application 记录，返回所有 application ID 数组。

#### Scenario: 正常创建
- **WHEN** 提交包含 2 个 sub_account 的合法请求
- **THEN** 创建 2 张 ITSM 审批单和 2 条 application 记录，返回 2 个 ID

#### Scenario: 无业务权限
- **WHEN** 当前用户对 bk_biz_id 无业务访问权限
- **THEN** 返回权限拒绝错误

### Requirement: Handler 参数校验
`ApplicationOfAddSubAccount.CheckReq()` SHALL：
1. 调用 `SubAccountItem.Validate()` 校验字段合法性
2. 验证 `account_id` 对应的二级账号存在
3. 验证三级账号 `name` 在同一 `account_id` 下不重复（查询 data-service sub_account 表）

#### Scenario: account_id 不存在
- **WHEN** 传入一个不存在的 account_id
- **THEN** `CheckReq()` 返回 account not found 错误

#### Scenario: name 重复
- **WHEN** 传入的 name 在该 account_id 下已存在
- **THEN** `CheckReq()` 返回名称重复错误

### Requirement: ITSM 单据渲染
系统 SHALL 实现 `RenderItsmTitle()` 返回格式为 `"申请新增[{云厂商中文名}]三级账号({name})"`。

系统 SHALL 实现 `RenderItsmForm()` 返回包含二级账号名称、三级账号名称、接收邮箱、手机号、备注等信息的文本表单。

#### Scenario: 渲染标题
- **WHEN** vendor 为 tcloud，name 为 "test-user"
- **THEN** 返回 "申请新增[腾讯云]三级账号(test-user)"

#### Scenario: 渲染表单含完整信息
- **WHEN** 请求包含 email、phone_num、memo 字段
- **THEN** 表单文本中包含这些字段的 label 和 value

### Requirement: TCloud AddUser adaptor
系统 SHALL 在 `pkg/adaptor/tcloud/account.go` 新增 `AddUser(kt *kit.Kit, opt *AddUserOption) (*AddUserResult, error)` 方法。

该方法 SHALL 调用腾讯云 CAM `AddUser` API（reference: https://cloud.tencent.com/document/product/598/34595），传入 Name、UseApi=1（生成密钥）、及可选的 Remark、ConsoleLogin、Password、PhoneNum、CountryCode、Email 等参数。

返回 `AddUserResult` SHALL 包含 Uin、Name、SecretId、SecretKey、Password（如自动生成）、Uid。

#### Scenario: 正常创建子用户
- **WHEN** 传入有效的 Name 和合法参数
- **THEN** 返回包含 Uin、SecretId、SecretKey 的结果

#### Scenario: 用户名已存在
- **WHEN** 腾讯云返回 InvalidParameter.SubUserNameInUse 错误
- **THEN** 方法返回对应错误

### Requirement: hc-service 创建子用户 API
系统 SHALL 在 hc-service 注册 `POST /vendors/tcloud/sub_accounts/create` 路由。

请求参数 SHALL 包含 account_id（用于获取凭证）和子用户创建所需字段。

处理流程：通过 account_id 获取二级账号密钥 → 构建 TCloud adaptor → 调用 `AddUser` → 返回云上创建结果。

#### Scenario: 正常创建
- **WHEN** 传入有效的 account_id 和用户信息
- **THEN** 调用腾讯云 AddUser 成功，返回子用户信息

#### Scenario: account_id 凭证无效
- **WHEN** account_id 对应的账号密钥已失效
- **THEN** 返回云 API 调用失败错误

### Requirement: hc-service client CreateSubAccount
系统 SHALL 在 `pkg/client/hc-service/tcloud/account.go` 新增 `CreateSubAccount` 方法，封装对 hc-service 创建子用户 API 的 HTTP 调用。

#### Scenario: 正常调用
- **WHEN** cloud-server handler 调用 `CreateSubAccount` 传入合法参数
- **THEN** 向 hc-service 发送 POST 请求并返回结果

### Requirement: Deliver 交付逻辑
`ApplicationOfAddSubAccount.Deliver()` SHALL：
1. 按 vendor switch-case 分发（本期仅 tcloud，其他返回 not supported 错误）
2. 调用 hc-service `CreateSubAccount` 创建云上子用户
3. 调用 data-service `SubAccount.BatchCreate` 写入本地 sub_account 表
4. 调用 `SendMail` 将 SecretId 和 SecretKey 发送到 `receive_email`
5. 邮件发送失败 SHALL 记录错误日志但不影响 deliver 状态
6. 返回 `Completed` 状态和包含 sub_account ID 的 detail

#### Scenario: TCloud 正常交付
- **WHEN** vendor 为 tcloud，hc-service 和 data-service 调用均成功
- **THEN** 返回 `Completed` 状态，密钥邮件发送到 receive_email

#### Scenario: 云上创建失败
- **WHEN** hc-service 调用 AddUser 返回错误
- **THEN** 返回 `DeliverError` 状态和错误信息

#### Scenario: 本地写入失败
- **WHEN** hc-service 成功但 data-service BatchCreate 失败
- **THEN** 返回 `DeliverError` 状态，日志记录云上已创建但本地写入失败

#### Scenario: 邮件发送失败
- **WHEN** 云上创建和本地写入均成功，但 SendMail 失败
- **THEN** deliver 仍返回 `Completed` 状态，错误记录到日志

#### Scenario: 不支持的 vendor
- **WHEN** vendor 为 aws（本期未实现）
- **THEN** 返回 `DeliverError` 和 not supported 错误

### Requirement: approve.go handler 分发
系统 SHALL 在 `getHandlerByApplication` 中新增 `OperateSubAccount` 类型的 case，解析 content 中的 `action` 字段，根据 action 值构造对应的 handler（本期仅实现 `action = "add"`）。

#### Scenario: 审批回调分发（新增子账号）
- **WHEN** ITSM 审批通过，application.Type 为 "operate_sub_account"，content.action 为 "add"
- **THEN** 构建 `ApplicationOfAddSubAccount` handler 并执行交付

#### Scenario: 不支持的 action
- **WHEN** application.Type 为 "operate_sub_account"，content.action 为 "delete"（本期未实现）
- **THEN** 返回 not supported 错误
