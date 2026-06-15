## Why

当前三级账号和三级密钥的审计记录创建存在一个关键问题：在 data-service 层无法获取 cloud-server 路由参数中的 `bk_biz_id`，导致审计记录无法正确填写业务 ID。现有审计记录创建逻辑在 dao 层直接操作数据库，缺少上层封装，使得需要在 data-service 中处理审计时无法获取完整的上下文信息（如 `bk_biz_id`）。

此外，三级账号和三级密钥的操作审计场景多样（创建、修改、删除），且三级密钥的创建不走审批流，与其他操作路径不同，需要统一、可复用的审计记录封装方案。

经过分析，现有 data-service 提供的审计接口（如 `CloudResourceUpdateAudit`、`CloudResourceDeleteAudit` 等）不支持直接传递 `bk_biz_id` 等完整审计字段，只支持特定的资源类型和操作类型。因此，需要新增一个**通用的审计批量创建接口**，支持完整的 `AuditTable` 结构，解决当前问题并为未来类似场景提供通用解决方案。

## What Changes

### 核心变更：新增通用审计批量创建接口

- 在 data-service 层新增通用的审计批量创建接口 `POST /data-service/api/v1/cloud/audits/batch/create`
- 该接口支持接收完整的 `AuditTable` 结构数组，允许 cloud-server 层构建包含所有必要字段（包括 `bk_biz_id`）的审计记录
- 接口支持批量创建（最多 100 条），提高性能

### 上层封装：三级账号和三级密钥审计封装

- 在 cloud-server 层新增审计记录创建封装模块（`cmd/cloud-server/logics/audit/`），提供针对三级账号和三级密钥的审计记录创建方法
- 新增 `SubAccountAudit` 接口及实现，封装三级账号创建、更新、删除的审计记录创建逻辑
- 新增 `SubAccountSecretAudit` 接口及实现，封装三级密钥创建、更新、删除的审计记录创建逻辑
- 所有封装方法接收 `bk_biz_id` 作为参数，构建完整的 `AuditTable` 数据，调用新增的通用批量创建接口

## Capabilities

### New Capabilities

- `sub-account-audit-wrapper`: 三级账号审计记录封装能力，提供三级账号创建、更新、删除操作的审计记录创建方法，支持传入 `bk_biz_id`，构建完整审计记录并调用通用批量创建接口
- `sub-account-secret-audit-wrapper`: 三级密钥审计记录封装能力，提供三级密钥创建、更新、删除操作的审计记录创建方法，支持传入 `bk_biz_id`，构建完整审计记录并调用通用批量创建接口

### Modified Capabilities

无现有 capabilities 需要修改，这是新增的封装层和接口，不改变现有审计逻辑。

## Impact

**Affected Code:**
- `pkg/api/data-service/audit/audit.go`: 新增通用审计批量创建请求结构体和接口定义
- `pkg/client/data-service/global/audit.go`: 新增通用审计批量创建客户端方法
- `cmd/data-service/service/audit/cloud/audit.go`: 新增通用审计批量创建接口实现
- `cmd/cloud-server/logics/audit/audit.go`: 新增三级账号和三级密钥审计封装方法
- `cmd/cloud-server/service/application/handlers/sub-account/create-sub-account/deliver.go`: 在三级账号创建交付后调用审计封装方法记录审计
- `cmd/cloud-server/service/application/handlers/sub-account/update-sub-account/deliver.go`: 在三级账号更新交付后调用审计封装方法记录审计
- `cmd/cloud-server/service/application/handlers/sub-account/delete-sub-account/deliver.go`: 在三级账号删除交付后调用审计封装方法记录审计
- `cmd/cloud-server/service/subaccount-secret/create.go`: 在三级密钥创建后调用审计封装方法记录审计
- `cmd/cloud-server/service/application/handlers/sub-account/delete-secret-key/deliver.go`: 在三级密钥删除后调用审计封装方法记录审计
- `cmd/cloud-server/service/application/handlers/sub-account/update-secret-status/deliver.go`: 在三级密钥状态更新后调用审计封装方法记录审计

**Dependencies:**
- 依赖现有的审计表结构和 DAO 层实现（无变更）

**Systems:**
- 影响 data-service 层（新增接口）和 cloud-server 层（新增封装方法和调用逻辑），对 hc-service 无影响
