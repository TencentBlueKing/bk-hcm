## Context

项目已有完整的三级账号审批流处理模式（创建/更新/删除三级账号），通过 `SubAccountAction` 枚举 + `RegisterActionHandler` 注册机制实现。密钥状态变更需要复用此模式，同时保证云上密钥状态与本地 DB 数据的一致性。

hc-service 已实现 `TCloudUpdateAccessKey` 接口（`POST /vendors/tcloud/sub_accounts/secrets/update`），data-service 已实现 `BatchUpdateSubAccountSecret` 方法。本次变更需要在 cloud-server application 层编排这两个下游调用。

## Goals / Non-Goals

**Goals:**
- 实现密钥状态变更的审批流 handler，支持批量操作（每个密钥创建独立申请单）
- 校验密钥存在性、所属三级账号存在性、父级账号存在性
- 交付时先更新云上再更新本地，保证数据一致性
- 复用现有工具函数和 handler 基类，避免重复造轮子

**Non-Goals:**
- 不实现其他云厂商的密钥状态变更（仅实现 TCloud，但预留扩展点）
- 不修改 hc-service 或 data-service 层的已有逻辑

## Decisions

### 1. 使用新的 SubAccountAction 而非复用 enable/disable

现有枚举已有 `enable_secret_key` 和 `disable_secret_key`，但 API 设计为单一端点同时支持启用和禁用。使用单一 action `update_secret_key_status` 更贴合 API 语义，且 handler 实现更简洁（一个 handler 存储目标状态，而非两个几乎相同的 handler）。

### 2. 交付策略：先云上后本地

在 deliver 阶段，先调用 hc-service 更新云上密钥状态，成功后再更新本地 DB。若云上成功但本地失败，记录错误日志并返回 `DeliverError`，便于人工介入修复。

### 3. 查询密钥详情采用带扩展字段的查询

deliver 阶段需要密钥的 `cloud_secret_id`（存储在 extension 中），因此使用 `TCloud.SubAccountSecret.ListSubAccountSecretWithExtension` 获取完整信息。check 阶段仅需验证存在性，使用 `Global.SubAccountSecret` 即可。

### 4. hc-service 客户端新增 UpdateAccessKey 方法

cloud-server 调用 hc-service 需要通过 client 封装，当前缺少 `UpdateAccessKey` 方法，需要补充。

## Risks / Trade-offs

- [云上成功本地失败] → 记录详细日志，返回 `DeliverError` 状态，支持人工修复或重试
- [密钥在审批期间被删除] → deliver 阶段的 check 会再次验证密钥存在性，失败则返回错误
