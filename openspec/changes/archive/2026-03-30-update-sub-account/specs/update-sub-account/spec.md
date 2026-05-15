## ADDED Requirements

### Requirement: SubAccountUpdateReq 请求结构体
系统 SHALL 定义 `SubAccountUpdateReq` 结构体，包含以下字段：
- `action`: `enumor.SubAccountAction`，必填，操作类型（固定为 `"update"`）
- `vendor`: `enumor.Vendor`，必填
- `bk_biz_id`: int64，必填
- `sub_accounts`: `[]SubAccountUpdateItem` 数组，必填，长度 1~100

`SubAccountUpdateItem` 包含（可选字段使用指针类型，nil 表示不修改、非 nil 表示修改为该值，防止未传字段被零值覆盖）：
- `id`: string，必填，本地 sub_account 表 ID
- `name`: *string，可选，修改 HCM 平台显示名称（仅本地）
- `email`: *string，可选，三级账号邮箱（同步云上 + 本地）
- `phone_num`: *string，可选，手机号（同步云上 + 本地）
- `country_code`: *string，可选，手机区域代码（同步云上 + 本地）
- `managers`: []string，可选，账号管理者（仅本地），nil 不修改，空数组清空
- `memo`: *string，可选，备注（仅本地）

nil 语义 SHALL 在整条链路中严格传递：构建 hc-service 请求时只设置非 nil 字段，构建 adaptor UpdateUserOption 时只设置非 nil/非空字段，构建 data-service UpdateField 时只设置非 nil 字段。

#### Scenario: 正常校验通过
- **WHEN** 提交包含 1 个 sub_account 的请求，vendor 为 tcloud，id 填写，至少有一个可选字段为非 nil
- **THEN** `Validate()` 返回 nil

#### Scenario: sub_accounts 为空
- **WHEN** 提交 sub_accounts 为空数组
- **THEN** `Validate()` 返回参数错误

#### Scenario: sub_accounts 超过 100 条
- **WHEN** 提交 sub_accounts 长度为 101
- **THEN** `Validate()` 返回参数错误

#### Scenario: id 为空
- **WHEN** 提交 sub_accounts 中某项 id 为空字符串
- **THEN** `Validate()` 返回参数错误

#### Scenario: 未传的可选字段不修改原值
- **WHEN** 提交请求仅包含 id 和 email（非 nil），其余可选字段均未传（nil）
- **THEN** `Validate()` 返回 nil，Deliver 阶段仅更新 email 字段，其余字段保持原值不变

