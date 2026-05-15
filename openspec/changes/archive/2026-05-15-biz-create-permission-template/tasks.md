## 1. 枚举与结构体

- [x] 1.1 `pkg/criteria/enumor/permission_template.go`：新增 `OperatePermTemplateAction` 枚举及 `PermTemplateActionCreate` 常量，实现 `Validate()`
- [x] 1.2 `pkg/criteria/enumor/application.go`：新增 `OperatePermissionTemplate ApplicationType = "operate_permission_template"`，加入 `Validate()` 的 case

## 2. API 结构体

- [x] 2.1 `pkg/api/cloud-server/application/permission_template.go`（新建）：定义 `BizCreatePermissionTemplateReq`（含 `Validate()`）和 `OperatePermTemplateContent`

## 3. applier.go 函数签名修改

- [x] 3.1 `cmd/cloud-server/service/permission-policy-library/applier.go`：`TCloudCreateCAMPolicy` 新增 `name string, memo *string` 入参，更新方法体
- [x] 3.2 同文件：`TCloudCreateLocalTemplate` 新增 `name string, memo *string` 入参，更新方法体
- [x] 3.3 同文件：更新 `tcloudApplyCreateForAccount` 内的两处调用，传入 `library.Name, library.Memo`

## 4. permission-template handler 体系

- [x] 4.1 `handlers/permission-template/base.go`（新建）：定义 `ActionHandlerFactory`、`actionHandlerRegistry`、`RegisterActionHandler`、`NewHandlerFromApplication`、`ApplicationBasePermissionTemplate` 及公共方法（`PrepareReq`、`PrepareReqFromContent`、`GetBkBizIDs`、`GetItsmApproverByTemplateID`）；其中 `GetItsmApproverByTemplateID(kt, id string)` 按模板 ID 查询账号后委托给 `GetAccountApprover`，替代原先固定返回 account_manager 的 `GetItsmApprover`
- [x] 4.2 `handlers/permission-template/create/init.go`（新建）：定义 `createPermTemplateContent`、`ApplicationOfCreatePermTemplate`、`NewApplicationOfCreatePermTemplate`、`newHandlerFromContent`，在 `init()` 中注册 `PermTemplateActionCreate`；并覆写 `GetItsmApprover(kt, managers)` 直接通过 `content.AccountID` 调用 `GetAccountApprover`
- [x] 4.3 `handlers/permission-template/create/check.go`（新建）：实现 `CheckReq()`，依次执行账号校验、biz 一致性校验、策略库校验、biz scope 校验、重复创建校验
- [x] 4.4 `handlers/permission-template/create/deliver.go`（新建）：实现 `GenerateApplicationContent()` 和 `Deliver()`，调用 `GetPolicyLibraryDetail`、`TCloudCreateCAMPolicy`、`TCloudCreateLocalTemplate`、`RecordApplyAudit`
- [x] 4.5 `handlers/permission-template/create/create_itsm_ticket.go`（新建）：实现 `RenderItsmTitle()` 和 `RenderItsmForm()`

## 5. 服务层注册

- [x] 5.1 `cmd/cloud-server/service/application/create.go`：新增 `CreateBizForCreatePermissionTemplate` handler（`meta.PermissionTemplate/meta.Create` 鉴权含 BizID、vendor 校验、请求解析、content 构造、调用 `a.create()`）
- [x] 5.2 同文件 `createApplication()`：在 bkBizIDs 赋值判断中加入 `enumor.OperatePermissionTemplate`
- [x] 5.3 `cmd/cloud-server/service/application/approve.go`：`getHandlerByApplication()` 新增 `case enumor.OperatePermissionTemplate`，调用 `permissiontemplate.NewHandlerFromApplication`；添加 blank import 触发 init() 注册
- [x] 5.4 `cmd/cloud-server/service/application/init.go`：`bizService()` 中注册路由 `POST /vendors/{vendor}/applications/types/create_permission_template`
