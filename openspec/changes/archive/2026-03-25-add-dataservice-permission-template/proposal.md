## Why

权限模板（permission_template）是混合云权限管理的核心资源，用于记录从云上同步或从权限策略库分发到二级账号的具体策略实例。当前缺少该表的 DataService CRUD 接口，需要补齐以支撑后续的权限模板同步、创建、更新、删除等业务流程。

## What Changes

- 新增 `permission_template` 数据库表（含 DDL）
- 新增完整的 DataService 分层 CRUD 代码：Table 定义 → DAO → API Model → Service Handler → Client
- 新增 `PermissionTemplateAuditResType` 审计资源类型，DAO Create 写 Create 审计，DS audit 模块补 Update/Delete 审计
- 新增 TCloud vendor 扩展字段定义（`cloud_type`）

## Capabilities

### New Capabilities
- `permission-template-crud`: 权限模板表的 DataService 层 CRUD 全链路实现，包括 SQL DDL、Table、DAO、API Model、Service Handler、Client 及审计

### Modified Capabilities

（无已有 spec 需要修改）

## Impact

- **新增文件**: 约 14 个新文件（SQL、Table、DAO、API、Service、Client、Audit）
- **修改文件**: 约 7 个已有文件（table.go、dao.go、service.go、client.go、audit.go 等注册点）
- **API 变更**: 新增 4 个 DataService HTTP 接口（Create/Update/Delete/List）
- **依赖**: 无新外部依赖，复用已有的 `types.JsonField`、`audit.Interface`、`idgenerator` 等基础设施
