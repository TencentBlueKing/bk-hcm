# 高级日志特性

## V 级别日志

### 概述

使用 `logs.V(n)` 控制详细日志输出级别，级别越高越详细。V 级别日志在生产环境默认不输出，可通过配置动态开启。

### 级别约定

| 级别 | 用途 | 说明 |
|------|------|------|
| V(2) | 调试级别请求/响应日志 | 一般调试信息 |
| V(4) | 详细请求体日志 | 请求详情 |
| V(5) | 授权跳过等详细逻辑日志 | 业务逻辑详情 |
| V(6) | 第三方服务详细日志 | 如 etcd 操作 |
| V(7) | 服务状态检查等高频日志 | 高频健康检查 |

### 两种使用方式

#### 方式一：链式调用（仅限 Info 级别）

```go
logs.V(5).Infof("skip authorization for resource: %+v, rid: %s", resource, kt.Rid)

logs.V(7).Infof("current service(%s) master state: %v", srvKey, cr == firstCR)
```

#### 方式二：布尔判断包裹（任何级别）

```go
// From: pkg/rest/handler.go:195-197
if logs.V(2) {
    logs.Errorf("do restful request %s failed, err: %v, rid: %s", action.Alias, err, cts.Kit.Rid)
}
```

### 选择建议

- **链式调用**：用于简单的 Info 级别详细日志
- **布尔判断**：用于需要多条日志或 Error/Warn 级别的场景

---

## 系统告警标记 (SysAlarmFlag)

### 概述

对于需要触发**系统告警**的严重错误，在日志消息开头添加 `errf.SysAlarmFlag`。

### 使用场景

| 场景 | 说明 |
|------|------|
| 外部 API 调用失败 | OBS、Plan 等第三方服务 |
| 定时任务 panic | 任务异常终止 |
| 定时任务运行时错误 | 任务返回错误 |

### 正确示例

```go
if err != nil {
    logs.Errorf("%s, call obs api failed, req: %+v, source: %s, schemeID: %s, err: %+v",
        errf.SysAlarmFlag, req, req.Source, ar.config.SchemeID, err)
    return nil, err
}

if r := recover(); r != nil {
    logs.Errorf("%s, cron job %s panic recovered after %d ms, panic: %v", errf.SysAlarmFlag, cj.Name, cost, r)
}

if err := cj.jobExecutor.GetRuntimeError(); err != nil {
    logs.Errorf("%s, cron job %s returns runtime err: %s", errf.SysAlarmFlag, cj.Name, err)
}
```

### 格式说明

- `SysAlarmFlag` 放在消息**最前面**
- 后续内容使用**逗号分隔**
- 包含足够的**上下文信息**便于排查

---

## ErrorJson 特殊用法

### 概述

`logs.ErrorJson` 用于需要序列化复杂对象的场景。会有性能损耗，**非必要不使用**。

### 使用场景

- 需要记录 filter 表达式
- 需要记录复杂请求参数
- 需要记录嵌套对象结构

### 正确示例

```go
if err != nil {
    logs.ErrorJson("obs bill ext count failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
    return nil, err
}
```

### 性能说明

ErrorJson 会对复杂对象进行 JSON 序列化，在高频调用场景下会有性能影响。优先考虑：

1. 是否可以只记录关键字段
2. 是否可以使用 `%+v` 格式
3. 是否真的需要完整对象信息

---

## 定时任务日志规范

### 标准模式

```go
func (cj *cronJob) Run() {
    // 从节点跳过
    if !cj.sd.IsMaster() {
        logs.Infof("skip run job %s, for current node is slave", cj.Name)
        return
    }

    start := time.Now()
    logs.Infof("cron job %s start", cj.Name)

    defer func() {
        cost := time.Since(start).Milliseconds()

        // panic 恢复
        if r := recover(); r != nil {
            logs.Errorf("%s, cron job %s panic recovered after %d ms, panic: %v",
                errf.SysAlarmFlag, cj.Name, cost, r)
        }

        // 任务结束
        logs.Infof("cron job %s end, takes %d ms", cj.Name, cost)

        // 运行时错误
        if err := cj.jobExecutor.GetRuntimeError(); err != nil {
            logs.Errorf("%s, cron job %s returns runtime err: %s",
                errf.SysAlarmFlag, cj.Name, err)
        }
    }()

    // 执行任务
    cj.jobExecutor.Execute(kt)
}
```

### 关键要点

1. **从节点跳过**：Info 级别记录
2. **任务开始/结束**：Info 级别，包含耗时
3. **panic 恢复**：Error 级别 + SysAlarmFlag
4. **运行时错误**：Error 级别 + SysAlarmFlag

---

## 批量操作日志控制

### 问题场景

循环内大量日志可能导致日志爆炸：

```go
// ❌ 不推荐：每条记录都打日志
for _, item := range items {
    logs.Infof("processing item: %s, rid: %s", item.ID, kt.Rid)
    // ...
}
```

### 推荐做法

```go
// ✅ 推荐：只在关键节点打日志
logs.Infof("start to process %d items, rid: %s", len(items), kt.Rid)

successCount := 0
failCount := 0
for _, item := range items {
    if err := process(item); err != nil {
        logs.Warnf("process item %s failed, err: %v, rid: %s", item.ID, err, kt.Rid)
        failCount++
        continue
    }
    successCount++
}

logs.Infof("finish processing items, success: %d, fail: %d, rid: %s",
    successCount, failCount, kt.Rid)
```

### 原则

1. **开始和结束**记录总数
2. **失败项**单独记录（Warn 级别）
3. **成功项**不逐条记录
4. **汇总统计**在结束时记录
