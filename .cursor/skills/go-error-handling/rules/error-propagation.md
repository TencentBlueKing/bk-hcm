# 错误传播模式规范

## 1. 向上传递模式

**规则**：Service 层记录日志后传递错误，由上层统一处理。

```go
// Handler 层 - 添加错误码
items, err := svc.cs.DataService().GetAreaItem(cts.Kit, req)
if err != nil {
    logs.Errorf("get area item failed, err: %v, rid: %s", err, cts.Kit.Rid)
    return nil, errf.NewFromErr(errf.Aborted, err)  // 添加错误码，向上传递
}

// DAO 层 - 直接传递
if err := s.oi.Do().SelectNamedC(kt.Ctx, &list, sql, whereValue); err != nil {
    logs.Errorf("query failed, sql: %s, err: %v, rid: %s", sql, err, kt.Rid)
    return nil, err  // 直接传递
}
```

## 2. 终止传播模式

**规则**：非关键错误可以记录日志后继续执行。

```go
// 使用 warn 日志，继续执行
cdnDescs, err := svc.getCdnTypeDesc(cts.Kit)
if err != nil {
    logs.Warnf("get cdn type description failed, err: %v, rid: %s", err, cts.Kit.Rid)
    // 不返回，继续使用默认值
}

// 循环中跳过错误数据
for _, item := range items {
    if err := process(item); err != nil {
        logs.Warnf("process item failed, item: %+v, err: %v, rid: %s", item, err, kt.Rid)
        continue  // 跳过当前项，继续处理
    }
}
```

## 3. 添加上下文信息

**规则**：传递错误时可以添加上下文信息。

```go
// 使用 fmt.Errorf 添加上下文
if err != nil {
    return nil, fmt.Errorf("query raws failed, err: %v", err)
}

// 使用 errf 包装并添加错误码
if err != nil {
    return nil, errf.NewFromErr(errf.DBExecCmdFailed, err)
}
```

## 4. 错误处理流程图

```
┌─────────────────────────────────────────────────────────────────┐
│                        API Handler                               │
├─────────────────────────────────────────────────────────────────┤
│  1. DecodeInto() → err → errf.NewFromErr(DecodeRequestFailed)   │
│  2. Validate()   → err → errf.NewFromErr(InvalidParameter)      │
│  3. Service()    → err → errf.NewFromErr(Aborted)               │
│  4. return data, nil                                             │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Service Layer                             │
├─────────────────────────────────────────────────────────────────┤
│  • 记录错误日志 (logs.Errorf)                                    │
│  • 添加业务错误码 (errf.NewFromErr)                              │
│  • 向上传递错误                                                   │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         DAO Layer                                │
├─────────────────────────────────────────────────────────────────┤
│  • 记录错误日志 (logs.Errorf)                                    │
│  • 直接返回原始错误                                               │
│  • 特殊情况添加 DBExecCmdFailed / InvalidParameter               │
└─────────────────────────────────────────────────────────────────┘
```

## 5. 各层错误处理职责

| 层级 | 职责 | 错误码添加 |
|-----|------|-----------|
| Handler | 记录日志、添加错误码、返回客户端 | 所有业务错误码 |
| Service | 记录日志、业务逻辑判断 | 可选（通常由 Handler 添加） |
| DAO | 记录日志、直接返回 | `DBExecCmdFailed`、`InvalidParameter` |
| 工具层 | 返回原始错误 | 不添加 |

## 6. 批处理错误传播

**规则**：批处理中的错误可以收集后统一返回。

```go
var errs []error
for _, item := range items {
    if err := process(item); err != nil {
        errs = append(errs, fmt.Errorf("process %s failed: %v", item.ID, err))
        continue
    }
}

if len(errs) > 0 {
    return nil, fmt.Errorf("batch process failed: %v", errs)
}
```

## 最佳实践

### DO

```go
// ✓ Handler 层添加错误码
if err != nil {
    logs.Errorf("get data failed, err: %v, rid: %s", err, kt.Rid)
    return nil, errf.NewFromErr(errf.Aborted, err)
}

// ✓ 非关键错误继续执行
if err != nil {
    logs.Warnf("optional step failed, err: %v, rid: %s", err, kt.Rid)
    // continue with default
}

// ✓ 循环中跳过错误
for _, item := range items {
    if err := process(item); err != nil {
        logs.Warnf("skip item, err: %v, rid: %s", err, kt.Rid)
        continue
    }
}
```

### DON'T

```go
// ✗ 忽略错误
_ = process(item)

// ✗ 重复添加错误码
if err != nil {
    err = errf.NewFromErr(errf.Aborted, err)
    return nil, errf.NewFromErr(errf.Aborted, err)  // 重复包装
}

// ✗ 关键错误不返回
if err != nil {
    logs.Errorf("critical error, err: %v", err)
    // 缺少 return
}
```
