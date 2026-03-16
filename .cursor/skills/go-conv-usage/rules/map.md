# Map 操作函数

## SliceToMap

**函数签名**
```go
func SliceToMap[T any, K comparable, V any](source []T, kvFunc func(T) (K, V)) map[K]V
```

**使用场景**
- 将切片转为 map，便于按 key 查找
- 构建 ID 到对象的映射

**正确示例**

```go
userMap := converter.SliceToMap(users, func(u User) (string, User) {
    return u.ID, u
})

idNameMap := converter.SliceToMap(items, func(item Item) (int64, string) {
    return item.ID, item.Name
})
```

**注意事项**
- kvFunc 返回 key-value 对
- 遇到重复 key 时后面的会覆盖前面的
- key 类型需为 `comparable`

---

## MapKeyToSlice / MapValueToSlice

**函数签名**
```go
func MapKeyToSlice[K comparable, V any](source map[K]V) []K
func MapValueToSlice[KeyType comparable, ValType any](source map[KeyType]ValType) []ValType
```

**使用场景**
- 提取 map 的键或值

**正确示例**

```go
allKeys := converter.MapKeyToSlice(usageDataMap)

return converter.MapKeyToSlice(viewNameMap)

return converter.MapValueToSlice(keyBillMap)

deduplicatedModels := converter.MapValueToSlice(uniqueModels)
```

**注意事项**
- 键类型需为 `comparable`
- 返回顺序不固定（Go map 遍历顺序不确定）

---

## StructToMap

**函数签名**
```go
func StructToMap(source interface{}) (map[string]interface{}, error)
```

**使用场景**
- 结构体转 map，用于动态处理或序列化

**注意事项**
- 使用 JSON tag 作为 key
- 支持指针类型（自动解引用）
- 非结构体返回错误
