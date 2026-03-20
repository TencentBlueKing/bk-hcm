## ADDED Requirements

### Requirement: Detail响应类型使用name+raw_data格式
GPU需求Excel导入预览接口的detail响应 SHALL 使用通用格式：Name（sheet名称）、RawData（原始行数据数组）和ValidateResult（校验结果）。不再使用DemandType、DemandYear、DemandMonth、Extension等结构化字段。

#### Scenario: detail包含正确的name字段
- **GIVEN** 解析"混元精调"sheet中的一行数据
- **WHEN** 系统构建该行的detail
- **THEN** detail的name字段值为"混元精调"

#### Scenario: detail的raw_data不包含hidden列的值
- **GIVEN** 某sheet的fix_headers中"预算卡数"列定义为hidden:true
- **WHEN** 系统构建该行的detail
- **THEN** raw_data中不包含"预算卡数"列的值

#### Scenario: detail的raw_data不包含公式计算列的值
- **GIVEN** 某sheet的fix_headers中"QPM峰值"列定义为field:"-"（公式计算列）
- **WHEN** 系统构建该行的detail
- **THEN** raw_data中不包含"QPM峰值"列的值

### Requirement: raw_data按fix_headers+headers中可见列顺序排列
raw_data数组 SHALL 按fix_headers中可见列在前、headers中可见列在后的顺序排列。可见列定义为field不为"-"且hidden不为true的列。

#### Scenario: 混合fix_headers和headers的raw_data排列顺序
- **GIVEN** Sheet定义fix_headers=[{name:"年份",field:"A",hidden:false},{name:"月份",field:"B",hidden:false},{name:"预算卡数",field:"C",hidden:true}]，headers=[{name:"使用场景",field:"D",hidden:false},{name:"卡型",field:"E",hidden:false}]
- **WHEN** excel数据行A列=2026、B列="9月"、C列=12、D列="文生图"、E列="H20"
- **THEN** raw_data为[2026, "9月", "文生图", "H20"]，不包含hidden的"预算卡数"值

### Requirement: raw_data中的值按列类型转换
raw_data中每个值 SHALL 按对应header的type定义进行类型转换：int类型转为整数、float类型转为浮点数、string和enum类型保持字符串。转换失败时保留原始字符串值。

#### Scenario: int类型列值正确转换
- **GIVEN** 某列type为"int"，excel中值为"2026"
- **WHEN** 系统构建raw_data
- **THEN** raw_data中该位置值为整数2026

#### Scenario: float类型列值正确转换
- **GIVEN** 某列type为"float"，excel中值为"12.5"
- **WHEN** 系统构建raw_data
- **THEN** raw_data中该位置值为浮点数12.5

#### Scenario: 类型转换失败保留原始字符串
- **GIVEN** 某列type为"int"，excel中值为"abc"
- **WHEN** 系统构建raw_data
- **THEN** raw_data中该位置保留原始字符串"abc"
