## 1. Header结构体扩展

- [x] 1.1 在 `pkg/tools/excel/schema.go` 的 Header 结构体中新增 `Min *int64` 和 `Max *int64` 字段，JSON tag 为 `min,omitempty` 和 `max,omitempty`，添加注释说明语义（string类型表示长度范围，int类型表示数值范围）

## 2. 校验逻辑实现

- [x] 2.1 新建 `pkg/tools/excel/validate.go`，定义中文错误模版常量，实现 `ValidateCellValue(val string, header Header) []string` 函数，返回中文错误描述列表
- [x] 2.2 实现 required 必填校验：TrimSpace后为空则返回 `"{name}: 必填项不能为空"`，required 失败时跳过后续校验
- [x] 2.3 实现 int 类型校验：使用 `strconv.ParseInt` 校验是否为合法整数，失败返回 `"{name}: 必须为整数"`
- [x] 2.4 实现 int 类型 min/max 数值范围校验：解析成功后检查值是否在 [min, max] 范围内，越界返回 `"{name}: 值{v}超出范围[{min}, {max}]"`
- [x] 2.5 实现 float 类型校验：使用 `strconv.ParseFloat` 校验是否为合法数值，失败返回 `"{name}: 必须为数字"`
- [x] 2.6 实现 enum 类型校验：根据 header.Value 第一个元素类型推断目标类型（float64→数值型，string→字符串型）；数值型枚举先校验用户输入能否解析为数值（不能则返回 `"{name}: 值'{val}'类型不匹配，应为数字"`），再校验值是否在列表中（不在则返回 `"{name}: 值'{val}'不在允许范围[...]内"`）
- [x] 2.7 实现 string 类型 min/max 长度校验：以 `utf8.RuneCountInString` 计算长度，超过 max 返回 `"{name}: 长度{n}超过最大长度{max}"`，小于 min 返回 `"{name}: 长度{n}小于最小长度{min}"`

## 3. 集成校验到buildDetails

- [x] 3.1 修改 `cmd/woa-server/logics/plan/gpu_demand_order.go` 中的 `buildDetails` 函数，在遍历可见列构建 raw_data 的同时调用 `ValidateCellValue`，将返回的错误追加到该行 detail 的 `ValidateResult`

## 4. 单元测试

- [x] 4.1 新建 `pkg/tools/excel/validate_test.go`，为 `ValidateCellValue` 编写单元测试，覆盖：int正常值、int小数值、int非数值、int范围越界、float正常值、float非数值、enum匹配/不匹配（字符串和数值枚举）、enum类型不匹配、string长度越界、required空值、空值非required跳过校验
