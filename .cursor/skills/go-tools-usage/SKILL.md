---
name: go-tools-usage
description: tools 包完整使用指南，用于构建类型安全的数据库查询条件。当需要了解 JSON 操作符、组合字段、ES 查询等高级特性时使用。
---

# Filter 包使用指南

> 核心规则见 Rule: `tools-usage.mdc`，本 Skill 包含完整规范和高级特性。

## 快速参考

| 场景 | 规则文件 |
|------|----------|
| Expression 构建基础 | [rules/expression-building.md](rules/expression-building.md) |
| 操作符使用 | [rules/operators.md](rules/operators.md) |

## 核心概念

| 概念 | 说明 |
|------|------|
| `Expression` | 查询表达式，包含逻辑操作符（And/Or）和规则列表 |
| `AtomRule` | 原子规则，包含字段、操作符和值 |
| `RuleFactory` | 规则工厂接口，Expression 和 AtomRule 都实现该接口 |
| `OpType` | 操作符类型（eq、neq、gt、gte、lt、lte、in、nin、cs、cis 等） |

## 核心原则

1. **使用快捷函数优先**：优先使用 `AllExpression()`、`RuleEqual()` 等快捷函数，避免直接构建结构体
2. **动态构建用 AllExpression()**：需要根据条件动态追加规则时，以 `AllExpression()` 为起点
3. **参数绑定防注入**：filter 包自动使用参数绑定，无需担心 SQL 注入
4. **时间过滤用 DateRange**：推荐使用 `DateRangeExpression` 或 `RuleGreaterThanEqual`/`RuleLessThanEqual` 组合

## 快速示例

### 简单查询

```go
// 单条件
expr := tools.ContainersExpression("op_product_id", productIDs)

// 多条件动态构建
expr := tools.AllExpression()
expr.Rules = append(expr.Rules, tools.RuleEqual("status", 1))
if len(types) > 0 {
    expr.Rules = append(expr.Rules, tools.RuleIn("type", types))
}

// 生成 SQL
whereExpr, whereValue, err := expr.SQLWhereExpr(filter.DefaultSqlWhereOption)
```

### 时间范围查询

```go
// 使用 DateRangeExpression（推荐）
dateExpr, err := tools.DateRangeExpression("fee_date", dateRange)
if err != nil {
    return nil, err
}
expr.Rules = append(expr.Rules, dateExpr)

// 直接使用 RuleGreaterThanEqual / RuleLessThanEqual
expr.Rules = append(expr.Rules, tools.RuleGreaterThanEqual("stat_date", startTime))
expr.Rules = append(expr.Rules, tools.RuleLessThanEqual("stat_date", endTime))
```

### 规则合并

```go
// And 合并
expr, err := tools.ExpressionAnd(baseExpr, tools.RuleEqual("status", 1))

// Or 合并
orExpr, err := tools.ExpressionOr(
    tools.RuleEqual("type", "A"),
    tools.RuleEqual("type", "B"),
)
```

## 常见问题

### Q: Expression 和 Rule 快捷函数如何选择？

- **Expression 快捷函数**（如 `EqualExpression`）：返回完整的 `*Expression`，可直接用于查询
- **Rule 快捷函数**（如 `RuleEqual`）：返回 `RuleFactory`，需要添加到 Expression 中使用

推荐：单条件用 Expression，多条件组合用 Rule。

### Q: IN 操作符有数量限制吗？

有，默认最大 200 个元素。可通过 `ExprOption` 配置：

```go
exprOpt := filter.NewExprOption(filter.MaxInLimit(500))
```

### Q: 如何处理空条件？

使用 `AllExpression()` 创建空表达式，空表达式生成的 SQL 不包含 WHERE 子句。

## 关键文件

| 文件 | 说明 |
|------|------|
| `pkg/runtime/filter/expression.go` | Expression 和 AtomRule 定义 |
| `pkg/runtime/filter/operator.go` | 操作符定义和实现 |
