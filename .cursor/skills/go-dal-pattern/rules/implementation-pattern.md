# 实现模式规范

## 1. 基础结构体（简单实体）

简单实体使用 `dao` 作为结构体名称：

```go
// From: pkg/dal/dao/dao.go:695-701
func (s *set) LoadBalancer() loadbalancer.LoadBalancerInterface {
    return &loadbalancer.LoadBalancerDao{
        Orm:   s.orm,
        IDGen: s.idGen,
        Audit: s.audit,
    }
}

// From: pkg/dal/dao/cloud/load-balancer/load_balancer.go:47-63
// LoadBalancerInterface only used for load balancer.
type LoadBalancerInterface interface {
    BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablelb.LoadBalancerTable) ([]string, error)
    Update(kt *kit.Kit, expr *filter.Expression, model *tablelb.LoadBalancerTable) error
    UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablelb.LoadBalancerTable) error
    List(kt *kit.Kit, opt *types.ListOption) (*typeslb.ListLoadBalancerDetails, error)
    DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ LoadBalancerInterface = new(LoadBalancerDao)

// LoadBalancerDao load balancer dao.
type LoadBalancerDao struct {
    Orm   orm.Interface
    IDGen idgen.IDGenInterface
    Audit audit.Interface
}
```

## 2. 描述性结构体（复杂实体）

复杂实体使用更具描述性的名称（如 `xxxDao`）：

```go
// From: pkg/dal/dao/cloud/load-balancer/load_balancer.go:47-63
// LoadBalancerInterface only used for load balancer.
type LoadBalancerInterface interface {
    BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablelb.LoadBalancerTable) ([]string, error)
    Update(kt *kit.Kit, expr *filter.Expression, model *tablelb.LoadBalancerTable) error
    UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablelb.LoadBalancerTable) error
    List(kt *kit.Kit, opt *types.ListOption) (*typeslb.ListLoadBalancerDetails, error)
    DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ LoadBalancerInterface = new(LoadBalancerDao)

// LoadBalancerDao load balancer dao.
type LoadBalancerDao struct {
    Orm   orm.Interface
    IDGen idgen.IDGenInterface
    Audit audit.Interface
}
```

## 3. 结构体命名选择

| 场景 | 命名 | 示例 |
|-----|------|------|
| 简单 CRUD 实体 | `dao` | `cost_optimization_ticket` |
| 带审计的实体 | `xxxDao` | `systemConfigDao` |
| 复杂业务实体 | `xxxDao` | `obsBillDao`, `obsMetaDao` |
| 需要缓存的实体 | `xxxDao` | `obsBillDao` |

审计使用示例：

```go
// create and init an audit operation
auditOp, err := sc.audit.NewAuditOp(kt, tx, enumor.APIActionCreate, []string{id})
if err != nil {
    logs.Errorf("system config new audit op failed, err: %v, rid: %s", err, kt.Rid)
    return "", err
}

// ... 执行业务逻辑 ...

// save audit records
if err = auditOp.Save(); err != nil {
    logs.Errorf("system config save audit records failed, err: %v, rid: %s", err, kt.Rid)
    return "", err
}
```

## 5. 带缓存的 DAO

```go
// New create instance.
func New(oi orm.Interface, idGen idgen.IDGenInterface, sd *share.Share) Interface {
    client := gcache.New(50).
        LRU().
        Expiration(time.Duration(3) * time.Hour).
        Build()

    return &obsBillDao{
        oi:           oi,
        idGen:        idGen,
        sd:           sd,
        obsBillCache: client,
    }
}

var _ Interface = new(obsBillDao)

type obsBillDao struct {
    obsBillCache gcache.Cache
    oi           orm.Interface
    idGen        idgen.IDGenInterface
    sd           *share.Share
}
```

### 5.1 缓存使用场景

✅ **适合缓存**：
- 查询频繁但变更少的数据
- 计算成本高的聚合结果
- 跨服务调用的元数据

❌ **不适合缓存**：
- 频繁变更的数据
- 实时性要求高的数据
- 数据量小且查询简单

## 6. ID 生成

### 6.1 ID 生成接口

```go
// From: pkg/dal/dao/id-generator/id_generator.go:36-42
type IDGenInterface interface {
    Batch(kt *kit.Kit, resource table.Name, count int) ([]string, error)
    One(kt *kit.Kit, resource table.Name) (string, error)
}
```

### 6.2 使用示例

```go
id, err := d.idGen.One(kt, table.CostOptimTicketTable)
if err != nil {
    logs.Errorf("generate cost optimization ticket id failed, err: %v, rid: %s", err, kt.Rid)
    return "", errf.NewFromErr(errf.DBExecCmdFailed, err)
}
```

## 7. 错误处理

### 7.1 错误日志

```go
if err != nil {
    logs.Errorf("system config count failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
    return nil, err
}
```

### 7.2 错误封装

```go
logs.Errorf("generate cost optimization ticket id failed, err: %v, rid: %s", err, kt.Rid)
return "", errf.NewFromErr(errf.DBExecCmdFailed, err)
```

## 8. SQL 构建

### 8.1 SQL 辅助函数

```go
// 使用 fmt.Sprintf 构建 SQL
sql := fmt.Sprintf(`
    SELECT %s
    FROM %s
    %s
    %s
`, table.SystemConfigColumns.FieldsNamedExpr(opt.Fields),
    table.SystemConfigTable, whereExpr, pageExpr)
```

### 8.2 SQL 构建示例

```go
sql := fmt.Sprintf(`
    SELECT %s
    FROM %s
    %s
    %s
`, table.SystemConfigColumns.FieldsNamedExpr(opt.Fields),
    table.SystemConfigTable, whereExpr, pageExpr)
```

### 8.3 Filter 表达式

```go
expr := tools.AllExpression()
expr.Rules = append(expr.Rules, tools.RuleEqual("type", typ))
expr.Rules = append(expr.Rules, tools.RuleEqual("name", name))
```

## 9. 实现检查清单

✅ **实现时应确认**：
- [ ] 构造函数命名为 `New`
- [ ] 使用 `var _ Interface = new(dao)` 进行接口校验
- [ ] 所有方法使用 `kt.Ctx` 传递上下文
- [ ] 错误日志包含 `rid: %s, kt.Rid`
- [ ] 使用 `fmt.Sprintf` 构建 SQL
- [ ] 分页时先检查 `opt.Page.Count`
