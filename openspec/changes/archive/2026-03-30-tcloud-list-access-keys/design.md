## Context

hc-service 已有密钥管理的 CRUD + 使用情况查询接口，均在 `secret.go` 中实现。ListAccessKeys 接口遵循相同的分层架构。

## Goals / Non-Goals

**Goals:**
- 实现腾讯云 ListAccessKeys API 的 hc-service 封装

**Non-Goals:**
- 不涉及密钥信息的本地持久化

## Decisions

请求需要 `AccountID`（获取 adaptor 客户端）和 `TargetUin`（指定子用户）。返回 AccessKey 列表包含 `AccessKeyId`、`Status`、`CreateTime`、`Description` 字段。
