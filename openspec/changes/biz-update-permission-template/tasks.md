## 1. 枚举与 API 结构体

- [x] 1.1 在 `pkg/criteria/enumor/permission_template.go` 中新增 `PermTemplateActionUpdate = "update"` 常量，并更新 `Validate()` 方法
- [x] 1.2 在 `pkg/api/cloud-server/application/permission_template.go` 中新增 `BizUpdatePermissionTemplateReq` 结构体（字段：`ID`、`PolicyLibraryID`、`Memo`）及其 `Validate()` 方法

## 2. applier.go 扩展

- [x] 2.1 在 `applier.go` 中新增 `UpdateTmplBaseInfo` 结构体（含 `Memo *string`），作为 update 场景可覆盖字段的载体
- [x] 2.2 修改 `updateTCloudCAMPolicy` 方法签名，第三个参数改为 `tmplInfo UpdateTmplBaseInfo`，使用 `tmplInfo.Memo` 作为 CAM Policy description；现有 `ApplyUpdate` 调用路径传 `UpdateTmplBaseInfo{Memo: library.Memo}` 保持向后兼容
- [x] 2.3 修改 `updateTCloudLocalTemplate` 方法签名，第三个参数改为 `tmplInfo UpdateTmplBaseInfo`，使用 `tmplInfo.Memo` 更新 memo 字段；同时始终写入 `policy_library_id` 字段；现有 `ApplyUpdate` 调用路径传 `UpdateTmplBaseInfo{Memo: library.Memo}` 保持向后兼容
- [x] 2.4 在 `applier.go` 中新增 `ApplyUpdateWithTmplInfo(kt, vendor, libraryID string, templateIDs []string, tmplInfo UpdateTmplBaseInfo)` 方法，调用 `CheckPermTmplUpdatability` → `GetPolicyLibraryDetail` → `CheckAccountsBizInScope` → `applyTCloudUpdate`（传 tmplInfo）
- [x] 2.5 扩展 `CheckPermTmplUpdatability` / `checkTCloudPermTmplUpdatability`，在「自定义模板（policy_library_id=nil AND cloud_type=TCloudCustomPolicy）」之外，同时允许「已绑定同一策略库（policy_library_id == targetPolicyLibraryID）」的模板进行更新

## 3. update handler 实现

- [x] 3.1 创建 `handlers/permission-template/update/init.go`：定义 `updatePermTemplateContent` 结构体，实现 `NewApplicationOfUpdatePermTemplate` 和 `newHandlerFromContent`，在 `init()` 中注册 `PermTemplateActionUpdate`；覆写 `GetItsmApprover(kt, managers)` 委托给 `GetItsmApproverByTemplateID(kt, content.ID)`
- [x] 3.2 创建 `handlers/permission-template/update/check.go`：实现 `CheckReq()`，按 `id` 查出模板，校验 biz 归属，校验自定义模板条件（`policy_library_id=nil` AND `cloud_type=TCloudCustomPolicy`），校验新策略库在 biz scope 内
- [x] 3.3 创建 `handlers/permission-template/update/create_itsm_ticket.go`：实现 `RenderItsmTitle()`（格式：`申请更新云权限模板(<template_id>)`）和 `RenderItsmForm()`（包含业务名称、云厂商、云账号名称、权限模版 ID、权限策略库名称、策略库 ID、策略内容）
- [x] 3.4 创建 `handlers/permission-template/update/deliver.go`：实现 `GenerateApplicationContent()` 和 `Deliver()`，构造 `UpdateTmplBaseInfo{Memo: content.Memo}` 后调用 `ApplyUpdateWithTmplInfo(kt, vendor, libraryID, []string{templateID}, tmplInfo)` 执行云端+本地更新，检查 `resp.Results[0].Status`

## 4. 路由与审批注册

- [x] 4.1 在 `application/create.go` 中新增 `CreateBizForUpdatePermissionTemplate` handler 函数（参数校验、`meta.PermissionTemplate/meta.Update` 鉴权含 BizID、解析请求、调用 `NewApplicationOfUpdatePermTemplate`）
- [x] 4.2 在 `application/init.go` 中注册路由：`POST /vendors/{vendor}/applications/types/update_permission_template`
- [x] 4.3 在 `application/approve.go` 中添加 blank import `_ "hcm/cmd/cloud-server/service/application/handlers/permission-template/update"`，并在 `getHandlerByApplication` switch 中确认 `OperatePermissionTemplate` 分支已能通过 `permissiontemplate.NewHandlerFromApplication` 分发到 update handler（无需额外修改，因为 update 的 init 会自动注册到 registry）
