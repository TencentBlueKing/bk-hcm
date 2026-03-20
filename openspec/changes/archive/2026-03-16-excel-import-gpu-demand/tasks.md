## 1. 通用Excel Schema解析器 (pkg/tools/excel/)

- [x] 1.1 在 `pkg/tools/excel/schema.go` 中定义通用的 `Schema`、`Sheet`、`Header` 结构体，JSON tag与接口文档(`docs/api-docs/web-server/docs/biz/scr/resource-plan/excel_import_gpu_demand.md`)中tpl_schema结构一一对应，Header包含name、type、field、value、formula、readonly、required全部字段
- [x] 1.2 在 `pkg/tools/excel/reader.go` 中实现 `ValidateFileIntegrity(excelFile *excelize.File, schema *Schema) error`：先校验sheet名称/数量是否与schema一致（不一致返回"缺少sheet: X"或"多余sheet: X"的明确错误），再校验每个sheet的列头行（取第start-1行）是否与headers.name匹配（不匹配返回"sheet[X]列头不匹配, 缺少列: Y"）；任一校验失败直接返回error
- [x] 1.3 在 `reader.go` 中实现 `ParseSheetRows(excelFile *excelize.File, sheet *Sheet) ([][]string, error)`：按sheet.Start行号开始逐行读取数据，跳过空行，列数不足补空字符串、多余截断，返回原始字符串二维数组

## 2. API类型定义 (pkg/api/woa-server/)

- [x] 2.1 在 `pkg/api/woa-server/res_plan_gpu_excel_import.go` 中定义响应结构体，严格对齐接口文档：`GpuDemandExcelImportResp` 包含 `TplSchema *excel.Schema json:"tpl_schema"` 和 `Details []GpuDemandExcelImportDetail json:"details"`；`GpuDemandExcelImportDetail` 包含 `DemandType string json:"demand_type"`、`DemandYear int json:"demand_year"`、`Extension []interface{} json:"extension"`、`ValidateResult []string json:"validate_result"`

## 3. Excel导入Logic层 (cmd/woa-server/logics/plan/)

- [x] 3.1 在 `gpu_excel_import.go` 中实现 `getLatestGpuTplSchema`：调用data-service的ListDemandGpuTemplate接口，按created_at倒序取第一条，将TplSchema字段反序列化为 `excel.Schema`；无模版记录时返回错误
- [x] 3.2 重构 `buildDetails` 方法：接收 `sheet *excel.Sheet` 和原始行数据 `[][]string`，根据sheet.Headers中每列的Type定义做类型转换构建extension（int→int64, float(1)→float64, string/enum→string，转换失败保留原字符串）；DemandType取sheet.Name，DemandYear取当前年份，ValidateResult为空数组
- [x] 3.3 `ExcelImportGpuDemand` 主编排方法：获取模版tpl_schema → excelize.OpenReader打开Excel → ValidateFileIntegrity校验文件完整性（失败直接返回error） → 遍历每个sheet调用ParseSheetRows+buildDetails → 组装GpuDemandExcelImportResp返回

## 4. Handler层注册 (cmd/woa-server/service/plan/)

- [x] 4.1 在 `gpu_excel_import.go` 中实现handler：从Request.FormFile("file")获取上传文件，调用logic层ExcelImportGpuDemand，错误通过errf.NewFromErr(errf.InvalidParameter, err)返回确保code!=0
- [x] 4.2 在 `service.go` 的 `initBizPlanService` 中注册路由（路径与接口文档一致）：`h.Add("ExcelImportGpuDemand", http.MethodPost, "/plans/gpu/excel/import", s.ExcelImportGpuDemand)`，路由挂在bizH下自动拼接`/bizs/{bk_biz_id}`前缀

## 5. Logic层接口声明

- [x] 5.1 在 `cmd/woa-server/logics/plan/plan.go` 的Logics接口中声明 `ExcelImportGpuDemand(kt *kit.Kit, reader io.Reader) (*woaapi.GpuDemandExcelImportResp, error)`
