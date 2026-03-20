## Why

程序生成的 xlsx 文件中，公式单元格只有 `<f>` 元素（公式定义）而没有 `<v>` 元素（缓存计算结果）。excelize 库的 `rows.Columns()` 读取的是缓存值，当缓存缺失时返回空字符串，导致 GPU Excel 导入接口对公式列报出 "必填项不能为空" 的误判。用户必须手动用 Excel/WPS 重新保存后才能成功导入，体验不佳。

## What Changes

- 在 `ParseSheetRowsAndFormulas` 的行解析流程中，对 schema 定义了 Formula 的列，当缓存值为空时，**先校验公式与 schema 一致**，再通过 `excelize.CalcCellValue` 补算缺失值，填入行数据
- 调整公式校验与值补算的执行顺序：公式校验前置，确保只有公式正确的单元格才会被补算，避免错误公式产生错误计算结果
- `CalcCellValue` 补算失败时不做特殊跳过，值仍为空，照常走 required 等值校验并报错

## Capabilities

### New Capabilities
- `excel-formula-value-fallback`: 当公式单元格缺少缓存值时，先校验公式正确性，再通过 CalcCellValue 补算缺失值；补算失败时值保持为空，照常走后续校验报错

### Modified Capabilities
- `gpu-demand-excel-import`: 行解析流程调整——公式校验前置，补算后的值参与后续数据解析和校验

## Impact

- **代码**: `pkg/tools/excel/reader.go` 的 `ParseSheetRowsAndFormulas` 函数逻辑调整
- **代码**: `pkg/tools/excel/validate.go` 无需改动，补算成功后值非空正常校验，补算失败值仍为空照常报错
- **依赖**: 无新增外部依赖，`CalcCellValue` 是 excelize v2 内置方法
- **API**: 接口入参出参不变，仅内部解析行为变更
- **影响层**: 仅 pkg 工具层（excel 包），不涉及 service/data-service 层
