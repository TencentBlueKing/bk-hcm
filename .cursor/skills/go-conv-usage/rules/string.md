# 字符串操作函数

## JoinQuotes

**函数签名**
```go
func JoinQuotes(elems []string, sep string) string
```

**使用场景**
- SQL IN 子句拼接（带单引号）
- 生成带引号的字符串列表

**正确示例**

```go
converter.JoinQuotes(secProdIDs, ","))

where = append(where, fmt.Sprintf("service_dept_name IN (%s)", converter.JoinQuotes(req.ServiceDeptNames, ",")))
where = append(where, fmt.Sprintf("secondary_product_id IN (%s)", converter.JoinQuotes(secProdIDs, ",")))

eleSQL = append(eleSQL, fmt.Sprintf("area_code IN (%s)", converter.JoinQuotes(filter.Areas, ",")))
eleSQL = append(eleSQL, fmt.Sprintf("vendor IN (%s)", converter.JoinQuotes(filter.Vendors, ",")))
eleSQL = append(eleSQL, fmt.Sprintf("category_code IN (%s)", converter.JoinQuotes(filter.Categories, ",")))
```

**注意事项**
- 每个元素用单引号包裹
- 空切片返回空字符串
- 主要用于 SQL 拼接

---

## InterfaceToString

**函数签名**
```go
func InterfaceToString(in interface{}) (string, error)
```

**使用场景**
- 将 interface{} 转为字符串
- 处理动态类型数据（如 Elasticsearch 聚合结果）

**正确示例**

```go
rawDate, err = converter.InterfaceToString(bucket.Key)

service3Name, err := converter.InterfaceToString(genericResult.FieldValue)

gpuType, err := converter.InterfaceToString(genericResult.FieldValue)

detail.DeviceType, err = converter.InterfaceToString(appBucket.DeviceType.Buckets[0].Key)
detail.Operator, err = converter.InterfaceToString(appBucket.Operator.Buckets[0].Key)
detail.IP, err = converter.InterfaceToString(appBucket.IP.Buckets[0].Key)
```

**注意事项**
- 支持基本类型、`fmt.Stringer`、指针
- 不支持的类型返回错误
- nil 返回错误

---

## InterfaceToStringPtr

**函数签名**
```go
func InterfaceToStringPtr(in interface{}) *string
```

**使用场景**
- 将 interface{} 转为字符串指针
- 转换失败返回 nil（不返回错误）

**注意事项**
- 失败时返回 nil，需判断 nil
- 适合可选字段场景
