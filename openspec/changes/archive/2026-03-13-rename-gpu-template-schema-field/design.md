## Context

`ResPlanDemandGpuTemplateTable` 的 DB 列名为 `tpl_schema`，Go 字段名已改为 `TplSchema`，但 JSON tag 仍为 `"schema"`。上游 API 层 (`DemandGpuTemplateCreateReq`、`DemandGpuTemplateUpdateReq`) 的字段名仍为 `Schema`，JSON tag 也是 `"schema"`。data-service 的 create/update 逻辑中使用 `Schema` 赋值。此外 `InsertValidate()` 中引用了 `r.Schema`（实际字段已是 `TplSchema`），存在编译错误。

涉及层级：API 层（pkg/api）、Table 层（pkg/dal/table）、Service 层（cmd/data-service）。不涉及 client 层的 Schema 字段直接引用，但 JSON 序列化格式有变更。

## Goals / Non-Goals

**Goals:**
- 将 JSON tag 统一为 `"tpl_schema"`，与 DB 列名一致
- 将 API 层字段名从 `Schema` 改为 `TplSchema`，与 table 层一致
- 修复 `InsertValidate()` 中的编译错误
- 确保 create/update service 逻辑中字段赋值正确

**Non-Goals:**
- 不涉及数据库 DDL 变更（DB 列名已经是 `tpl_schema`）
- 不涉及前端代码变更（前端目前无 gpu template 相关引用）
- 不涉及 client 层接口签名变更（client 仅透传请求体）

## Decisions

**逐层重命名，自底向上**：Table 层 → API 层 → Service 层。这是最简单直接的方式，所有变更都是机械性的 find-and-replace，无需引入任何新逻辑。

## Risks / Trade-offs

- **[Breaking Change]** API JSON 字段从 `"schema"` 改为 `"tpl_schema"` → 调用方需要同步更新请求体字段名。该功能较新（gpu template CRUD 刚落地），影响面可控。
