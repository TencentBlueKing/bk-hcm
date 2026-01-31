# 错误日志记录规范

## 1. 使用 `logs.Errorf()` 记录错误日志

**规则**：错误返回前记录错误日志，必须包含 rid（请求ID）。

```go
logs.Errorf("get hc area item decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
logs.Errorf("list obs bill ext failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
```

## 2. 日志格式要求

| 要素 | 说明 | 示例 |
|-----|------|------|
| 动词开头 | 描述操作 | get、list、create、update、delete |
| 上下文信息 | 请求参数等 | `req: %+v` |
| 错误信息 | error 对象 | `err: %v` |
| 请求 ID | **必须** | `rid: %s` |

**格式模板**：

```go
logs.Errorf("<action> <resource> failed, <context>, err: %v, rid: %s", ..., err, kt.Rid)
```

**示例**：

```go
// 标准格式
logs.Errorf("get area item failed, err: %v, rid: %s", err, cts.Kit.Rid)

// 带请求参数
logs.Errorf("list obs bill ext failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)

// 带 SQL 信息
logs.Errorf("query failed, sql: %s, err: %v, rid: %s", sql, err, kt.Rid)
```

## 3. 使用 `logs.Warnf()` 记录警告日志

**规则**：不中断流程的异常情况使用 warn 日志。

```go
// 可降级处理的错误
cdnDescs, err := svc.getCdnTypeDesc(cts.Kit)
if err != nil {
    logs.Warnf("get cdn type description failed, err: %v, rid: %s", err, cts.Kit.Rid)
    // 继续执行，使用默认值
}

// 数据异常但不影响主流程
if len(staff.Departments) == 0 {
    logs.Warnf("get no department info by user %s", staff.Username)
    continue
}
```

## 4. 日志级别选择原则

| 场景 | 日志级别 | 后续动作 |
|------|---------|---------|
| 错误后 return | `logs.Errorf` | 流程终止 |
| 错误后 continue/fallthrough | `logs.Warnf` | 流程继续 |
| 可降级处理的错误 | `logs.Warnf` | 使用默认值或跳过 |
| 系统告警 | `logs.Errorf` + `errf.SysAlarmFlag` | 需要运维关注 |

## 5. 系统告警标记

**规则**：关键业务错误需要加上告警标记。

```go
logs.Warnf("%s, obs_bill_timing_sync_error, convert failed, ..., rid: %s", 
    errf.SysAlarmFlag, feeDate, ...)
```

## 6. 格式化占位符选择

| 占位符 | 用途 | 示例 |
|-------|------|------|
| `%v` | error 对象 | `err: %v` |
| `%+v` | 结构体（详细） | `req: %+v` |
| `%s` | 字符串 | `rid: %s` |
| `%d` | 数字 | `count: %d` |

## 最佳实践

### DO

```go
// ✓ 完整的错误日志
logs.Errorf("get area item failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)

// ✓ 使用 warn 记录非致命错误
logs.Warnf("cache miss, will query from db, key: %s, rid: %s", key, kt.Rid)

// ✓ 关键业务加告警标记
logs.Errorf("%s, sync failed, err: %v, rid: %s", errf.SysAlarmFlag, err, kt.Rid)
```

### DON'T

```go
// ✗ 缺少 rid
logs.Errorf("get area item failed, err: %v", err)

// ✗ 缺少错误详情
logs.Errorf("failed, rid: %s", kt.Rid)

// ✗ 流程终止却用 warn
if err != nil {
    logs.Warnf("critical error, err: %v, rid: %s", err, kt.Rid)
    return nil, err  // 应该用 Errorf
}
```
