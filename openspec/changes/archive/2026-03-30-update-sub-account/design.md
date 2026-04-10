## Context

HCM 已有成熟的 Application → ITSM → Deliver 审批交付框架。三级账号(sub_account)创建和删除已实现完整链路（`handlers/sub-account/create-sub-account/` 和 `handlers/sub-account/delete-sub-account/`）。

现有架构分层：
- **cloud-server**：入口路由 + ApplicationHandler 抽象（check → prepare → ITSM ticket → deliver），子账号操作通过 `base.go` 中的 `RegisterActionHandler` 机制按 action 自注册
- **hc-service**：对接公有云 API（通过 adaptor 层封装 SDK 调用），已有 `TCloudCreateSubAccount` 和 `TCloudDeleteSubAccount`
- **data-service**：本地 DB CRUD（已有 sub_account 表、`BatchUpdate` 和 `UpdateField` 结构体）
- **adaptor**：腾讯云 CAM SDK 封装（已有 `AddUser`、`DeleteUser`，缺 `UpdateUser`）

腾讯云更新子用户 API：`cam:UpdateUser`（reference: https://cloud.tencent.com/document/product/598/34583），支持修改 Remark、ConsoleLogin、Password、NeedResetPassword、PhoneNum、CountryCode、Email 等字段。该 API 通过 Name 标识目标用户。

修改接口与创建/删除的核心差异：修改需要同时更新云上数据和本地 sub_account 表中的对应字段，且每个三级账号独立一张审批单。

## Goals / Non-Goals

**Goals:**
- 实现三级账号修改审批全链路（cloud-server handler → hc-service → adaptor → data-service）
- 复用现有 `OperateSubAccount` 审批类型和 `SubAccountActionUpdate` action 枚举，通过 `RegisterActionHandler` 自注册
- 先实现腾讯云(TCloud)，架构上预留多云扩展能力
- 审批通过后更新云上子用户信息 → 更新本地 sub_account 表
- CheckReq 阶段校验三级账号 ID 在 sub_account 表中存在，并获取所属二级账号信息

**Non-Goals:**
- 本期不实现 AWS/HuaWei/GCP/Azure 子用户修改（仅预留 vendor switch 分支）
- 不涉及三级账号密钥相关操作（密钥的增删启停属于独立 action）
- 不涉及三级账号同步逻辑调整
- 不修改 Name 字段（腾讯云 UpdateUser API 以 Name 作为用户标识，不支持修改 Name）

## Decisions

### 1. 复用 OperateSubAccount 审批类型 + SubAccountActionUpdate

与创建/删除保持一致，使用已有的 `OperateSubAccount` ApplicationType，content 中 `action = "update"`。`SubAccountActionUpdate` 枚举值已存在于 `enumor/sub_account_action.go` 中，无需新增。通过 `init()` 调用 `RegisterActionHandler(SubAccountActionUpdate, factory)` 自注册到 action handler registry。

### 2. 结构体复用分析

创建(`SubAccountItem`) 和更新(`SubAccountUpdateItem`) 的 API 请求结构体存在 `email`、`phone_num`、`country_code`、`managers`、`memo` 等同名字段，但因语义不同导致类型不同：创建使用值类型（必填/可选），更新使用指针类型（nil = 不修改）。抽取公共基类会迫使创建接口改为指针类型，破坏已有 API 契约，因此**不做字段级公共基类抽取**。

同理，hc-service 层（`CreateSubAccountReq` vs `UpdateSubAccountReq`）和 adaptor 层（`AddUserOption` vs `UpdateUserOption`）也存在值类型/指针类型差异，不适合抽取。

真正可复用的公共基类已存在于 handler 层：
- `BaseSubAccountContent`：审批单 content 的公共头部（Action、Vendor、BkBizID），update content 结构体直接嵌入
- `ApplicationBaseSubAccount`：handler 公共基类（含 BaseApplicationHandler、action、bkBizID、accountID），update handler 直接嵌入
- `RegisterActionHandler` 注册机制：update handler 通过 `init()` 自注册，无需修改基类

### 3. Handler 结构：update-sub-account 子包

```
handlers/sub-account/update-sub-account/
├── init.go               # ApplicationOfUpdateSubAccount 结构体定义、构造、init() 注册
├── check.go              # CheckReq: 校验参数、验证三级账号 ID 存在、获取二级账号信息
├── prepare.go            # GenerateApplicationContent / updateSubAccountContent
├── create_itsm_ticket.go # RenderItsmTitle / RenderItsmForm
└── deliver.go            # Deliver: 按 vendor 分发更新、写 DB
```

遵循 `create-sub-account/` 和 `delete-sub-account/` 的文件组织和命名模式。

### 4. SubAccountUpdateItem 请求结构体设计 — 指针语义区分"不修改"与"清空"

修改接口的可选字段**必须**使用指针类型，用于精确表达三种状态：
- **nil（未传 / JSON 中不存在该字段）**：不修改该字段，保持原值
- **非 nil 且非空字符串**：修改为新值
- **非 nil 且空字符串**：清空该字段

