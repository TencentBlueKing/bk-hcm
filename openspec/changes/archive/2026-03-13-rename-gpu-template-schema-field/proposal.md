## Why

`ResPlanDemandGpuTemplateTable` 的 `TplSchema` 字段 Go struct 名已经是 `TplSchema`，但 JSON tag 仍为 `"schema"`，导致上游 API 层（`DemandGpuTemplateCreateReq`、`DemandGpuTemplateUpdateReq`）的字段名也是 `Schema`。需要统一为 `TplSchema`/`"tpl_schema"`，保持 DB 列名、Go 字段名、JSON 字段名三者一致。

## What Changes

- **BREAKING**: 将 `ResPlanDemandGpuTemplateTable.TplSchema` 的 JSON tag 从 `"schema"` 改为 `"tpl_schema"`
- 将 API 层 `DemandGpuTemplateCreateReq.Schema` 和 `DemandGpuTemplateUpdateReq.Schema` 字段名改为 `TplSchema`，JSON tag 改为 `"tpl_schema"`
- 修正 table 层 `InsertValidate()` 中 `r.Schema.IsEmpty()` 为 `r.TplSchema.IsEmpty()`（当前编译期已报错）
- 更新 data-service create/update 逻辑中的字段赋值

## Capabilities

### New Capabilities

（无新增能力）

### Modified Capabilities

（无 spec 级别的需求变更，仅字段重命名）

## Impact

- **API 层**: `pkg/api/data-service/resource-plan/res_plan_demand_gpu_template.go` — Create/Update 请求体 JSON 字段从 `schema` → `tpl_schema`
- **Table 层**: `pkg/dal/table/resource-plan/res-plan-demand-gpu-template/res_plan_demand_gpu_template.go` — JSON tag 修改 + InsertValidate 修复
- **Service 层**: `cmd/data-service/service/resource-plan/res-plan-demand-gpu-template/create.go`、`update.go` — 字段赋值跟随改名
- **客户端影响**: 调用方需使用 `tpl_schema` 作为 JSON key，属于 **breaking change**
