## Context

cloud-server 层已有资源视角下的 `GetApplication` 接口，业务视角下的 `ListBizApplications` 也已实现。本次需要在业务视角下补充查看单据明细的接口 `GetBizApplication`。

参考实现：`get.go`（同目录，资源视角下的单据详情）、`list.go`（业务视角下列表接口的鉴权和归属校验模式）。

## Goals / Non-Goals

**Goals:**
- 在 cloud-server 层实现业务视角下查看单据明细接口 `GET /bizs/{bk_biz_id}/applications/{application_id}`
- 使用业务访问权限（`meta.Biz` + `meta.Access`）进行鉴权
- 进行归属校验：检查 `bk_biz_id` 是否在单据的 `bk_biz_ids` 列表中
- 统一 NotFound 错误策略，避免泄露信息
- 抽取公共方法 `buildApplicationGetResp`，复用响应构建逻辑

**Non-Goals:**
- 修改 data-service 层
- 新增数据库字段
- 修改 IAM 权限定义

## Decisions

### 1. 路由注册在 bizService 中

在 `init.go` 的 `bizService` 函数中添加路由：

路径前缀 `/bizs/{bk_biz_id}` 由 `bizH.Path("/bizs/{bk_biz_id}")` 设置，完整路径为 `/api/v1/cloud/bizs/{bk_biz_id}/applications/{application_id}`。

### 2. 业务访问权限鉴权

使用 `meta.Biz` + `meta.Access` 组合进行鉴权：

**替代方案**：使用 `meta.Application` + `meta.Find` —— 拒绝，因为业务视角下应使用业务级别的权限，而非单据管理权限。

### 3. 归属校验使用 slice.IsItemInSlice

检查请求的 `bk_biz_id` 是否在单据的 `bk_biz_ids` 列表中：

### 4. 统一 NotFound 错误策略

无论是权限不足、单据不存在还是不归属当前业务，均返回 `errf.RecordNotFound` 错误，避免泄露敏感信息。

### 5. 抽取公共方法复用代码

将响应体构建逻辑抽取为 `buildApplicationGetResp` 方法，供 `GetApplication` 和 `GetBizApplication` 共同使用，减少代码重复。

## Risks / Trade-offs

- [信息隐藏] 统一返回 NotFound 可能导致调试困难 → 通过 `logs.Warnf` 记录详细原因，便于排查问题
