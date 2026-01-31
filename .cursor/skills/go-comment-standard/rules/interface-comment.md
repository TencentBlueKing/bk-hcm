# 接口注释规范

## 接口整体注释

### 规则

1. 注释**必须**以接口名称开头
2. 使用 `holds`、`defines` 或 `describes` 描述
3. 说明接口定义的操作范围

### 标准格式

```go
// InterfaceName holds/defines all the supported operations for xxx.
type InterfaceName interface {
```

### 代码示例

```45:53:pkg/dal/dao/cloud/cvm/cvm.go
// Interface only used for cvm.
type Interface interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablecvm.Table) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *tablecvm.Table) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablecvm.Table) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListCvmDetails, error)
	ListWithTx(kt *kit.Kit, tx *sqlx.Tx, opt *types.ListOption) (*types.ListCvmDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ Interface = new(Dao)
```

```74:77:pkg/dal/dao/dao.go
// Set defines all the DAO to be operated.
type Set interface {
	Audit() audit.Interface
	Auth() auth.Auth
```

```38:44:pkg/client/api.go
// ClientSet defines all server's api client set.
type ClientSet struct {
	version      string
	client       client.HTTPClient
	apiDiscovery map[cc.Name]*discovery.APIDiscovery
}
```

## 接口方法注释

### 规则

1. 接口定义处的方法**通常不添加注释**
2. 实现接口的方法**需要添加注释**
3. 复杂方法可在接口定义处添加简要说明

### 代码示例（接口定义处 - 无注释）

```74:150:pkg/dal/dao/dao.go
// Set defines all the DAO to be operated.
type Set interface {
	Audit() audit.Interface
	Auth() auth.Auth
	// ... 方法无单独注释

}
```

### 代码示例（接口实现处 - 有注释）

```go
// List retrieves GPU small model usage records for the specified date.
func (d *gpuSMUsageDao) List(kt *kit.Kit, date *times.VarTime, opt *types.ListOption) (
	*types.ListResult[*table.GPUSMUsage], error) {
	// implementation
}
```

## 常见动词

| 动词 | 使用场景 | 示例 |
|-----|---------|------|
| `holds` | 方法集合 | `// Interface holds all operations.` |
| `defines` | 契约定义 | `// Set defines all the DAO to be operated.` |
| `describes` | 行为描述 | `// Handler describes HTTP request handling.` |

## DAO 层接口

### 规则

1. DAO 接口通常命名为 `Interface`
2. 注释说明操作的数据实体

### 代码示例

```go
// Interface holds all the supported operations for the user entity.
type Interface interface {
	Get(kt *kit.Kit, id string) (*table.User, error)
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListResult[*table.User], error)
	Create(kt *kit.Kit, user *table.User) error
	Update(kt *kit.Kit, user *table.User) error
	Delete(kt *kit.Kit, id string) error
}
```

## Client 接口

### 规则

1. Client 接口通常命名为 `Client` 或 `Interface`
2. 注释说明客户端连接的服务

### 代码示例

```go
// Client defines the operations for interacting with the ETL service.
type Client interface {
	// RefreshData triggers data refresh for the specified date range.
	RefreshData(kt *kit.Kit, req *RefreshDataReq) error
}
```
