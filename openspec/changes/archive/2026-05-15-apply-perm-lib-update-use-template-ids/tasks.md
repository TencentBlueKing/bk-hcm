## 1. Proto 结构体变更

- [x] 1.1 `pkg/api/cloud-server/application/permission_policy_library.go`：新增 `ApplyPermPolicyLibBaseContent`（Action/Vendor/BkBizID）
- [x] 1.2 同文件：新增 `ApplyPermPolicyLibUpdateContent`（embed base + PolicyLibraryID + PermissionTemplateID）
- [x] 1.3 同文件：将 `ApplyPermPolicyLibContent` 改名为 `ApplyPermPolicyLibCreateContent`（保留原字段，embed base）
- [x] 1.4 同文件：`BizApplyPermissionPolicyLibraryUpdateReq` 中 `AccountIDs` 改为 `PermissionTemplateIDs`
- [x] 1.5 `pkg/api/cloud-server/permission_policy_library.go`：新增 `ApplyPermissionPolicyLibraryUpdateReq`（含 `permission_template_ids`）
- [x] 1.6 同文件：新增 `ApplyTemplateResult`（permission_template_id / status / reason）
- [x] 1.7 同文件：新增 `ApplyPermissionPolicyLibraryUpdateResult`（results []ApplyTemplateResult）

## 2. Handler 基础设施调整

- [x] 2.1 `handlers/permission-policy-library/base.go`：`ActionHandlerFactory` 签名改为 `func(opt, base *ApplyPermPolicyLibBaseContent, content string) (handlers.ApplicationHandler, error)`
- [x] 2.2 同文件：`NewHandlerFromApplication` 改为先反序列化 `BaseContent`，再以 rawContent 调用工厂
- [x] 2.3 同文件：`ApplicationBasePermissionPolicyLibrary` 的 `Content` 字段类型改为 `*ApplyPermPolicyLibBaseContent`（或根据 create/update 分别持有各自 content）

## 3. apply-create handler 适配（小改）

- [x] 3.1 `apply-create/init.go`：工厂函数签名适配，内部二次反序列化为 `ApplyPermPolicyLibCreateContent`
- [x] 3.2 同文件：`BuildContent` 返回类型改为 `*ApplyPermPolicyLibCreateContent`
- [x] 3.3 `apply-create/check.go`：引用类型改为 `ApplyPermPolicyLibCreateContent`
- [x] 3.4 `apply-create/deliver.go`：引用类型改为 `ApplyPermPolicyLibCreateContent`
- [x] 3.5 `apply-create/create_itsm_ticket.go`：引用类型改为 `ApplyPermPolicyLibCreateContent`

## 4. apply-update handler 改造

- [x] 4.1 `apply-update/init.go`：工厂函数签名适配，内部反序列化为 `ApplyPermPolicyLibUpdateContent`；`BuildContent` 入参改为 `PermissionTemplateID`，返回 `*ApplyPermPolicyLibUpdateContent`；`ApplicationOfApplyPermPolicyLibUpdate` 存 `*ApplyPermPolicyLibUpdateContent`
- [x] 4.2 `apply-update/check.go`：改为验证 PermissionTemplateID → `GetTemplateByID` 查模版 → 校验 policy_library_id 匹配 → 校验账号 biz 在 scope 内；删除 `CheckAccountApplied` 调用
- [x] 4.3 `apply-update/deliver.go`：改为调用 `applier.ApplyUpdateByTemplateIDs`；Deliver 结果使用 `ApplyTemplateResult`；detail map 改为 `permission_template_id`
- [x] 4.4 `apply-update/create_itsm_ticket.go`：`RenderItsmTitle`/`RenderItsmForm` 改为基于 PermissionTemplateID 和模版信息渲染（通过 GetTemplateByID 获取账号信息）

## 5. applier 新增方法

- [x] 5.1 `applier.go`：新增 `GetTemplateByID(kt, templateID)`，通过 ListPermissionTemplateExt 按 id 查询，不存在返回 nil
- [x] 5.2 同文件：新增 `tcloudApplyUpdateForTemplate(kt, library, templateID)`，从模版获取 CloudID + AccountID，调 TCloudUpdateCAMPolicy + TCloudUpdateLocalTemplate，返回 `ApplyTemplateResult`
- [x] 5.3 同文件：新增 `ApplyUpdateByTemplateIDs(kt, vendor, libraryID, templateIDs)`，获取库详情 → 遍历调 `tcloudApplyUpdateForTemplate` → 返回 `ApplyPermissionPolicyLibraryUpdateResult`
- [x] 5.4 同文件：删除旧 `ApplyUpdate(accountIDs)` 方法及 `tcloudApplyUpdate`、`tcloudApplyUpdateForAccount`

## 6. Service 层调整

- [x] 6.1 `service/application/create.go`：`createBizForApplyPermPolicyLibUpdate` 遍历 `PermissionTemplateIDs`，调 `applyupdate.BuildContent(bizID, vendor, req, templateID)`
- [x] 6.2 `service/permission-policy-library/apply.go`：`ApplyPermissionPolicyLibraryUpdate` 改用 `ApplyPermissionPolicyLibraryUpdateReq` 和 `applier.ApplyUpdateByTemplateIDs`

## 7. 接口文档更新

- [x] 7.1 `docs/api-docs/web-server/docs/biz/permission-policy-library/apply_permission_policy_library_update.md`：输入参数 `account_ids` → `permission_template_ids`，响应 `ids` 描述保持不变
- [x] 7.2 `docs/api-docs/web-server/docs/resource/permission-policy-library/apply_permission_policy_library_update.md`：输入参数 `account_ids` → `permission_template_ids`；响应结果从 `account_id` 改为 `permission_template_id`
