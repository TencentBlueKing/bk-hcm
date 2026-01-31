# 结构体和字段注释规范

## 结构体注释

### 规则

1. 注释**必须**以结构体名称开头
2. 使用 `is`、`defines`、`represents` 或 `holds` 描述
3. 请求/响应结构体说明是什么类型的请求/响应

### 标准格式

```go
// StructName is/defines/represents description.
type StructName struct {
```

### 代码示例

```
// ListDeptParentRelReq is the list dept parent relation request.
type ListDeptParentRelReq struct {
	DeptIDs []string `json:"dept_ids" validate:"required,dive,gt=0"`
	// IncludeSelf 用于指定返回的父部门列表中是否包含输入部门自身的信息
	IncludeSelf bool `json:"include_self"`
}
```

```
// Client is the ETL service's client instance.
type Client struct {
```

```
// OpProduct defines the op product database table.
// TODO: remove ObsDeptID and ObsDeptName fields, which are redundant with PlanProduct table.
type OpProduct struct {
	// ID is the OpProduct's unique identity id.
	ID string `db:"id" json:"id" validate:"max=16"`
```

### 常见动词

| 动词 | 使用场景 | 示例 |
|-----|---------|------|
| `is` | 通用描述 | `// Client is the HTTP client.` |
| `defines` | 数据库表、配置 | `// OpProduct defines the op product database table.` |
| `represents` | 数据模型 | `// User represents a system user.` |
| `holds` | 容器类结构体 | `// Cache holds cached data.` |

## 字段注释

### 规则

1. 字段注释**必须**放在字段上方
2. 以字段名开头（如 `// ID is ...`）
3. 枚举类型字段应说明可选值
4. 仅在特殊说明时使用行尾注释

### 标准格式

```go
type Example struct {
	// FieldName is the field description.
	FieldName string `json:"field_name"`
	
	// Status is the status, enumeration values such as: pending/running/completed.
	Status string `json:"status"`
}
```

### 代码示例（字段上方注释）

```74:91:pkg/dal/table/op_product.go
	// ID is the OpProduct's unique identity id.
	ID string `db:"id" json:"id" validate:"max=16"`
	// OpProductID is the operating product ID.
	OpProductID int64 `db:"op_product_id" json:"op_product_id"`
	// Name is the product name.
	Name string `db:"name" json:"name" validate:"max=64"`
	// CostEnv is the cost environment, enumeration values such as: idc/oversea/global/unassigned.
	CostEnv enumor.BizCostEnvironment `db:"cost_env" json:"cost_env" validate:"max=64"`
```

### 代码示例（行尾注释 - 仅特殊情况）

```
	VirtualDeptID   int64  `json:"VirtualDeptId"` // 虚拟部门ID（注意查找规划产品对应的部门ID时使用此字段，而不是用上面的DeptId）
```

> 注意：行尾注释仅用于需要特别强调的补充说明，且允许使用中文（参见语言策略）

## 请求/响应结构体

### 规则

1. 请求结构体以 `Req` 后缀命名，注释说明是什么类型的请求
2. 响应结构体以 `Resp` 后缀命名，注释说明是什么类型的响应

### 代码示例

```go
// CreateUserReq is the create user request.
type CreateUserReq struct {
	// Name is the user name.
	Name string `json:"name" validate:"required"`
	// Email is the user email address.
	Email string `json:"email" validate:"required,email"`
}

// CreateUserResp is the create user response.
type CreateUserResp struct {
	// ID is the created user ID.
	ID string `json:"id"`
}
```

## 嵌入字段

### 规则

1. 嵌入字段通常不需要注释
2. 如需说明嵌入原因，可添加注释

### 代码示例

```go
type User struct {
	// BaseModel provides common fields like ID, CreatedAt, UpdatedAt.
	BaseModel
	
	// Name is the user name.
	Name string `json:"name"`
}
```
