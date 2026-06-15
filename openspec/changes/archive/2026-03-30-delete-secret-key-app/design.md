## Context

项目已有完整的三级账号审批流处理模式，并且刚完成了密钥状态变更的 handler（`update-secret-key-status`）。密钥删除 handler 与之结构高度相似，复用相同的 `ApplicationBaseSubAccount` 基类和 `RegisterActionHandler` 注册机制。

hc-service 已实现 `TCloudDeleteAccessKey` 接口（`POST /vendors/tcloud/sub_accounts/secrets/delete`），data-service 已实现 `Global.SubAccountSecret.BatchDelete` 方法。本次变更在 cloud-server application 层编排这两个下游调用。

已有的 `listSecretBasicInfo` / `listTCloudSecretBasicInfo` 辅助函数可直接复用，无需重复实现。

## Goals / Non-Goals

**Goals:**
- 实现密钥删除的审批流 handler，支持批量操作（每个密钥创建独立申请单）
- 校验密钥存在性、所属三级账号存在性、父级账号存在性
- 交付时先删除云上再删除本地，保证数据一致性
- 复用 `update-secret-key-status` 中已建立的辅助函数（如 `listSecretBasicInfo`）

**Non-Goals:**
- 不实现其他云厂商的密钥删除（仅实现 TCloud，预留扩展点）

## Decisions

### 1. 复用已有的 SubAccountActionDeleteSecretKey 枚举

现有枚举已有 `SubAccountActionDeleteSecretKey = "delete_secret_key"`，直接使用，无需新增。

### 2. 交付策略：先云上后本地

先调用 hc-service 删除云上密钥，成功后再删除本地 DB。若云上成功但本地失败，记录错误日志并返回 `DeliverError`。

### 3. 本地删除使用 BatchDelete + filter

data-service `Global.SubAccountSecret.BatchDelete` 接受 filter 表达式，使用密钥表记录主键 `id` 作为删除条件。

## Risks / Trade-offs

- [云上删除成功本地删除失败] → 记录详细日志，返回 `DeliverError` 状态
- [密钥在审批期间被删除] → deliver 阶段的查询会发现不存在，返回错误
