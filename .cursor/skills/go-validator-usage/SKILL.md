---
name: go-validator-usage
description: validator 工具包使用指南，包括 Struct Tag 验证、自定义验证函数、验证模式组合。当编写请求结构体验证、数据表结构体验证、或使用 validator 包进行数据校验时使用。
---

# Validator 工具包使用规范

## 快速参考

| 场景 | 规则文件 |
|-----|---------|
| Struct Tag 验证 | [rules/struct-tag-validation.md](rules/struct-tag-validation.md) |
| 自定义验证函数 | [rules/custom-validation.md](rules/custom-validation.md) |
| 验证模式与最佳实践 | [rules/validation-patterns.md](rules/validation-patterns.md) |

## 核心原则

1. **统一验证入口**：所有请求结构体都应实现 `Validate()` 方法，并在其中统一调用 `validator.Validate.Struct()`
2. **分层验证**：先进行 Struct Tag 验证（基础规则），再进行业务逻辑验证（复杂规则）
3. **合理选择验证方式**：通用规则用 Struct Tag，特定格式用自定义验证函数
4. **错误信息清晰**：自定义验证函数应返回清晰的错误信息，包含字段名和具体规则

## 工具概览

`pkg/criteria/validator` 包提供两类验证功能：

1. **Struct Tag 验证**：基于 go-playground/validator 的结构体标签验证
   - 使用 `validator.Validate.Struct(req)` 进行验证
   - 支持 `required`、`max`、`min`、`gt`、`dive`、`omitempty` 等标签

2. **自定义验证函数**：针对项目特定格式的验证
   - `ValidateName()` - 验证名称格式（支持中文、英文、数字等）
   - `ValidateKey()` - 验证 Key 格式（仅英文、数字、下划线等）
   - `ValidateValEng()` - 验证英文值格式
   - `ValidateMemo()` / `ValidateMemoFmt()` - 验证备注格式

## 常见问题

### Q: 什么时候使用 Struct Tag，什么时候使用自定义验证函数？

**A**: 
- **Struct Tag**：适用于通用的验证规则（必填、长度、数值范围等），在请求结构体中广泛使用
- **自定义验证函数**：适用于项目特定的格式要求（如名称格式、Key 格式），通常在数据表结构体或需要特定格式验证的场景使用

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

### Q: ValidateMemo 为什么接受 `*string` 而不是 `string`？

**A**: 因为备注字段通常是可选的，使用指针可以区分"未设置"（nil）和"空字符串"（""）。在 `required=false` 时，两者都视为有效；在 `required=true` 时，两者都视为无效。

### Q: 验证失败时如何返回友好的错误信息？

**A**: `validator.Validate.Struct()` 返回的错误可以直接返回，go-playground/validator 会提供详细的字段级错误信息。对于自定义验证函数，应返回清晰的错误信息，例如：
```go
return fmt.Errorf("invalid name: %s, only allows to include chinese、english、numbers...", name)
```

## 相关规范

本规范与以下规范有关联，必要时请一并参考：
- [go-naming-convention](../go-naming-convention/SKILL.md) - 请求结构体命名规范
- [go-error-handling](../go-error-handling/SKILL.md) - 错误处理规范
