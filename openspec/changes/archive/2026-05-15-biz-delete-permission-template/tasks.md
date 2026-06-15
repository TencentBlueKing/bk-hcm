## 1. Adaptor 层 — TCloud DeletePolicy

- [x] 1.1 在 `pkg/adaptor/types/account/tcloud_policy.go` 新增 `TCloudDeletePolicyOption` 结构体（含 `PolicyIDs []uint64`，validate `required,min=1`）及 `Validate()` 方法
- [x] 1.2 在 `pkg/adaptor/tcloud/cam_policy.go` 新增 `DeletePolicy(kt *kit.Kit, opt *typeaccount.TCloudDeletePolicyOption) error` 方法，调用腾讯云 CAM SDK 删除策略

## 2. hc-service 层 — DeleteCAMPolicy 接口

- [x] 2.1 在 `pkg/api/hc-service/permission-template/permission_template.go` 新增 `DeleteCAMPolicyReq` 结构体（`AccountID string`、`PolicyIDs []uint64`，validate `required,min=1`）及 `Validate()` 方法
- [x] 2.2 在 `cmd/hc-service/service/permission-template/cam_policy.go` 新增 `TCloudDeleteCAMPolicy` handler，解码请求、调用 adaptor `DeletePolicy`
- [x] 2.3 在 `cmd/hc-service/service/permission-template/service.go` 注册路由 `DELETE /permission_templates/cam/delete_policy`

## 3. hc-service Client 封装

- [x] 3.1 在 `pkg/client/hc-service/tcloud/permission_template.go` 新增 `DeleteCAMPolicy(kt *kit.Kit, req *proto.DeleteCAMPolicyReq) error` 方法

## 4. Enumor & API Proto

- [x] 4.1 在 `pkg/criteria/enumor/permission_template.go` 新增 `PermTemplateActionDelete OperatePermTemplateAction = "delete"`，并在 `Validate()` 中包含该值
- [x] 4.2 在 `pkg/api/cloud-server/application/permission_template.go` 新增 `BizDeletePermissionTemplateReq` 结构体（`ID string validate:"required"`）及 `Validate()` 方法

## 5. Application Handler — delete 包

- [x] 5.1 新建 `cmd/cloud-server/service/application/handlers/permission-template/delete/init.go`：定义 `deletePermTemplateContent`（嵌入 `BasePermTemplateContent` + `ID string`）、`ApplicationOfDeletePermTemplate`（嵌入 `permissiontemplate.ApplicationBasePermissionTemplate`，公共方法由 base 继承，无需手写桩方法）；覆写 `GetItsmApprover(kt, managers)` 委托给 `GetItsmApproverByTemplateID(kt, content.ID)`，在 `init()` 注册 `PermTemplateActionDelete` handler
- [x] 5.2 新建 `cmd/cloud-server/service/application/handlers/permission-template/delete/check.go`：实现 `CheckReq()`，内部 `switch vendor`，`checkTCloud()` 完成：获取模板详情、校验 CloudType==TCloudCustomPolicy、校验 biz 归属、查询子账号关联数==0
- [x] 5.3 新建 `cmd/cloud-server/service/application/handlers/permission-template/delete/deliver.go`：实现 `GenerateApplicationContent()` 和 `Deliver()`，内部 `switch vendor`，`deleteTCloud()` 完成：调用 hc-service `DeleteCAMPolicy` + data-service `BatchDelete`，返回 `Completed` 或 `DeliverError`
- [x] 5.4 新建 `cmd/cloud-server/service/application/handlers/permission-template/delete/create_itsm_ticket.go`：实现 `RenderItsmTitle()`（格式：`申请删除云权限模板(<name>)`）和 `RenderItsmForm()`（包含业务、云厂商、云账号、权限模板 ID、权限模板名称）

## 6. Application Service — 路由注册

- [x] 6.1 在 `cmd/cloud-server/service/application/create.go` 新增 `CreateBizForDeletePermissionTemplate` 方法：校验 biz_id、`meta.PermissionTemplate/meta.Delete` 鉴权含 BizID、vendor 校验，构造 `BasePermTemplateContent{Action: PermTemplateActionDelete}`，调用 delete handler
- [x] 6.2 在 `cmd/cloud-server/service/application/init.go` 注册路由 `POST /vendors/{vendor}/applications/types/delete_permission_template`，绑定 `CreateBizForDeletePermissionTemplate`
- [x] 6.3 在 `cmd/cloud-server/service/application/handlers/permission-template/delete/init.go` 的 `import` 中确保 delete 包被 side-effect 导入（在 application service 的某个 init 导入处添加 `_ "hcm/cmd/cloud-server/service/application/handlers/permission-template/delete"`）
