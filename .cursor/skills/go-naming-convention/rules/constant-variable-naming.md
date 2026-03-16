# 常量和变量命名

## 基本规则

- 导出常量使用 **PascalCase**
- 私有常量/变量使用 **camelCase**
- 不使用 SCREAMING_CASE（全大写下划线）

---

## API 常量

```go
// From: pkg/criteria/constant/api.go:24-27
const (
    ApiPrefix  = "api"
    ApiVersion = "v1"
)
```

---

## 枚举常量

枚举常量使用 PascalCase，通常带有类型前缀：

```go
const (
    PlatCatDirectRes  ObsPlatCategoryType = 1
    PlatCatTEGService ObsPlatCategoryType = 2
    PlatCatBGService  ObsPlatCategoryType = 3
)
```

---

## 表名常量

表名常量使用 `*Table` 后缀，类型为 `Name`：

```go
const ObsBillExtTable Name = "obs_bill_ext"

const ObsBillTable Name = "obs_bill"
```

---

## 列描述符

列描述符变量使用 `*Columns` 和 `*ColumnsDescriptor` 后缀：

```go
var ObsBillExtColumns = helper.MergeColumns(nil, ObsBillExtColumnsDescriptor)

var ObsBillExtColumnsDescriptor = helper.ColumnDescriptors{
    {Column: "id", NamedC: "id", Type: reflect.TypeOf((*string)(nil))},
    {Column: "bill_id", NamedC: "bill_id", Type: reflect.TypeOf((*string)(nil))},
    // ...
}
```

---

## 私有包级变量

私有包级变量使用 camelCase：

```go
var viewNameMap = map[View]string{
    Ops:  "运维",
    Fin:  "财务",
    Prod: "产品",
}
```

---

## 错误变量

错误变量使用 `Err*` 前缀：

```go
// From: pkg/dal/dao/orm/orm.go:173
var ErrRetryTransaction = errors.New("retry transaction hit an error")

// From: pkg/criteria/errf/error.go
var ErrInvalidParameter = errors.New("invalid parameter")
```

---

## ❌ 错误示例

```go
// 错误：不使用 SCREAMING_CASE
const API_PREFIX = "api"

// 错误：导出常量应使用 PascalCase
const apiPrefix = "api"

// 错误：私有 map 不应导出
var ViewNameMap = map[View]string{...}
```

---

## Context Key

Context key 使用特定类型和常量：

```go
// From: pkg/runtime/kit/kit.go
type ctxKey string

const (
    RidKey ctxKey = "rid"
    UserKey ctxKey = "user"
)
```
