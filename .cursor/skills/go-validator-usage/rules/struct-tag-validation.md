# Struct Tag 验证

## 基本用法

在请求结构体中定义 `validate` 标签，然后在 `Validate()` 方法中调用 `validator.Validate.Struct()`。

**规则**：
- 请求结构体必须实现 `Validate()` 方法
- `Validate()` 方法中首先调用 `validator.Validate.Struct(req)` 进行基础验证
- 然后进行业务逻辑相关的自定义验证

**✅ 正确示例**：

```go
type OpDataRangeReq struct {
	OpProductIDs []int64                `json:"op_product_ids" validate:"required,max=20"`
	DateRange    *times.DateRange       `json:"date_range" validate:"required"`
	Type         enumor.PerUserCostType `json:"type" validate:"required"`
}

func (req *OpDataRangeReq) Validate() (*times.MonthRange, error) {
	if err := validator.Validate.Struct(req); err != nil {
		return nil, err
	}

	mr, err := req.DateRange.ValidateAsMonth()
	if err != nil {
		return nil, err
	}

	if err = req.Type.Validate(); err != nil {
		return nil, err
	}

	return mr, nil
}
```

**✅ 正确示例**（仅 Struct 验证）：

```go
type InfrastructureBillReq struct {
	FeeDate       string  `json:"fee_date" validate:"required"`
	BgID          int64   `json:"bg_id" validate:"required"`
	BgName        string  `json:"bg_name" validate:"required"`
	DeptID        int64   `json:"dept_id" validate:"required"`
	DeptName      string  `json:"dept_name" validate:"required"`
	PlProductID   int64   `json:"pl_product_id" validate:"required"`
	PlProductName string  `json:"pl_product_name" validate:"required"`
	OpProductID   int64   `json:"op_product_id" validate:"required"`
	OpProductName string  `json:"op_product_name" validate:"required"`
	Cost          float64 `json:"cost" validate:"required"`
	DeviceCost    float64 `json:"device_cost" validate:"required"`
	IdcCost       float64 `json:"idc_cost" validate:"required"`
	CdnCost       float64 `json:"cdn_cost" validate:"required"`
	OtherCost     float64 `json:"other_cost" validate:"required"`
}

func (req *InfrastructureBillReq) Validate() error {
	return validator.Validate.Struct(req)
}
```

**❌ 错误示例**（缺少 Validate 方法）：

```go
type SomeReq struct {
	Name string `json:"name" validate:"required"`
	// Missing Validate() method
}
```

## 常用 Struct Tag 规则

| Tag | 说明 | 示例 |
|-----|------|------|
| `required` | 必填字段 | `validate:"required"` |
| `max=n` | 最大长度/值 | `validate:"max=64"` |
| `min=n` | 最小长度/值 | `validate:"min=1"` |
| `gt=0` | 大于 0 | `validate:"gt=0"` |
| `dive` | 验证切片/数组元素 | `validate:"required,dive,gt=0"` |
| `omitempty` | 可选字段（为空时跳过验证） | `validate:"omitempty"` |

**✅ 正确示例**（组合使用）：

```go
type CommonNotificationReq struct {
	Recipients []string `json:"recipients" validate:"gt=0,dive,required"`
	Content    string   `json:"content" validate:"required,max=4096"`
}

func (r *CommonNotificationReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}
	return nil
}
```

```go
type CheckCostOptimAuditTicketReq struct {
	BizID   int64                           `json:"biz_id" validate:"required,gt=0"`
	Tickets []CostOptimAuditTicketSubTicket `json:"tickets" validate:"required,min=1,max=20,dive"`
}

func (r *CheckCostOptimAuditTicketReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for i := range r.Tickets {
		if err := r.Tickets[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}
```

## 在数据表结构体中的使用

数据表结构体的验证方法中也可以使用 `validator.Validate.Struct()`，但通常需要结合自定义验证函数。

**✅ 正确示例**：

```go
func (sc SystemConfig) ValidateUpdate() error {
	if err := validator.Validate.Struct(sc); err != nil {
		return err
	}
	// ... 其他验证逻辑
}
```

## 常见问题

### Q: `required` 和 `omitempty` 的区别？

**A**:
- `required`：字段必须存在且不能为空值（对于字符串不能为空字符串，对于指针不能为 nil）
- `omitempty`：如果字段为空值，则跳过该字段的验证（常用于可选字段）

### Q: 如何验证切片/数组元素？

**A**: 使用 `dive` 标签，例如：
```go
Recipients []string `json:"recipients" validate:"gt=0,dive,required"`
```
表示：切片长度 > 0，且每个元素必须非空。
