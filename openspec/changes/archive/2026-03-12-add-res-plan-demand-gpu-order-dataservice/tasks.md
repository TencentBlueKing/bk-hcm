## 1. 枚举类型与常量

- [x] 1.1 在 `pkg/criteria/enumor/` 下新增 `res_plan_demand_gpu_order_status.go`，定义 `ResPlanDemandGpuOrderStatus` 类型及全部枚举值（INIT/PENDING/DONE/REJECT/REJECT_ALL/TERMINATE），实现 `Validate()` 方法

## 2. Table 结构体

- [x] 2.1 在 `pkg/dal/table/table.go` 中添加 `ResPlanDemandGpuOrderTable Name = "res_plan_demand_gpu_order"` 常量，并在 `SupportedTables` map 中注册
- [x] 2.2 新建目录 `pkg/dal/table/resource-plan/res-plan-demand-gpu-order/`，创建 `res_plan_demand_gpu_order.go`，定义列描述符 `ResPlanDemandGpuOrderColumns`、`ResPlanDemandGpuOrderColumnDescriptor` 及 `ResPlanDemandGpuOrderTable` 结构体（注意 `op_product_name` 类型为 string）
- [x] 2.3 为 `ResPlanDemandGpuOrderTable` 实现 `TableName()`、`InsertValidate()`、`UpdateValidate()` 方法

## 3. DAO 层

- [x] 3.1 新建 `pkg/dal/dao/resource-plan/res_plan_demand_gpu_order.go`，定义 `ResPlanDemandGpuOrderInterface` 接口（`CreateWithTx`、`UpdateWithTx`、`List`、`DeleteWithTx`），并实现 `ResPlanDemandGpuOrderDao` 结构体
- [x] 3.2 在 `pkg/dal/dao/dao.go` 的 `Set` 接口中添加 `ResPlanDemandGpuOrder() resplan.ResPlanDemandGpuOrderInterface` 方法声明，并在 `set` 结构体上实现该方法

## 4. API 协议层

- [x] 4.1 新建 `pkg/api/data-service/resource-plan/res_plan_demand_gpu_order.go`，定义以下请求/响应结构体并实现 `Validate()`：
  - `ResPlanDemandGpuOrderBatchCreateReq` / `ResPlanDemandGpuOrderCreateReq`
  - `ResPlanDemandGpuOrderBatchUpdateReq` / `ResPlanDemandGpuOrderUpdateReq`
  - `ResPlanDemandGpuOrderListReq`
  - `ResPlanDemandGpuOrderListResult`

## 5. Service 层

- [x] 5.1 新建目录 `cmd/data-service/service/resource-plan/res-plan-demand-gpu-order/`，创建 `service.go`，定义 `service` 结构体与 `InitService` 函数，注册四个路由：`batch/create`（POST）、`batch`（PATCH 更新）、`list`（POST）、`batch`（DELETE）
- [x] 5.2 创建 `create.go`，实现 `BatchCreate` 方法（解析请求 → 事务 → 写 DAO → 返回 IDs）
- [x] 5.3 创建 `update.go`，实现 `BatchUpdate` 方法（解析请求 → 事务 → 逐条 UpdateWithTx）
- [x] 5.4 创建 `query.go`，实现 `List` 方法（解析请求 → 调用 DAO List → 返回结果）
- [x] 5.5 创建 `delete.go`，实现 `BatchDelete` 方法（解析请求 → 事务 → 调用 DAO DeleteWithTx）

## 6. 服务注册

- [x] 6.1 在 `cmd/data-service/service/resource-plan/resource_plan.go` 的 `InitService` 函数中 import 新包并调用 `resplandemandgpuorder.InitService(cap)`

## 7. 编译验证

- [x] 7.1 在项目根目录执行 `go build ./...`，确保无编译错误
