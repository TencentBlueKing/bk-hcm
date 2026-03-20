## MODIFIED Requirements

### Requirement: raw_data中的值按列类型转换
raw_data中每个值 SHALL 按对应header的type定义进行类型转换：int类型转为整数、float类型转为浮点数、enum类型根据value列表推断目标类型进行转换。转换失败时保留原始字符串值。构建raw_data的同时 SHALL 对每个可见列执行基于schema定义的类型和约束校验，校验错误以中文描述字符串追加到该行detail的validate_result数组中。

#### Scenario: int类型列值正确转换
- **GIVEN** 某列type为"int"，excel中值为"2026"
- **WHEN** 系统构建raw_data
- **THEN** raw_data中该位置值为整数2026，validate_result不包含该列的错误

#### Scenario: 类型转换失败保留原始字符串并追加校验错误
- **GIVEN** 某列type为"int"，name为"预算卡数"，excel中值为"12.5"
- **WHEN** 系统构建raw_data
- **THEN** raw_data中该位置保留原始字符串"12.5"，validate_result追加 `"预算卡数: 必须为整数"`
