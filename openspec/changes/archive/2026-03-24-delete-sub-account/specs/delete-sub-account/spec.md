## ADDED Requirements

### Requirement: 删除三级账号审批流 handler 注册

系统必须（SHALL）在 `delete-sub-account` 包的 `init()` 中，通过 `RegisterActionHandler(enumor.SubAccountActionDelete, ...)` 将删除 handler 工厂注册到 `actionHandlerRegistry`，使审批回调能自动分发到删除 handler。

#### Scenario: handler 注册成功
- **WHEN** 程序启动并执行 `init()` 函数
- **THEN** `actionHandlerRegistry` 中包含 `SubAccountActionDelete` 对应的 handler 工厂

#### Scenario: 审批回调分发到删除 handler
- **WHEN** 审批通过后 `NewHandlerFromApplication` 解析到 `action = "delete"`
- **THEN** 系统调用已注册的删除 handler 工厂创建 `ApplicationOfDeleteSubAccount` 实例

### Requirement: 删除前校验三级账号存在性

系统必须（SHALL）在 `CheckReq` 中校验待删除的三级账号 ID 列表中的每个账号在本地 DB 中确实存在，且所属的二级账号（account_id）有效。

#### Scenario: 三级账号存在且二级账号有效
- **WHEN** 提交的 IDs 列表中所有三级账号在 sub_account 表中存在，且对应的 account_id 在 account 表中有效
- **THEN** 校验通过，流程继续

#### Scenario: 三级账号不存在
- **WHEN** 提交的 IDs 列表中某个三级账号在 sub_account 表中不存在
- **THEN** 返回错误，阻止审批流创建

### Requirement: 删除前校验密钥已清理（TODO 占位）

系统必须（SHALL）在 `CheckReq` 中预留密钥校验逻辑位置，以 TODO 注释标记。待密钥管理功能实现后，此处须校验三级账号关联的所有密钥是否已删除。

#### Scenario: 密钥校验占位
- **WHEN** 执行 `CheckReq` 校验流程
- **THEN** 代码中存在 TODO 注释标记的密钥校验占位逻辑，当前直接通过

### Requirement: ITSM 审批单标题和表单渲染

系统必须（SHALL）实现 `RenderItsmTitle` 和 `RenderItsmForm` 方法，生成删除三级账号的 ITSM 审批单标题和表单内容。

#### Scenario: 渲染审批单标题
- **WHEN** 调用 `RenderItsmTitle`
- **THEN** 返回格式为 `"申请删除[{云厂商中文名}]三级账号({账号名称})"` 的标题

#### Scenario: 渲染审批单表单
- **WHEN** 调用 `RenderItsmForm`
- **THEN** 表单内容包含云厂商、所属二级账号名称、待删除三级账号名称等关键信息

### Requirement: 生成申请单内容

系统必须（SHALL）实现 `GenerateApplicationContent` 方法，生成包含 `BaseSubAccountContent`（action/vendor/bk_biz_id）及删除特有字段（account_id、待删除三级账号 IDs 及名称列表）的内容结构体，序列化后存入 DB。

#### Scenario: 内容包含必要字段
- **WHEN** 调用 `GenerateApplicationContent`
- **THEN** 返回的内容结构体包含 action=delete、vendor、bk_biz_id、account_id、三级账号 ID 列表及名称列表

### Requirement: 审批通过后先删除云上账号

系统必须（SHALL）在 `Deliver` 方法中，首先调用 hc-service TCloud 端点删除云上 CAM 子用户。如果云上删除失败，必须立即返回 `DeliverError`，不继续本地清理。

#### Scenario: 腾讯云账号删除成功
- **WHEN** 审批通过，调用 hc-service `DeleteSubAccount` 接口
- **THEN** 云上 CAM 用户被删除，流程继续本地清理

#### Scenario: 腾讯云账号删除失败
- **WHEN** 调用 hc-service `DeleteSubAccount` 接口返回错误
- **THEN** 返回 `DeliverError` 状态，记录错误日志，不进行本地清理

#### Scenario: 不支持的云厂商
- **WHEN** vendor 不是 TCloud
- **THEN** 返回 `DeliverError` 及不支持的厂商错误信息

### Requirement: 删除本地 sub_account 记录

系统必须（SHALL）在云上删除成功后，通过 data-service `SubAccount.BatchDelete` 删除本地 sub_account 表中的记录。

#### Scenario: 本地 sub_account 删除成功
- **WHEN** 云上 CAM 用户已删除，调用 `BatchDelete` 传入三级账号 ID 过滤条件
- **THEN** sub_account 表中对应记录被删除

#### Scenario: 本地 sub_account 删除失败
- **WHEN** `BatchDelete` 返回错误
- **THEN** 返回 `DeliverError`，记录日志（包含 cloud_id 便于人工修复）

### Requirement: 删除 account 表中的登记账号记录

系统必须（SHALL）在 sub_account 删除成功后，通过三级账号的 `cloud_id` 在 account 表中查找 `extension.cloud_sub_account_id` 匹配的登记账号记录并删除。

#### Scenario: 找到并删除登记账号
- **WHEN** account 表中存在 `cloud_sub_account_id` 等于三级账号 `cloud_id` 的登记账号记录
- **THEN** 该登记账号记录被删除

#### Scenario: 未找到登记账号记录
- **WHEN** account 表中不存在匹配的登记账号记录
- **THEN** 记录警告日志，不阻塞整体删除流程，视为成功

### Requirement: hc-service TCloud 删除子用户端点

系统必须（SHALL）在 hc-service 中新增 `TCloudDeleteSubAccount` 端点，接收 sub-user 名称参数，调用 TCloud CAM SDK `DeleteUser` API 执行云上删除。

#### Scenario: 成功删除云上子用户
- **WHEN** hc-service 接收到合法的删除请求（包含 account_id 和子用户名称）
- **THEN** 调用 CAM `DeleteUserWithContext` 删除子用户，返回成功

#### Scenario: CAM API 调用失败
- **WHEN** CAM `DeleteUserWithContext` 返回错误
- **THEN** hc-service 返回错误信息

### Requirement: TCloud adaptor 新增 DeleteUser 方法

系统必须（SHALL）在 `pkg/adaptor/tcloud/account.go` 中新增 `DeleteUser` 方法，封装 CAM `DeleteUser` API 调用。

#### Scenario: 调用 DeleteUser 成功
- **WHEN** 传入有效的子用户名称
- **THEN** 调用 `cam.NewDeleteUserRequest` 设置 `Name`，执行 `DeleteUserWithContext` 成功返回

### Requirement: hc-service client 新增 DeleteSubAccount 方法

系统必须（SHALL）在 `pkg/client/hc-service/tcloud/account.go` 中新增 `DeleteSubAccount` 方法，向 hc-service 发送删除请求。

#### Scenario: client 调用成功
- **WHEN** cloud-server handler 通过 client 调用 `DeleteSubAccount`
- **THEN** 请求发送到 hc-service 对应端点，返回结果

### Requirement: web-server 注册删除三级账号 API 路由

系统必须（SHALL）在 cloud-server web-service 层注册 `POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/delete_sub_account` 路由。

#### Scenario: 路由注册并可用
- **WHEN** 用户发送符合格式的删除三级账号申请请求
- **THEN** 系统创建 `OperateSubAccount` 类型申请单，action 为 `SubAccountActionDelete`
