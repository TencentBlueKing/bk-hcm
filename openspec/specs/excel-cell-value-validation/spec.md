# excel-cell-value-validation

## Purpose

Defines the schema-driven cell value validation rules applied during Excel import, covering type checking (int, float, enum, string), range/length constraints, required field enforcement, and multi-column error accumulation.

## Requirements

### Requirement: Header结构体扩展Min/Max约束字段
Header结构体 SHALL 新增 `Min` 和 `Max` 字段（类型为 `*int64`，JSON tag 为 `min` 和 `max`，omitempty）。语义随 `type` 不同而不同：当 type 为 `string` 时，min/max 代表字符串最小/最大长度；当 type 为 `int` 时，min/max 代表数值范围。

#### Scenario: Header中定义了min和max字段
- **GIVEN** tpl_schema JSON 中某列定义为 `{"name":"数量","type":"int","field":"C","min":0,"max":1000}`
- **WHEN** 系统将 JSON 反序列化为 Header 结构体
- **THEN** Header 的 Min 值为 0，Max 值为 1000

#### Scenario: Header中未定义min和max字段
- **GIVEN** tpl_schema JSON 中某列定义为 `{"name":"备注","type":"string","field":"D"}`
- **WHEN** 系统将 JSON 反序列化为 Header 结构体
- **THEN** Header 的 Min 和 Max 均为 nil，不参与校验

### Requirement: int类型值校验
当 header type 为 `int` 时，系统 SHALL 校验单元格值是否为合法整数。若值包含小数点或不是合法整数格式，SHALL 追加中文错误描述。

#### Scenario: int类型列填入整数值
- **GIVEN** 某列 type 为 "int"，excel 中值为 "2026"
- **WHEN** 系统校验该单元格值
- **THEN** 校验通过，validate_result 不追加错误

#### Scenario: int类型列填入带小数点的值
- **GIVEN** 某列 type 为 "int"，name 为 "预算卡数"，excel 中值为 "12.5"
- **WHEN** 系统校验该单元格值
- **THEN** validate_result 追加 `"预算卡数: 必须为整数"`

#### Scenario: int类型列填入非数值字符串
- **GIVEN** 某列 type 为 "int"，name 为 "单次训练数据量"，excel 中值为 "abc"
- **WHEN** 系统校验该单元格值
- **THEN** validate_result 追加 `"单次训练数据量: 必须为整数"`

#### Scenario: int类型值范围校验
- **GIVEN** 某列 type 为 "int"，name 为 "预算卡数"，min 为 0，max 为 1000，excel 中值为 "1500"
- **WHEN** 系统校验该单元格值
- **THEN** validate_result 追加 `"预算卡数: 值1500超出范围[0, 1000]"`

#### Scenario: int类型空值且非required
- **GIVEN** 某列 type 为 "int"，required 为 false，excel 中值为空
- **WHEN** 系统校验该单元格值
- **THEN** 校验通过，不追加错误

### Requirement: float类型值校验
当 header type 为 `float` 时，系统 SHALL 校验单元格值是否为合法数值（整数或浮点数均可）。若不是合法数值，SHALL 追加中文错误描述。

#### Scenario: float类型列填入合法浮点数
- **GIVEN** 某列 type 为 "float"，excel 中值为 "12.5"
- **WHEN** 系统校验该单元格值
- **THEN** 校验通过

#### Scenario: float类型列填入合法整数
- **GIVEN** 某列 type 为 "float"，excel 中值为 "100"
- **WHEN** 系统校验该单元格值
- **THEN** 校验通过

#### Scenario: float类型列填入非数值
- **GIVEN** 某列 type 为 "float"，name 为 "实际使用TPM"，excel 中值为 "abc"
- **WHEN** 系统校验该单元格值
- **THEN** validate_result 追加 `"实际使用TPM: 必须为数字"`

### Requirement: enum类型值校验
当 header type 为 `enum` 时，系统 SHALL 根据 value 列表的第一个元素推断目标类型（JSON反序列化后 float64 为数值型，string 为字符串型）。校验分两步：
1. **类型匹配校验**：若 value 列表为数值型，用户输入 SHALL 能被解析为数值（ParseInt 或 ParseFloat），否则追加类型不匹配错误。字符串型则无需类型前置校验。
2. **值匹配校验**：类型匹配通过后，SHALL 校验转换后的值是否存在于 value 列表中，不存在则追加不在允许范围错误。

#### Scenario: enum值为字符串列表且匹配
- **GIVEN** 某列 type 为 "enum"，value 为 ["H20", "L20"]，excel 中值为 "H20"
- **WHEN** 系统校验该单元格值
- **THEN** 校验通过

