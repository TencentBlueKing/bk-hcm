# 时间过滤模式

## 1. VarTime 系列（推荐）

VarTime 系列支持多粒度时间，是最灵活的时间过滤方式。

| 函数 | 用途 | 生成条件 |
|------|------|----------|
| `VarTimeExpression(field, vt)` | 单时间点 | `field >= start AND field < next` |
| `VarTimeRangeExpression(field, vtr)` | 时间范围 | `field >= start AND field < end+1` |
| `VarTimeBeforeRule(field, vt)` | 某时间之前 | `field < vt.Time()` |
| `VarTimeNotBeforeRule(field, vt)` | 不早于某时间 | `field >= vt.Time()` |
| `VarTimeAfterRule(field, vt)` | 某时间之后 | `field >= vt.AddByGran(1).Time()` |
| `VarTimeNotAfterRule(field, vt)` | 不晚于某时间 | `field < vt.AddByGran(1).Time()` |

### 1.1 基本用法

```go
// 时间范围过滤（最常用）
expr.Rules = append(expr.Rules, tools.VarTimeRangeExpression("fee_date", opt.DateRange))

// 单时间点过滤
expr.Rules = append(expr.Rules, tools.VarTimeExpression("stat_date", vt))
```

### 1.2 日期范围交叉查询

常用于查询有效期与给定范围有交集的记录：

```go
// 查询有效期与给定范围有交集的记录
// 条件：start_date <= 范围结束 AND end_date >= 范围开始
dateExpr := &filter.Expression{
    Op: filter.And,
    Rules: []filter.RuleFactory{
        filter.VarTimeNotAfterRule("start_date", r.DateRange.End),   // 开始日期 <= 范围结束
        filter.VarTimeNotBeforeRule("end_date", r.DateRange.Start), // 结束日期 >= 范围开始
    },
}
```

### 1.3 边界条件说明

VarTime 生成的是**左闭右开**区间 `[start, next)`：

```
VarTime = 2025-01
  -> [2025-01-01 00:00:00, 2025-02-01 00:00:00)

VarTime = 2025-01-15
  -> [2025-01-15 00:00:00, 2025-01-16 00:00:00)
```

## 2. Month/Day 系列

### 2.1 月级别函数

| 函数 | 用途 | 生成条件 |
|------|------|----------|
| `MonthExpression(field, month)` | 单月过滤 | `field >= month_start AND field < next_month_start` |
| `MonthRangeExpression(field, mr)` | 月范围过滤 | `field >= start AND field < end_next_month` |

```go
// 单月过滤
expr, err := tools.ExpressionAnd(opt, tools.MonthExpression("fee_date", month))

// 月范围过滤
expr.Rules = append(expr.Rules, tools.MonthRangeExpression("fee_date", monthRange))
```

### 2.2 日级别函数

| 函数 | 用途 | 生成条件 |
|------|------|----------|
| `DayExpression(field, day)` | 单日过滤 | `field >= day_start AND field < next_day_start` |
| `DayRangeExpression(field, dr)` | 日范围过滤 | `field >= start AND field < end_next_day` |

```go
// 单日过滤
expr.Rules = append(expr.Rules, tools.DayExpression("fee_date", day))

// 日范围过滤
expr.Rules = append(expr.Rules, tools.DayRangeExpression("fee_date", dayRange))
```

### 2.3 DateRange 函数

| 函数 | 用途 | 粒度 |
|------|------|------|
| `DateRangeExpression(field, dr)` | 月粒度日期范围 | 月 |
| `DayDateRangeExpression(field, dr)` | 日粒度日期范围 | 日 |

```go
// 月粒度日期范围
expr, err := tools.DateRangeExpression("fee_date", dateRange)
if err != nil {
    return nil, err
}

// 日粒度日期范围
expr, err := tools.DayDateRangeExpression("fee_date", dateRange)
```

## 3. Between 表达式

用于通用的范围查询（含边界）：

```go
// 基本范围查询（含两端边界）
expr, err := tools.ExpressionAnd(
    tools.BetweenExpression("fee_date", yr.Start.FormattedYear(), yr.End.FormattedYear()),
    f,
)
// SQL: fee_date >= :start AND fee_date <= :end
```

## 4. 时间过滤选择指南

| 场景 | 推荐函数 |
|------|----------|
| 通用时间范围查询 | `VarTimeRangeExpression` |
| 单个月份查询 | `MonthExpression` |
| 单个日期查询 | `DayExpression` |
| 多月范围查询 | `MonthRangeExpression` |
| 多日范围查询 | `DayRangeExpression` |
| 有效期交叉查询 | `VarTimeNotAfterRule` + `VarTimeNotBeforeRule` |
| 某时间点之前/之后 | `VarTimeBeforeRule` / `VarTimeAfterRule` |
| 通用数值范围 | `BetweenExpression` |

## 5. 时间格式要求

时间字符串格式必须为 `2006-01-02 15:04:05`（Go 标准格式）：

```go
// ✅ 正确
tools.RuleGreaterThanEqual("created_at", "2024-01-01 00:00:00")

// ❌ 错误
tools.RuleGreaterThanEqual("created_at", "2024-01-01")  // 缺少时间部分
tools.RuleGreaterThanEqual("created_at", "2024/01/01 00:00:00")  // 分隔符错误
```

## 6. 注意事项

1. **时区处理**：`created_at` 和 `updated_at` 字段会自动转换为 UTC 时间
2. **边界条件**：时间过滤函数默认生成**左闭右开**区间
3. **类型匹配**：确保 VarTime 对象的粒度与查询场景匹配
