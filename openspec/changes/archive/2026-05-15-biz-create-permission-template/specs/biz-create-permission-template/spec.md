## ADDED Requirements

### Requirement: OperatePermTemplateAction 枚举

系统 SHALL 在 `pkg/criteria/enumor/permission_template.go` 中定义 `OperatePermTemplateAction string` 类型及常量 `PermTemplateActionCreate = "create"`，并提供 `Validate()` 方法。该文件独立于 `permission_policy_library.go`。

#### Scenario: 合法 action 校验通过
- **WHEN** 调用 `PermTemplateActionCreate.Validate()`
- **THEN** 返回 nil

#### Scenario: 非法 action 校验失败
- **WHEN** 调用非法字符串值的 `OperatePermTemplateAction.Validate()`
- **THEN** 返回 unsupported 错误

### Requirement: OperatePermissionTemplate ApplicationType

系统 SHALL 在 `pkg/criteria/enumor/application.go` 中新增常量 `OperatePermissionTemplate ApplicationType = "operate_permission_template"`，并加入 `Validate()` 的 case 分支。

#### Scenario: ApplicationType 校验通过
- **WHEN** 调用 `OperatePermissionTemplate.Validate()`
- **THEN** 返回 nil

### Requirement: 请求与 Content 结构体

系统 SHALL 在 `pkg/api/cloud-server/application/permission_template.go`（新建文件）中定义：

`BizCreatePermissionTemplateReq`：
- `account_id` string（必填）：目标二级账号 ID
- `policy_library_id` string（必填）：权限策略库 ID
- `name` string（必填）：云权限模板名称
- `memo` string（可选）：备注
- 提供 `Validate()` 方法，调用 `validator.ValidatePermTmplName` 校验名称

`BasePermTemplateContent`：公共基础 Content，作为所有 action-specific content struct 的嵌入基础：
- `action` OperatePermTemplateAction：操作类型
- `vendor` enumor.Vendor：云厂商
- `bk_biz_id` int64：业务 ID

> 注意：不定义单一的 `OperatePermTemplateContent`；各 action 的完整 content 由各自的私有结构体在对应 handler 目录下定义（如 `createPermTemplateContent`）。

#### Scenario: 请求参数合法
- **WHEN** 传入 account_id、policy_library_id、name 均非空
- **THEN** `Validate()` 返回 nil

#### Scenario: 必填字段缺失
- **WHEN** 传入 account_id 为空
- **THEN** `Validate()` 返回 InvalidParameter 错误

### Requirement: permission-template handler base

系统 SHALL 在 `cmd/cloud-server/service/application/handlers/permission-template/base.go` 中提供：

- `ActionHandlerFactory` 类型：`func(opt *handlers.HandlerOption, base *proto.BasePermTemplateContent, content string) (handlers.ApplicationHandler, error)`；其中 `content` 为审批单中存储的原始 JSON 字符串，供 factory 内部解析 action-specific 字段。
- `actionHandlerRegistry`：`map[OperatePermTemplateAction]ActionHandlerFactory`
- `RegisterActionHandler(action, factory)`：供子 handler 的 `init()` 注册
- `NewHandlerFromApplication(opt, appContent string) (handlers.ApplicationHandler, error)`：先将 appContent 反序列化为 `BasePermTemplateContent` 以获取 action，再查找注册表调用对应 factory，并将 `base` 和原始 `appContent` 同时传给 factory。
- `ApplicationBasePermissionTemplate` struct，embed `handlers.BaseApplicationHandler` 和 `*permissionpolicylibrary.PolicyLibraryApplier`，持有 `bkBizID int64`（从 `BasePermTemplateContent` 中提取）；公开 `BkBizID() int64` 方法。
- `NewApplicationBasePermissionTemplate(opt, base *proto.BasePermTemplateContent) ApplicationBasePermissionTemplate`：构造函数，初始化 BaseApplicationHandler、PolicyLibraryApplier 及 bkBizID。
- 公共方法：`PrepareReq()`（no-op）、`PrepareReqFromContent()`（no-op）、`GetBkBizIDs() []int64`、`GetItsmApproverByTemplateID(kt *kit.Kit, id string) ([]itsm.VariableApprover, error)`（按模板 ID 查询 account_id 后委托给 `GetAccountApprover`）；各 action handler 自行覆写 `GetItsmApprover`

#### Scenario: 注册并分发 create action
- **WHEN** `create/init.go` 通过 `init()` 注册 `PermTemplateActionCreate`，之后调用 `NewHandlerFromApplication`，content 中 action 为 "create"
- **THEN** 返回 `ApplicationOfCreatePermTemplate` handler

#### Scenario: 未注册的 action
- **WHEN** 调用 `NewHandlerFromApplication`，content 中 action 为未注册的值
- **THEN** 返回 errf.InvalidParameter 错误

### Requirement: create action handler

系统 SHALL 在 `cmd/cloud-server/service/application/handlers/permission-template/create/` 下实现以下文件：

