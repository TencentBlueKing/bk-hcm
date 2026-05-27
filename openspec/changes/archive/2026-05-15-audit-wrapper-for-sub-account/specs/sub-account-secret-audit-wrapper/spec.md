## Capability: sub-account-secret-audit-wrapper

为三级密钥提供审计记录创建封装方法，支持在 cloud-server 层直接创建审计记录，解决 data-service 无法获取 bk_biz_id 的问题。

## Behavior

1. 提供 `SubAccountSecretCreateAudit` 方法，封装三级密钥创建审计记录创建逻辑
2. 提供 `SubAccountSecretUpdateAudit` 方法，封装三级密钥更新审计记录创建逻辑
3. 提供 `SubAccountSecretDeleteAudit` 方法，封装三级密钥删除审计记录创建逻辑
4. 所有方法接收 `bk_biz_id` 作为参数，构建完整的 `AuditTable` 数据结构
5. 调用 data-service 新增的通用审计批量创建接口完成记录创建
6. 统一的错误处理和日志记录

## Interfaces

### Go Interface
```go
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

### Dependencies
- data-service: 通用审计批量创建接口 `POST /data-service/api/v1/cloud/audits/batch/create`

## Requirements

### Requirement: 三级密钥创建审计记录创建

系统必须（SHALL）在三级密钥创建成功后，调用 `SubAccountSecretCreateAudit` 方法创建审计记录。

#### Scenario: 三级密钥创建成功后创建审计记录
- **WHEN** 三级密钥在云上创建成功并持久化到本地数据库
- **THEN** 系统调用 `SubAccountSecretCreateAudit` 方法创建审计记录
- **AND** 审计记录包含 `res_type=SubAccountSecret`、`action=create`、`bk_biz_id`、`account_id`、`sub_account_id`、`secret_id` 等字段
- **AND** 方法内部构建完整的 `AuditTable` 结构并调用通用批量创建接口

#### Scenario: 审计记录创建失败
- **WHEN** 调用 `SubAccountSecretCreateAudit` 创建审计记录失败
- **THEN** 系统记录错误日志，但不中断业务流程
- **AND** 审计记录缺失，但不影响三级密钥创建结果

### Requirement: 三级密钥更新审计记录创建

系统必须（SHALL）在三级密钥状态更新成功后，调用 `SubAccountSecretUpdateAudit` 方法创建审计记录。

#### Scenario: 三级密钥状态更新成功后创建审计记录
- **WHEN** 三级密钥状态在云上更新成功并同步到本地数据库
- **THEN** 系统调用 `SubAccountSecretUpdateAudit` 方法创建审计记录
- **AND** 审计记录包含 `res_type=SubAccountSecret`、`action=update`、`bk_biz_id`、`account_id`、`sub_account_id`、`secret_id`、`changed` 等字段
- **AND** `changed` 字段包含状态变更信息（如 `status: active -> inactive`）
- **AND** 方法内部构建完整的 `AuditTable` 结构并调用通用批量创建接口

### Requirement: 三级密钥删除审计记录创建

系统必须（SHALL）在三级密钥删除成功后，调用 `SubAccountSecretDeleteAudit` 方法创建审计记录。

#### Scenario: 三级密钥删除成功后创建审计记录
- **WHEN** 三级密钥在云上删除成功并从本地数据库删除
- **THEN** 系统调用 `SubAccountSecretDeleteAudit` 方法创建审计记录
- **AND** 审计记录包含 `res_type=SubAccountSecret`、`action=delete`、`bk_biz_id`、`account_id`、`sub_account_id`、`secret_id`、`detail` 等字段
- **AND** `detail` 字段包含删除前的资源信息
- **AND** 方法内部构建完整的 `AuditTable` 结构并调用通用批量创建接口

### Requirement: 审计资源类型定义

系统必须（SHALL）在 `pkg/criteria/enumor/audit_resource_type.go` 中定义 `SubAccountSecretAuditResType` 审计资源类型。

#### Scenario: 审计资源类型存在
- **WHEN** 系统启动并加载审计资源类型枚举
- **THEN** `SubAccountSecretAuditResType` 已定义且可用

### Requirement: 三级密钥创建不走审批流的特殊处理

系统必须（SHALL）在三级密钥创建的 service 方法中直接调用审计封装方法，而非在审批流的 handler 中调用。

#### Scenario: 三级密钥创建直接调用审计封装
- **WHEN** 用户在业务下创建三级密钥
- **THEN** cloud-server 的 `createTCloudSubAccountSecret` 方法在密钥创建成功后直接调用 `SubAccountSecretCreateAudit`
- **AND** 不经过审批流 handler 的 `Deliver()` 方法

### Requirement: 通用审计批量创建接口

系统必须（SHALL）在 data-service 层提供通用的审计批量创建接口。

#### Scenario: 调用通用批量创建接口
- **WHEN** cloud-server 的审计封装方法构建完成 `AuditTable` 数据
- **THEN** 系统调用 data-service 的通用批量创建接口 `POST /data-service/api/v1/cloud/audits/batch/create`
- **AND** 接口接收完整的 `AuditTable` 结构数组
- **AND** 接口支持批量创建（最多 100 条）
