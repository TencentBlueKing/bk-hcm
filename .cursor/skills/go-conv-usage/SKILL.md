---
name: go-conv-usage
description: converter 包使用规范完整指南。当需要了解类型转换函数的详细用法、使用场景、最佳实践时使用。
---

# converter 包使用规范

## 概述

`pkg/tools/converter` 提供类型转换工具函数，涵盖指针、切片、Map、字符串、数值和结构体解析。使用泛型，类型安全。

## 快速参考

| 场景 | 规则文件 |
|-----|---------|
| 指针转换 | [rules/pointer.md](rules/pointer.md) |
| 切片操作 | [rules/slice.md](rules/slice.md) |
| Map 操作 | [rules/map.md](rules/map.md) |
| 字符串操作 | [rules/string.md](rules/string.md) |
| 数值转换 | [rules/numeric.md](rules/numeric.md) |
| 结构体解析 | [rules/struct.md](rules/struct.md) |

## 函数分类速查表

| 分类 | 函数 | 用途 | 使用频率 |
|------|------|------|---------|
| **指针转换** | `ValToPtr` | 值转指针 | 常用 |
| | `PtrToVal` | 指针转值（nil→零值） | 常用 |
| | `SliceToPtr` | 切片值转指针切片 | 较少 |
| | `PtrToSlice` | 指针切片转值切片 | 较少 |
| | `IfNilF64` | float64 指针默认值 | 较少 |
| **切片操作** | `SliceToSet` | 切片转 set | 常用 |
| | `ArrayUnique` | 切片去重 | 常用 |
| | `FlattenSlice` | 二维切片展平 | 较少 |
| | `StringSliceToInt64Slice` | 字符串切片→int64 切片 | 常用 |
| | `Int64SliceToStringSlice` | int64 切片→字符串切片 | 常用 |
| | `IntSliceToStringSlice` | int 切片→字符串切片 | 较少 |
| | `SliceToInterfaceSlice` | 切片→interface{} 切片 | 常用 |
| | `InterfaceToInterfaceSlice` | interface{}→切片 | 较少 |
| | `InterfaceToStringSlice` | interface{}→字符串切片 | 较少 |
| | `StringToInt64Slice` | 字符串→int64 切片 | 常用 |
| **Map 操作** | `StructToMap` | 结构体转 map | 较少 |
| | `SliceToMap` | 切片转 map | 常用 |
| | `MapKeyToSlice` | map 键→切片 | 常用 |
| | `MapValueToSlice` | map 值→切片 | 常用 |
| | `MapToSlice` | map 通过函数转切片 | 较少 |
| **字符串操作** | `JoinQuotes` | 带引号拼接（SQL） | 常用 |
| | `InterfaceToString` | interface{}→string | 常用 |
| | `InterfaceToStringPtr` | interface{}→*string | 较少 |
| **数值转换** | `JoinNumerics` | 数值切片拼接 | 常用 |
| | `JoinNumericElem` | 按索引取值拼接 | 较少 |
| | `JoinNumeric` | 数值切片逗号拼接 | 常用 |
| | `InterfaceToFloat64` | interface{}→float64 | 常用 |
| | `InterfaceToFloat64Ptr` | interface{}→*float64 | 常用 |
| | `InterfaceToInt64` | interface{}→int64 | 常用 |
| | `InterfaceToInt32` | interface{}→int32 | 较少 |
| | `InterfaceToIntPtr` | interface{}→*int | 较少 |
| | `InterfaceToUint*` | interface{}→无符号整数 | 较少 |
| **结构体解析** | `ParseFieldByType` | 按类型解析字段 | 特殊场景 |
| | `FlattenStructToFields` | 展平结构体字段 | 特殊场景 |

## 核心原则

1. **错误处理**：返回 error 的函数必须检查错误
2. **nil 安全**：使用 `PtrToVal` 避免 nil panic
3. **类型安全**：优先使用泛型函数，避免 `interface{}`
4. **性能意识**：反射函数性能较低，避免在循环中使用
5. **SQL 安全**：使用 `JoinQuotes` 和 `JoinNumeric` 拼接 SQL

## 常见问题

### Q: 什么时候用 `ArrayUnique`，什么时候用 `SliceToSet`？
A: 需要去重后的切片用 `ArrayUnique`；需要 O(1) 查找或集合运算用 `SliceToSet`。

### Q: `MapToSlice` 和 `SliceToMap` 的区别？
A: `SliceToMap` 将切片转为 map，通过 kvFunc 提取 key-value 对；`MapToSlice` 将 map 转为切片，通过 mapFunc 将 key-value 转为切片元素。

### Q: 数值转换函数支持哪些格式的字符串？
A: 支持标准数值字符串（如 "123"）和千位分隔符格式（如 "1,234,567.89"），不支持科学计数法。

### Q: `SliceToMap` 遇到重复 key 会怎样？
A: 后出现的 key 会覆盖先前的值。如果需要严格去重，请在调用前对切片进行去重处理。
