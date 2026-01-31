# 日志级别选择规范

## 概述

日志级别的选择应基于错误后的**流程控制**，而非错误的严重程度。

## Error 日志

**规则**：Error 日志用于记录需要**返回错误**的情况，通常紧跟 `return` 语句。

### 标准模式

```go
if err != nil {
    logs.Errorf("{operation} failed, err: %v, rid: %s", err, kt.Rid)
    return err  // 或 return nil, err
}
```

### 正确示例

```go
file, _, err := cts.Request.Request.FormFile("file")
if err != nil {
    logs.Errorf("read file from request failed, err: %v, rid: %s", err, cts.Kit.Rid)
    return nil, times.VarTimeRange{}, errf.NewFromErr(errf.DecodeRequestFailed, err)
}

if err := e.cs.AnalysisServer().BatchUpsertProductOrgRel(kt, req); err != nil {
    logs.Errorf("batch upsert product org rel failed, err: %v, rid: %s", err, kt.Rid)
    return 0, err
}
```

### 错误示例

```go
// ❌ 错误后 continue 不应使用 Error
if err != nil {
    logs.Errorf("process item failed, err: %v, rid: %s", err, kt.Rid)
    continue  // 应该用 Warnf
}
```

---

## Warn 日志

**规则**：Warn 日志用于**可恢复**的异常情况，流程可以继续（`continue`、降级处理等）。

### 标准模式

```go
if err != nil {
    logs.Warnf("{operation} failed, err: %v, rid: %s", err, kt.Rid)
    continue  // 或降级处理
}
```

### 正确示例

```go
if opProductID <= 0 || sgName == "" {
    logs.Warnf("skip invalid SG mapping: op_product_id=%d, sg=%s, rid: %s", opProductID, sgName, kt.Rid)
    continue
}

// 降级处理场景
cdnDescs, err := svc.getCdnTypeDesc(cts.Kit)
if err != nil {
    logs.Warnf("get cdn type description failed, err: %v, rid: %s", err, cts.Kit.Rid)
    // 后续使用硬编码描述作为降级
}
```

### 错误示例

```go
// ❌ 需要 return 的错误不应使用 Warn
if err != nil {
    logs.Warnf("critical operation failed, err: %v, rid: %s", err, kt.Rid)
    return err  // 应该用 Errorf
}
```

---

## Info 日志

**规则**：Info 日志用于记录**正常业务流程**的关键节点。

### 使用场景

| 场景 | 消息模式 |
|------|----------|
| 任务开始 | `start to {operation}, date: %s, rid: %s` |
| 任务结束 | `finish to {operation}, date: %s, total: %d, rid: %s` |
| 操作成功 | `{operation} success, count: %d, rid: %s` |
| 正常跳过 | `skip {reason}, rid: %s` |

### 正确示例

```go
logs.Infof("cron job %s start", cj.Name)

logs.Infof("cron job %s end, takes %d ms", cj.Name, cost)

logs.Infof("batch upsert product org rel success, count: %d, rid: %s", len(items), kt.Rid)

if sgName == "其他" {
    logs.Infof("skip SG mapping for '其他': op_product_id=%d, rid: %s", opProductID, kt.Rid)
    continue
}

if !cj.sd.IsMaster() {
    logs.Infof("skip run job %s, for current node is slave", cj.Name)
    return
}
```

---

## 级别选择决策树

```
发生异常/错误
    │
    ├─ 后续执行 return？
    │   └─ 是 → Error
    │
    ├─ 后续执行 continue/降级？
    │   └─ 是 → Warn
    │
    └─ 正常业务流程记录？
        └─ 是 → Info
```

---

## 边界情况

### 循环内的错误处理

在循环中处理单个项目失败时，根据是否需要中断整个循环来选择级别：

```go
for _, item := range items {
    if err := process(item); err != nil {
        // 单项失败，继续处理其他项
        logs.Warnf("process item %s failed, err: %v, rid: %s", item.ID, err, kt.Rid)
        continue
    }
}
```

### 函数开头的校验失败

参数校验失败通常需要 return，使用 Error：

```go
if req == nil {
    logs.Errorf("request is nil, rid: %s", kt.Rid)
    return errf.New(errf.InvalidParameter, "request is nil")
}
```
