# 错误检查模式规范

## 1. 标准 `if err != nil` 处理

**规则**：检查错误后立即处理（记录日志并返回）。

```go
// 解码失败
if err := cts.DecodeInto(req); err != nil {
    logs.Errorf("decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
    return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
}

// 验证失败
if err := req.Validate(); err != nil {
    logs.Errorf("validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
    return nil, errf.NewFromErr(errf.InvalidParameter, err)
}

// 业务调用失败
items, err := svc.cs.DataService().GetAreaItem(cts.Kit, req)
if err != nil {
    logs.Errorf("get area item failed, err: %v, rid: %s", err, cts.Kit.Rid)
    return nil, errf.NewFromErr(errf.Aborted, err)
}
```

## 2. 使用 `errors.Is()` 判断特定错误类型

**规则**：判断特定错误类型时使用 `errors.Is()`。

```go
if errors.Is(err, orm.ErrRecordNotFound) {
    return nil, errf.New(errf.RecordNotFound, "record does not exist")
}

if !errors.Is(hitErr, gcache.KeyNotFoundError) {
    logs.Errorf("cache not found, err: %v, rid: %s", hitErr, kt.Rid)
}
```

**代码示例**：

```go
if hitErr != nil {
    if !errors.Is(hitErr, gcache.KeyNotFoundError) {
        // this is not found error, log it.
        logs.Errorf("get obs cost composition by region cache not found, month: %s, conds: %v, err: %v, rid: %s",
```

## 3. 使用 `errf.HasCode()` 判断业务错误码

**规则**：判断特定业务错误码时使用 `errf.HasCode()`。

```go
if errf.HasCode(err, errf.RecordNotFound) {
    // 处理记录不存在的情况
    return defaultValue, nil
}
```

**代码示例**（`pkg/criteria/errf/error.go:126-134`）：

```go
func HasCode(err error, code int32) bool {
    if err == nil {
        return false
    }

    return Error(err).Code == code
}
```

## 4. 使用 `errors.As()` 类型断言

**规则**：需要获取特定错误类型的字段时使用。

```go
var ef *errf.ErrorF
if errors.As(err, &ef) {
    // 可以访问 ef.Code, ef.Message
    if ef.Code == errf.RecordNotFound {
        return nil, nil
    }
}
```

## 5. nil error 处理

**规则**：在错误转换函数中需要判断 nil error。

```go
func NewFromErr(code int32, err error) error {
    if err == nil {
        return nil
    }
    // ...
}
```

## 错误检查模式对比

| 方法 | 用途 | 示例 |
|-----|------|------|
| `err != nil` | 基本错误检查 | `if err != nil { return err }` |
| `errors.Is()` | 判断错误类型/值 | `errors.Is(err, sql.ErrNoRows)` |
| `errors.As()` | 类型断言获取错误详情 | `errors.As(err, &customErr)` |
| `errf.HasCode()` | 判断业务错误码 | `errf.HasCode(err, errf.RecordNotFound)` |

## 最佳实践

### DO

```go
// ✓ 使用 errors.Is 检查特定错误
if errors.Is(err, orm.ErrRecordNotFound) {
    return nil, errf.New(errf.RecordNotFound, "not found")
}

// ✓ 使用 errf.HasCode 检查业务错误码
if errf.HasCode(err, errf.RecordNotFound) {
    return defaultData, nil
}
```

### DON'T

```go
// ✗ 直接字符串比较
if err.Error() == "record not found" { ... }

// ✗ 直接类型断言（可能丢失包装的错误）
if e, ok := err.(*errf.ErrorF); ok { ... }
```
