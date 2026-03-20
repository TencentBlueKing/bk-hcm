## Why

GPU需求Excel导入预览接口的`detail`记录中，`demand_year`当前硬编码为`time.Now().Year()`，且缺少`demand_month`字段。根据接口文档（`docs/api-docs/web-server/docs/biz/scr/resource-plan/excel_import_gpu_demand.md`），这两个字段应从用户填写的Excel数据行中提取。Excel模版中固定存在name为`"年份"`和`"月份"`的列头，对应detail的`demand_year`和`demand_month`字段。

## What Changes

- 在`GpuDemandExcelImportDetail`结构体中新增`DemandMonth int json:"demand_month"`字段
- 重构`buildDetails`逻辑：遍历每行数据时，识别header name为`"年份"`和`"月份"`的列，提取值到demand_year和demand_month固定字段，其余列进入extension
- 有固定映射的列（年份/月份）不再出现在extension数组中
- Header结构体无需变更（通过name识别，不需要新字段）

## Capabilities

### New Capabilities
- `gpu-excel-fixed-field-extract`: 从Excel数据行中按固定列名提取demand_year和demand_month到detail的顶层字段

### Modified Capabilities

## Impact

- **pkg/api/woa-server/res_plan_gpu_excel_import.go**: GpuDemandExcelImportDetail新增`DemandMonth`
- **cmd/woa-server/logics/plan/gpu_excel_import.go**: `buildDetails`和`convertCellValue`逻辑调整
- **接口文档**: 已有`demand_month`定义，无需变更
