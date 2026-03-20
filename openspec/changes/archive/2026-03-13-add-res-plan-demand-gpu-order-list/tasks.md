## 1. 请求/响应结构体

- [x] 1.1 在 `cmd/woa-server/types/plan/list_gpu_demand_order.go` 中定义 `ListGpuDemandOrderReq`（filter + page）及其 `Validate()` 方法
- [x] 1.2 定义 `GpuDemandOrderItem`（主单字段 + TotalGpuNum + TotalQpmMax）
- [x] 1.3 定义 `ListGpuDemandOrderResult`（Count + Details）

## 2. woa-server 核心查询逻辑

- [x] 2.1 在 `cmd/woa-server/service/plan/gpu_demand_order.go` 中实现 `fetchSubOrderStats(kt, orderIDs)` — 循环翻页查子单，返回 `map[orderID]{gpuNum, qpmMax}`
- [x] 2.2 实现 `assembleGpuOrderItems(orders, statsMap)` — 将主单与聚合数据拼装为 `[]GpuDemandOrderItem`
- [x] 2.3 实现 `listResPlanDemandGpuOrder(kt, filter, page)` — 协调两阶段查询，count 模式下短路

## 3. woa-server Handler

- [x] 3.1 实现 `ListResPlanDemandGpuOrder`（资源视角）— 鉴权 ZiYanResPlanGPUDemands，调用 `listResPlanDemandGpuOrder`
- [x] 3.2 实现 `ListBizResPlanDemandGpuOrder`（业务视角）— 从路径取 bk_biz_id，ListAuthorizedBiz 鉴权，注入 bk_biz_id filter，调用 `listResPlanDemandGpuOrder`

## 4. 路由注册

- [x] 4.1 在 `service.go` 的 `initPlanService` 中注册资源视角路由：`POST /plans/resources/gpu/demands/orders/list`
- [x] 4.2 在 `service.go` 的 `initBizPlanService` 中注册业务视角路由：`POST /plans/resources/gpu/demands/orders/list`
