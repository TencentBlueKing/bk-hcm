# Map 操作函数

## GroupBy

**函数签名**
```go
func GroupBy[K comparable, V any](source []V, keyFunc func(V) K) map[K][]V
```

**使用场景**
- 按 key 函数分组
- 数据聚合与统计

**正确示例**

```go
// From: cmd/hc-service/logics/res-sync/aws/region.go:188
delCloudMap := converter.StringSliceToMap(delCloudIDs)
```

**注意事项**
- key 函数需返回 `comparable` 类型
- 结果 map 的值为切片，可能为空切片

---

## SliceToMap

**函数签名**
```go
func SliceToMap[T any](S []T, fn func(item T) string, overwrite bool) (m map[string]T, err error)
```

**使用场景**
- 将切片转为 map，便于按 key 查找
- 构建 ID 到对象的映射

**正确示例**

```go
// From: cmd/analysis-server/service/plat/cost_optim.go:250-255
if ticketMap, err = converter.SliceToMap(ticketsResp, func(item *table.CostOptimTicket) string {
    return item.ID
}, false); err != nil {
    logs.Errorf("convert ticket slice to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
    return nil, errf.NewFromErr(errf.Aborted, err)
}

// From: cmd/analysis-server/service/plat/cost_optim.go:289-291
aiApprovalMap, err := converter.SliceToMap(dsResp, func(item dsset.CostOptimTicketAIApprovalItem) string {
    return item.ID
}, false)

// From: cmd/analysis-server/service/logics/idc/bandwidth_change_rate.go:192-199
curRegionMap, err := converter.SliceToMap(cur, func(item anset.BandwidthCostDistRegionLevel) string {
    return item.Region
}, false)
prevRegionMap, err := converter.SliceToMap(prev, func(item anset.BandwidthCostDistRegionLevel) string {
    return item.Region
}, false)
```

**注意事项**
- `overwrite=false` 时遇到重复 key 返回错误
- `overwrite=true` 时覆盖旧值
- key 函数需返回 string

---

## MapKeyToSlice / MapValueToSlice

**函数签名**
```go
func MapKeyToSlice[K cmp.Ordered, V any](source map[K]V) []K
func MapValueToSlice[K cmp.Ordered, V any](source map[K]V) []V
```

**使用场景**
- 提取 map 的键或值
- 需要有序结果时使用（键会被排序）

**正确示例**

```go
allKeys := converter.MapKeyToSlice(usageDataMap)

dsReq := &dsset.CostOptimTicketAIApprovalReq{IDs: converter.MapKeyToSlice(ticketMap)}

return converter.MapKeyToSlice(viewNameMap)

return converter.MapValueToSlice(keyBillMap)

deduplicatedModels := converter.MapValueToSlice(uniqueModels)
```

**注意事项**
- 键类型需为 `cmp.Ordered`（可排序）
- `MapKeyToSlice` 返回有序键
- `MapValueToSlice` 按有序键顺序返回值

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
