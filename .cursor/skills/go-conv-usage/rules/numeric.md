# 数值转换函数

## JoinNumeric / JoinNumerics

**函数签名**
```go
func JoinNumeric[T base.Numeric](source []T) string
func JoinNumerics[T base.Numeric](elems []T, sep string) string
```

**使用场景**
- 数值切片拼接为字符串（用于 SQL IN 子句）

**正确示例**

```go
whereConds = append(whereConds, fmt.Sprintf("op_product_id IN (%s)", converter.JoinNumeric(opProdIDs)))

fmt.Sprintf("platform_category_id IN (%s)", converter.JoinNumerics(opt.PlatformCategoryIDs, ","))

bizCond := fmt.Sprintf("biz_id IN (%s)", converter.JoinNumeric(bizIDs))
```

**注意事项**
- `JoinNumeric` 使用逗号分隔
- `JoinNumerics` 可自定义分隔符
- 类型需实现 `base.Numeric` 接口

---

## JoinNumericElem

**函数签名**
```go
func JoinNumericElem[T base.Numeric](length int, value func(index int) T, sep string) string
```

**使用场景**
- 按索引从数据结构中取值并拼接
- 特殊场景：需要从复杂结构按索引提取数值

**注意事项**
- `value` 函数需自行保证不越界，否则会 panic
- 生产代码中较少使用
- 一般场景推荐直接构造切片后使用 `JoinNumerics`

---

## InterfaceToFloat64 / InterfaceToFloat64Ptr

**函数签名**
```go
func InterfaceToFloat64(in interface{}) (float64, error)
func InterfaceToFloat64Ptr(in interface{}) *float64
```

**使用场景**
- 将 interface{} 转为 float64
- 处理 Elasticsearch 聚合结果中的数值

**正确示例**

```go
totalAmount, err := converter.InterfaceToFloat64(bucket.TotalAmountSum.Value)
substandardAmount, err := converter.InterfaceToFloat64(bucket.SubstandardAmountSum.Value)
validLogicAmount, err := converter.InterfaceToFloat64(bucket.ValidLogicAmountSum.Value)
dailyPeakGPUUsage := converter.InterfaceToFloat64Ptr(bucket.DailyPeakGPUWeightUsageSum.Value)
dailyAvgGPUUsage := converter.InterfaceToFloat64Ptr(bucket.DailyAvgGPUWeightUsageSum.Value)

metricValue, err = converter.InterfaceToFloat64(bucket.TotalDevice.Value)
metricValue, err = converter.InterfaceToFloat64(bucket.TotalMem.Value)
metricValue, err = converter.InterfaceToFloat64(bucket.TotalCore.Value)
```

**注意事项**
- 支持数值类型和字符串（含千位分隔符）
- NaN 会返回错误
- `InterfaceToFloat64Ptr` 失败返回 nil

---

## InterfaceToInt64 / InterfaceToInt32

**函数签名**
```go
func InterfaceToInt64(in interface{}) (int64, error)
func InterfaceToInt32(in interface{}) (int32, error)
```

**使用场景**
- 将 interface{} 转为整数
- 处理动态类型数据中的 ID

**正确示例**

```go
productID, err := converter.InterfaceToInt64(genericResult.FieldValue)

bizID, err := converter.InterfaceToInt64(genericResult.FieldValue)
```

**注意事项**
- 支持数值类型和字符串（含千位分隔符）
- `InterfaceToInt32` 会检查范围
- 字符串可解析为浮点数时会截断小数部分

---

## InterfaceToIntPtr

**函数签名**
```go
func InterfaceToIntPtr(in interface{}) *int
```

**使用场景**
- 将 interface{} 转为 int 指针
- 失败返回 nil

**注意事项**
- 基于 `InterfaceToInt64`，需注意 int 范围

---

## InterfaceToUint64/Uint32/Uint16/Uint8/Uint

**函数签名**
```go
func InterfaceToUint64(in interface{}) (uint64, error)
func InterfaceToUint32(in interface{}) (uint32, error)
func InterfaceToUint16(in interface{}) (uint16, error)
func InterfaceToUint8(in interface{}) (uint8, error)
func InterfaceToUint(in interface{}) (uint, error)
```

**使用场景**
- 将 interface{} 转为无符号整数

**注意事项**
- 负数会返回错误
- 会检查范围溢出
- 字符串解析失败返回错误
