## Why

业务视角下的申请单列表接口（`ListBizApplications`）已实现，但缺少查看单据明细的接口。用户无法在业务视角下查看单据的详细信息，需要补齐该功能。

## What Changes

- 在 cloud-server 层新增 `GET /api/v1/cloud/bizs/{bk_biz_id}/applications/{application_id}` 接口
- 使用业务访问权限（`meta.Biz` + `meta.Access`）进行鉴权
- 进行归属校验：检查 `bk_biz_id` 是否在单据的 `bk_biz_ids` 列表中
- 统一 NotFound 错误策略：无权限/不存在/不归属均返回 `RecordNotFound`
- 复用 `GetApplication` 的响应构建逻辑，抽取公共方法 `buildApplicationGetResp`
- 响应体包含 `Source` 字段，对 `Content` 进行脱敏处理

## Capabilities

### New Capabilities
- `biz-application-detail`: 业务视角下查看单据明细接口，包含参数解析、业务访问权限鉴权、归属校验、脱敏处理

### Modified Capabilities
- `GetApplication`: 复用公共方法 `buildApplicationGetResp`，同时修复缺少 `Source` 字段的问题

## Impact

- `cmd/cloud-server/service/application/init.go`：在 `bizService` 中注册 `GetBizApplication` 路由
- `cmd/cloud-server/service/application/get.go`：新增 `GetBizApplication` handler 和 `buildApplicationGetResp` 公共方法
