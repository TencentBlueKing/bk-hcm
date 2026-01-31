# 事务处理规范

## 1. 事务传递模式

事务通过 `*sqlx.Tx` 参数传递：

```go
// BatchCreate batch create cost optimization ticket.
func (d *dao) BatchCreate(kt *kit.Kit, tx *sqlx.Tx, models []table.CostOptimTicket) ([]string, error) {
    // ...
    if err = d.oi.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
        logs.Errorf("insert %s failed, err: %v, rid: %s", table.CostOptimTicketTable, err, kt.Rid)
        return nil, fmt.Errorf("insert %s failed, err: %v", table.CostOptimTicketTable, err)
    }
    // ...
}
```

## 2. 智能选择插入函数

使用 `GetInsertFunc` 根据事务状态自动选择：

```go
// GetInsertFunc get insert function according to whether tx is nil or not
func GetInsertFunc(oi orm.Interface, tx *sqlx.Tx) func(ctx context.Context, expr string, data interface{}) error {
    insertFunc := oi.Do().Insert
    if tx != nil {
        insertFunc = oi.Txn(tx).Insert
    }

    return insertFunc
}
```

使用示例：

```go
insertFunc := dhelper.GetInsertFunc(d.oi, tx)
if err := insertFunc(kt.Ctx, sql, model); err != nil {
    logs.Errorf("insert failed, err: %v, rid: %s", err, kt.Rid)
    return "", err
}
```

## 3. 自动事务

```go
// From: pkg/dal/dao/orm/orm.go:175-214
// AutoTxn is a wrapper to do all the transaction operations as follows:
// 1. auto launch the transaction
// 2. process the logics, which is a callback run function
// 3. rollback the transaction if 'run' hit an error automatically.
// 4. commit the transaction if no error happens.
func (o *runtimeOrm) AutoTxn(kit *kit.Kit, run TxnFunc) (interface{}, error) {
    // ...
}
```

使用示例：

```go
result, err := s.dao.Txn().AutoTxn(kt, func(tx *sqlx.Tx) (interface{}, error) {
    // 创建主记录
    id, err := s.dao.Entity().Create(kt, tx, model)
    if err != nil {
        return nil, err
    }
    
    // 创建关联记录
    if err := s.dao.Related().BatchCreate(kt, tx, relatedModels); err != nil {
        return nil, err
    }
    
    return id, nil
})
```

## 4. DAO Set 事务接口

```go
// From: pkg/dal/dao/dao.go:513-516
// Txn define dao set Txn.
type Txn struct {
    orm orm.Interface
}

// AutoTxn auto Txn.
func (t *Txn) AutoTxn(kt *kit.Kit, run orm.TxnFunc) (interface{}, error) {
    return t.orm.AutoTxn(kt, run)
}

// Txn return Txn.
func (s *set) Txn() *Txn {
    return &Txn{
        orm: s.orm,
    }
}
```

## 5. 何时使用事务

✅ **需要事务**：
- 多表写入（创建主记录 + 关联记录）
- 先删后增（替换操作）
- 需要原子性保证的业务逻辑
- 带审计的写入操作

❌ **不需要事务**：
- 单表单条记录操作
- 只读查询
- 幂等的批量操作

## 6. 事务方法命名

带 `Txn` 后缀表示在事务中执行：

```go
// ListTxn returns a list of system config.
func (sc *systemConfigDao) ListTxn(kt *kit.Kit, tx *sqlx.Tx, opt *types.ListOption) (*types.ListResult[*table.SystemConfig], error)
```

## 7. ORM 事务操作对比

| 场景 | 方法 | 说明 |
|-----|------|------|
| 无事务操作 | `d.oi.Do().Method()` | 直接执行 |
| 事务内操作 | `d.oi.Txn(tx).Method()` | 在事务内执行 |
| 自动事务 | `d.oi.AutoTxn(kt, func)` | 自动管理事务 |

## 8. 事务处理检查清单

✅ **使用事务时应确认**：
- [ ] CUD 方法包含 `tx *sqlx.Tx` 参数
- [ ] 使用 `d.oi.Txn(tx)` 执行事务内操作
- [ ] 错误时立即返回（自动回滚）
- [ ] 事务内操作逻辑完整
- [ ] 考虑使用 `GetInsertFunc` 简化代码
