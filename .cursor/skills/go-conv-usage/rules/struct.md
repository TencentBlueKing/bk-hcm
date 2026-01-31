# 结构体解析函数

> **注意**：这些函数使用反射，性能较低，主要用于 Excel/CSV 解析等特殊场景。

## ParseFieldByType

**函数签名**
```go
func ParseFieldByType(field reflect.StructField, fieldValue reflect.Value, fieldName, rawData string) error
```

**使用场景**
- Excel/CSV 解析时按字段类型解析
- 动态结构体字段赋值

**正确示例**

```go
if err := converter.ParseFieldByType(field, fieldValue, title, rawData); err != nil {
    return &CellFault{
        Field:   title,
        Message: getParseErrorMessage(err),
    }
}

err := converter.ParseFieldByType(field, fieldValue, field.Name, inputValue)
```

**注意事项**
- 使用反射，性能较低，避免在循环中频繁使用
- 支持指针类型（自动处理）
- 支持实现 `json.Unmarshaler` 的类型
- 返回 `*ParseError`，包含详细错误信息

**内部实现说明**

`ParseFieldByType` 内部调用以下函数（不建议直接使用）：
- `ParseValue` - 解析值（支持 json.Unmarshaler）
- `ParseBasicValue` - 解析基本类型值

---

## FlattenStructToFields

**函数签名**
```go
func FlattenStructToFields(value reflect.Value) []reflect.StructField
```

**使用场景**
- 获取结构体的所有字段（含匿名字段）
- Excel/CSV 解析时获取字段列表

**正确示例**

```go
fields := converter.FlattenStructToFields(zeroValue)

fields := converter.FlattenStructToFields(structValue)

fields := converter.FlattenStructToFields(val)
```

**注意事项**
- 递归处理匿名字段
- 支持指针类型（自动解引用）
- 返回的字段顺序与定义顺序一致

---

## ParseError

**类型定义**
```go
type ParseError struct {
    FieldName string
    FieldType reflect.Type
    Value     string
    ErrorType ErrorType
    Cause     error
}

type ErrorType int

const (
    ErrorTypeUnknown ErrorType = iota
    ErrorTypeInvalidInteger
    ErrorTypeInvalidFloat
    ErrorTypeInvalidBoolean
    ErrorTypeInvalidUnsignedInteger
    ErrorTypeOutOfRange
    ErrorTypeNilPointer
    ErrorTypeUnsupportedType
    ErrorTypeInvalidJSON
)
```

**使用场景**
- 获取解析错误的详细信息
- 用于生成用户友好的错误提示

**示例**

```go
err := conv.ParseFieldByType(field, fieldValue, title, rawData)
if parseErr, ok := err.(*conv.ParseError); ok {
    switch parseErr.ErrorType {
    case conv.ErrorTypeInvalidInteger:
        // 处理整数解析错误
    case conv.ErrorTypeOutOfRange:
        // 处理范围溢出错误
    }
}
```
