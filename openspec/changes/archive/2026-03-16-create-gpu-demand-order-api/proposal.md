## Why

前端通过Excel导入预览接口获取GPU需求解析数据后，需要一个接口将解析结果持久化为GPU需求提报主单和子单。目前已有Excel导入预览接口（`ExcelImportGpuDemand`）和data-service层的批量创建接口，但woa-server层缺少将前端传入的details转换并创建主单+子单的业务编排接口。

## What Changes

- 新增woa-server业务层接口 `POST /api/v1/woa/bizs/{bk_biz_id}/plans/resource/gpu/order/create`，接收前端传入的运营产品信息和details列表
- 新增woa-server请求类型 `CreateGpuDemandOrderReq`，包含 `op_product_id`、`op_product_name`、`details` 字段
- 新增service层handler `CreateGpuDemandOrder`，负责请求解析、校验、鉴权
- 新增logics层方法 `CreateGpuDemandOrder`，负责：
  - 获取最新GPU模板ID
  - 调用data-service创建GPU需求主单（状态为INIT）
  - 将前端传入的details转换为子单（extension字段JSON序列化、关联主单order_id、状态为INIT）
  - 调用data-service批量创建GPU需求子单
- 返回主单ID给前端

## Capabilities

### New Capabilities
- `create-gpu-demand-order`: woa-server层创建GPU需求提报主单及子单的业务编排接口，包含请求类型定义、service handler、logics业务逻辑

### Modified Capabilities
- `res-plan-demand-gpu-order-dataservice`: 无需求级别变更，仅作为下游依赖被调用
- `res-plan-demand-gpu-suborder-data-service`: 无需求级别变更，仅作为下游依赖被调用

## Impact

- **新增文件**：
  - `cmd/woa-server/service/plan/gpu_demand_order.go` — service层handler
  - `cmd/woa-server/logics/plan/gpu_demand_order.go` — logics层业务逻辑
- **修改文件**：
  - `cmd/woa-server/service/plan/service.go` — 在 `initBizPlanService` 中注册新路由
  - `cmd/woa-server/logics/plan/plan.go` — 在 `Logics` 接口中新增方法签名
  - `pkg/api/woa-server/` — 新增或扩展woa-server请求/响应类型定义
- **依赖**：
  - 复用已有的 `ResourcePlanClient.BatchCreateResPlanDemandGpuOrder` 和 `BatchCreateResPlanDemandGpuSubOrder` 客户端方法
  - 复用已有的data-service批量创建接口
  - 复用 `enumor.ResPlanDemandGpuOrderStatusInit` 和 `enumor.RPDemandGPUSubOrderStatusInit` 状态枚举
  - 需要获取最新GPU模板ID（复用 `getLatestGpuTplSchema` 或类似逻辑）
