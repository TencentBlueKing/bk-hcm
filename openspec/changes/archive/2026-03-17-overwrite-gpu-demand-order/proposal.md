## Why

GPU需求提报单被"全部驳回"后，业务方需要修正数据后重新提交，但目前没有覆盖上传的能力，只能终止旧单后重新创建新单，流程繁琐且主单ID发生变化，不利于追溯。新增覆盖上传接口可以在保留主单ID的前提下，用新的Excel数据替换已驳回的子单。

## What Changes

- 新增 `PATCH /api/v1/woa/bizs/{bk_biz_id}/plans/resources/gpu/order/overwrite` 接口
- 接收主单ID（order_id）和新的details列表（来自Excel解析）
- 校验主单状态必须为 `REJECT_ALL`（全部已驳回），否则拒绝操作
- 校验该主单下所有子单状态均为 `REJECT` 或 `TERMINATE`，否则拒绝操作
- 对新传入的details执行与创建时相同的extension校验
- 校验通过后，删除该主单下所有旧子单，根据新details重新创建子单
- 覆盖完成后，将主单状态重置为 `INIT`（待评审）

## Capabilities

### New Capabilities
- `overwrite-gpu-demand-order`: 覆盖上传GPU需求提报单，在主单被全部驳回后允许用新的子单数据替换旧子单

### Modified Capabilities

（无现有spec需要修改）

## Impact

- **Service Layer**: `woa-server` 新增路由和handler
- **Logic Layer**: `cmd/woa-server/logics/plan/gpu_demand_order.go` 新增覆盖逻辑
- **API Types**: `pkg/api/woa-server/res_plan_gpu.go` 新增请求类型 `OverwriteGpuDemandOrderReq`
- **Data Service Client**: 使用已有的 `DeleteResPlanDemandGpuSubOrder` 和 `BatchCreateResPlanDemandGpuSubOrder` 方法
- **Route Registration**: `cmd/woa-server/service/plan/service.go` 注册新路由
