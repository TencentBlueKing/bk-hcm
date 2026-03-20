## 1. API类型更新

- [x] 1.1 在 `pkg/api/woa-server/res_plan_gpu_excel_import.go` 的 `GpuDemandExcelImportDetail` 中新增 `DemandMonth int json:"demand_month"` 字段

## 2. buildDetails 逻辑重构

- [x] 2.1 在 `cmd/woa-server/logics/plan/gpu_excel_import.go` 中定义包级常量 `headerNameDemandYear = "年份"` 和 `headerNameDemandMonth = "月份"`
- [x] 2.2 重构 `buildDetails` 方法：遍历每行数据时，根据 `header.Name` 判断——匹配 `"年份"` 则通过 `strconv.Atoi` 转换为 demand_year（转换失败取0），匹配 `"月份"` 则转换为 demand_month（转换失败取0），其余列经 `convertCellValue` 类型转换后追加到 extension；删除原有的 `time.Now().Year()` 硬编码
