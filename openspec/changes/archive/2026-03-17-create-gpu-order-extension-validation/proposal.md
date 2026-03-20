## Why

CreateGpuDemandOrder 接口接收前端回传的 details，其中 extension 字段是按 schema headers（动态列）顺序排列的值数组。当前仅做基础结构校验（required、类型长度等由 validate tag 完成），但不对 extension 中的每个值按 schema header 定义的 type/value/min/max/required 规则校验。需要复用已有的 `ValidateCellValue` 校验逻辑，在创建工单前对 extension 做基于 schema 的类型和约束校验。

## What Changes

- 在 `pkg/tools/excel/validate.go` 新增 `ValidateExtension(values []interface{}, headers []Header) []string`，将 `interface{}` 转为字符串后逐个调用 `ValidateCellValue`
- 在 `CreateGpuDemandOrder` 的 controller 层调用时，先获取 schema，按 demand_type 匹配 sheet，取该 sheet 的 headers 校验每条 detail 的 extension
- 校验失败返回错误，不创建工单

## Capabilities

### New Capabilities
- `extension-value-validation`: 对 CreateGpuDemandOrder 请求中 extension 数组的值按 schema headers 定义进行类型和约束校验，复用 ValidateCellValue

### Modified Capabilities

## Impact

- **代码**：`pkg/tools/excel/validate.go`（新增 ValidateExtension）、`cmd/woa-server/logics/plan/gpu_demand_order.go`（CreateGpuDemandOrder 增加校验步骤）
- **API**：CreateGpuDemandOrder 接口行为变更 — extension 值不合规时返回错误拒绝创建
- **服务层**：仅影响 woa-server
