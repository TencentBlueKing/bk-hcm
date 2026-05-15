## Why

权限模板的数据层（CRUD）已实现，但业务层缺少创建入口。用户需要通过审批流（ITSM）为指定二级账号创建云权限模板，以便将策略库的策略内容实际应用到云账号上，并支持自定义模板名称和备注。

## What Changes

- 新增 `ApplicationType = operate_permission_template`，用于承载权限模板操作类审批单
- 新增 `OperatePermTemplateAction` 枚举（create），为后续 update/delete 预留扩展点
- 新增云权限模板创建请求结构体 `BizCreatePermissionTemplateReq` 及 Content 结构体
- 新增 `permission-template` 独立 handler 体系（base.go + create/），参照 `permission-policy-library` 模式
- 修改 `applier.go` 的 `TCloudCreateCAMPolicy` 和 `TCloudCreateLocalTemplate`，新增 `name`/`memo` 入参，支持自定义命名
- 新增 biz 路由 `POST /bizs/{bk_biz_id}/vendors/{vendor}/applications/types/create_permission_template`
- 在审批回调中注册 `OperatePermissionTemplate` 类型的 handler 分发

## Capabilities

### New Capabilities

- `biz-create-permission-template`: 业务层创建云权限模板的 ITSM 审批流接口，包括请求结构、handler 体系（base + create action）、服务层路由注册和审批回调分发

### Modified Capabilities

- `permission-template-crud`: `TCloudCreateCAMPolicy` 和 `TCloudCreateLocalTemplate` 函数签名新增 `name string, memo *string` 入参，调用方（`tcloudApplyCreateForAccount`）传入 `library.Name, library.Memo` 保持原有行为不变

## Impact

- `pkg/criteria/enumor/application.go`：新增 `OperatePermissionTemplate` ApplicationType
- `pkg/criteria/enumor/permission_template.go`（新建）：定义 `OperatePermTemplateAction` 枚举
- `pkg/api/cloud-server/application/permission_template.go`（新建）：请求/Content 结构体
- `cmd/cloud-server/service/permission-policy-library/applier.go`：修改两个方法签名，更新内部调用
- `cmd/cloud-server/service/application/handlers/permission-template/`（新建目录）：base.go + create/ 子目录
- `cmd/cloud-server/service/application/create.go`：新增 `CreateBizForCreatePermissionTemplate`，更新 bkBizIDs 判断
- `cmd/cloud-server/service/application/approve.go`：新增 `OperatePermissionTemplate` case
- `cmd/cloud-server/service/application/init.go`：注册 biz 路由