#### Scenario: enum值为字符串列表且不匹配
- **GIVEN** 某列 type 为 "enum"，name 为 "卡型"，value 为 ["H20", "L20"]，excel 中值为 "A100"
- **WHEN** 系统校验该单元格值
- **THEN** validate_result 追加 `"卡型: 值'A100'不在允许范围[H20, L20]内"`

#### Scenario: enum值为整型列表且匹配
- **GIVEN** 某列 type 为 "enum"，value 为 [2026, 2027, 2028]（JSON反序列化后为 float64），excel 中值为 "2026"
- **WHEN** 系统将 "2026" 转换为数值后与 value 列表比较
- **THEN** 校验通过

#### Scenario: enum值为整型列表且不匹配
- **GIVEN** 某列 type 为 "enum"，name 为 "年份"，value 为 [2026, 2027, 2028]，excel 中值为 "2025"
- **WHEN** 系统校验该单元格值
- **THEN** validate_result 追加 `"年份: 值'2025'不在允许范围[2026, 2027, 2028]内"`

#### Scenario: enum值为数值列表但用户填入非数值字符串
- **GIVEN** 某列 type 为 "enum"，name 为 "月份"，value 为 [1,2,3,4,5,6,7,8,9,10,11,12]（数值型），excel 中值为 "一月"
- **WHEN** 系统尝试将 "一月" 解析为数值
- **THEN** 解析失败，validate_result 追加 `"月份: 值'一月'类型不匹配，应为数字"`

#### Scenario: enum空值且非required
- **GIVEN** 某列 type 为 "enum"，required 为 false，excel 中值为空
- **WHEN** 系统校验该单元格值
- **THEN** 校验通过

### Requirement: string类型长度校验
当 header type 为 `string` 且定义了 min/max 时，系统 SHALL 将 min/max 解释为字符串的最小/最大长度。长度按中文字符计算（即 Go 中的 rune 计数，使用 `utf8.RuneCountInString`），每个中文字符计为1，每个英文字符也计为1。

#### Scenario: 中文字符串长度在范围内
- **GIVEN** 某列 type 为 "string"，min 为 1，max 为 10，excel 中值为 "文生图场景"（rune长度为6）
- **WHEN** 系统校验该单元格值
- **THEN** 校验通过

#### Scenario: 中文字符串长度超过max
- **GIVEN** 某列 type 为 "string"，name 为 "使用场景"，max 为 5，excel 中值为 "这是一个超长的描述"（rune长度为9）
- **WHEN** 系统校验该单元格值
- **THEN** validate_result 追加 `"使用场景: 长度9超过最大长度5"`

#### Scenario: 字符串长度小于min
- **GIVEN** 某列 type 为 "string"，name 为 "业务逻辑"，min 为 10，excel 中值为 "短"（rune长度为1）
- **WHEN** 系统校验该单元格值
- **THEN** validate_result 追加 `"业务逻辑: 长度1小于最小长度10"`

#### Scenario: 中英混合字符串长度计算
- **GIVEN** 某列 type 为 "string"，max 为 10，excel 中值为 "H20卡型说明abc"（rune长度为10）
- **WHEN** 系统校验该单元格值
- **THEN** 校验通过，rune长度10等于max，在范围内

#### Scenario: string类型空值且非required不校验长度
- **GIVEN** 某列 type 为 "string"，min 为 1，required 为 false，excel 中值为空
- **WHEN** 系统校验该单元格值
- **THEN** 校验通过，不进行长度校验

### Requirement: required必填校验
当 header 的 required 为 true 时，系统 SHALL 校验单元格值不为空（TrimSpace后）。此校验在类型校验之前执行，required 失败时跳过后续校验。

#### Scenario: required字段值为空
- **GIVEN** 某列 required 为 true，name 为 "预算卡数"，excel 中值为空或仅空格
- **WHEN** 系统校验该单元格值
- **THEN** validate_result 追加 `"预算卡数: 必填项不能为空"`

#### Scenario: required字段值非空
- **GIVEN** 某列 required 为 true，excel 中值为 "12"
- **WHEN** 系统校验该单元格值
- **THEN** required 校验通过，继续执行类型校验

### Requirement: 单行多列校验错误
同一行的多个列可能各自产生校验错误，所有错误 SHALL 追加到该行 detail 的 validate_result 数组中。

#### Scenario: 单行多列校验错误
- **GIVEN** 某行的"预算卡数"列值为 "12.5"（int类型错误），"卡型"列值为 "A100"（enum不匹配）
- **WHEN** 系统逐列校验该行
- **THEN** validate_result 为 `["预算卡数: 必须为整数", "卡型: 值'A100'不在允许范围[H20, L20]内"]`