### Requirement: CreateBizForUpdateSubAccount 入口
系统 SHALL 实现 `CreateBizForUpdateSubAccount` 方法，处理 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/update_sub_account` 请求。

系统 SHALL 校验 bk_biz_id 路径参数有效性并进行三级账号操作权限鉴权。

系统 SHALL 对 `sub_accounts` 数组中的每个元素独立创建一张 ITSM 审批单和 application 记录，返回所有 application ID 数组。

#### Scenario: 正常创建
- **WHEN** 提交包含 2 个 sub_account 的合法请求
- **THEN** 创建 2 张 ITSM 审批单和 2 条 application 记录，返回 2 个 ID

#### Scenario: 无业务权限
- **WHEN** 当前用户对 bk_biz_id 无三级账号操作权限
- **THEN** 返回权限拒绝错误

### Requirement: Handler 参数校验
`ApplicationOfUpdateSubAccount.CheckReq()` SHALL：
1. 调用 `SubAccountUpdateItem.Validate()` 校验字段合法性
2. 通过 data-service `SubAccount.List` 按 ID 查询，验证三级账号在 sub_account 表中存在
3. 从查询结果中获取 `account_id`（所属二级账号 ID），设置到 handler 的 accountID 字段
4. 通过 `GetAccount(accountID)` 验证所属二级账号存在

#### Scenario: 三级账号 ID 不存在
- **WHEN** 传入一个在 sub_account 表中不存在的 ID
- **THEN** `CheckReq()` 返回 sub account not found 错误

#### Scenario: 所属二级账号不存在
- **WHEN** 三级账号存在但其所属二级账号已被删除
- **THEN** `CheckReq()` 返回 account not found 错误

#### Scenario: 校验通过
- **WHEN** 传入存在的三级账号 ID，且所属二级账号有效
- **THEN** `CheckReq()` 返回 nil，handler 的 accountID 字段已设置

### Requirement: ITSM 单据渲染
系统 SHALL 实现 `RenderItsmTitle()` 返回格式为 `"申请修改[{云厂商中文名}]三级账号({name})"`。

系统 SHALL 实现 `RenderItsmForm()` 返回包含三级账号名称、修改字段明细（邮箱、手机号、备注等）的文本表单。

#### Scenario: 渲染标题
- **WHEN** vendor 为 tcloud，三级账号 name 为 "test-user"
- **THEN** 返回 "申请修改[腾讯云]三级账号(test-user)"

#### Scenario: 渲染表单含修改字段
- **WHEN** 请求包含 email、phone_num、memo 字段
- **THEN** 表单文本中包含这些字段的 label 和修改后的 value

### Requirement: TCloud UpdateUser adaptor
系统 SHALL 在 `pkg/adaptor/tcloud/account.go` 新增 `UpdateUser(kt *kit.Kit, opt *UpdateUserOption) error` 方法。

该方法 SHALL 调用腾讯云 CAM `UpdateUser` API（reference: https://cloud.tencent.com/document/product/598/34583），通过 Name 标识目标用户，支持修改 Remark、ConsoleLogin、Password、NeedResetPassword、PhoneNum、CountryCode、Email 等参数。

`UpdateUserOption` 结构体 SHALL 包含 Name（必填）及上述可选修改字段。

该方法 SHALL 返回 error（UpdateUser API 无实质性返回数据）。

系统 SHALL 在 `pkg/adaptor/tcloud/interface.go` 的 TCloud 接口中新增 `UpdateUser` 方法签名。

#### Scenario: 正常更新子用户
- **WHEN** 传入有效的 Name 和合法修改参数
- **THEN** 调用腾讯云 UpdateUser API 成功，返回 nil

#### Scenario: 用户不存在
- **WHEN** 腾讯云返回 ResourceNotFound.UserNotExist 错误
- **THEN** 方法返回对应错误

### Requirement: hc-service 更新子用户 API
系统 SHALL 在 hc-service 注册 `POST /vendors/tcloud/sub_accounts/update` 路由。

请求结构体 `UpdateSubAccountReq` SHALL 包含：
- `account_id`: string，必填，用于获取凭证
- `name`: string，必填，云上子用户名，用于标识 UpdateUser API 的目标用户
- `remark`: *string，可选
- `email`: *string，可选
- `phone_num`: *string，可选
- `country_code`: *string，可选

可选字段使用指针类型，nil 表示不修改。hc-service 构建 adaptor `UpdateUserOption` 时 SHALL 只将非 nil 字段传递给腾讯云 API。

处理流程：通过 account_id 获取二级账号密钥 → 构建 TCloud adaptor → 调用 `UpdateUser`（仅传入非 nil 字段） → 返回结果。

#### Scenario: 正常更新
- **WHEN** 传入有效的 account_id、name 和修改字段
- **THEN** 调用腾讯云 UpdateUser 成功，返回 nil

#### Scenario: account_id 凭证无效
- **WHEN** account_id 对应的账号密钥已失效
- **THEN** 返回云 API 调用失败错误

### Requirement: hc-service client UpdateSubAccount
系统 SHALL 在 `pkg/client/hc-service/tcloud/account.go` 新增 `UpdateSubAccount` 方法，封装对 hc-service 更新子用户 API 的 HTTP 调用。

#### Scenario: 正常调用
- **WHEN** cloud-server handler 调用 `UpdateSubAccount` 传入合法参数
- **THEN** 向 hc-service 发送 POST 请求并返回结果

### Requirement: Deliver 交付逻辑
`ApplicationOfUpdateSubAccount.Deliver()` SHALL：
1. 按 vendor switch-case 分发（本期仅 tcloud，其他返回 not supported 错误）
2. 调用 hc-service `UpdateSubAccount` 更新云上子用户信息（email、phone_num、country_code 等云上支持的字段）
3. 调用 data-service `SubAccount.BatchUpdate` 更新本地 sub_account 表（所有修改字段，包括 managers、memo 等仅本地存储的字段）
4. 云上更新失败 SHALL 直接返回 `DeliverError`，不更新本地
5. 云上更新成功但本地更新失败 SHALL 记录错误日志并返回 `DeliverError`
6. 返回 `Completed` 状态和包含 sub_account ID 的 detail

#### Scenario: TCloud 正常交付
- **WHEN** vendor 为 tcloud，hc-service 和 data-service 调用均成功
- **THEN** 返回 `Completed` 状态

#### Scenario: 云上更新失败
- **WHEN** hc-service 调用 UpdateUser 返回错误
- **THEN** 返回 `DeliverError` 状态和错误信息，不更新本地

#### Scenario: 本地写入失败
- **WHEN** hc-service 成功但 data-service BatchUpdate 失败
- **THEN** 返回 `DeliverError` 状态，日志记录云上已更新但本地写入失败

#### Scenario: 不支持的 vendor
- **WHEN** vendor 为 aws（本期未实现）
- **THEN** 返回 `DeliverError` 和 not supported 错误
