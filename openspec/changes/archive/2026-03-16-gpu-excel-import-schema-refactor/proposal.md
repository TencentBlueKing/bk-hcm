## Why

GPU需求Excel导入预览接口的API文档（`excel_import_gpu_demand.md`）已进行重大调整，引入了`fix_headers`/`headers`分离、`row_start`、`order_key`、`hidden`等新概念，同时detail响应格式从结构化字段（`demand_type`/`demand_year`/`demand_month`/`extension`）变更为通用的`name` + `raw_data`格式。当前代码（包括Schema结构体、解析逻辑、响应类型）仍基于旧接口定义，需要全面对齐新文档。此外，之前两个相关变更（`excel-import-gpu-demand`和`gpu-excel-detail-fixed-fields`）中生成的代码日志和错误信息存在中文，需统一为英文。

## What Changes

- **Schema结构体重构**（`pkg/tools/excel/schema.go`）：
  - `Sheet.Start` 重命名为 `Sheet.RowStart`（json tag: `row_start`）
  - 新增 `Sheet.FixHeaders []Header`（json tag: `fix_headers`），与 `Sheet.Headers` 分离
  - `Header` 新增 `OrderKey string`（json tag: `order_key`）和 `Hidden bool`（json tag: `hidden`）字段
  - 调整辅助方法以同时覆盖 `FixHeaders` 和 `Headers`

- **Detail响应类型重构**（`pkg/api/woa-server/res_plan_gpu_excel_import.go`）：**BREAKING**
  - `GpuDemandExcelImportDetail` 移除 `DemandType`、`DemandYear`、`DemandMonth`、`Extension` 字段
  - 新增 `Name string`（对应sheet名称）和 `RawData []interface{}`（原始行数据）
  - `raw_data` 按 `fix_headers` + `headers` 中 `field != "-"` 且 `hidden != true` 的列顺序排列

- **解析逻辑重构**（`cmd/woa-server/logics/plan/gpu_excel_import.go`）：
  - `buildDetails` 不再提取固定字段到顶层，改为按新规则构建 `raw_data`
  - 移除 `headerNameDemandYear`、`headerNameDemandMonth` 常量
  - `convertCellValue` 适配新的列遍历逻辑

- **Reader校验逻辑更新**（`pkg/tools/excel/reader.go`）：
  - `validateHeaders` 同时校验 `fix_headers` 和 `headers` 中有实际excel列（`field != "-"`）的列头
  - `ParseSheetRows` 使用 `RowStart` 代替 `Start`，解析时考虑 `fix_headers` + `headers` 的列映射
  - 列数对齐逻辑根据有实际excel列的 header 数量计算

- **日志和错误信息英文化**：确保所有日志和返回的错误信息均为英文

## Capabilities

### New Capabilities
- `gpu-excel-rawdata-detail`: 基于raw_data格式的Excel导入预览detail构建能力，按fix_headers+headers中可见列顺序返回原始行数据

### Modified Capabilities
- `gpu-demand-excel-import`: 整体导入预览流程适配新Schema结构（fix_headers/headers分离、row_start、order_key、hidden等）和新detail响应格式（name + raw_data）

## Impact

- **pkg/tools/excel/schema.go**: Schema、Sheet、Header结构体重构，辅助方法调整
- **pkg/tools/excel/reader.go**: 校验和解析逻辑适配新结构
- **pkg/api/woa-server/res_plan_gpu_excel_import.go**: 响应类型breaking change
- **cmd/woa-server/logics/plan/gpu_excel_import.go**: 解析和构建逻辑重写
- **下游影响**: `create_resource_plan_gpu_demand_order` 和 `overwrite_resource_plan_gpu_demand_order` 接口的前端调用需适配新的detail格式（前端负责从raw_data中根据order_key提取结构化字段再提交创建/覆盖接口）
