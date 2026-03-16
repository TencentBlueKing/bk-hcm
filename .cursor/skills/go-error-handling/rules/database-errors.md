# 数据库错误处理规范

## 1. 记录不存在处理

**规则**：使用 `orm.ErrRecordNotFound` 判断记录是否存在。

```go
var ErrRecordNotFound = sql.ErrNoRows

// 使用示例
if errors.Is(err, orm.ErrRecordNotFound) {
    return nil, errf.New(errf.RecordNotFound, "record does not exist")
}
```

**代码位置**：`pkg/dal/dao/orm/sqlx.go:36-41`

```go
var (
    // ErrRecordNotFound returns a "record not found error".
    // Occurs only when attempting to query the database with a struct,
    // querying with a slice won't return this error.
    ErrRecordNotFound = sql.ErrNoRows
)
```

## 2. 数据库操作错误

**规则**：数据库执行错误使用 `errf.DBExecCmdFailed` 错误码。

```go
if err := ob.oi.Do().SelectNamedC(kt.Ctx, &elements, sql, whereValue); err != nil {
    logs.Errorf("get platform cost trend by platform failed, sql: %s, err: %v, rid: %s", sql, err, kt.Rid)
    return nil, errf.NewFromErr(errf.DBExecCmdFailed, err)
}
```

## 3. 事务错误处理

**规则**：使用 `AutoTxn` 自动管理事务，错误时自动回滚。

```go
_, err := o.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
    // 事务内操作
    if _, err := o.dao.ObsBillExt().BatchCreate(kt, txn, part); err != nil {
        logs.Errorf("batch create obs bill ext failed, err: %v, rid: %s", err, kt.Rid)
        return nil, err  // 返回错误，事务自动回滚
    }
    return nil, nil
})
if err != nil {
    logs.Errorf("batch create data failed, err: %v, rid: %s", err, kt.Rid)
    return err
}
```

**代码示例**：

```go
func (o *obs) saveObsBillModels(kt *kit.Kit, models []table.ObsBillExt) error {
    _, err := o.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
        for _, part := range slice.Split(models, int(ctypes.DefaultMaxPageLimit)) {
            if _, err := o.dao.ObsBillExt().BatchCreate(kt, txn, part); err != nil {
                logs.Errorf("batch create obs bill ext failed, err: %v, rid: %s", err, kt.Rid)
                return nil, err
            }
        }
        return nil, nil
    })
    if err != nil {
        logs.Errorf("batch create data failed, err: %v, rid: %s", err, kt.Rid)
        return err
    }
    return nil
}
```

## 4. 批量操作错误处理

**规则**：批量操作中的错误可以选择跳过或收集。

### 跳过错误记录

```go
for _, item := range items {
    if err := insert(item); err != nil {
        logs.Warnf("insert item failed, item: %+v, err: %v, rid: %s", item, err, kt.Rid)
        continue
    }
}
```

### 收集错误统一返回

```go
var errs []error
for _, item := range items {
    if err := insert(item); err != nil {
        errs = append(errs, err)
        continue
    }
}
if len(errs) > 0 {
    return nil, errf.Newf(errf.PartialFailed, "batch insert failed: %d errors", len(errs))
}
```

## 5. 查询错误处理

### 单条查询

```go
var result Model
if err := dao.Do().SelectOne(kt.Ctx, &result, sql, args); err != nil {
    if errors.Is(err, orm.ErrRecordNotFound) {
        return nil, errf.New(errf.RecordNotFound, "record not found")
    }
    logs.Errorf("query failed, sql: %s, err: %v, rid: %s", sql, err, kt.Rid)
    return nil, errf.NewFromErr(errf.DBExecCmdFailed, err)
}
```

### 列表查询

```go
var results []Model
if err := dao.Do().SelectList(kt.Ctx, &results, sql, args); err != nil {
    logs.Errorf("list failed, sql: %s, err: %v, rid: %s", sql, err, kt.Rid)
    return nil, errf.NewFromErr(errf.DBExecCmdFailed, err)
}
// 列表查询空结果不是错误，直接返回空切片
return results, nil
```

## 6. 更新/删除错误处理

```go
result, err := dao.Do().Exec(kt.Ctx, sql, args)
if err != nil {
    logs.Errorf("update failed, sql: %s, err: %v, rid: %s", sql, err, kt.Rid)
    return errf.NewFromErr(errf.DBExecCmdFailed, err)
}

affected, _ := result.RowsAffected()
if affected == 0 {
    return errf.New(errf.RecordNotFound, "no record updated")
}
```

## 最佳实践

### DO

```go
// ✓ 区分记录不存在和数据库错误
if errors.Is(err, orm.ErrRecordNotFound) {
    return nil, errf.New(errf.RecordNotFound, "not found")
}
return nil, errf.NewFromErr(errf.DBExecCmdFailed, err)

// ✓ 事务使用 AutoTxn
_, err := dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
    return nil, operation(txn)
})

// ✓ 日志包含 SQL 信息
logs.Errorf("query failed, sql: %s, err: %v, rid: %s", sql, err, kt.Rid)
```

### DON'T

```go
// ✗ 不区分错误类型
if err != nil {
    return nil, err  // 可能是记录不存在或数据库错误
}

// ✗ 手动管理事务
tx, _ := db.Begin()
// ... 容易忘记 commit/rollback
tx.Commit()

// ✗ 空结果当作错误
if len(results) == 0 {
    return nil, errors.New("no data")  // 应该返回空切片
}
```
