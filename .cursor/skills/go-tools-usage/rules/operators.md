# 操作符使用指南

## 1. 基本操作符

| 操作符 | 常量 | SQL 表示 | 值类型要求 |
|--------|------|----------|------------|
| 等于 | `Equal` (`eq`) | `=` | 基本类型 |
| 不等于 | `NotEqual` (`neq`) | `!=` | 基本类型 |
| 大于 | `GreaterThan` (`gt`) | `>` | 数值或时间字符串 |
| 大于等于 | `GreaterThanEqual` (`gte`) | `>=` | 数值或时间字符串 |
| 小于 | `LessThan` (`lt`) | `<` | 数值或时间字符串 |
| 小于等于 | `LessThanEqual` (`lte`) | `<=` | 数值或时间字符串 |
| 包含 | `In` (`in`) | `IN` | 数组/切片 |
| 不包含 | `NotIn` (`nin`) | `NOT IN` | 数组/切片 |

### 1.1 使用示例

```go
// 等于
tools.RuleEqual("name", "test")
// SQL: name = :name_xxx

// 不等于
tools.RuleNotEqual("status", 0)
// SQL: status != :status_xxx

// 大于等于
tools.RuleGreaterThanEqual("fee_date", "2024-01-01 00:00:00")
// SQL: fee_date >= :fee_date_xxx

// 小于等于
tools.RuleLessThanEqual("util_rate", 20)
// SQL: util_rate <= :util_rate_xxx

// IN
tools.RuleIn("id", []int64{1, 2, 3})
// SQL: id IN (:id_xxx)

// NOT IN
tools.RuleNotIn("status", []int{0, 1})
// SQL: status NOT IN (:status_xxx)
```

## 2. 模糊查询操作符

| 操作符 | 常量 | SQL 表示 | 说明 |
|--------|------|----------|------|
| 包含（大小写敏感） | `ContainsSensitive` (`cs`) | `field LIKE BINARY :placeholder` | 区分大小写 |
| 包含（大小写不敏感） | `ContainsInsensitive` (`cis`) | `LOWER(field) LIKE :placeholder` | 不区分大小写（推荐） |

> **注意**：模糊查询的值会自动添加前后 `%` 通配符，如传入 `"test"` 会变成 `"%test%"`

### 2.1 使用示例

```go
// 大小写敏感的模糊匹配（较少使用）
expr.Rules = append(expr.Rules, &filter.AtomRule{
    Field: "name",
    Op:    filter.ContainsSensitive.Factory(),
    Value: "Test",
})
// SQL: name LIKE BINARY :name_xxx
// 值: "%Test%"
```

## 3. JSON 操作符

| 操作符 | 常量 | 用途 | 值类型 |
|--------|------|------|--------|
| JSON 等于 | `JSONEqual` (`json_eq`) | JSON 字段等于某值 | 基本类型 |
| JSON 包含 | `JSONIn` (`json_in`) | JSON 字段值在数组中 | 数组 |
| JSON 数组包含 | `JSONContains` (`json_contains`) | JSON 数组包含某元素 | 基本类型 |
| JSON 重叠 | `JSONOverlaps` (`json_overlaps`) | JSON 数组有交集 | 数组 |
| JSON 路径存在 | `JSONContainsPath` (`json_contains_path`) | JSON 包含某路径 | 字符串 |
| JSON 路径不存在 | `JSONNotContainsPath` (`json_not_contains_path`) | JSON 不包含某路径 | 字符串 |
| JSON 数组长度 | `JSONLength` (`json_length`) | JSON 数组长度等于 | 数值 |

### 3.1 JSON 字段名规则

使用 `.` 分隔嵌套字段，如 `extension.vpc.id`

### 3.2 使用示例

```go
// JSON 字段等于
expr := &filter.Expression{
    Op: filter.And,
    Rules: []filter.RuleFactory{
        &filter.AtomRule{
            Field: "extension.vpc.id",
            Op:    filter.JSONEqual.Factory(),
            Value: 3,
        },
    },
}
// SQL: extension->>"$.vpc.id" = :extensionvpcid

// JSON 数组包含某元素
expr := &filter.Expression{
    Op: filter.And,
    Rules: []filter.RuleFactory{
        &filter.AtomRule{
            Field: "managers",
            Op:    filter.JSONContains.Factory(),
            Value: "Jim",
        },
    },
}
// SQL: JSON_CONTAINS(managers, JSON_ARRAY(:managers_xxx))

// JSON 数组交集（使用快捷函数）
expr.Rules = append(expr.Rules, tools.RuleJsonOverlaps("tags", []string{"tag1", "tag2"}))
// SQL: JSON_OVERLAPS(tags, JSON_ARRAY(:xxx))

// JSON 路径存在
expr.Rules = append(expr.Rules, &filter.AtomRule{
    Field: "extension",
    Op:    filter.JSONContainsPath.Factory(),
    Value: "vpc_id",
})
// SQL: JSON_CONTAINS_PATH(extension, 'one', '$.vpc_id')
```

## 5. 操作符限制

### 5.1 IN/NOT IN 元素数量限制

默认最大 200 个元素，可通过 ExprOption 配置：

```go
exprOpt := filter.NewExprOption(
    filter.MaxInLimit(500),        // 自定义 IN 限制
    filter.MaxNotInLimit(500),     // 自定义 NOT IN 限制
)
```

### 5.2 规则数量限制

默认最大 10 条规则，ES 查询默认最大 100 条：

```go
exprOpt := filter.NewExprOption(
    filter.MaxRulesLimit(50),      // 自定义规则数量限制
)
```

## 6. 操作符值类型说明

| 值类型 | 说明 | 适用操作符 |
|--------|------|------------|
| 基本类型 | string, int, int64, float64, bool | eq, neq, json_eq, json_contains |
| 数值或时间字符串 | 数值类型或 `2006-01-02 15:04:05` 格式的时间字符串 | gt, gte, lt, lte |
| 数组/切片 | `[]T` 其中 T 为基本类型 | in, nin, json_in, json_overlaps |
| 字符串 | 非空字符串 | cs, cis, json_contains_path |

## 7. 时间字段特殊处理

`created_at` 和 `updated_at` 字段会自动转换为 UTC 时间处理（兼容 MySQL 8.0.19 之前版本）。

```go
// 时间字段会自动处理时区转换
expr.Rules = append(expr.Rules, tools.RuleGreaterThanEqual("created_at", "2024-01-01 00:00:00"))
```
