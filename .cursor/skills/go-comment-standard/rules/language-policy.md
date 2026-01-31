# 注释语言策略

## 核心原则

**代码注释必须使用英文**，仅在特定场景允许使用中文。

## 英文注释（默认）

### 适用范围

- 函数/方法注释
- 结构体/字段注释
- 接口注释
- 包注释
- 常量/变量注释
- TODO/FIXME/Deprecated 注释
- 代码块注释

### 代码示例

```
// Less reports whether x is less than y (partial order).
func Less[T cmp.Ordered](x, y T, reverse bool) bool {
```

```go
// OpProduct defines the op product database table.
type OpProduct struct {
	// ID is the OpProduct's unique identity id.
	ID string `db:"id" json:"id"`
}
```

## 中文注释（例外情况）

### 允许使用中文的场景

1. **业务数据映射**
   - 中文名称
   - 中文错误提示
   - 中文枚举值

2. **复杂业务逻辑的补充说明**
   - 使用 `Note:` 前缀
   - 解释特殊的业务规则

3. **第三方系统字段说明**
   - 对接外部系统时的字段映射说明
   - 需要强调与外部系统的对应关系

### 代码示例

#### 业务数据映射

```
// objectTypeNameMap is the map from DetObjectType to its name.
var objectTypeNameMap = map[DetObjectType]string{
	DetObjectTypeBiz:       "业务",
	DetObjectTypeOpProduct: "运营产品",
```

#### 复杂业务逻辑补充说明

```
	// IncludeSelf 用于指定返回的父部门列表中是否包含输入部门自身的信息
	IncludeSelf bool `json:"include_self"`
```

#### 第三方系统字段说明

```32:37:pkg/thirdparty/api-gateway/cmdb/types.go
// SearchBizParams is cmdb search business parameter.
type SearchBizParams struct {
    Fields            []string     `json:"fields"`
    Page              BasePage     `json:"page"`
    BizPropertyFilter *QueryFilter `json:"biz_property_filter,omitempty"`
}
```

## 语言选择指南

| 场景 | 语言 | 示例 |
|-----|------|------|
| 函数说明 | 英文 | `// NewClient creates a new client.` |
| 参数说明 | 英文 | `// The timeout parameter specifies...` |
| 业务名称映射值 | 中文 | `"业务"`, `"运营产品"` |
| 字段注释 | 英文 | `// Name is the user name.` |
| 业务字段补充说明 | 中文 | `// 用于指定返回的父部门列表中...` |
| TODO/FIXME | 英文优先，中文可接受 | `// TODO: add validation` |
| Note 说明 | 中文可接受 | `// Note: 此处使用 VirtualDeptID` |

## 常见英文表达

### 函数注释

| 中文意图 | 英文表达 |
|---------|---------|
| 创建 | creates, creates a new |
| 获取 | gets, retrieves, fetches |
| 设置 | sets, updates |
| 验证 | validates, checks |
| 转换 | converts, transforms |
| 解析 | parses |
| 处理 | handles, processes |
| 初始化 | initializes |
| 返回 | returns |

### 字段注释

| 中文意图 | 英文表达 |
|---------|---------|
| 是 | is the |
| 表示 | represents, indicates |
| 包含 | contains, holds |
| 指定 | specifies |
| 定义 | defines |

### 常见句式

```go
// FuncName creates a new instance of Type.
// FuncName retrieves the data from database.
// FuncName validates the input parameters.
// FieldName is the unique identifier.
// FieldName specifies the configuration option.
// Package xxx provides functionality for yyy.
```

## 注意事项

1. **不要混用语言**：一个注释块内应只使用一种语言
2. **优先英文**：如果不确定，使用英文
3. **业务术语**：专业业务术语可使用中文，便于理解
4. **历史代码**：新代码不应使用中文注释，但无需修改历史中文注释（除非重构）
