## Why

GPU需求提报主单创建后需要经历评审流程（待评审→评审中→已驳回/已终止），目前 woa-server 缺少对应的状态变更接口，运营人员和业务用户无法通过平台触发这些操作。

## What Changes

- 新增资源下（运营管理员）批量将主单状态改为"评审中"的接口
- 新增资源下批量驳回主单（整单驳回）的接口
- 新增资源下批量终止主单的接口
- 新增业务下批量终止主单的接口
- 所有接口同步更新关联子单的状态，并记录审计日志

## Capabilities

### New Capabilities

- `gpu-order-status-ops`: GPU需求提报主单批量状态变更操作（评审中/驳回/终止），包含 woa-server 入口层、权限鉴权、审计记录，以及复用 data-service 现有 CRUD 接口完成主单和子单的状态更新

### Modified Capabilities

（无需求层面变更，仅新增能力）

## Impact

- **woa-server**：新增 `cmd/woa-server/service/plan/gpu_demand_order.go`，注册 4 条新路由（3 条资源下 + 1 条业务下）
- **不修改 data-service**：复用现有 `BatchUpdateResPlanDemandGpuOrder` 和 `BatchUpdateResPlanDemandGpuSubOrder` 接口
- **审计**：使用已有的 `ResPlanGPUDemandsOrderAuditResType` 审计类型，action 为 `Update`
- **权限**：资源下使用 `meta.ZiYanResPlanGPUDemands`，业务下使用业务访问权限校验
