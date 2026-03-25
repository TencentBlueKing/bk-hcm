## 1. SQL DDL 与 Table 层

- [x] 1.1 新建 `scripts/sql/9999_{date}_1001_permission_template.sql`，包含建表语句、id_generator 插入、hcm_version 视图
- [x] 1.2 新建 `pkg/dal/table/cloud/permission_template.go`，定义 ColumnDescriptor、PermissionTemplateTable struct、InsertValidate、UpdateValidate
- [x] 1.3 在 `pkg/dal/table/table.go` 中添加 `PermissionTemplateTable` 常量和 TableMap 配置（EnableTenant: true）

## 2. DAO 层

- [x] 2.1 新建 `pkg/dal/dao/types/permission_template.go`，定义 ListPermissionTemplateDetails
- [x] 2.2 新建 `pkg/dal/dao/cloud/permission-template/permission_template.go`，实现 PermissionTemplate interface（BatchCreateWithTx 含 Create 审计、BatchUpdate、BatchDelete、List）
- [x] 2.3 在 `pkg/dal/dao/dao.go` 中添加 PermissionTemplate() 接口方法和 set 实现（含 Audit 字段）

## 3. 审计注册

- [x] 3.1 在 `pkg/criteria/enumor/audit.go` 中添加 `PermissionTemplateAuditResType` 常量和 AuditResourceTypeMap 注册
- [x] 3.2 新建 `cmd/data-service/service/audit/cloud/permission_template.go`，实现 permissionTemplateUpdateAuditBuild、permissionTemplateDeleteAuditBuild、listPermissionTemplate
- [x] 3.3 在 `cmd/data-service/service/audit/cloud/create_resource_update_audit.go` 中添加 PermissionTemplateAuditResType case
- [x] 3.4 在 `cmd/data-service/service/audit/cloud/create_resource_delete_audit.go` 中添加 PermissionTemplateAuditResType case

## 4. API Model 层

- [x] 4.1 新建 `pkg/api/core/cloud/permission_template.go`，定义 BasePermissionTemplate（含 TCloudPermissionTemplateExtension）
- [x] 4.2 新建 `pkg/api/data-service/cloud/permission_template.go`，定义 Create/Update/Delete/List 的 Req/Resp 结构体

## 5. Service Handler 层

- [x] 5.1 新建 `cmd/data-service/service/cloud/permission-template/service.go`，注册 4 个路由（Create/Update 带 vendor，Delete/List 不带）
- [x] 5.2 新建 `cmd/data-service/service/cloud/permission-template/create.go`，实现 CreatePermissionTemplate（含 extension 序列化、policy_hash 计算）
- [x] 5.3 新建 `cmd/data-service/service/cloud/permission-template/update.go`，实现 BatchUpdatePermissionTemplate（含 policy_document 变更时重算 hash）
- [x] 5.4 新建 `cmd/data-service/service/cloud/permission-template/list.go`，实现 ListPermissionTemplate（含 extension 反序列化、Table→Core 转换）
- [x] 5.5 新建 `cmd/data-service/service/cloud/permission-template/delete.go`，实现 BatchDeletePermissionTemplate
- [x] 5.6 在 `cmd/data-service/service/service.go` 中添加 permissiontemplate.InitService(capability) 调用

## 6. Client 层

- [x] 6.1 新建 `pkg/client/data-service/tcloud/permission_template.go`，实现 PermissionTemplateClient（BatchCreate、BatchUpdate）
- [x] 6.2 新建 `pkg/client/data-service/global/permission_template.go`，实现 PermissionTemplateClient（ListPermissionTemplate、BatchDelete）
- [x] 6.3 在 `pkg/client/data-service/tcloud/client.go` 中添加 PermissionTemplate 字段和 NewClient 初始化
- [x] 6.4 在 `pkg/client/data-service/global/client.go` 中添加 PermissionTemplate 字段和 NewClient 初始化
