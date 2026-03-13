## 1. Table定义层

- [x] 1.1 在 `pkg/dal/table/table.go` 中新增 `ResPlanDemandGpuTemplateTable` 表名常量，并在表名映射map中注册
- [x] 1.2 创建 `pkg/dal/table/resource-plan/res-plan-demand-gpu-template/res_plan_demand_gpu_template.go`，定义 `ResPlanDemandGpuTemplateTable` 结构体（含列描述符、InsertValidate、UpdateValidate）

## 2. DAO层

- [x] 2.1 在 `pkg/dal/dao/resource-plan/` 新增 `res_plan_demand_gpu_template.go`，定义 `DemandGpuTemplateInterface` 接口和 `DemandGpuTemplateDao` 实现（CreateWithTx、UpdateWithTx、List、DeleteWithTx）
- [x] 2.2 在 `pkg/dal/dao/dao.go` 的 `Set` 接口中新增 `ResPlanDemandGpuTemplate()` 方法声明
- [x] 2.3 在 `pkg/dal/dao/dao.go` 的 `set` 结构体上新增 `ResPlanDemandGpuTemplate()` 方法实现

## 3. API类型定义层

- [x] 3.1 在 `pkg/api/data-service/resource-plan/` 新增 `res_plan_demand_gpu_template.go`，定义批量创建请求、批量更新请求、列表查询请求和列表结果类型（含Validate方法）

## 4. Data-Service Handler层

- [x] 4.1 创建 `cmd/data-service/service/resource-plan/res-plan-demand-gpu-template/service.go`，注册CRUD路由（InitService）
- [x] 4.2 创建 `cmd/data-service/service/resource-plan/res-plan-demand-gpu-template/create.go`，实现 BatchCreateDemandGpuTemplate handler
- [x] 4.3 创建 `cmd/data-service/service/resource-plan/res-plan-demand-gpu-template/query.go`，实现 ListDemandGpuTemplate handler
- [x] 4.4 创建 `cmd/data-service/service/resource-plan/res-plan-demand-gpu-template/update.go`，实现 BatchUpdateDemandGpuTemplate handler
- [x] 4.5 创建 `cmd/data-service/service/resource-plan/res-plan-demand-gpu-template/delete.go`，实现 DeleteDemandGpuTemplate handler

## 5. 注册与集成

- [x] 5.1 在 `cmd/data-service/service/resource-plan/resource_plan.go` 的 InitService 中注册新service
