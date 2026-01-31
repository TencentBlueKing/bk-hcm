# 自定义验证函数

## ValidateName - 验证名称格式

**用途**：验证资源名称格式，支持中文、英文、数字、括号、下划线、连字符。

**规则**：
- 长度必须 >= 1
- 只能包含：中文、英文、数字、中英文括号、下划线(_)、连字符(-)
- 必须以中文/英文/数字开头，以中文/英文/数字/右括号结尾

**函数签名**：
```go
func ValidateName(name string) error
```

**✅ 正确示例**：

```go
if err := validator.ValidateName(r.Operation); err != nil {
	return err
}
```

**✅ 有效格式示例**（来自测试用例）：

```go
// Valid English Name: "validName123"
// Valid Chinese Name: "有效名123"
// Valid Name with Hyphen: "name-with-hyphen"
// Valid Name with Underscore: "name_with_underscore"
```

**❌ 无效格式示例**：

```go
// From: pkg/criteria/validator/name_test.go:48-50
// Invalid Format with Special Characters: "!Invalid#"
// Empty Name: ""
```

## ValidateKey - 验证 Key 格式

**用途**：验证资源 Key 格式，仅支持英文、数字、下划线、连字符、点、空格。

**规则**：
- 如果 `required=true`，则不能为空
- 如果 `required=false`，空字符串视为有效
- 只能包含：英文、数字、下划线(_)、连字符(-)、点(.)、空格( )
- 必须以英文/数字开头和结尾

**函数签名**：
```go
func ValidateKey(key string, required bool) error
```

**✅ 正确示例**：

```go
if err := validator.ValidateKey(sc.Type, true); err != nil {
	return err
}
```

**✅ 有效格式示例**：

```go
// Valid Key: "dadadadad_-.sadd"
// Valid Key with Numbers: "123"
// Valid Key with English: "abc"
```

**❌ 无效格式示例**：

```go
// Invalid Key with Questionmark: "a?_-.b"
// Invalid Key Starting with Special Character: "-.abc"
// Invalid Key Ending with Special Character: "abc-."
```

## ValidateValEng - 验证英文值格式

**用途**：验证英文值格式，支持英文、数字、特殊字符。

**规则**：
- 如果 `required=true`，则不能为空
- 如果 `required=false`，空字符串视为有效
- 只能包含：英文、数字、特殊字符（`-_=+:;'"~!@#$%^&*[]{}.,?<>`）、空格
- 必须以英文/数字开头和结尾

**函数签名**：
```go
func ValidateValEng(valEng string, required bool) error
```

**✅ 正确示例**：

```go
if err := validator.ValidateValEng(om.CategoryName, true); err != nil {
	return err
}
```

**✅ 有效格式示例**：

```go
// Valid English Value: "zxczxkc!-?.'jzk"
// Valid Numbers Only: "123"
// Valid English Only: "abc"
```

**❌ 无效格式示例**：

```go
// From: pkg/criteria/validator/val_test.go:42-46
// Invalid Value with Chinese Characters: "zxczxkc!-?.'你好jzk"
// Value Starting with Special Character: "!abc"
// Value Ending with Special Character: "abc!"
```

## ValidateMemo / ValidateMemoFmt - 验证备注格式

**用途**：验证备注格式，支持中文、英文、数字、标点符号。

**规则**：
- `ValidateMemo` 包含格式验证和长度验证（<= 256 字符）
- `ValidateMemoFmt` 仅验证格式
- 如果 `required=true`，nil 或空字符串视为无效
- 如果 `required=false`，nil 或空字符串视为有效
- 只能包含：中文、英文、数字、中英文标点符号、空格
- 必须以中文/英文/数字开头和结尾

**函数签名**：
```go
func ValidateMemo(memo *string, required bool) error
func ValidateMemoFmt(memo *string, required bool) error
```

**✅ 正确示例**：

```go
if err := validator.ValidateMemo(sc.Memo, false); err != nil {
	return err
}
```

**✅ 有效格式示例**：

```go
// Valid Chinese Memo: "这是一个有效的备注"
// Valid English Memo: "This is a valid memo"
// Valid Mixed Memo: "这是一个valid混合memo"
// Include Space: "包含中文空格 ano"
```

**❌ 无效格式示例**：

```go
// From: pkg/criteria/validator/memo_test.go:90-94
// Invalid Format Special Characters Only: "!@#$%^&*()"
// Too Long Memo: (超过 256 字符)
// Only Chinese Punctuation: "[，。！？`'《》...]"
```

**注意**：`ValidateMemo` 和 `ValidateMemoFmt` 接受 `*string` 类型，需要处理 nil 情况。

## 常见问题

### Q: ValidateMemo 为什么接受 `*string` 而不是 `string`？

**A**: 因为备注字段通常是可选的，使用指针可以区分"未设置"（nil）和"空字符串"（""）。在 `required=false` 时，两者都视为有效；在 `required=true` 时，两者都视为无效。

### Q: 什么时候使用 ValidateMemo，什么时候使用 ValidateMemoFmt？

**A**: 
- `ValidateMemo`：需要同时验证格式和长度（<= 256 字符）时使用
- `ValidateMemoFmt`：只需要验证格式，不需要验证长度时使用
