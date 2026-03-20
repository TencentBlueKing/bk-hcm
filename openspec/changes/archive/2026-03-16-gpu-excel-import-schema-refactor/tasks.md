## 1. Schema结构体重构

- [x] 1.1 修改 `pkg/tools/excel/schema.go`：Sheet结构体 `Start int` 重命名为 `RowStart int`，json tag 改为 `row_start`；新增 `FixHeaders []Header` 字段（json tag: `fix_headers`）
- [x] 1.2 修改 `pkg/tools/excel/schema.go`：Header结构体新增 `OrderKey string`（json tag: `order_key`）和 `Hidden bool`（json tag: `hidden`）字段
- [x] 1.3 修改 `pkg/tools/excel/schema.go`：新增 `Sheet.AllExcelHeaders()` 方法，返回 fix_headers + headers 中 field != "-" 的所有Header
- [x] 1.4 修改 `pkg/tools/excel/schema.go`：新增 `Sheet.AllVisibleHeaders()` 方法，返回 fix_headers + headers 中 field != "-" 且 hidden != true 的Header
- [x] 1.5 移除 `Sheet.HeaderNames()`（已无调用方），新增 `AllExcelHeaders()` 和 `AllVisibleHeaders()` 替代

## 2. Reader校验和解析逻辑适配

- [x] 2.1 修改 `pkg/tools/excel/reader.go`：`validateHeaders` 使用 `AllExcelHeaders()` 获取需要校验的列头，替代原来的 `sheet.HeaderNames()`
- [x] 2.2 修改 `pkg/tools/excel/reader.go`：`validateHeaders` 中 `sheet.Start` 替换为 `sheet.RowStart`
- [x] 2.3 修改 `pkg/tools/excel/reader.go`：`ParseSheetRows` 中 `sheet.Start` 替换为 `sheet.RowStart`
- [x] 2.4 修改 `pkg/tools/excel/reader.go`：`ParseSheetRows` 按header的field属性（A/B/C列号）映射读取excel列值；新增 `buildFieldIndices` 和 `extractFieldValues` 替代 `normalizeRow`
- [x] 2.5 移除 `normalizeRow`，解析逻辑完全基于field列索引映射

## 3. Detail响应类型重构

- [x] 3.1 修改 `pkg/api/woa-server/res_plan_gpu_excel_import.go`：`GpuDemandExcelImportDetail` 移除 `DemandType`、`DemandYear`、`DemandMonth`、`Extension` 字段，新增 `Name string`（json: `name`）和 `RawData []interface{}`（json: `raw_data`）

## 4. 导入逻辑重构

- [x] 4.1 修改 `cmd/woa-server/logics/plan/gpu_excel_import.go`：移除 `headerNameDemandYear`、`headerNameDemandMonth` 常量
- [x] 4.2 重写 `buildDetails` 函数：遍历 `AllExcelHeaders()` 跳过hidden列，构建 `Name` + `RawData` + `ValidateResult` 格式的detail
- [x] 4.3 确保 `ExcelImportGpuDemand` 主流程适配新的Schema字段名和detail格式

## 5. 日志和错误信息英文化

- [x] 5.1 检查 `cmd/woa-server/logics/plan/gpu_excel_import.go` 中所有日志和错误信息，确认均为英文
- [x] 5.2 检查 `pkg/tools/excel/reader.go` 中所有错误信息，确认均为英文
