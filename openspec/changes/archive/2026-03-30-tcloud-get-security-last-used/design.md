## Context

hc-service 已有完整的密钥管理接口（创建/删除/更新），均在 `cmd/hc-service/service/account/secret.go` 中实现。本次新增的 GetSecurityLastUsed 接口遵循完全相同的分层架构和编码模式。

## Goals / Non-Goals

**Goals:**
- 实现腾讯云 GetSecurityLastUsed API 的 hc-service 封装
- 完全复用现有架构分层模式

**Non-Goals:**
- 不涉及密钥使用情况的本地存储或缓存
- 不涉及定时轮询或自动化清理

## Decisions

### 请求参数为 AccountID + SecretIdList

`AccountID` 用于获取 adaptor 客户端，`SecretIdList` 透传给腾讯云 API。腾讯云限制最多 10 个密钥 ID，在 Validate 中校验。

### 返回结构直接映射腾讯云响应

返回 `SecretIdLastUsed` 数组，包含 `SecretId`、`LastUsedDate`、`LastSecretUsedDate` 三个字段，与腾讯云 API 响应一一对应。

## Risks / Trade-offs

- **[数据延迟]** 腾讯云文档说明 LastUsedDate 有 1 天延迟 → 接口文档中需说明。
