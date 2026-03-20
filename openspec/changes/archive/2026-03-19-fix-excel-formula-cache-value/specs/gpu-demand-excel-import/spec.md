## MODIFIED Requirements

### Requirement: 数据行解析
系统 SHALL 按tpl_schema中每个sheet定义的start行号开始读取数据行，将每行数据构建为一条detail记录。每条detail包含demand_type（sheet名称）、demand_year（当前年份）、extension（按headers定义的列顺序提取的值数组）。对于 schema 中定义了 Formula 的列，当缓存值为空时，系统 SHALL 先执行公式校验，校验通过后调用 CalcCellValue 补算值填入行数据，使补算后的值参与后续校验和前端预览展示。

#### Scenario: 按start行号解析数据
- **GIVEN** tpl_schema定义某sheet的start为2
- **WHEN** 系统解析该sheet的数据
- **THEN** 系统从Excel的第2行开始读取数据（第1行为列头），每行生成一条detail记录

#### Scenario: start行号大于2的sheet
- **GIVEN** tpl_schema定义某sheet的start为3（表示第1-2行为表头区域）
- **WHEN** 系统解析该sheet的数据
- **THEN** 系统从Excel的第3行开始读取数据，跳过前2行

#### Scenario: sheet中无数据行
- **GIVEN** 某个sheet在start行号之后没有任何数据行
- **WHEN** 系统解析该sheet的数据
- **THEN** 该sheet不产生任何detail记录，不报错

#### Scenario: 数据行中存在空行
- **GIVEN** 数据区域中存在完全空白的行
- **WHEN** 系统解析该行数据
- **THEN** 系统跳过空行，不生成对应的detail记录

#### Scenario: 公式列缓存缺失的数据行正常解析
- **GIVEN** 某数据行的公式列（如 C 列）缓存值为空，但公式正确且 CalcCellValue 成功补算
- **WHEN** 系统解析该行数据
- **THEN** 该行的 detail 记录中，公式列的值为 CalcCellValue 的计算结果（非空），rawData 展示补算后的值，validate_result 不因该列报 required 错误

#### Scenario: 公式列缓存缺失且补算失败的数据行
- **GIVEN** 某数据行的公式列缓存值为空，公式正确但 CalcCellValue 补算失败
- **WHEN** 系统解析该行数据
- **THEN** 该行的 detail 记录中，公式列的值为空字符串，validate_result 中包含该列的 "必填项不能为空" 错误