这是为了防止 JSON 反序列化时，未传的字段被置为零值（空字符串），导致不应修改的字段被错误地清空。

定义 `SubAccountUpdateItem`：
- `id`: string，必填 — 本地 sub_account 表 ID
- `name`: *string，可选 — 修改三级账号在 HCM 平台上的显示名称（仅更新本地，不更新云上 Name）
- `email`: *string，可选 — 三级账号邮箱（同步云上 + 本地）
- `phone_num`: *string，可选 — 手机号（同步云上 + 本地）
- `country_code`: *string，可选 — 手机区域代码（同步云上 + 本地）
- `managers`: []string，可选 — 账号管理者（仅本地），nil 表示不修改，空数组表示清空
- `memo`: *string，可选 — 备注（仅本地）

在整条链路中，nil 语义需要被严格传递：
- **cloud-server handler**：Deliver 阶段构建 hc-service 请求时，只设置非 nil 字段到 `UpdateSubAccountReq`
- **hc-service**：构建 `UpdateUserOption` 时，只设置非 nil/非空字段到腾讯云 API 请求参数
- **data-service**：构建 `UpdateField` 时，只设置非 nil 字段

`SubAccountUpdateReq` 包含 `Vendor`、`BkBizID`、`SubAccounts []SubAccountUpdateItem`、`Action`，与 `SubAccountAddReq` 结构对齐。

### 5. CheckReq 校验逻辑

1. 调用 `SubAccountUpdateItem.Validate()` 校验字段合法性
2. 通过 data-service `SubAccount.List` 按 ID 查询，验证三级账号存在
3. 从查询结果中获取 `account_id`（所属二级账号 ID），设置到 handler 的 `accountID` 字段
4. 通过 `GetAccount(accountID)` 验证二级账号存在

### 6. hc-service 层新增 UpdateSubAccount API

路由：`POST /vendors/tcloud/sub_accounts/update`

请求通过 hc-service → adaptor 调用腾讯云 `cam:UpdateUser`。UpdateUser API 以 Name（子用户用户名）标识目标用户，因此 hc-service 请求中需包含 `Name` 字段（从 sub_account 记录的 cloud name 获取）。

hc-service 层职责单一——只对接云 API；本地持久化由 cloud-server handler 的 Deliver 方法调用 data-service 完成。与现有 create/delete 模式一致。

### 7. adaptor 层新增 UpdateUser 方法

在 `pkg/adaptor/tcloud/account.go` 中新增 `UpdateUser`，封装 `cam.NewUpdateUserRequest()` + `camClient.UpdateUserWithContext()`。

入参结构体 `UpdateUserOption` 包含 Name（必填，用于标识用户）、Remark、ConsoleLogin、Password、NeedResetPassword、PhoneNum、CountryCode、Email 等字段，与腾讯云 UpdateUser API 参数对齐。

UpdateUser API 无实质性返回数据（仅 RequestId），因此方法返回 `error`。

### 8. 多云扩展性设计

- Deliver 方法中按 `vendor` switch-case 分发，目前仅实现 `tcloud` 分支，其他 vendor 返回 `not supported` 错误
- hc-service API 路径带 vendor 前缀，天然支持多云
- adaptor 接口定义在 vendor 各自的包中，不做抽象接口（与现有模式一致）

### 9. 交付流程：先云上后本地

Deliver 执行顺序：
1. 调用 hc-service `UpdateSubAccount` 更新云上子用户信息（email、phone_num、country_code 等云上支持的字段）
2. 调用 data-service `SubAccount.BatchUpdate` 更新本地 sub_account 表（所有修改字段，包括 managers、memo 等仅本地存储的字段）

如果云上更新失败，直接返回 `DeliverError`，不更新本地。如果云上更新成功但本地更新失败，记录错误日志并返回 `DeliverError`（后续可通过 sync 机制补齐）。

### 10. 批量提交设计

API 支持批量提交（`sub_accounts` 数组，上限 100），但每个三级账号独立创建一张 ITSM 审批单，独立交付。与创建/删除保持一致的审批粒度。

返回值为批量创建的 application ID 数组。

## Risks / Trade-offs

- **[风险] 云上更新成功但本地写入失败** → 记录错误日志，deliver 状态为 `deliver_error`，运维可通过 sync 机制补齐数据
- **[风险] 腾讯云 UpdateUser API 以 Name 标识用户** → 若本地记录的 Name 与云上不一致（如手动在云上修改过），可能更新错误用户。可通过 CheckReq 阶段查询云上实际用户信息缓解
- **[权衡] 不修改云上 Name** → 腾讯云 UpdateUser 不支持修改 Name，仅本地 name 字段可修改作为 HCM 平台显示名称
- **[权衡] 每个三级账号独立审批单** → 牺牲审批效率换取操作粒度和故障隔离，与创建/删除保持一致
