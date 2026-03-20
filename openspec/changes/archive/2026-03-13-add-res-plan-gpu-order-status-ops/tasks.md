## 1. woa-server 新增 GPU 主单状态变更接口

- [x] 1.1 新建 `cmd/woa-server/service/plan/gpu_demand_order.go`，实现私有公共方法 `batchUpdateGpuOrderStatus(kt *kit.Kit, orderIDs []string, allowedFromStatuses []enumor.ResPlanDemandGpuOrderStatus, targetOrderStatus enumor.ResPlanDemandGpuOrderStatus, targetSubOrderStatus enumor.RPDemandGPUSubOrderStatus) error`，包含：查主单→校验前置状态→写审计→更新主单→查子单→分批更新子单
- [x] 1.2 在同文件中实现 `BatchSetResPlanDemandGpuOrderPending`（资源下评审中）handler，权限校验：`meta.ZiYanResPlanGPUDemands + meta.Update`，调用 `batchUpdateGpuOrderStatus([INIT], PENDING, PENDING)`
- [x] 1.3 在同文件中实现 `BatchRejectResPlanDemandGpuOrder`（资源下整单驳回）handler，权限校验：`meta.ZiYanResPlanGPUDemands + meta.Update`，调用 `batchUpdateGpuOrderStatus([PENDING], REJECT_ALL, REJECT)`
- [x] 1.4 在同文件中实现 `BatchTerminateResPlanDemandGpuOrder`（资源下终止）handler，权限校验：`meta.ZiYanResPlanGPUDemands + meta.Update`，调用 `batchUpdateGpuOrderStatus([PENDING], TERMINATE, TERMINATE)`
- [x] 1.5 在同文件中实现 `BatchTerminateBizResPlanDemandGpuOrder`（业务下终止）handler，从 path 取 `bk_biz_id`，使用 `bizLogics.ListAuthorizedBiz` 校验业务权限，调用 `batchUpdateGpuOrderStatus([INIT, REJECT_ALL], TERMINATE, TERMINATE)`

## 2. 路由注册

- [x] 2.1 在 `cmd/woa-server/service/plan/service.go` 的 `initPlanService` 中注册 3 条资源下路由：`POST /plans/resources/gpu/demands/orders/batch/pending`、`/batch/reject`、`/batch/terminate`
- [x] 2.2 在 `cmd/woa-server/service/plan/service.go` 的 `initBizPlanService` 中注册 1 条业务下路由：`POST /plans/gpu/demands/orders/batch/terminate`
