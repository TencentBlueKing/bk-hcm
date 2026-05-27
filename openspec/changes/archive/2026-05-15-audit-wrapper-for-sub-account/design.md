## Context

HCM 平台的审计记录创建目前有两种路径：
1. **通过 data-service 的审计接口**：适用于 cloud-server 层，可以获取完整的上下文信息（如 `bk_biz_id`）
2. **通过 dao 层直接操作数据库**：适用于 data-service 内部，但无法获取 cloud-server 路由参数

当前三级账号和三级密钥的审计记录创建存在以下问题：
- 在 data-service 层无法获取 cloud-server 路由参数中的 `bk_biz_id`
- 三级账号的创建、更新、删除都走审批流，在 handler 的 `Deliver()` 方法中交付资源后需要记录审计
- 三级密钥的创建不走审批流，直接在 cloud-server 的 service 方法中创建，需要独立的审计记录创建点
- 缺少统一的上层封装，代码重复且难以维护

现有架构中，`cmd/cloud-server/logics/audit/audit.go` 已经提供了云资源审计的封装方法（如 `ResDeleteAudit`、`ResUpdateAudit` 等），但缺少针对三级账号和三级密钥的专门封装。

**现有审计接口的局限性：**
现有的 data-service 审计接口（如 `CloudResourceUpdateAudit`、`CloudResourceDeleteAudit` 等）存在以下限制：
- 不支持直接传递 `bk_biz_id` 字段
- 接口参数结构固定，只支持特定的资源类型和操作类型
- 无法满足三级账号和三级密钥需要自定义完整审计字段的需求

因此，需要新增一个**通用的审计批量创建接口**，支持完整的 `AuditTable` 结构。

## Goals / Non-Goals

**Goals:**
- 在 data-service 层新增通用的审计批量创建接口，支持完整的 `AuditTable` 结构
- 在 cloud-server 层新增三级账号审计记录创建封装方法，支持创建、更新、删除操作
- 在 cloud-server 层新增三级密钥审计记录创建封装方法，支持创建、更新、删除操作
- 所有封装方法接收 `bk_biz_id` 作为参数，构建完整的审计记录数据
- 设计统一接口，便于复用和维护
- 在三级账号创建、更新、删除的交付流程中集成审计记录创建
- 在三级密钥创建、删除、状态更新的流程中集成审计记录创建

**Non-Goals:**
- 不修改现有的 data-service 审计接口实现（保持兼容）
- 不修改 dao 层的审计记录创建逻辑
- 不涉及审计记录查询、列表等读取操作
- 不涉及其他云资源的审计记录封装（仅针对三级账号和三级密钥）

## Decisions

### 1. 核心方案：新增通用审计批量创建接口

**选择**：在 data-service 层新增一个通用的审计批量创建接口，支持接收完整的 `AuditTable` 结构数组。

**原因**：
- 现有的审计接口参数结构固定，不支持自定义 `bk_biz_id` 等字段
- 通用的批量创建接口可以为未来类似的审计场景提供解决方案，不仅限于三级账号和三级密钥
- 支持批量创建，提高性能
- 与现有接口保持兼容，不影响现有功能

**接口设计**：
```go
// 请求结构体
type BatchCreateAuditReq struct {
    Audits []*AuditTable `json:"audits" validate:"required,min=1,max=100"`
}

// AuditTable 包含所有审计字段
type AuditTable struct {
    ID         uint64                 `json:"id"`
    ResID      string                 `json:"res_id" validate:"required"`
    ResName    string                 `json:"res_name"`
    ResType    enumor.AuditResourceType `json:"res_type" validate:"required"`
    Action     enumor.AuditAction     `json:"action" validate:"required"`
    Vendor     enumor.Vendor          `json:"vendor"`
    AccountID  string                 `json:"account_id"`
    BkBizID    int64                  `json:"bk_biz_id"`  // 支持直接传递 bk_biz_id
    Operator   string                 `json:"operator" validate:"required"`
    Source     string                 `json:"source"`
    Rid        string                 `json:"rid"`
    AppCode    string                 `json:"app_code"`
    Detail     *BasicDetail           `json:"detail,omitempty"`
    CreatedAt  time.Time              `json:"created_at"`
}

// 接口路径
POST /data-service/api/v1/cloud/audits/batch/create
```

### 2. 审计封装位置：cloud-server/logics/audit

**选择**：在 `cmd/cloud-server/logics/audit/audit.go` 中新增三级账号和三级密钥审计封装方法。

**原因**：
- 已有的云资源审计封装都在此文件中，保持一致性
- cloud-server 层可以获取完整的上下文信息（`bk_biz_id`、`vendor`、`account_id` 等）
- 与现有的 `ResDeleteAudit`、`ResUpdateAudit` 等方法保持相同的设计模式
- 方便 cloud-server 的各个 service 和 handler 调用

### 3. 接口设计：专门的审计方法 vs 通用方法

**选择**：为三级账号和三级密钥提供专门的审计方法，而非扩展现有的通用方法。

**原因**：
- 三级账号和三级密钥的审计字段有特殊性（如 `account_id`、`sub_account_id`、`secret_id` 等）
- 专门的方法可以提供更清晰的参数签名和类型安全
- 便于后续扩展和维护
- 不影响现有的通用审计方法

