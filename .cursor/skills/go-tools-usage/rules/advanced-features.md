# 高级特性

## 1. 组合字段查询

支持多字段组合查询，字段名使用括号包裹。

### 1.1 组合字段 IN 查询

```go
// 组合字段 IN 查询
catPlatPairs := make([][]interface{}, len(opt.PlatformInfos))
for i, item := range opt.PlatformInfos {
    catPlatPairs[i] = []interface{}{item.PlatformCategoryID, item.PlatformID}
}
expr.Rules = append(expr.Rules, filter.ContainsRule("(platform_category_id,platform_id)", catPlatPairs))
// SQL: (platform_category_id,platform_id) IN ((:a,:b),(:c,:d))
```

### 1.2 组合字段等于/不等于

```go
// 组合字段不等于
f.Rules = append(f.Rules, filter.NotEqualExpression("(platform_category_id,platform_id)",
    []interface{}{enumor.PlatCatOtherCost, enumor.PlatSpecialAdj}))
// SQL: (platform_category_id,platform_id) != (:a,:b)

// 组合字段等于
f.Rules = append(f.Rules, filter.EqualExpression("(org_id,dept_id)",
    []interface{}{orgID, deptID}))
// SQL: (org_id,dept_id) = (:a,:b)
```

## 2. 辅助方法

### 2.1 IsEmpty() - 检查表达式是否为空

```go
if opt.Filter == nil || opt.Filter.IsEmpty() {
    opt.Filter = filter.AllExpression()
}
```

### 2.2 HasAnyField() - 检查是否包含特定字段

**用途**：根据查询字段选择不同的表或执行不同逻辑

```go
tableName := table.ObsBillBriefTable
if expr.HasAnyField("statistics_type_id", "statistics_type") {
    tableName = table.ObsBillTable  // 需要详细字段时使用完整表
}
```

### 2.3 LogMarshal() - 日志打印

```go
logs.Infof("query filter: %s", expr.LogMarshal())
```

### 2.4 Validate() - 表达式校验

```go
exprOpt := filter.NewExprOption(
    filter.RuleFields(table.ObsBillColumns.ColumnTypes()),  // 允许的字段及类型
    filter.MaxInLimit(constant.OpProductIDMaxLimit),        // IN 操作符最大元素数
    filter.MaxNotInLimit(200),                              // NOT IN 最大元素数
    filter.MaxRulesLimit(20),                               // 规则数量限制
)
if err := expr.Validate(exprOpt); err != nil {
    return errf.NewFromErr(errf.InvalidParameter, err)
}
```

## 3. SQLWhereOption 配置

### 3.1 默认配置

大多数场景使用默认配置即可：

```go
whereExpr, whereValue, err := expr.SQLWhereExpr(filter.DefaultSqlWhereOption)
```

### 3.2 自定义 Priority

Priority 定义字段优先级，影响 SQL 条件顺序（让查询条件顺序匹配数据库索引）：

```go
opt := &filter.SQLWhereOption{
    Priority: filter.Priority{"id", "fee_date", "op_product_id"},
}
whereExpr, whereValue, err := expr.SQLWhereExpr(opt)
```

### 3.3 CrownedOption

用于添加额外的查询条件：

```go
opt := &filter.SQLWhereOption{
    Priority: filter.Priority{"id"},
    CrownedOption: &filter.CrownedOption{
        CrownedOp: filter.And,
        Rules: []filter.RuleFactory{
            filter.EqualRule("is_deleted", false),
        },
    },
}
```

## 4. Elasticsearch 查询生成

filter 包支持生成 Elasticsearch 查询。

### 4.1 生成 ES 查询

```go
query, err := expr.GenEsQuery()
if err != nil {
    return nil, err
}
// query 是 elastic.Query 类型，可直接用于 ES 查询
```

### 4.2 支持的操作符

| 操作符 | ES 支持 | ES Query 类型 |
|--------|---------|---------------|
| `eq` | ✅ | TermQuery |
| `neq` | ✅ | BoolQuery.MustNot(TermQuery) |
| `gt`, `gte`, `lt`, `lte` | ✅ | RangeQuery |
| `in` | ✅ | TermsQuery |
| `nin` | ✅ | BoolQuery.MustNot(TermsQuery) |
| `cs`, `cis` | ❌ | 不支持 |
| JSON 操作符 | ❌ | 不支持 |

### 4.3 ES 查询示例

```go
expr := filter.AllExpression()
expr.Rules = append(expr.Rules, filter.EqualRule("status", "active"))
expr.Rules = append(expr.Rules, filter.ContainsRule("type", []string{"A", "B"}))

query, err := expr.GenEsQuery()
if err != nil {
    return nil, err
}

// 使用 ES 客户端执行查询
result, err := esClient.Search().
    Index("my_index").
    Query(query).
    Do(ctx)
```

## 5. ExprOption 配置项

| 配置函数 | 用途 | 默认值 |
|----------|------|--------|
| `RuleFields(fields)` | 设置允许的字段及其类型 | 无限制 |
| `MaxInLimit(limit)` | IN 操作符最大元素数 | 200 |
| `MaxNotInLimit(limit)` | NOT IN 操作符最大元素数 | 200 |
| `MaxRulesLimit(limit)` | 规则数量限制 | 10（ES: 100） |

```go
exprOpt := filter.NewExprOption(
    filter.RuleFields(map[string]enumor.ColumnType{
        "id":         enumor.Integer,
        "name":       enumor.String,
        "fee_date":   enumor.Time,
        "is_deleted": enumor.Boolean,
    }),
    filter.MaxInLimit(500),
    filter.MaxNotInLimit(500),
    filter.MaxRulesLimit(50),
)
```

## 6. 常见问题

### Q1: 如何避免 SQL 注入？

Filter 包使用参数绑定（Named Parameters），自动防止 SQL 注入：

```go
// 生成: name = :name_xxx
// 值通过 whereValue map 安全传递
whereExpr, whereValue, err := expr.SQLWhereExpr(filter.DefaultSqlWhereOption)
db.SelectNamedC(ctx, &result, sql, whereValue)
```

### Q2: 如何处理空条件？

使用 `AllExpression()` 创建空表达式，空表达式生成的 SQL 不包含 WHERE 子句：

```go
expr := filter.AllExpression()
whereExpr, whereValue, err := expr.SQLWhereExpr(filter.DefaultSqlWhereOption)
// whereExpr 为空字符串
```

### Q3: 如何调试 Expression？

使用 `LogMarshal()` 打印 Expression 内容：

```go
logs.Debugf("filter expression: %s", expr.LogMarshal())
```

### Q4: 如何在 JSON 字段上查询？

使用 `.` 分隔嵌套路径：

```go
// 查询 extension.vpc.id = 3
expr.Rules = append(expr.Rules, &filter.AtomRule{
    Field: "extension.vpc.id",
    Op:    filter.JSONEqual.Factory(),
    Value: 3,
})
```
