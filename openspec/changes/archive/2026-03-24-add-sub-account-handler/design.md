## Context

HCM 已有成熟的 Application → ITSM → Deliver 审批交付框架。二级账号(account)创建已实现完整链路（`handlers/account/`），三级账号(sub_account)目前仅有路由注册和 stub 函数。

现有架构分层：
- **cloud-server**：入口路由 + ApplicationHandler 抽象（check → prepare → ITSM ticket → deliver）
- **hc-service**：对接公有云 API（通过 adaptor 层封装 SDK 调用）
- **data-service**：本地 DB CRUD（已有 sub_account 表和 BatchCreate 等 client）
- **adaptor**：腾讯云 CAM SDK 封装（已有 ListAccount、GetAccountInfoBySecret，缺 AddUser）

腾讯云创建子用户 API：`cam:AddUser`，返回 Uin、SecretId、SecretKey 等密钥信息，需发送到用户指定的 `receive_email`。

## Goals / Non-Goals

**Goals:**
- 实现三级账号创建审批全链路（cloud-server handler → hc-service → adaptor → data-service）
- 设计统一审批流 `operate_sub_account`，兼容后续 update/delete 扩展
- 先实现腾讯云(TCloud)，架构上预留多云扩展能力
- 审批通过后创建云上子用户 → 写入本地 DB → 发送密钥邮件

**Non-Goals:**
- 本期不实现 update/delete 三级账号操作（后续在 `OperateSubAccount` 下通过 `action` 字段扩展）
- 不实现 AWS/HuaWei/GCP/Azure 子用户创建（仅预留 vendor switch 分支）
- 不实现权限模板关联（属于后续迭代）
- 不涉及三级账号同步逻辑调整

## Decisions

### 1. 单一 ApplicationType + content 区分操作

定义一个统一的 `ApplicationType`：
- `OperateSubAccount = "operate_sub_account"`

`approval_process` 表一条 `operate_sub_account` 记录即可，`string(OperateSubAccount)` 直接匹配，无需额外的映射方法，不改动 `getApprovalProcessInfo` 现有逻辑。

具体操作类型（新增/更新/删除）通过 application content 中的 `action` 字段区分。`getHandlerByApplication` 中只需一个 `case enumor.OperateSubAccount`，在内部根据 `action` 值构造对应的 Handler。本期仅实现 `action = "add"`，后续 update/delete 在同一 case 分支内扩展即可。

### 2. Handler 结构：按操作拆分文件

`handlers/sub_account/` 包结构：
```
handlers/sub_account/
├── init.go              # ApplicationOfAddSubAccount 结构体定义与构造
├── check.go             # CheckReq: 参数校验、名称去重、account_id 有效性
├── prepare.go           # PrepareReq / PrepareReqFromContent / GenerateApplicationContent
├── create_itsm_ticket.go # RenderItsmTitle / RenderItsmForm
└── deliver.go           # Deliver: 按 vendor 分发创建、写 DB、发邮件
```

遵循 `handlers/account/` 的文件组织和命名模式，后续 update/delete 可新增 handler struct 复用同一包。

### 3. hc-service 层新增 CreateSubAccount API

路由：`POST /vendors/tcloud/sub_accounts/create`

请求通过 hc-service → adaptor 调用腾讯云 `cam:AddUser`，返回云上子用户信息（Uin、Name、SecretId、SecretKey 等）。hc-service 不负责写入 data-service，由 cloud-server handler 的 Deliver 方法在收到 hc-service 响应后分别调用 data-service 写入和 CMSI 发邮件。

**理由**：hc-service 职责单一——只对接云 API；本地持久化和通知由 cloud-server 编排。与现有 account deliver 模式一致。

### 4. adaptor 层新增 AddUser 方法

在 `pkg/adaptor/tcloud/account.go` 中新增 `AddUser`，封装 `cam.NewAddUserRequest()` + `camClient.AddUserWithContext()`。

入参结构体 `AddUserOption` 包含 Name、Remark、ConsoleLogin、UseApi、Password、NeedResetPassword、PhoneNum、CountryCode、Email 等字段，与腾讯云 AddUser API 参数对齐。

返回结构体 `AddUserResult` 包含 Uin、Name、SecretId、SecretKey、Password、Uid。

### 5. 多云扩展性设计

- Deliver 方法中按 `vendor` switch-case 分发，目前仅实现 `tcloud` 分支，其他 vendor 返回 `not supported` 错误
- hc-service API 路径带 vendor 前缀（`/vendors/{vendor}/sub_accounts/create`），天然支持多云
- adaptor 接口定义在 vendor 各自的包中，不做抽象接口（与现有模式一致）
- 请求结构体中 vendor 特有字段通过 Extension（`json.RawMessage`）传递

### 6. 密钥邮件发送

在 Deliver 成功后，使用 `BaseApplicationHandler.SendMail()` 发送密钥到 `receive_email`（非用户自身邮箱，而是请求参数中指定的开通接收邮箱）。邮件内容包含 SecretId 和 SecretKey，不包含 Password（如自动生成密码则一并发送）。

### 7. 批量提交设计

API 支持批量提交（`sub_accounts` 数组，上限 100），但每个子账号独立创建一张 ITSM 审批单，独立交付。原因：
- 审批流程中每个子账号可能需要单独审批/驳回
- 云上创建是逐个调用，单个失败不应影响其他

返回值为批量创建的 application ID 数组。

## Risks / Trade-offs

- **[风险] 云上创建成功但本地写入失败** → 记录错误日志，deliver 状态为 `deliver_error`，运维可通过 sync 机制补齐数据
- **[风险] 邮件发送失败** → 不影响 deliver 状态（视为非关键路径），记录错误日志，管理员可手动查看密钥信息
- **[权衡] 批量每个独立审批单 vs 一张审批单含多个子账号** → 选择独立审批单，牺牲审批效率换取操作粒度和故障隔离
- **[权衡] hc-service 不写 DB** → cloud-server 多一次 RPC 调用，但职责清晰，与现有 account 模式一致
