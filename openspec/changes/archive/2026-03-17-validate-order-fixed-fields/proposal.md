## Why

创建和覆盖 GPU 需求提报接口中，前端传入的固定字段（demand_year、demand_month、gpu_num、qpm_max 等）目前只做了基础的结构体校验（required、min/max 等 validate tag），没有依据模版 schema 的 fixed_headers 进行业务规则校验。例如 demand_year 在 schema 中定义为 enum 类型、可选值为 [2026, 2027, 2028]，但当前传入 2025 也能通过校验。需要补充基于 schema fixed_headers 的校验逻辑，与 extension 动态字段的校验保持一致。

## What Changes

- 在创建（CreateGpuDemandOrder）和覆盖（OverwriteGpuDemandOrder）流程中，增加对每个 detail 的固定字段按 schema fixed_headers 进行校验
- 利用 fixed_headers 中的 db_field 将 schema 定义映射到请求结构体中的 JSON 字段，提取对应值后调用已有的 ValidateTypedValue 校验器
- 校验逻辑复用 pkg/tools/excel/validate.go 中已有的类型校验和范围校验能力

## Capabilities

### New Capabilities
- `validate-order-fixed-fields`: 基于 schema fixed_headers 对 GPU 需求提报接口的固定字段进行校验

### Modified Capabilities

（无）

## Impact

- **Service Layer**: cmd/woa-server/logics/plan/gpu_demand_order.go - 在 validateOrderExtensions 或新增函数中补充 fixed_headers 校验逻辑
- **Tool Layer**: pkg/tools/excel/ - 可能需要在 schema.go 或 validate.go 中新增辅助方法
- **API 行为变更**: 创建和覆盖接口将对固定字段返回更严格的校验错误，不兼容之前传入非法值的调用
