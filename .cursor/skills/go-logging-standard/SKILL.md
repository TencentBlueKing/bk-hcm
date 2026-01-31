---
name: go-logging-standard
description: Go 日志规范完整指南。当需要了解日志级别使用、rid 追踪、日志消息格式等详细规范时使用。
---

# Go 日志规范

SKILL.md 本身并没有提供具体的规范内容，必须根据场景和下方的规则索引，进一步读取对应的规则文件。

## 快速参考

| 场景 | 规则文件 |
|-----|---------|
| 日志级别选择 | [rules/log-level-selection.md](rules/log-level-selection.md) |
| 日志消息格式 | [rules/message-format.md](rules/message-format.md) |
| 高级特性 | [rules/advanced-features.md](rules/advanced-features.md) |

## 核心原则

1. **使用项目日志库**：`hcm/pkg/logs`（基于 glog 封装）
2. **必须包含 rid**：除启动阶段外，所有日志必须包含 `rid: %s` 用于链路追踪
3. **级别选择原则**：Error + return，Warn + continue/降级
4. **消息风格**：小写开头，使用 `failed`/`success` 描述结果

## 日志级别速查

| 级别 | 函数 | 使用场景 |
|------|------|----------|
| Info | `logs.Infof()` | 任务开始/结束、操作成功、正常跳过 |
| Warn | `logs.Warnf()` | 可恢复错误后 continue、降级处理 |
| Error | `logs.Errorf()` | 错误后 return |
| V(n) | `logs.V(n).Infof()` | 调试级别详细日志 |

## 标准消息模式

```go
// 解码失败
logs.Errorf("{operation} decode request failed, err: %v, rid: %s", err, kt.Rid)

// 验证失败
logs.Errorf("{operation} validate request failed, err: %v, rid: %s", err, kt.Rid)

// 操作失败
logs.Errorf("{operation} failed, err: %v, rid: %s", err, kt.Rid)

// 操作成功
logs.Infof("{operation} success, count: %d, rid: %s", count, kt.Rid)

// 任务开始
logs.Infof("start to {operation}, date: %s, rid: %s", date, kt.Rid)

// 任务结束
logs.Infof("finish to {operation}, date: %s, total: %d, rid: %s", date, total, kt.Rid)
```

## 常见问题

### Q: 什么时候用 Error，什么时候用 Warn？
A: 判断后续流程：`return` 用 Error，`continue`/降级用 Warn。

### Q: rid 从哪里获取？
A: Handler 层用 `cts.Kit.Rid`，Service/DAO 层用 `kt.Rid`。

### Q: 什么时候需要 SysAlarmFlag？
A: 外部 API 调用失败、定时任务 panic、定时任务运行时错误。

### Q: V 级别如何选择？
A: V(2) 调试日志，V(5) 详细逻辑，V(7) 高频状态检查。