新增方法签名：
```go
// SubAccountCreateAudit 三级账号创建审计
SubAccountCreateAudit(kt *kit.Kit, bizID int64, vendor enumor.Vendor, 
    accountID string, subAccountID string, subAccountName string, detail interface{}) error

// SubAccountUpdateAudit 三级账号更新审计
SubAccountUpdateAudit(kt *kit.Kit, bizID int64, vendor enumor.Vendor, 
    accountID string, subAccountID string, subAccountName string, 
    updateFields map[string]interface{}) error

// SubAccountDeleteAudit 三级账号删除审计
SubAccountDeleteAudit(kt *kit.Kit, bizID int64, vendor enumor.Vendor, 
    accountID string, subAccountID string, subAccountName string, detail interface{}) error

// SubAccountSecretCreateAudit 三级密钥创建审计
SubAccountSecretCreateAudit(kt *kit.Kit, bizID int64, vendor enumor.Vendor, 
    accountID string, subAccountID string, secretID string, detail interface{}) error

// SubAccountSecretUpdateAudit 三级密钥更新审计
SubAccountSecretUpdateAudit(kt *kit.Kit, bizID int64, vendor enumor.Vendor, 
    accountID string, subAccountID string, secretID string, 
    updateFields map[string]interface{}) error

// SubAccountSecretDeleteAudit 三级密钥删除审计
SubAccountSecretDeleteAudit(kt *kit.Kit, bizID int64, vendor enumor.Vendor, 
    accountID string, subAccountID string, secretID string, detail interface{}) error
```

### 4. 审计记录创建流程

**选择**：封装方法内部构建完整的 `AuditTable` 数据，调用新增的通用批量创建接口。

**流程**：
1. 接收审计参数（包含 `bk_biz_id`、资源信息、操作详情等）
2. 构建完整的 `AuditTable` 结构体数组，填充所有必要字段
3. 调用 data-service 的通用审计批量创建接口
4. 处理错误并记录日志

**原因**：
- 使用通用接口，不再受限于特定资源类型的审计接口
- 支持完整的审计字段填充，包括 `bk_biz_id`
- 统一的错误处理和日志记录
- 支持批量创建，提高性能

### 5. 审计类型（ResType）定义

**选择**：使用已有的审计资源类型枚举，或新增专门的类型。

现有的审计资源类型（`enumor.AuditResourceType`）中可能已包含三级账号和三级密钥的类型，需要确认：
- 如果已存在，直接使用
- 如果不存在，需要在 `pkg/criteria/enumor/audit_resource_type.go` 中新增

预期类型：
- `SubAccountAuditResType`: 三级账号审计资源类型
- `SubAccountSecretAuditResType`: 三级密钥审计资源类型

### 6. 审计操作（Action）定义

**选择**：使用已有的审计操作枚举（`create`、`update`、`delete`）。

现有的审计操作（`enumor.AuditAction`）已涵盖：
- `create`: 创建
- `update`: 更新
- `delete`: 删除

直接复用即可，无需新增。

### 7. 与现有审批流的集成

**三级账号**：
- **创建**：在 `create-sub-account/deliver.go` 的 `Deliver()` 方法中，资源交付成功后调用 `SubAccountCreateAudit`
- **更新**：在 `update-sub-account/deliver.go` 的 `Deliver()` 方法中，资源更新成功后调用 `SubAccountUpdateAudit`
- **删除**：在 `delete-sub-account/deliver.go` 的 `Deliver()` 方法中，资源删除成功后调用 `SubAccountDeleteAudit`

**三级密钥**：
- **创建**：在 `subaccount-secret/create.go` 的 `createTCloudSubAccountSecret()` 方法中，密钥创建成功后调用 `SubAccountSecretCreateAudit`
- **删除**：在 `delete-secret-key/deliver.go` 的 `Deliver()` 方法中，密钥删除成功后调用 `SubAccountSecretDeleteAudit`
- **状态更新**：在 `update-secret-status/deliver.go` 的 `Deliver()` 方法中，状态更新成功后调用 `SubAccountSecretUpdateAudit`

## Risks / Trade-offs

**风险：**
1. **审计记录丢失**：如果资源操作成功但审计记录创建失败，会导致审计记录缺失
   - **缓解措施**：在资源操作成功后立即创建审计记录，失败时记录详细日志，但不回滚资源操作（因为云上操作通常无法回滚）
   
2. **性能影响**：每次操作都创建审计记录，可能影响响应时间
   - **缓解措施**：审计记录创建是轻量级操作，影响可控；可以考虑异步创建（本期不做）

3. **接口兼容性**：新增通用接口需要确保不影响现有接口
   - **缓解措施**：通用接口是新接口，不修改现有接口，保持兼容性

**权衡：**
1. **通用接口 vs 专用接口**：选择通用接口，牺牲了一些特定场景的简化，但换来了更好的扩展性和复用性
2. **封装粒度**：选择专门的方法而非通用方法，牺牲了一些灵活性，但换来了更清晰的接口和更好的类型安全
3. **错误处理**：选择记录日志而非返回错误中断流程，避免因审计失败导致业务操作失败

## Implementation Steps

1. 在 `pkg/api/data-service/audit/audit.go` 中新增通用审计批量创建请求结构体 `BatchCreateAuditReq`
2. 在 `pkg/client/data-service/global/audit.go` 中新增通用审计批量创建客户端方法 `BatchCreateAudit`
3. 在 `cmd/data-service/service/audit/cloud/audit.go` 中新增通用审计批量创建接口实现
4. 在 `pkg/criteria/enumor/audit_resource_type.go` 中确认或新增 `SubAccountAuditResType` 和 `SubAccountSecretAuditResType`
5. 在 `cmd/cloud-server/logics/audit/audit.go` 中新增三级账号审计封装方法（3 个）
6. 在 `cmd/cloud-server/logics/audit/audit.go` 中新增三级密钥审计封装方法（3 个）
7. 在三级账号创建、更新、删除的 handler 中集成审计记录创建
8. 在三级密钥创建、删除、状态更新的流程中集成审计记录创建
9. 编写单元测试验证审计记录创建逻辑
10. 进行集成测试，验证审计记录完整性
