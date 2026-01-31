# 指针转换函数

## ValToPtr

**函数签名**
```go
func ValToPtr[T any](val T) *T
```

**使用场景**
- 将值转为指针，用于可选字段或响应结构体
- 构造测试数据时创建指针字段

**正确示例**

```go
finalResult.DailyPeakGPUUsage = converter.ValToPtr(maths.SafeDivideZeroF64(dailyUsageNumerator,
    dailyUsageDenominator))
finalResult.WeeklyPeakAvgGPUUsage = converter.ValToPtr(maths.SafeDivideZeroF64(weeklyUsageNumerator,
    weeklyUsageDenominator))

Count: converter.ValToPtr(uint64(len(todOrgs))),

return converter.ValToPtr(SafeDivideZeroF64R3(src-pre, math.Abs(pre)))
```

**注意事项**
- 返回的指针指向新分配的值，不是原值的地址
- 适用于需要指针类型的 API 响应字段

---

## PtrToVal

**函数签名**
```go
func PtrToVal[T any](ptr *T) T
```

**使用场景**
- 安全解引用指针，nil 时返回零值
- 日志输出时避免 nil 指针解引用

**正确示例**

```go
converter.PtrToVal(obsdata.Memo), kt.Rid)

obsdata, converter.PtrToVal(obsdata.Memo), kt.Rid)
```

**注意事项**
- nil 指针返回零值，需判断是否需要零值语义
- 不适合需要区分 nil 和零值的场景

---

## SliceToPtr / PtrToSlice

**函数签名**
```go
func SliceToPtr[T any](slice []T) []*T
func PtrToSlice[T any](slice []*T) []T
```

**使用场景**
- `SliceToPtr`：将值切片转为指针切片，用于需要指针切片的 API
- `PtrToSlice`：将指针切片转为值切片，nil 指针会被转为零值

**正确示例**

```go
tmp := converter.SliceToSet(conv.PtrToSlice(a))
```

**注意事项**
- `PtrToSlice` 中 nil 指针会转为零值，可能丢失信息
- `SliceToPtr` 会为每个元素创建新指针

---

## IfNilF64

**函数签名**
```go
func IfNilF64(src *float64, candidate float64) float64
```

**使用场景**
- float64 指针的默认值处理

**注意事项**
- 仅适用于 float64 指针，其他类型需自行判断
