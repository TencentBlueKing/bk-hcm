# 接口命名

## 基本规则

- DAO 接口统一命名为 `Interface`
- 客户端接口命名为 `Client`
- 集合类接口使用描述性名称
- 避免 `-er` 后缀（除非是单方法接口且符合 Go 惯例）

## DAO 层接口

DAO 层接口统一命名为 `Interface`，因为接口名称已由包名限定：

```go
type Interface interface {
    List(kt *kit.Kit, opt *types.ListOption) (*types.ListObsBillResult, error)
    BatchCreate(kt *kit.Kit, tx *sqlx.Tx, models []table.ObsBill) ([]string, error)
    BatchDelete(kt *kit.Kit, tx *sqlx.Tx, expr filter.Expression) error
    GetCostTrend(kt *kit.Kit, opt *types.ObsBillDateRangeOption) ([]types.ObsBillCostTrendElement, error)
    // ...
}
```

调用时通过包名区分：

```go
import obsbill "hcm/pkg/dal/dao/obs-bill"

var dao obsbill.Interface
```

### 例外情况

当包内有多个功能接口时，可使用描述性名称：

```go
type AllocInterface interface {
    // 资源分配相关方法
}

// From: pkg/dal/dao/id-generator/id_generator.go
type IDGenInterface interface {
    // ID 生成相关方法
}
```

---

## 客户端接口

第三方客户端接口命名为 `Client`：

```go
// From: pkg/thirdparty/itsm/itsm.gopkg/thirdparty/api-gateway/itsm/itsm.go:37-51
// Client Itsm api.
type Client interface {
    // CreateTicket 创建单据。
    CreateTicket(kt *kit.Kit, params *CreateTicketParams) (string, error)
    // GetTicketResult 获取单据结果。
    GetTicketResult(kt *kit.Kit, sn string) (TicketResult, error)
    // WithdrawTicket 撤销单据。
    WithdrawTicket(kt *kit.Kit, sn string, operator string) error
    // VerifyToken 校验Token。
    VerifyToken(kt *kit.Kit, token string) (bool, error)
    // GetTicketsByUser 获取用户的单据。
    GetTicketsByUser(kt *kit.Kit, req *GetTicketsByUserReq) (*GetTicketsByUserRespData, error)
    // Approve 审批单据。
    Approve(kt *kit.Kit, req *ApproveReq) error
}
```

---

## 集合类接口

多个客户端的集合使用 `ClientSet`：

```go
// From: pkg/client/client.go:39
// ClientSet defines all server's api client set.
type ClientSet struct {
    version      string
    client       client.HTTPClient
    apiDiscovery map[cc.Name]*discovery.APIDiscovery
    // TODO add flow control option
}
```

---

## ORM 接口

ORM 相关接口使用描述性名称：

```go
// From: pkg/dal/dao/orm/orm.go:40
type DoOrm interface {
    Select(dest interface{}, expr string, args ...interface{}) error
    Insert(tableName string, data interface{}) error
    Update(tableName string, data interface{}) error
    // ...
}

// From: pkg/dal/dao/orm/orm.go
type Interface interface {
    Do() DoOrm
    Txn(tx *sqlx.Tx) DoOrm
}
```

---

## 接口一致性检查

在实现文件中添加编译时接口一致性检查：

```go
var _ Interface = new(obsBillDao)
```

这确保 `obsBillDao` 实现了 `Interface` 接口的所有方法。
