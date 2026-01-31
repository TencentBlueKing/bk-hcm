# ES DAO 规范

## 1. ES DAO 结构

```go
// New creates a new instance.
func New(oi ormes.Interface) Interface {
    return &dao{
        oi: oi,
    }
}

var _ Interface = new(dao)

type dao struct {
    oi ormes.Interface
}
```

**与 MySQL DAO 的区别**：
- 使用 `ormes.Interface` 而非 `orm.Interface`
- 通常不需要 ID 生成器（使用文档唯一键）
- 不支持事务

## 2. 索引管理

### 2.1 按月分索引

```go
// getIndex gets index by date.
func (d *dao) getIndex(date *times.VarTime) string {
    return fmt.Sprintf("%s_%s", table.GPUSMUsageTable, date.Time().Format(constant.MonthLayout))
}
```

### 2.2 动态索引创建

```go
// 检查索引是否存在
indexExists, err := d.oi.Do().IndexExists(kt.Ctx, []string{index})
if err != nil {
    logs.Errorf("check index exists failed, err: %v, rid: %s", err, kt.Rid)
    return err
}

// 不存在则创建
if !indexExists {
    mapping, err := helper.ConvertToEsMapping(table.GPUSMUsageColumnsDescriptor)
    if err != nil {
        logs.Errorf("convert columns to mapping failed, err: %v, rid: %s", err, kt.Rid)
        return err
    }

    if err := d.oi.Do().CreateIndex(kt.Ctx, index, mapping); err != nil {
        logs.Errorf("create index failed, err: %v, rid: %s", err, kt.Rid)
        return err
    }
}
```

## 3. 批量写入

```go
docs := make([]ormes.BulkDoc, len(models))
for i, model := range models {
    if model == nil {
        logs.Warnf("model is nil, index: %s, rid: %s", index, kt.Rid)
        continue
    }

    docs[i] = ormes.BulkDoc{
        ID:  model.UniqueKey(),
        Doc: model,
    }
}

if err := d.oi.Do().BulkUpsert(kt.Ctx, index, docs, 1000); err != nil {
    logs.Errorf("bulk upsert failed, err: %v, rid: %s", err, kt.Rid)
    return err
}
```

**关键点**：
- 使用 `model.UniqueKey()` 作为文档 ID
- 使用 `BulkUpsert` 进行批量写入
- 指定批量大小（如 1000）

## 4. 聚合查询

ES DAO 聚合方法通常使用 `Agg*` 前缀：

```go
// 示例：按产品聚合使用量
func (d *dao) AggUsageByProduct(kt *kit.Kit, opt *types.AggOption) ([]types.ProductUsage, error) {
    // 构建聚合查询
    // ...
}

// 示例：按日期范围聚合趋势
func (d *dao) AggUsageTrendByDateRange(kt *kit.Kit, opt *types.TrendOption) ([]types.UsageTrend, error) {
    // 构建时间序列聚合
    // ...
}
```

## 5. ES vs MySQL DAO 对比

| 特性 | MySQL DAO | ES DAO |
|-----|-----------|--------|
| ORM 接口 | `orm.Interface` | `ormes.Interface` |
| ID 生成 | `idgenerator.IDGenInterface` | 文档唯一键 |
| 事务支持 | ✅ `*sqlx.Tx` | ❌ 不支持 |
| 索引管理 | 无需 | 需要处理动态索引 |
| 聚合方法命名 | `Get*` | `Agg*` |
| 批量操作 | `BulkInsert` | `BulkUpsert` |

## 6. 文档唯一键

ES 文档使用唯一键而非自增 ID：

```go
// 表结构定义中实现 UniqueKey 方法
func (g *GPUSMUsage) UniqueKey() string {
    return fmt.Sprintf("%s_%s_%d", g.DataDate, g.GPUID, g.Timestamp)
}
```

## 7. ES DAO 检查清单

✅ **实现 ES DAO 时应确认**：
- [ ] 使用 `ormes.Interface`
- [ ] 实现 `getIndex` 方法处理动态索引
- [ ] 批量写入前检查索引是否存在
- [ ] 使用 `BulkUpsert` 而非 `BulkInsert`
- [ ] 文档模型实现 `UniqueKey()` 方法
- [ ] 聚合方法使用 `Agg*` 前缀
