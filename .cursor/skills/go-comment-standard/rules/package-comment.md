# 包注释规范

## 基本规则

1. 包注释以 `// Package xxx` 开头（xxx 为包名）
2. 描述包提供的功能
3. 包注释紧跟 `package` 语句
4. 项目不使用 `doc.go` 文件

## 标准格式

### 单行包注释

```go
// Package xxx provides description.
package xxx
```

### 多行包注释

```go
// Package xxx provides description.
// It offers feature1, feature2, and feature3.
// The package supports various use cases.
package xxx
```

## 代码示例

### 单行包注释

```
// Package sorts provides the sort related functions.
package sorts
```

```
// Package parser provides the methods for Excel parsing, which translate an Excel sheet into target objects.
package parser
```

### 多行包注释

```19:23:pkg/tools/times/time.go
// Package times provides advanced time handling and date range manipulation functionality.
// It offers flexible date range operations, time conversions, and specialized time granularity
// handling for year, month, and day levels. The package supports various time formats,
// comparisons, and SQL-compatible date range operations.
package times
```

## 包注释内容指南

### 应包含的内容

1. 包的主要功能
2. 核心能力（如有多个）
3. 适用场景（可选）

### 不应包含的内容

1. 实现细节
2. 版本历史
3. 作者信息

## 常见动词

| 动词 | 使用场景 |
|-----|---------|
| `provides` | 提供功能或服务 |
| `implements` | 实现某个接口或标准 |
| `contains` | 包含某类定义 |
| `defines` | 定义数据结构或常量 |
| `handles` | 处理某类请求或事件 |

## 示例模板

### 工具包

```go
// Package times provides advanced time handling and date range manipulation.
package times
```

### 业务逻辑包

```go
// Package cost implements cost calculation and optimization logic.
package cost
```

### DAO 包

```go
// Package user provides database operations for user entity.
package user
```

### Handler 包

```go
// Package handler handles HTTP requests for the API server.
package handler
```

### 类型定义包

```go
// Package types defines common types and structures used across the application.
package types
```

## 注意事项

1. 包注释是包文档的入口，应简洁明了
2. 复杂包可使用多行注释详细说明功能
3. 避免在包注释中引用其他包或具体类型
4. 包注释应与包名保持一致（如 `Package xxx` 中的 xxx 应与实际包名相同）
