## 1. 重构 validateRowFormulas 为 validateAndFillFormulas

- [x] 1.1 将 `validateRowFormulas` 替换为 `validateAndFillFormulas`，单次循环内完成公式校验和缓存值补算，公式校验通过且缓存值为空时调用 `CalcCellValue` 补算并写入 row
- [x] 1.2 更新 `formula_test.go` 中所有 `validateRowFormulas` 测试用例适配 `validateAndFillFormulas` 新签名

## 2. 集成到 ParseSheetRowsAndFormulas

- [x] 2.1 `validateAndFillFormulas` 已合并校验和补算逻辑，无需单独的 `fillMissingFormulaValues` 函数
- [x] 2.2 调整 `ParseSheetRowsAndFormulas` 执行顺序：`extractFieldValues` → `validateAndFillFormulas`（公式校验+补算）→ append，移除 `formulaHeaders` 预过滤

## 3. 单元测试

- [x] 3.1 `TestValidateAndFillFormulas_FillMissingValue`：无缓存值+公式正确 → CalcCellValue 补算成功
- [x] 3.2 `TestValidateAndFillFormulas_NoFillWhenMismatch`：公式被篡改 → 不补算，值保持为空
- [x] 3.3 `TestValidateAndFillFormulas_NoFillWhenCached`：有缓存值 → 不触发补算，保留原值
- [x] 3.4 `TestValidateAndFillFormulas_FillRoundUp`：ROUNDUP 公式端到端补算验证
