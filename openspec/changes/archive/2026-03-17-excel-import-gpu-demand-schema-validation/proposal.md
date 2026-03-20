## Why

GPU需求Excel导入预览接口目前仅做结构校验（sheet匹配、列头匹配）和值类型转换，但不对单元格值进行基于schema定义的类型和约束校验。用户填入不合规数据（如int字段填小数、enum字段填非法值、字符串超长等）时，前端无法在预览阶段提示具体错误，需要在导入预览阶段根据schema中headers和fixed_headers的定义进行基础类型校验，将校验错误写入每行detail的validate_result数组。

## What Changes

- 新增单元格值校验逻辑：在`buildDetails`构建detail时，对每个可见列的值按schema定义的type进行校验
  - `int`类型：值必须为整数，不能包含小数点；`min`/`max`代表数值范围
  - `float`类型：值必须为合法数值
  - `enum`类型：值必须在value列表中，区分字符串/整型/浮点型枚举值
  - `string`类型：`min`/`max`代表最小长度和最大长度
  - 通用：`required`字段为true时，值不能为空
- Header结构体扩展：新增`Min`、`Max`（`*int64`）约束字段，语义随type不同：string类型表示最小/最大长度，int类型表示数值范围
- `validate_result` 保持 `[]string` 不变，直接返回中文错误描述，格式为 `"{列名}: {中文错误描述}"`
- 中文错误模版定义在 `pkg/tools/excel/validate.go` 中作为常量集中管理

## Capabilities

### New Capabilities
- `excel-cell-value-validation`: 基于schema header定义的type/value/required/min/max等约束，对Excel单元格值进行基础类型和范围校验，错误写入detail的validate_result

### Modified Capabilities
- `gpu-excel-rawdata-detail`: validate_result从始终为空数组变为包含具体校验错误描述

## Impact

- **代码**：`pkg/tools/excel/schema.go`（Header结构体扩展）、`cmd/woa-server/logics/plan/gpu_demand_order.go`（buildDetails增加校验逻辑）、可能新增`pkg/tools/excel/validator.go`
- **API**：响应结构不变，validate_result字段语义增强（从空数组变为可能包含校验错误）
- **服务层**：仅影响woa-server服务层
