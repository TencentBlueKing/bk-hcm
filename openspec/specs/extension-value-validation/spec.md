# Spec: Extension Value Validation

## Purpose

提供通用的 extension 字段校验能力，供 GPU 需求工单等场景在创建/导入时对 extension 数组按 schema headers 定义进行类型和合法性校验。

## Requirements

### Requirement: ValidateExtension 通用校验函数
系统 SHALL 在 `pkg/tools/excel/validate.go` 提供 `ValidateExtension(values []interface{}, headers []Header) []string` 函数。该函数将 values 数组与 headers 按索引一一对应，对每个值调用 `ValidateTypedValue` 直接按 Go 类型校验，汇总返回所有校验错误。

#### Scenario: extension 长度与 headers 长度一致且全部校验通过
- **GIVEN** extension 为 `["文生图", "H20", float64(12)]`，headers 为 3 个定义（string、enum、int）
- **WHEN** 调用 ValidateExtension
- **THEN** 返回空错误列表

#### Scenario: extension 中某个值类型不匹配
- **GIVEN** extension 为 `["文生图", "A100", float64(12)]`，headers[1] type 为 enum，value 为 ["H20", "L20"]
- **WHEN** 调用 ValidateExtension
- **THEN** 返回包含 headers[1].Name 相关的枚举校验错误

#### Scenario: extension 长度小于 headers 长度
- **GIVEN** extension 只有 2 个元素但 headers 有 3 个
- **WHEN** 调用 ValidateExtension
- **THEN** 第 3 个 header 按 nil 值处理，若 required 则报必填错误

### Requirement: ValidateTypedValue 直接类型校验
系统 SHALL 提供 `ValidateTypedValue(val interface{}, header Header) []string`，根据值的 Go 类型和 header.Type 直接校验，不做 string 转换。

#### Scenario: int 类型期望收到整数值的 float64
- **GIVEN** header.Type 为 "int"，val 为 float64(12)
- **WHEN** 系统检查 `v == float64(int64(v))`
- **THEN** 校验通过

#### Scenario: int 类型收到带小数的 float64
- **GIVEN** header.Type 为 "int"，name 为 "预算卡数"，val 为 float64(12.5)
- **WHEN** 系统检查整数判定
- **THEN** 校验失败，返回 `"预算卡数: 必须为整数"`

#### Scenario: int 类型收到字符串
- **GIVEN** header.Type 为 "int"，name 为 "预算卡数"，val 为 "abc"
- **WHEN** 系统检查值类型
- **THEN** 校验失败，返回 `"预算卡数: 必须为整数"`

#### Scenario: enum 类型直接与 value 列表比较
- **GIVEN** header.Type 为 "enum"，value 为 [float64(2026), float64(2027)]，val 为 float64(2026)
- **WHEN** 系统直接比较
- **THEN** 校验通过

#### Scenario: string 类型校验长度
- **GIVEN** header.Type 为 "string"，max 为 5，val 为 "超过五个字的描述"
- **WHEN** 系统校验 rune 长度
- **THEN** 返回长度超限错误

#### Scenario: nil 值且 required
- **GIVEN** header.Required 为 true，val 为 nil
- **WHEN** 调用 ValidateTypedValue
- **THEN** 返回 `"{name}: 必填项不能为空"`

### Requirement: CreateGpuDemandOrder 集成 extension 校验
CreateGpuDemandOrder controller 层 SHALL 在创建工单前获取最新 schema，对每条 detail 按 demand_type 匹配 sheet，取该 sheet 的 headers 校验 extension。任意 detail 校验失败 SHALL 返回错误，不创建工单。

#### Scenario: 所有 detail 的 extension 校验通过
- **GIVEN** 请求中所有 detail 的 extension 值符合对应 sheet headers 定义
- **WHEN** 调用 CreateGpuDemandOrder
- **THEN** 正常创建主单和子单

#### Scenario: 某条 detail 的 extension 校验失败
- **GIVEN** 请求中 details[1] 的 extension 某个值不符合 schema 定义
- **WHEN** 调用 CreateGpuDemandOrder
- **THEN** 返回错误，错误信息包含 detail 索引和具体校验错误，不创建任何工单

#### Scenario: demand_type 在 schema 中找不到对应 sheet
- **GIVEN** 某条 detail 的 demand_type 为 "不存在的sheet"
- **WHEN** 系统按 demand_type 匹配 sheet
- **THEN** 返回错误，提示找不到对应的 sheet 定义
