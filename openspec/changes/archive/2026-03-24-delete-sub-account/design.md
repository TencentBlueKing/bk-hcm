## Context

当前平台已实现三级账号创建的完整审批流（`create-sub-account` 包），采用 `OperateSubAccount` 申请单类型 + `SubAccountAction` 分发机制。每个 action 子包通过 `init()` 自注册到 `actionHandlerRegistry`，审批回调时由 `NewHandlerFromApplication` 根据 content 中的 `action` 字段分发到对应 handler。

删除流程需要复用这套审批架构，action 为已定义的 `SubAccountActionDelete`。删除操作涉及三层清理：云上 CAM 用户、本地 sub_account 表、account 表中的登记账号记录。

现有相关基础设施：
- `ApplicationBaseSubAccount`（base.go）提供公共基础能力（BkBizID、AccountID、ITSM 审批人等）
- `BaseSubAccountContent` 提供 action/vendor/bk_biz_id 公共字段
- data-service 已有 `SubAccount.BatchDelete` 和 `Account.Delete` 客户端
- TCloud CAM SDK 提供 `DeleteUser` API（`cam.NewDeleteUserRequest`）
- hc-service 已有 `TCloudCreateSubAccount` 端点模式可参照

## Goals / Non-Goals

**Goals:**
- 实现腾讯云三级账号删除的完整审批流（校验 → ITSM 审批 → 交付删除）
- 复用现有 `OperateSubAccount` 审批架构和 `SubAccountActionDelete` action
- 保持 vendor-switch 扩展性，后续可便捷添加其他云厂商
- 抽取 create/delete 共用的结构体到公共层

**Non-Goals:**
- 不实现密钥删除校验的实际逻辑（以 TODO 占位）
- 不实现 AWS/Azure/GCP/HuaWei 等其他云厂商的删除
- 不修改现有审批流框架

## Decisions

### 1. 复用 OperateSubAccount 审批流而非新增申请单类型

**选择**：复用已有的 `OperateSubAccount` + `SubAccountActionDelete` 分发机制。

**原因**：
- `SubAccountActionDelete` 已在 `enumor/sub_account_action.go` 中定义
- `actionHandlerRegistry` 分发机制天然支持新增 action
- 审批人逻辑（二级账号管理员）与创建流程一致，无需特殊处理

**替代方案**：新增独立的 `DeleteSubAccount` ApplicationType → 需要修改 approve.go 分发逻辑，破坏现有设计模式，复杂度更高。

### 2. 删除交付顺序：云上 → sub_account 表 → account 表

**选择**：先删云上 CAM 用户，再删本地 sub_account 记录，最后删 account 表登记记录。

**原因**：
- 云上删除是最关键的操作，失败时应立即中断而非留下孤立的本地记录
- sub_account 是核心实体，account 表的登记记录依赖于 sub_account 的 cloud_id
- 如果云上删除成功但本地删除失败，日志记录 cloud_id 便于人工修复

### 3. hc-service 新增 DeleteSubAccount 端点

**选择**：在 hc-service 新增 `TCloudDeleteSubAccount` 端点，封装 CAM `DeleteUser` 调用。

**原因**：
- 与 `TCloudCreateSubAccount` 保持一致的分层模式
- cloud-server handler 不直接访问 adaptor，通过 hc-service 中转
- 便于后续添加删除前的云上资源检查等逻辑

### 4. 通过 cloud_sub_account_id 匹配删除 account 表登记记录

**选择**：查询 account 表中 `extension.cloud_sub_account_id` 等于三级账号 `cloud_id` 且类型为 `registration` 的记录，进行删除。

**原因**：
- 创建流程中通过 `registerAccountForTCloud` 将三级账号注册为登记账号，`cloud_sub_account_id` 是唯一关联字段
- 使用 filter 表达式匹配比硬编码 account_id 更可靠

### 5. 密钥校验 TODO 占位

**选择**：在 `CheckReq` 中以 TODO 注释标记密钥校验逻辑。

**原因**：密钥管理功能尚未实现，但校验位置和逻辑预留明确，后续补充实现时无需调整流程结构。

## Risks / Trade-offs

- **[风险] 云上删除成功但本地删除失败** → 在日志中记录 cloud_id，交付状态标记为 `DeliverError`，运维可根据日志手动清理。
- **[风险] account 表中未找到对应登记记录** → 视为非致命错误，仅记录警告日志，不阻塞整体删除流程（可能该三级账号未注册为登记账号）。
- **[风险] 批量删除中部分账号失败** → 当前 API 接受 IDs 列表，每个 ID 独立创建审批单，单个失败不影响其他。
- **[权衡] 密钥校验暂缺** → 接受风险，通过 TODO 标记确保后续可追踪补充。
