## Context

权限模板（permission_template）存储从云上同步或从权限策略库分发到二级账号的具体策略实例。它与已有的 `permission_policy_library` 表存在业务关联：模板可通过 `policy_library_id` 追溯来源策略库。

当前 DataService 已有完整的 CRUD 分层模式（SQL → Table → DAO → API → Service → Client），`permission_policy_library` 是最近新增的同类资源，可作为直接参考。

## Goals / Non-Goals

**Goals:**
- 按项目已有分层架构为 `permission_template` 表生成完整的 DataService CRUD 接口
- 支持 TCloud vendor 的 extension 扩展字段
- 支持 Create/Update/Delete 审计记录
- `policy_document` 变更时自动计算 `policy_hash`（SHA256）

**Non-Goals:**
- 不实现 cloud-server 层的接口（由后续变更覆盖）
- 不实现策略库到模板的同步逻辑
- 不添加数据库索引（后续按需添加）
- 不实现除 TCloud 以外的 vendor 支持

## Decisions

### 1. 可空时间戳字段使用 `*string`

`policy_library_sync_time` 是项目中首个可空 timestamp 字段。`types.Time` 底层为 `string`，选择 `*string` 而非 `*types.Time`，避免 sqlx scan 对自定义类型指针的兼容风险。ColumnDescriptor 中仍用 `enumor.Time`。

### 2. extension 字段 Create 时必填

对 TCloud vendor，`cloud_type`（1=自定义策略，2=预设策略）总是已知的。DS Create 请求中 extension 设为 `validate:"required"`。Update 时为可选。

### 3. Update 无需自增版本

与 `permission_policy_library` 不同，`permission_template` 没有自身的 version 字段。`policy_library_version` 仅为快照。Update handler 中 `policy_document` 变更时只需重算 hash，不需要查已有记录做版本对比。

### 4. DAO AddBlankedFields 列表

以下字段在 Update 时允许置空：`memo`, `extension`, `policy_library_id`, `policy_library_version`, `policy_library_sync_time`。

### 5. 审计模式复用 permission_policy_library

- DAO `BatchCreateWithTx` 内直接写 Create 审计（`enumor.PermissionTemplateAuditResType`）
- DS audit 模块新建 `permission_template.go`，实现 Update/Delete 审计构建
- `create_resource_update_audit.go` 和 `create_resource_delete_audit.go` 增加 case 分支

## Risks / Trade-offs

- [无唯一索引] `cloud_id` + `account_id` 无联合唯一约束，依赖上层业务保证不重复 → 后续可按需追加
- [可空时间用 *string] 丢失类型语义 → 影响可控，字段仅用于记录，无时间运算需求
