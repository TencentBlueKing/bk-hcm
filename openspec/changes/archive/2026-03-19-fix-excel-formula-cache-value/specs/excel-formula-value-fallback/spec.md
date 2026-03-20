## ADDED Requirements

### Requirement: 公式列缓存值缺失时补算
当 `ParseSheetRowsAndFormulas` 解析数据行时，对 schema 中定义了 Formula 的列，若 `rows.Columns()` 返回的缓存值为空（TrimSpace 后），系统 SHALL 先校验该单元格的实际公式与 schema 定义一致，校验通过后调用 `excelize.CalcCellValue` 计算公式结果并填入行数据。

#### Scenario: 公式正确且缓存缺失时成功补算
- **GIVEN** Excel 文件中 C5 单元格有公式 `ROUNDUP(M5*1000000000/Q5/3600/P5,0)` 且无缓存值（`<v>` 元素缺失），schema 中该列 Formula 定义为 `ROUNDUP(M5*1000000000/Q5/3600/P5,0)`
- **WHEN** 系统解析该行数据
- **THEN** 系统检测到缓存值为空，校验公式与 schema 一致后，调用 CalcCellValue 计算得到结果（如 "28"），填入行数据的对应位置

#### Scenario: 公式正确且缓存已存在时不触发补算
- **GIVEN** Excel 文件中 C5 单元格有公式且缓存值为 "28"（`<v>` 元素存在）
- **WHEN** 系统解析该行数据
- **THEN** 系统直接使用缓存值 "28"，不调用 CalcCellValue

#### Scenario: 非公式列不触发补算
- **GIVEN** Excel 文件中 A5 单元格无公式，schema 中该列 Formula 为空
- **WHEN** 系统解析该行数据
- **THEN** 系统直接使用 `rows.Columns()` 返回的值，不触发任何补算逻辑

### Requirement: 公式校验必须前置于补算
系统 SHALL 确保公式正确性校验在 CalcCellValue 补算之前执行。只有公式校验通过（实际公式与 schema 定义一致）的单元格才会执行补算。

#### Scenario: 公式被篡改时不补算
- **GIVEN** Excel 文件中 C5 单元格的实际公式为 `=1+1`（与 schema 定义的 `ROUNDUP(M5*1000000000/Q5/3600/P5,0)` 不一致），缓存值为空
- **WHEN** 系统解析该行数据
- **THEN** 公式校验失败，系统不调用 CalcCellValue，C5 的值保持为空字符串，后续 required 校验照常报 "必填项不能为空"，公式校验错误也一并记录

#### Scenario: 公式校验通过后补算
- **GIVEN** Excel 文件中 C5 单元格的实际公式与 schema 定义一致，缓存值为空
- **WHEN** 系统执行公式校验
- **THEN** 公式校验通过，随后调用 CalcCellValue 补算该单元格的值

### Requirement: CalcCellValue 补算失败时照常报错
当 CalcCellValue 返回 error 或空值时，系统 SHALL 保持该单元格的值为空字符串，不做特殊处理，后续 required 等值校验照常执行并报错。

#### Scenario: CalcCellValue 计算失败
- **GIVEN** Excel 文件中 C5 单元格公式校验通过，但 CalcCellValue 因不支持的函数返回 error
- **WHEN** 系统尝试补算该单元格
- **THEN** 补算失败，C5 的值保持为空字符串，后续 ValidateCellValue 对 required 列报 "必填项不能为空"

#### Scenario: CalcCellValue 返回空值
- **GIVEN** Excel 文件中 C5 单元格公式校验通过，CalcCellValue 返回空字符串（如公式引用的单元格也为空）
- **WHEN** 系统尝试补算该单元格
- **THEN** C5 的值为空字符串，后续校验照常执行
