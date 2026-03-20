## 1. ValidateTypedValue 和 ValidateExtension

- [x] 1.1 在 `pkg/tools/excel/validate.go` 新增 `ValidateTypedValue(val interface{}, header Header) []string`，直接对 interface{} 做类型校验：nil/空字符串按 required 校验；int 检查 float64 整数判定 + 范围；float 检查 float64；enum 直接与 value 列表比较；string 检查类型 + 长度
- [x] 1.2 在 `pkg/tools/excel/validate.go` 新增 `ValidateExtension(values []interface{}, headers []Header) []string`，按索引逐个调用 ValidateTypedValue

## 2. Controller 层集成

- [x] 2.1 修改 `cmd/woa-server/logics/plan/gpu_demand_order.go` 中的 `CreateGpuDemandOrder`，在创建主单前调用 `getLatestGpuTplSchema` 获取 schema
- [x] 2.2 新增 `validateOrderExtensions` 函数，遍历 req.Details，按 demand_type 匹配 sheet，取 headers 调用 `ValidateExtension` 校验 extension，返回汇总错误
- [x] 2.3 在 CreateGpuDemandOrder 中调用 validateOrderExtensions，校验失败返回错误

## 3. 单元测试

- [x] 3.1 在 `pkg/tools/excel/validate_test.go` 新增 TestValidateTypedValue 和 TestValidateExtension 测试用例
