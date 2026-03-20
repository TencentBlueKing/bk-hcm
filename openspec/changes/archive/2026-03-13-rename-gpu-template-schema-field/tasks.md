## 1. Table 层修改

- [x] 1.1 修改 `pkg/dal/table/resource-plan/res-plan-demand-gpu-template/res_plan_demand_gpu_template.go` 中 `TplSchema` 字段的 JSON tag 从 `json:"schema"` 改为 `json:"tpl_schema"`
- [x] 1.2 修复 `InsertValidate()` 中 `r.Schema.IsEmpty()` 为 `r.TplSchema.IsEmpty()`

## 2. API 层修改

- [x] 2.1 修改 `pkg/api/data-service/resource-plan/res_plan_demand_gpu_template.go` 中 `DemandGpuTemplateCreateReq.Schema` 字段名改为 `TplSchema`，JSON tag 改为 `"tpl_schema"`
- [x] 2.2 修改 `pkg/api/data-service/resource-plan/res_plan_demand_gpu_template.go` 中 `DemandGpuTemplateUpdateReq.Schema` 字段名改为 `TplSchema`，JSON tag 改为 `"tpl_schema"`

## 3. Service 层修改

- [x] 3.1 修改 `cmd/data-service/service/resource-plan/res-plan-demand-gpu-template/create.go` 中 `item.Schema` 赋值改为 `item.TplSchema`
- [x] 3.2 修改 `cmd/data-service/service/resource-plan/res-plan-demand-gpu-template/update.go` 中 `updateReq.Schema` 赋值改为 `updateReq.TplSchema`
