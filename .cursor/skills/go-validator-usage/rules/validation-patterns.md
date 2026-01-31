# 验证模式与最佳实践

## 验证方法组合使用

### Struct Tag + 自定义验证

在 `Validate()` 方法中，通常先进行 Struct Tag 验证，然后进行自定义验证。

**✅ 正确示例**：

```go
func (req *BaseDeviceReq) Validate() error {
	// 1. Struct Tag 验证
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	// 2. 业务逻辑验证
	for _, bizID := range req.BizIDs {
		if bizID <= 0 {
			return errors.New("invalid biz id, should be > 0")
		}
	}
	// ... 更多验证逻辑

	return nil
}
```

### Struct Tag + 自定义验证函数

结合使用 Struct Tag 和自定义验证函数，先进行基础验证，再进行格式验证。

**✅ 正确示例**：

```go
func (r *CostOptimAuditTicketSubTicket) Validate() error {
	// 1. Struct Tag 验证
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	// 2. 业务逻辑验证
	if r.DateRange != nil {
		opt := &times.DayDateRangeValidateOption{MaxPeriodMonth: 9}
		if err := r.DateRange.Validate(constant.MonthTimeGrans, opt); err != nil {
			return err
		}
	}

	// 3. 自定义格式验证
	if err := validator.ValidateName(r.Operation); err != nil {
		return err
	}
	if utf8.RuneCountInString(r.Operation) > 64 {
		return errors.New("operation length should be <= 64")
	}

	if err := validator.ValidateName(r.Title); err != nil {
		return err
	}
	if utf8.RuneCountInString(r.Title) > 64 {
		return errors.New("title length should be <= 64")
	}

	return nil
}
```

## 嵌套结构体验证

对于包含嵌套结构体的请求，需要递归调用子结构体的 `Validate()` 方法。

**✅ 正确示例**：

```go
func (r *CheckCostOptimAuditTicketReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	// 递归验证嵌套结构体
	for i := range r.Tickets {
		if err := r.Tickets[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}
```

**✅ 正确示例**（嵌套结构体数组）：

```go
for _, childBiz := range req.ChildBizs {
	if err := childBiz.Validate(); err != nil {
		return err
	}
}
```

## 数据表结构体验证模式

数据表结构体的验证方法通常结合 Struct Tag 和自定义验证函数。

**✅ 正确示例**：

```go
func (sc SystemConfig) ValidateCreate() error {
	// 1. Struct Tag 验证
	if err := validator.Validate.Struct(sc); err != nil {
		return err
	}

	// 2. 业务逻辑验证
	if len(sc.Type) == 0 {
		return errors.New("invalid type, can not be empty")
	}

	// 3. 自定义格式验证
	if err := validator.ValidateKey(sc.Type, true); err != nil {
		return err
	}

	if err := validator.ValidateKey(sc.Name, true); err != nil {
		return err
	}

	if err := validator.ValidateMemo(sc.Memo, false); err != nil {
		return err
	}

	return nil
}
```

## 最佳实践

### 1. 统一验证入口

所有请求结构体都应实现 `Validate()` 方法，并在其中统一调用 `validator.Validate.Struct()`。

**✅ 正确**：
```go
func (req *SomeReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}
	// 其他验证...
	return nil
}
```

**❌ 错误**：
```go
// 缺少 Validate() 方法，直接在业务代码中验证
func SomeHandler(req *SomeReq) {
	// 直接验证，没有统一入口
	if req.Name == "" {
		return errors.New("name is required")
	}
}
```

### 2. 分层验证

先进行 Struct Tag 验证（基础规则），再进行业务逻辑验证（复杂规则）。

**✅ 正确顺序**：
1. Struct Tag 验证（`validator.Validate.Struct()`）
2. 业务逻辑验证（自定义规则）
3. 格式验证（自定义验证函数）

### 3. 错误信息清晰

自定义验证函数应返回清晰的错误信息，包含字段名和具体规则。

**✅ 正确**：
```go
return fmt.Errorf("invalid name: %s, only allows to include chinese、english、numbers...", name)
```

**❌ 错误**：
```go
return errors.New("invalid name") // 信息不够清晰
```

### 4. 合理使用 required

区分必填字段和可选字段，可选字段使用 `omitempty`。

**✅ 正确**：
```go
type SomeReq struct {
	Name string `json:"name" validate:"required"`
	Memo *string `json:"memo" validate:"omitempty"`
}
```

### 5. 嵌套结构体递归验证

对于嵌套结构体，在父结构体的 `Validate()` 方法中递归调用子结构体的 `Validate()` 方法。

**✅ 正确**：
```go
func (req *ParentReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for i := range req.Children {
		if err := req.Children[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}
```

## 常见问题

### Q: 什么时候使用 Struct Tag，什么时候使用自定义验证函数？

**A**: 
- **Struct Tag**：适用于通用的验证规则（必填、长度、数值范围等），在请求结构体中广泛使用
- **自定义验证函数**：适用于项目特定的格式要求（如名称格式、Key 格式），通常在数据表结构体或需要特定格式验证的场景使用

### Q: 验证失败时如何返回友好的错误信息？

**A**: `validator.Validate.Struct()` 返回的错误可以直接返回，go-playground/validator 会提供详细的字段级错误信息。对于自定义验证函数，应返回清晰的错误信息，例如：
```go
return fmt.Errorf("invalid name: %s, only allows to include chinese、english、numbers...", name)
```
