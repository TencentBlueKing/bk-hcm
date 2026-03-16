# 切片操作函数

## ArrayUnique

**函数签名**
```go
func ArrayUnique[T comparable](source []T) []T
```

**使用场景**
- 切片去重
- 合并多个列表后去重

**正确示例**

```go
thirdLevelDeptNames = converter.ArrayUnique(thirdLevelDeptNames)

comp.Creators = converter.ArrayUnique(operatorMap["creators"])
comp.Committers = converter.ArrayUnique(operatorMap["committers"])

operators = converter.ArrayUnique(operators)

dates = converter.ArrayUnique(dates)

moduleIDs = converter.ArrayUnique(moduleIDs)
```

**注意事项**
- 元素类型需为 `comparable`
- 保持首次出现的元素顺序
- 空切片返回原切片

---

## SliceToSet

**函数签名**
```go
func SliceToSet[T comparable](source []T) map[T]struct{}
```

**使用场景**
- 快速成员判断（O(1)）
- 集合运算（交集、差集等）

**正确示例**

```go
partIDMap := converter.SliceToSet(partIDs)

tmp := converter.SliceToSet(converter.PtrToSlice(a))
```

**注意事项**
- 用于频繁查找，不用于去重（去重用 `ArrayUnique`）

---

## StringSliceToInt64Slice / Int64SliceToStringSlice

**函数签名**
```go
func StringSliceToInt64Slice(source []string) ([]int64, error)
func Int64SliceToStringSlice(source []int64) []string
```

**使用场景**
- 字符串与 int64 切片互转
- 处理 API 参数或数据库查询

**正确示例**

```go
prodIDs, err := converter.StringSliceToInt64Slice(req.OrgFilter.OpProductIDs)

deptIDs := converter.Int64SliceToStringSlice(obstypes.DeptIDsWithFinOps)

updateProject.OpProductIDs = strings.Join(converter.Int64SliceToStringSlice(*req.OpProductIDs),
    constant.SepSemicolon)
```

**注意事项**
- `StringSliceToInt64Slice` 空字符串会返回错误
- `Int64SliceToStringSlice` 不会失败，可安全使用

---

## IntSliceToStringSlice

**函数签名**
```go
func IntSliceToStringSlice(source []int) []string
```

**使用场景**
- int 切片转字符串切片（较少使用，推荐使用 `Int64SliceToStringSlice`）

**注意事项**
- 不会失败，可安全使用
- 生产代码中较少使用

---

## StringToInt64Slice

**函数签名**
```go
func StringToInt64Slice(source, sep string) ([]int64, error)
```

**使用场景**
- 解析分隔符分隔的 ID 字符串（如分号、逗号）

**正确示例**

```go
opProductIDs, err := converter.StringToInt64Slice(project.Details[0].OpProductIDs, constant.SepSemicolon)

bizIDsInt64, err := converter.StringToInt64Slice(bizIDs, constant.SepSemicolon)

opProductIDs, err := converter.StringToInt64Slice(results.Details[0].OpProductIDs, constant.SepSemicolon)
```

**注意事项**
- 空字符串返回空切片，不返回错误
- 分隔符需与数据格式一致

---

## SliceToInterfaceSlice

**函数签名**
```go
func SliceToInterfaceSlice[T any](source []T) []interface{}
```

**使用场景**
- 将类型切片转为 `[]interface{}`，用于需要 `[]interface{}` 的 API（如 Elasticsearch）

**正确示例**

```go
queryList = queryList.Must(elastic.NewTermsQuery("app_name", converter.SliceToInterfaceSlice(opts.AppNames)...))
queryList = queryList.Must(elastic.NewTermsQuery("sync_date", converter.SliceToInterfaceSlice(opts.Dates)...))

queryList = queryList.Must(elastic.NewTermsQuery("app_name", converter.SliceToInterfaceSlice(appNames)...))
queryList = queryList.Must(elastic.NewTermsQuery("sync_date", converter.SliceToInterfaceSlice(dates)...))
```

**注意事项**
- 主要用于 Elasticsearch 等需要 `[]interface{}` 的库
- 有类型转换开销

---

## InterfaceToInterfaceSlice / InterfaceToStringSlice

**函数签名**
```go
func InterfaceToInterfaceSlice(value interface{}) ([]interface{}, error)
func InterfaceToStringSlice(source interface{}) ([]string, error)
```

**使用场景**
- 处理动态类型数据（如 JSON 解析结果）
- 将 interface{} 转为切片

**正确示例**

```go
values, err := converter.InterfaceToInterfaceSlice(value)
```

**注意事项**
- 输入需为数组或切片，否则返回错误
- 支持指针类型（自动解引用）

---

## FlattenSlice

**函数签名**
```go
func FlattenSlice[T any](s [][]T) []T
```

**使用场景**
- 将二维切片展平为一维

**注意事项**
- 输入需为二维切片
- 空切片返回空切片