**init.go**：
- 定义私有 `createPermTemplateContent` struct，inline embed `proto.BasePermTemplateContent`，额外字段：`account_id` string、`policy_library_id` string、`name` string、`memo` *string。该结构体为审批单 content 的实际存储格式。
- 定义 `ApplicationOfCreatePermTemplate` struct，embed `ApplicationBasePermissionTemplate`，持有 `content *createPermTemplateContent`
- `NewApplicationOfCreatePermTemplate(opt *handlers.HandlerOption, base *proto.BasePermTemplateContent, req *proto.BizCreatePermissionTemplateReq) *ApplicationOfCreatePermTemplate`：从 HTTP 请求构造 handler，将 req 字段填入 createPermTemplateContent
- 私有 `newHandlerFromContent(opt, base, content string) (handlers.ApplicationHandler, error)`：从 JSON content 字符串反序列化 createPermTemplateContent 并构造 handler，作为注册的 factory
- `init()` 中调用 `RegisterActionHandler(PermTemplateActionCreate, newHandlerFromContent)`

**check.go** — `CheckReq()` SHALL 执行以下校验（顺序）：
1. 账号存在性校验（`GetAccount`）
2. 账号 biz_id 与请求 biz_id 一致校验
3. 策略库存在性校验（`GetPolicyLibraryDetail`）
4. 账号 biz 在策略库 biz 范围内校验（`CheckAccountsBizInScope`）
5. 账号未对该策略库创建过模板校验（`CheckAccountApplied`，已应用则返回错误）

**deliver.go** — `GenerateApplicationContent()` 返回 content struct；`Deliver()` SHALL：
1. 构造 `TmplBaseInfo{Name: content.Name, Memo: content.Memo}`
2. 调用 `ApplyCreateWithTmplInfo(kt, vendor, libraryID, []string{accountID}, tmplInfo)` 执行创建（该方法内部处理 GetPolicyLibraryDetail、TCloudCreateCAMPolicy、TCloudCreateLocalTemplate、RecordApplyAudit）
3. 校验响应 Results 长度为 1
4. 若 result.Status 为 Failed，返回 `enumor.DeliverError` 及错误信息
5. 成功返回 `enumor.Completed` 及包含 policy_library_id、account_id 的 deliveryDetail map

**create_itsm_ticket.go**：
- `RenderItsmTitle()`：返回 `"申请创建云权限模板({content.Name})到账号({content.AccountID})"`
- `RenderItsmForm()`：渲染包含业务名、云厂商、云账号名、策略库名、策略库 ID、策略内容、模板名称、模板描述的表单字符串

#### Scenario: CheckReq 账号已应用该策略库
- **WHEN** 目标账号已有来自同一策略库的权限模板记录
- **THEN** `CheckReq()` 返回"已创建"相关错误，阻止重复提交

#### Scenario: Deliver 成功
- **WHEN** CAM Policy 创建成功且本地模板写入成功
- **THEN** 返回 `enumor.Completed` 及包含 policy_library_id、account_id 的 deliveryDetail map

#### Scenario: CAM Policy 创建失败
- **WHEN** hc-service 调用失败
- **THEN** 返回 `enumor.DeliverError` 及包含 error 信息的 map

#### Scenario: 本地模板写入失败
- **WHEN** CAM Policy 已创建但 DataService 写入失败
- **THEN** 返回 `enumor.DeliverError`，错误信息中包含 cloudPolicyID，便于人工补录

### Requirement: cloud-server 服务层注册

系统 SHALL 在以下位置完成注册：

**`cmd/cloud-server/service/application/init.go`** — `bizService()` 新增路由：
`POST /vendors/{vendor}/applications/types/create_permission_template` → `CreateBizForCreatePermissionTemplate`

**`cmd/cloud-server/service/application/create.go`** — 新增 `CreateBizForCreatePermissionTemplate`，该 handler SHALL：
1. 校验 `bk_biz_id > 0`
2. 校验云权限模板操作权限（`meta.PermissionTemplate` / `meta.Create`，含 BizID）
3. 解析并校验 `vendor`
4. 解析 `BizCreatePermissionTemplateReq` 并校验
5. 构造 `BasePermTemplateContent{Action: PermTemplateActionCreate, Vendor: vendor, BkBizID: bizID}`，调用 `NewApplicationOfCreatePermTemplate(opt, base, req)` 创建 handler，调用 `a.create(cts, &proto.CreateCommonReq{}, handler)`
6. `createApplication()` 的 bkBizIDs 判断中加入 `OperatePermissionTemplate`

**`cmd/cloud-server/service/application/approve.go`** — `getHandlerByApplication()` 新增 case：
`case enumor.OperatePermissionTemplate` → 调用 `permissiontemplate.NewHandlerFromApplication(opt, application.Content)`
以 blank import 触发 create handler 的 `init()` 注册

#### Scenario: 正常创建申请单
- **WHEN** 业务用户 POST /bizs/1/vendors/tcloud/applications/types/create_permission_template，提供合法参数
- **THEN** 创建 ITSM 单据，写入审批单记录，返回 `{"id": "<application_id>"}`

#### Scenario: vendor 不支持
- **WHEN** vendor 不是 tcloud
- **THEN** 返回 InvalidParameter 错误

#### Scenario: 权限不足
- **WHEN** 用户无 PermissionTemplate Create 权限
- **THEN** 返回 PermissionDenied 错误

#### Scenario: 审批通过后 deliver
- **WHEN** ITSM 审批通过，回调触发 deliver
- **THEN** `getHandlerByApplication` 正确分发到 create handler，执行 `CheckReq` + `Deliver`
