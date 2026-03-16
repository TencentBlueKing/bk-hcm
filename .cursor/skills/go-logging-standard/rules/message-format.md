# 日志消息格式规范

## 概述

统一的日志消息格式有助于日志检索和问题排查。

## rid 追踪

### 基本规则

所有日志**必须**包含 `rid`（Request ID），用于请求链路追踪。

- **格式**：`rid: %s`
- **位置**：消息末尾
- **来源**：Handler 层用 `cts.Kit.Rid`，Service/DAO 层用 `kt.Rid`

### 正确示例

```go
logs.Errorf("read file from request failed, err: %v, rid: %s", err, cts.Kit.Rid)

logs.Errorf("check GPU small model usage index exists failed, err: %v, rid: %s", err, kt.Rid)
```

### 错误示例

```go
// ❌ 缺少 rid
logs.Errorf("read file from request failed, err: %v", err)

// ❌ rid 位置不对
logs.Errorf("rid: %s, read file from request failed, err: %v", cts.Kit.Rid, err)
```

### 例外情况

仅在**启动阶段**或**无 context** 的场景下可以不带 rid：

```go
// From: cmd/web-server/app/app.go:81
logs.Infof("load settings from config file success.")

// From: cmd/web-server/app/app.go:86
logs.Errorf("init profiling failed, err: %v", err)
```

---

## 消息风格

### 基本规则

1. **小写字母开头**（句子不用大写）
2. **描述清晰具体**的操作
3. **使用 `failed`/`success`** 描述结果
4. **错误信息使用 `err: %v`** 格式

### 正确示例

```go
// ✅ 小写开头，描述清晰
logs.Errorf("get hc area item decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
logs.Errorf("get hc area item validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
logs.Errorf("get hc area item failed, err: %v, rid: %s", err, cts.Kit.Rid)
```

### 错误示例

```go
// ❌ 大写开头
logs.Errorf("Get hc area item failed, err: %v, rid: %s", err, cts.Kit.Rid)

// ❌ 消息不清晰
logs.Errorf("error occurred, err: %v, rid: %s", err, cts.Kit.Rid)

// ❌ 错误格式不一致
logs.Errorf("get hc area item failed, error=%v, rid: %s", err, cts.Kit.Rid)
```

---

## 标准消息模板

### 请求处理场景

| 场景 | 消息模板 |
|------|----------|
| 解码请求失败 | `{operation} decode request failed, err: %v, rid: %s` |
| 验证请求失败 | `{operation} validate request failed, err: %v, rid: %s` |
| 操作失败 | `{operation} failed, err: %v, rid: %s` |
| 操作成功 | `{operation} success, count: %d, rid: %s` |

### 任务执行场景

| 场景 | 消息模板 |
|------|----------|
| 任务开始 | `start to {operation}, date: %s, rid: %s` |
| 任务结束 | `finish to {operation}, date: %s, total: %d, rid: %s` |
| 定时任务开始 | `cron job %s start` |
| 定时任务结束 | `cron job %s end, takes %d ms` |

### 跳过场景

| 场景 | 消息模板 |
|------|----------|
| 跳过无效数据 | `skip invalid {item}: {details}, rid: %s` |
| 跳过特定条件 | `skip {operation} for {reason}: {details}, rid: %s` |
| 从节点跳过 | `skip run job %s, for current node is slave` |

---

## 上下文信息

### 推荐包含的信息

| 类型 | 说明 | 示例 |
|------|------|------|
| **rid** | 必须（除启动阶段） | `rid: %s` |
| **count** | 批量操作 | `count: %d` |
| **date** | 时间相关操作 | `date: %s` |
| **cost** | 耗时信息 | `takes %d ms` |
| **业务 ID** | 关键标识 | `op_product_id=%d` |

### 示例

```go
// 批量操作
logs.Infof("batch upsert product org rel success, count: %d, rid: %s", len(items), kt.Rid)

// 时间相关
logs.Infof("start to load domain cdn usage monthly, date: %s, rid: %s", date, kt.Rid)

// 耗时信息
logs.Infof("cron job %s end, takes %d ms", cj.Name, cost)

// 多字段
logs.Warnf("skip invalid SG mapping: op_product_id=%d, sg=%s, rid: %s", opProductID, sgName, kt.Rid)
```

---

## 敏感信息处理

### 避免记录的信息

- 用户密码、Token
- 完整的认证信息
- 敏感业务数据

### 脱敏处理

```go
// ❌ 错误：记录完整 Token
logs.Infof("auth token: %s, rid: %s", token, kt.Rid)

// ✅ 正确：仅记录部分信息
logs.Infof("auth token length: %d, rid: %s", len(token), kt.Rid)
```
