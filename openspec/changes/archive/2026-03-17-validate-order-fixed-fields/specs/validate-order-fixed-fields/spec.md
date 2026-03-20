## ADDED Requirements

### Requirement: ValidateFixedFields 校验函数

系统 SHALL 在 `pkg/tools/excel/validate.go` 中提供 `ValidateFixedFields(values map[string]interface{}, headers []Header) []string` 函数，遍历 headers 列表，按 `DBField` 从 values map 中查找对应值，调用 `ValidateTypedValue` 进行校验，返回所有错误信息列表。

当 Header 的 `DBField` 为空时，SHALL 跳过该 header 不校验。
当 values map 中不存在对应 key 时，SHALL 视为空值，由 `ValidateTypedValue` 的 required 逻辑处理。

#### Scenario: 固定字段枚举值校验通过

- **GIVEN** schema fixed_header 定义 demand_year 为 enum 类型，value 为 [2026, 2027, 2028]
- **WHEN** values map 中 demand_year 为 2026
- **THEN** 返回空错误列表

#### Scenario: 固定字段枚举值校验失败

- **GIVEN** schema fixed_header 定义 demand_year 为 enum 类型，value 为 [2026, 2027, 2028]
- **WHEN** values map 中 demand_year 为 2025
- **THEN** 返回包含 "需求年份: 值'2025'不在允许范围[2026, 2027, 2028]内" 的错误信息

#### Scenario: 固定字段范围校验失败

- **GIVEN** schema fixed_header 定义 gpu_num 为 int 类型，min 为 0
- **WHEN** values map 中 gpu_num 为 -1
- **THEN** 返回包含 "GPU预算卡数: 值-1不能小于0" 的错误信息

#### Scenario: db_field 为空时跳过校验

- **GIVEN** schema fixed_header 的 DBField 为空字符串
- **WHEN** 执行校验
- **THEN** 跳过该 header，不产生任何错误

#### Scenario: 非必填固定字段为空时跳过校验

- **GIVEN** schema fixed_header 定义某字段 required 为 false
- **WHEN** values map 中该字段不存在或为 nil
- **THEN** 跳过该字段校验，不产生错误

### Requirement: 创建接口校验固定字段

系统 SHALL 在 CreateGpuDemandOrder 流程中，对每个 detail 的固定字段按 schema 对应 sheet 的 fixed_headers 进行校验。校验逻辑为：将 detail 通过 JSON 序列化转为 `map[string]interface{}`，然后调用 `ValidateFixedFields`。

校验失败时 SHALL 返回错误，阻止创建。

#### Scenario: 创建接口固定字段校验通过

- **GIVEN** 模版 schema 定义 demand_year 枚举值为 [2026, 2027, 2028]
- **WHEN** 用户提交创建请求，detail 的 demand_year 为 2026
- **THEN** 固定字段校验通过，继续后续创建流程

#### Scenario: 创建接口固定字段校验失败

- **GIVEN** 模版 schema 定义 demand_year 枚举值为 [2026, 2027, 2028]
- **WHEN** 用户提交创建请求，detail 的 demand_year 为 2025
- **THEN** 返回校验错误，不创建任何记录

### Requirement: 覆盖接口校验固定字段

系统 SHALL 在 OverwriteGpuDemandOrder 流程中，对每个 detail 的固定字段按 schema 对应 sheet 的 fixed_headers 进行校验，逻辑与创建接口一致。

#### Scenario: 覆盖接口固定字段校验通过

- **GIVEN** 模版 schema 定义 demand_year 枚举值为 [2026, 2027, 2028]
- **WHEN** 用户提交覆盖请求，detail 的 demand_year 为 2027
- **THEN** 固定字段校验通过，继续后续覆盖流程

#### Scenario: 覆盖接口固定字段校验失败

- **GIVEN** 模版 schema 定义 demand_month 枚举值为 [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12]
- **WHEN** 用户提交覆盖请求，detail 的 demand_month 为 13
- **THEN** 返回校验错误，不执行覆盖操作
