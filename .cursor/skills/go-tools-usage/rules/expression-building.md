# Expression 构建规则

## 1. 构建方式概述

| 方式 | 适用场景 | 推荐度 |
|------|----------|--------|
| `AllExpression()` + 动态追加 | 需要根据条件动态构建 | ⭐⭐⭐ 推荐 |
| 快捷函数直接构建 | 简单的单条件或固定条件 | ⭐⭐⭐ 推荐 |
| `ExpressionAnd()`/`ExpressionOr()` | 合并多个已有表达式或规则 | ⭐⭐⭐ 推荐 |
| 直接构建结构体 | 静态的、固定的过滤条件 | ⭐⭐ 可选 |

## 2. 推荐方式：AllExpression() + 动态追加

**适用场景**：需要根据条件动态构建过滤器

```go
expr := tools.AllExpression()
dateExpr, err := tools.DateRangeExpression("fee_date", opt.DateRange)
if err != nil {
    return nil, err
}
expr.Rules = append(expr.Rules, dateExpr)

if len(opt.OpProductIDs) > 0 {
    expr.Rules = append(expr.Rules, tools.ContainersExpression("op_product_id", opt.OpProductIDs))
}

if len(opt.PlatformInfos) > 0 {
    catPlatPairs := make([][]interface{}, len(opt.PlatformInfos))
    for i, item := range opt.PlatformInfos {
        catPlatPairs[i] = []interface{}{item.PlatformCategoryID, item.PlatformID}
    }
    expr.Rules = append(expr.Rules, tools.ContainersExpression("(platform_category_id,platform_id)", catPlatPairs))
}
```

## 3. 推荐方式：快捷函数直接构建

**适用场景**：简单的单条件或固定条件过滤

```go
// 单条件查询
expr := tools.ContainersExpression("op_product_id", opProductIDs)

// 等于查询
expr := tools.EqualExpression("status", 1)
```

## 4. 推荐方式：tools.And()/tools.ExpressionAnd()/tools.ExpressionOr() 合并规则

**适用场景**：需要合并多个已有的表达式或规则

### 4.1 tools.And

`And` 接受 `filter.RuleFactory` 变参，可以同时传入 Expression 和 AtomRule：

```go
expr, err := tools.And(
    filt,
    tools.RuleEqual("bk_biz_id", c.bkBizID),
)
if err != nil {
    return nil, err
}
```

**特性**：
- 如果传入的 Expression 的 Op 也是 And，会展平其 Rules（避免不必要的嵌套）
- 返回 error，需要处理错误

### 4.2 tools.ExpressionAnd / tools.ExpressionOr

`ExpressionAnd` 和 `ExpressionOr` 只接受 `*filter.AtomRule`，用于将多个 AtomRule 组合为一个 Expression：

```go
// And 组合
expr := tools.ExpressionAnd(
    tools.RuleEqual("type", "A"),
    tools.RuleIn("status", []int{1, 2}),
)

// Or 组合
orExpr := tools.ExpressionOr(
    tools.RuleEqual("type", "A"),
    tools.RuleEqual("type", "B"),
    tools.RuleEqual("type", "C"),
)
```
```

### 4.3 嵌套 Or 条件

```go
// 复杂的 OR 条件作为子表达式
f.Rules = append(f.Rules, &filter.Expression{
    Op: filter.Or,
    Rules: []filter.RuleFactory{
        tools.RuleNotEqual("platform_category_id", enumor.PlatCatBGService),
        tools.RuleNotIn("platform_id", opt.InternalAllocPlatIDs),
    },
})
```

## 5. 可选方式：直接构建结构体

**适用场景**：静态的、固定的过滤条件

```go
expr := &filter.Expression{
    Op: filter.And,
    Rules: []filter.RuleFactory{
        tools.RuleGreaterThanEqual("stat_date", month),
        tools.RuleGreaterThan("os", 0),
    },
}
```

## 6. 错误示例

### ❌ Rules 未初始化就追加

```go
// 错误：会导致 panic: nil pointer
var expr *filter.Expression
expr.Rules = append(expr.Rules, tools.RuleEqual("name", "test"))
```

### ✅ 正确做法

```go
// 先使用 AllExpression() 初始化
expr := tools.AllExpression()
expr.Rules = append(expr.Rules, tools.RuleEqual("name", "test"))
```

## 7. Expression 快捷函数列表

| 函数 | 用途 | 返回类型 |
|------|------|----------|
| `AllExpression()` | 创建空的 And 表达式 | `*Expression` |
| `EqualExpression(field, value)` | 等于表达式 | `*Expression` |
| `ContainersExpression(field, values)` | IN 表达式 | `*Expression` |

## 8. Rule 快捷函数列表

| 函数 | 用途 | 返回类型 |
|------|------|----------|
| `RuleEqual(field, value)` | 等于规则 | `*filter.AtomRule` |
| `RuleNotEqual(field, value)` | 不等于规则 | `*filter.AtomRule` |
| `RuleGreaterThanEqual(field, value)` | 大于等于规则 | `*filter.AtomRule` |
| `RuleLessThanEqual(field, value)` | 小于等于规则 | `*filter.AtomRule` |
| `RuleIn(field, values)` | IN 规则 | `*filter.AtomRule` |
| `RuleNotIn(field, values)` | NOT IN 规则 | `*filter.AtomRule` |
| `RuleJsonOverlaps(field, values)` | JSON 数组交集规则 | `*filter.AtomRule` |
| `RuleJSONContains[T any](fieldName string, values T)` | JSON 数组包含规则 | `*filter.AtomRule` |

## 9. SQL 生成

### 9.1 使用默认配置（推荐）

```go
whereExpr, whereValue, err := expr.SQLWhereExpr(filter.DefaultSqlWhereOption)
if err != nil {
    return nil, err
}

// whereExpr: "WHERE field1 = :field1_xxx AND field2 IN (:field2_yyy)"
// whereValue: map[string]interface{}{"field1_xxx": value1, "field2_yyy": values2}
```

### 9.2 表达式校验

```go
exprOpt := filter.NewExprOption(
    filter.RuleFields(table.ObsBillColumns.ColumnTypes()),  // 允许的字段及类型
    filter.MaxInLimit(constant.OpProductIDMaxLimit),        // IN 操作符最大元素数
)
if err := listOpt.Validate(exprOpt, ctypes.NewUnlimitedPageOption()); err != nil {
    return nil, errf.NewFromErr(errf.InvalidParameter, err)
}
```

## 10. Expression vs Rule 选择原则

- **Expression 快捷函数**（如 `EqualExpression`）：返回完整的 `*Expression`，可直接用于查询
- **Rule 快捷函数**（如 `RuleEqual`）：返回 `RuleFactory`，需要添加到 Expression 中使用

**推荐模式**：

```go
// 单条件直接使用 Expression
expr := tools.ContainersExpression("op_product_id", ids)

// 多条件组合使用 Rule
expr := tools.AllExpression()
expr.Rules = append(expr.Rules, tools.RuleEqual("status", 1))
expr.Rules = append(expr.Rules, tools.RuleIn("type", types))
```
