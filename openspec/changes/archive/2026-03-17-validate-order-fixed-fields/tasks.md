## 1. 校验工具函数

- [x] 1.1 在 `pkg/tools/excel/validate.go` 中新增 `ValidateFixedFields(values map[string]interface{}, headers []Header) []string` 函数，遍历 headers 按 DBField 从 map 取值后调用 ValidateTypedValue，DBField 为空时跳过

## 2. 业务校验集成

- [x] 2.1 在 `cmd/woa-server/logics/plan/gpu_demand_order.go` 中将 `validateOrderExtensions` 重命名为 `validateOrderDetails`，增加固定字段校验逻辑：将 detail 通过 json.Marshal/Unmarshal 转为 map，调用 ValidateFixedFields 校验 sheet.FixedHeaders
- [x] 2.2 更新 CreateGpuDemandOrder 和 OverwriteGpuDemandOrder 中对该函数的调用处
