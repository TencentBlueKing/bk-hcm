## Why

资源计划模块已完成 GPU 需求提报主单（`res_plan_demand_gpu_order`）的数据服务层建设，但尚未对外暴露列表查询接口。前端和业务方需要通过资源视角（SCR）和业务视角（BIZ）两个维度查询主单列表，并在列表中展示由关联子单汇总得出的 `total_gpu_num` 和 `total_qpm_max` 聚合字段。

## What Changes

- 新增 woa-server 资源视角接口：`POST /api/v1/woa/plans/resources/gpu/demands/orders/list`
  - 权限：平台-GPU需求（ZiYanResPlanGPUDemands）
  - 支持通用 filter + page 查询
- 新增 woa-server 业务视角接口：`POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/gpu/demands/orders/list`
  - 权限：业务访问，自动注入 bk_biz_id 过滤条件
- 两个接口响应中均包含聚合字段 `total_gpu_num`、`total_qpm_max`（由关联子单 SUM 汇总）
- 采用两阶段查询方案：先查主单，再查子单明细在 woa-server 内存中聚合，不修改 data-service 和 DAO 层

## Capabilities

### New Capabilities

- `res-plan-demand-gpu-order-list`: GPU 需求提报主单列表查询能力，支持资源视角和业务视角，响应包含子单聚合字段

### Modified Capabilities

## Impact

- **新增文件**：
  - `cmd/woa-server/types/plan/list_gpu_demand_order.go`（请求/响应结构体）
- **修改文件**：
  - `cmd/woa-server/service/plan/gpu_demand_order.go`（新增两个 handler 和核心查询逻辑）
  - `cmd/woa-server/service/plan/service.go`（注册两条路由）
- **不涉及**：data-service、DAO、数据库表结构、client 层（全部复用现有接口）
