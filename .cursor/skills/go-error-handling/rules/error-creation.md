# 错误创建方式规范

## 1. 使用 `errf.New()` / `errf.Newf()` 创建业务错误

**规则**：当需要返回带有业务错误码的错误时，使用 `errf` 包创建错误。

```go
// 简单错误 - 静态消息
return nil, errf.New(errf.InvalidParameter, "device family is required")

// 格式化错误 - 动态消息
return nil, errf.Newf(errf.InvalidParameter, "unsupported dimension: %s", dim)
```

**适用场景**：
- API 层返回给客户端的错误
- 需要携带错误码的业务错误
- 参数校验失败、资源不存在等业务逻辑错误

**代码示例**：

```go
func (svc *service) GetEstRegion(cts *rest.Contexts) (interface{}, error) {
    deviceFamily := cts.PathParameter("device_family").String()
    if deviceFamily == "" {
        return nil, errf.New(errf.InvalidParameter, "device family is required")
    }
```

## 2. 使用 `errf.NewFromErr()` 包装底层错误

**规则**：当需要将底层错误转换为带业务错误码的错误时使用。

```go
if err := req.Validate(); err != nil {
    return nil, errf.NewFromErr(errf.InvalidParameter, err)
}
```

**代码示例**：

```go
func (svc *service) GetAreaItem(cts *rest.Contexts) (interface{}, error) {
    req := new(types.BizDateRangeKindOption)
    if err := cts.DecodeInto(req); err != nil {
        logs.Errorf("get hc area item decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
        return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
    }
```

## 3. 使用 `errors.New()` 创建简单错误

**规则**：用于内部函数、工具函数中不需要错误码的场景。

```go
if kt == nil {
    return errors.New("kit is nil")
}

if arg == nil {
    return errors.New("update args is required")
}
```

**代码示例**：

```go
func (do *do) Scan(ctx context.Context, expr string, scanToRaw func() []any) (err error) {
    if scanToRaw == nil {
        return errors.New("callback scan to row function is nil")
    }
```

## 4. 使用 `fmt.Errorf()` 创建格式化错误

**规则**：用于需要动态信息但不需要业务错误码的内部错误。

```go
return nil, fmt.Errorf("init sharding failed, err: %v", err)
return nil, fmt.Errorf("per user cost type: %s not support", costType)
```

**代码示例**（`pkg/dal/dao/dao.go:577-582`）：

```go
func Connect(opt cc.ResourceDB) (*sqlx.DB, error) {
    db, err := sqlx.Connect("mysql", uri(opt))
    if err != nil {
        return nil, fmt.Errorf("connect to mysql failed, err: %v", err)
    }
```

## 5. 使用 `%w` 进行错误包装（有限使用）

**规则**：项目中较少使用 `%w` 进行错误包装，主要在 objectstore 等基础设施层使用。

```go
return fmt.Errorf("failed to copy object from %s to %s: %w", srcPath, dstPath, err)
```

**何时使用 `%w`**：
- 基础设施层需要保留原始错误链时
- 调用方需要使用 `errors.Is()` 或 `errors.As()` 检查原始错误时

## 选择指南

| 场景 | 方法 | 示例 |
|-----|------|------|
| API 返回给客户端 | `errf.New/Newf` | `errf.New(errf.InvalidParameter, "xxx")` |
| 包装底层错误返回客户端 | `errf.NewFromErr` | `errf.NewFromErr(errf.Aborted, err)` |
| 内部函数参数校验 | `errors.New` | `errors.New("param is nil")` |
| 内部函数格式化错误 | `fmt.Errorf` | `fmt.Errorf("op failed: %v", err)` |
| 需要保留错误链 | `fmt.Errorf + %w` | `fmt.Errorf("wrap: %w", err)` |
