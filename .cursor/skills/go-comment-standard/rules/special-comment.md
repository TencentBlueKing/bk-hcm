# 特殊注释规范

## TODO 注释

### 规则

1. 格式：`// TODO: description`
2. 冒号后加空格
3. 说明待完成的功能或改进

### 标准格式

```go
// TODO: description of what needs to be done
```

### 代码示例

```72:73:pkg/dal/table/op_product.go
// OpProduct defines the op product database table.
// TODO: remove ObsDeptID and ObsDeptName fields, which are redundant with PlanProduct table.
type OpProduct struct {
```

```
// TODO: 评估是否能统一迁移到新接口
```

### 使用场景

- 标记需要后续实现的功能
- 标记需要优化的代码
- 标记需要移除的临时代码
- 标记待讨论的设计决策

## FIXME 注释

### 规则

1. 格式：`// FIXME: description`
2. 冒号后加空格
3. 说明需要修复的问题

### 标准格式

```go
// FIXME: description of the bug or issue
```

### 代码示例

```
	// FIXME: max in limit
```

### 使用场景

- 标记已知的 bug
- 标记需要修复的逻辑错误
- 标记性能问题
- 标记安全隐患

## Deprecated 注释

### 规则

1. 格式：`// Deprecated: reason and alternative`
2. 说明弃用原因和替代方案
3. 放在被弃用项的注释上方或后面

### 标准格式

```go
// FuncName does something.
//
// Deprecated: Use NewFuncName instead. This function will be removed in v2.0.
func FuncName() {}
```

或

```go
// Deprecated: Use NewFuncName instead.
func FuncName() {}
```

### 代码示例

```go
// CreateUser creates a new user.
//
// Deprecated: Use CreateUserV2 instead, which supports more options.
func CreateUser(name string) (*User, error) {
	return CreateUserV2(name, nil)
}
```

### 使用场景

- 标记即将移除的函数/方法
- 标记即将移除的类型/常量
- 指导用户迁移到新 API

## NOTE 注释

### 规则

1. 格式：`// Note: important information`
2. 用于强调重要信息或注意事项
3. 可使用中文说明复杂业务逻辑

### 标准格式

```go
// Note: this function is not thread-safe, caller must handle synchronization.
```

### 代码示例

```go
// Note: 此处使用 VirtualDeptID 而非 DeptID，因为规划产品对应的部门需要通过虚拟部门查找
```

### 使用场景

- 说明重要的实现细节
- 说明使用限制
- 说明特殊的业务逻辑

## 行内注释

### 规则

1. 用于简短说明代码行的目的
2. 与代码至少保持两个空格
3. 避免冗余说明

### 代码示例

```go
count := 0  // T-2 day adjustment
```

```
	// Refresh ETL data
	if err := e.refreshGPUProductDataETLData(cts.Kit, &dateRange, providerData, amountResult); err != nil {
```

### 使用场景

- 解释复杂的计算逻辑
- 说明魔法数字的含义
- 简短说明代码块的目的

## 代码块注释

### 规则

1. 放在代码块上方
2. 说明后续代码块的目的
3. 使用单行 `//` 格式

### 代码示例

```go
// Validate input parameters
if req.Name == "" {
	return errors.New("name is required")
}

// Query user from database
user, err := d.userDAO.GetByName(kt, req.Name)
if err != nil {
	return err
}

// Update user status
user.Status = StatusActive
if err := d.userDAO.Update(kt, user); err != nil {
	return err
}
```

## 注释优先级

当需要同时使用多种特殊注释时，建议按以下顺序排列：

1. 函数/类型注释（功能说明）
2. Deprecated（弃用说明）
3. Note（注意事项）
4. TODO/FIXME（待办事项）

### 代码示例

```go
// CreateUser creates a new user with the given name.
//
// Deprecated: Use CreateUserV2 instead.
//
// Note: This function does not validate email format.
//
// TODO: Add email validation before removal.
func CreateUser(name string) (*User, error) {
```
