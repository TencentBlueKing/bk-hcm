## Context

当前 `ParseSheetRowsAndFormulas` 的执行流程为：
1. `rows.Columns()` 读取行的所有缓存值
2. `extractFieldValues` 按 schema 列索引提取各列的值
3. `validateRowFormulas` 校验公式列的公式是否与 schema 一致
4. 上层 `buildDetails` 对提取的值逐列调用 `ValidateCellValue` 进行 required/类型/范围校验

问题出在第 1 步：`rows.Columns()` 读取的是 xlsx XML 中的 `<v>` 缓存值。程序生成的文件公式单元格可能没有 `<v>`，导致步骤 2 提取到空字符串，步骤 4 对 required 公式列误报 "必填项不能为空"。

现有代码中公式校验（步骤 3）与值提取（步骤 2）是独立的，公式校验结果没有被用来辅助值的补全。

## Goals / Non-Goals

**Goals:**
- 对缺少缓存值的公式单元格，在公式校验通过后通过 `CalcCellValue` 补算值，消除误报
- 确保公式校验前置于补算，只有公式正确才补算，避免错误公式产生错误计算结果
- 改动集中在 `reader.go` 的 `ParseSheetRowsAndFormulas`，不影响 validate.go 及上层调用方

**Non-Goals:**
- 不实现通用的 "重新保存" / "全量公式重算" 能力
- 不改变 `ValidateCellValue` 的校验逻辑，补算失败值为空时照常报错
- 不处理非 schema 定义的公式列（只处理 Header.Formula 非空的列）

## Decisions

### Decision 1: 在 `ParseSheetRowsAndFormulas` 内补算，而非在 `buildDetails` 校验时跳过

**选择**: 在 `ParseSheetRowsAndFormulas` 提取行数据后，对公式列缓存值为空的单元格调用 `CalcCellValue` 补算，将结果填入 row 数组。

**理由**: row 数组同时用于前端预览展示（rawData）和值校验（ValidateCellValue）。如果只在校验时跳过 required，前端预览仍会显示空值，用户体验不一致。在源头补算可以让后续所有消费方都拿到正确的值。

**替代方案**: 在 `ValidateCellValue` 中对 Formula 非空的 Header 跳过 required 校验。但这会导致前端预览显示空值，且 CalcCellValue 失败时无法照常报错。

### Decision 2: 公式校验前置于补算

**选择**: 先对该行所有公式列执行 `validateRowFormulas`，收集公式校验结果。只对公式校验通过（无 error）的列执行 `CalcCellValue` 补算。

**理由**: 用户要求必须先确认公式正确才能补算。如果用户篡改了公式（如改成 `=1+1`），补算会得到错误值，后续校验和入库的数据都是错的。公式校验前置可以阻断这种情况。

**实现**: `validateRowFormulas` 改为返回 per-header 的校验结果（而非只返回错误列表），以便调用方知道哪些公式列通过了校验。

### Decision 3: 使用 excelize 内置 `CalcCellValue` 而非外部引擎

**选择**: 调用 `excelize.File.CalcCellValue(sheetName, cellRef)` 补算。

**理由**: 零外部依赖；excelize v2 支持约 400 个公式函数，覆盖当前 schema 使用的 ROUNDUP 等常见函数；性能开销极小（仅对缺缓存的单元格调用）。

**替代方案**: LibreOffice headless 全量重算——完美兼容但需要安装额外软件、启动进程，对服务端部署要求过高。

## Risks / Trade-offs

- **[Risk] CalcCellValue 不支持某些公式函数** → 补算失败返回 error，值保持空字符串，照常走 required 校验报错。用户需手动用 Excel 重新保存。这是可接受的降级，因为当前 schema 使用的公式（ROUNDUP）均已支持。
- **[Risk] CalcCellValue 依赖其他单元格的值** → 公式引用的单元格（如 M5、Q5、P5）是用户输入列，有缓存值。只有公式列本身缺缓存，被引用的输入列不受影响。
- **[Trade-off] 每行公式列多一次 CalcCellValue 调用** → 仅对缓存值为空的单元格触发，正常文件（有缓存）不受影响，性能影响可忽略。
