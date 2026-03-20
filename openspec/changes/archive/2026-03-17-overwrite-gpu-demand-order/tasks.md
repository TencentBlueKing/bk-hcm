## 1. API Types

- [x] 1.1 在 `pkg/api/woa-server/res_plan_gpu.go` 中新增 `OverwriteGpuDemandOrderReq` 请求结构体，包含 `OrderID string` 和 `Details []CreateGpuDemandOrderDetail`，并实现 `Validate()` 方法
- [x] 1.2 在 `Logics` 接口（`cmd/woa-server/logics/plan/plan.go`）中新增 `OverwriteGpuDemandOrder` 方法签名

## 2. Logic Layer

- [x] 2.1 在 `cmd/woa-server/logics/plan/gpu_demand_order.go` 中实现 `OverwriteGpuDemandOrder` 方法：获取schema → 校验extension → 获取templateID → 删除旧子单 → 创建新子单 → 重置主单状态为INIT
- [x] 2.2 实现子单状态校验辅助方法 `validateSubOrderStatuses`：查询主单下所有子单，校验每个子单状态必须为 `REJECT` 或 `TERMINATE`

## 3. Service Layer

- [x] 3.1 在 `cmd/woa-server/service/plan/gpu_demand_excel.go` 中新增 `OverwriteGpuDemandOrder` handler：解码请求 → 校验参数 → 鉴权(biz access) → 调用logic层执行覆盖（含状态校验）
- [x] 3.2 在 `cmd/woa-server/service/plan/service.go` 的 `initBizPlanService` 中注册路由 `PATCH /plans/resources/gpu/order/overwrite`
