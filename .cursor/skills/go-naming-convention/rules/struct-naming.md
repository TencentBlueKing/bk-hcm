# 结构体命名

## 基本规则
- 使用 **PascalCase**
- 根据用途选择合适的后缀
- 私有实现结构体使用小写

## 后缀约定

| 后缀 | 用途 | 示例 |
|-----|-----|------|
| `*Req` | 请求参数 | `ListDeptParentRelReq`, `SetCacheReq` |
| `*Resp` | 响应数据 | `CacheResp`, `ListResp` |
| `*Elem` | 列表元素 | `GetTopoOrgChildrenElem`, `ObsCostTrendElement` |
| `*Result` | 查询结果包装 | `ListObsBillResult` |
| `*Option` | 配置选项 | `ListOption`, `TraceOption` |
| `*Handler` | 请求处理器 | `OrgObsCostTrendExporterHandler` |

---

## 请求结构体 (*Req)

```go
// From: pkg/client/data-service/plat_types.go:35
type ListDeptParentRelReq struct {
    DeptIDs     []string `json:"dept_ids" validate:"required,dive,gt=0"`
    IncludeSelf bool     `json:"include_self"`
}

// From: pkg/client/data-service/plat_types.go:51
type GetTopoOrgChildrenReq struct {
    View          enumor.View `json:"view" validate:"required,oneof=ops fin prod"`
    OrgIDs        []string    `json:"org_ids" validate:"required,dive,gt=0"`
    IncludeSelf   bool        `json:"include_self"`
    IncludeParent bool        `json:"include_parent"`
}
```

> **注意**：统一使用 `*Req` 后缀，不使用 `*Request`（项目中 741 个 `*Req` vs 0 个 `*Request`）

---

## 响应结构体 (*Resp)

```go
// From: pkg/client/data-service/plat_types.go:179
type CacheResp struct {
    ID        string      `json:"id"`
    CacheID   string      `json:"cache_id"`
    UpdatedAt helper.Time `json:"updated_at"`
}
```

---

## 元素结构体 (*Elem)

用于列表中的单个元素：

```go
// From: pkg/client/data-service/plat_types.go:89
type GetTopoOrgChildrenElem struct {
    IDName   `json:",inline"`
    Children []IDName `json:"children"`
}
```

---

## 处理器结构体 (*Handler)

```go
// From: cmd/web-server/logics/excel/exporter/org_obs_cost_trend_exporter.go:46
type OrgObsCostTrendExporterHandler struct {
    cs  clientset.ClientSet
    req *anset.ExportOrgCostTrendReq
}
```

---

## 私有实现结构体

私有实现结构体使用小写，不导出：

```go
// From: pkg/dal/dao/obs-bill/obs_bill.go:213
type obsBillDao struct {
    orm   orm.Interface
    idGen idgen.IDGenInterface
}

// From: cmd/account-server/service/biz/biz.go
type service struct {
    cs clientset.ClientSet
}

// From: cmd/web-server/service/logics/excel/excel.go
type excelSvc struct {
    cs clientset.ClientSet
}
```

---

## 配置结构体

配置相关结构体通常不带特定后缀：

```go
// From: pkg/cc/types.go
type DataBase struct {
    Endpoints []string `yaml:"endpoints"`
    Database  string   `yaml:"database"`
    User      string   `yaml:"user"`
    Password  string   `yaml:"password"`
}
```

---

## 泛型结构体

泛型类型参数使用单字母大写或描述性名称：

```go
// From: pkg/thirdparty/itsm/types.go
type ItemResp[T any] struct {
    Data T `json:"data"`
}

// From: pkg/dal/dao/types/types.go
type ListResult[T any] struct {
    Count uint64 `json:"count"`
    Items []T    `json:"items"`
}
```
