## Implementation Tasks

### 1. 新增通用审计批量创建接口（data-service层）

- [x] 1.1 在 `pkg/api/data-service/audit/audit.go` 中新增 `BatchCreateAuditReq` 请求结构体，包含 `Audits []*AuditTable` 字段
- [x] 1.2 在 `pkg/api/data-service/audit/audit.go` 中新增 `AuditTable` 结构体定义（如果尚未定义），包含所有审计字段（`res_id`、`res_name`、`res_type`、`action`、`bk_biz_id`、`vendor`、`account_id`、`operator`、`source`、`rid`、`app_code`、`detail` 等）
- [x] 1.3 在 `pkg/client/data-service/global/audit.go` 中新增 `BatchCreateAudit` 客户端方法，调用 `POST /data-service/api/v1/cloud/audits/batch/create` 接口
- [x] 1.4 在 `cmd/data-service/service/audit/cloud/audit.go` 中新增 `BatchCreateAudit` 接口实现，接收请求并调用 DAO 层批量创建审计记录
- [x] 1.5 在 `cmd/data-service/service/audit/cloud/audit.go` 的路由注册中添加批量创建接口的路由

### 2. 审计资源类型定义

- [x] 2.1 在 `pkg/criteria/enumor/audit_resource_type.go` 中确认或新增 `SubAccountAuditResType` 审计资源类型
- [x] 2.2 在 `pkg/criteria/enumor/audit_resource_type.go` 中确认或新增 `SubAccountSecretAuditResType` 审计资源类型

### 3. 审计封装方法实现（三级账号）

- [x] 3.1 在 `cmd/cloud-server/logics/audit/audit.go` 的 `Interface` 接口中新增 `SubAccountCreateAudit` 方法签名
- [x] 3.2 实现 `SubAccountCreateAudit` 方法，构建完整的 `AuditTable` 数据并调用通用批量创建接口
- [x] 3.3 在 `cmd/cloud-server/logics/audit/audit.go` 的 `Interface` 接口中新增 `SubAccountUpdateAudit` 方法签名
- [x] 3.4 实现 `SubAccountUpdateAudit` 方法，构建包含 `changed` 字段的审计记录并调用通用批量创建接口
- [x] 3.5 在 `cmd/cloud-server/logics/audit/audit.go` 的 `Interface` 接口中新增 `SubAccountDeleteAudit` 方法签名
- [x] 3.6 实现 `SubAccountDeleteAudit` 方法，构建包含删除前资源信息的审计记录并调用通用批量创建接口

### 4. 审计封装方法实现（三级密钥）

- [x] 4.1 在 `cmd/cloud-server/logics/audit/audit.go` 的 `Interface` 接口中新增 `SubAccountSecretCreateAudit` 方法签名
- [x] 4.2 实现 `SubAccountSecretCreateAudit` 方法，构建三级密钥创建审计记录并调用通用批量创建接口
- [x] 4.3 在 `cmd/cloud-server/logics/audit/audit.go` 的 `Interface` 接口中新增 `SubAccountSecretUpdateAudit` 方法签名
- [x] 4.4 实现 `SubAccountSecretUpdateAudit` 方法，构建三级密钥状态更新审计记录并调用通用批量创建接口
- [x] 4.5 在 `cmd/cloud-server/logics/audit/audit.go` 的 `Interface` 接口中新增 `SubAccountSecretDeleteAudit` 方法签名
- [x] 4.6 实现 `SubAccountSecretDeleteAudit` 方法，构建三级密钥删除审计记录并调用通用批量创建接口

### 5. 三级账号审计集成

- [x] 5.1 在 `cmd/cloud-server/service/application/handlers/sub-account/create-sub-account/deliver.go` 的 `deliverForTCloud` 方法中，三级账号创建成功后调用 `SubAccountCreateAudit`
- [x] 5.2 在 `cmd/cloud-server/service/application/handlers/sub-account/update-sub-account/deliver.go` 的 `Deliver` 方法中，三级账号更新成功后调用 `SubAccountUpdateAudit`
- [x] 5.3 在 `cmd/cloud-server/service/application/handlers/sub-account/delete-sub-account/deliver.go` 的 `Deliver` 方法中，三级账号删除成功后调用 `SubAccountDeleteAudit`

### 6. 三级密钥审计集成

- [x] 6.1 在 `cmd/cloud-server/service/subaccount-secret/create.go` 的 `createTCloudSubAccountSecret` 方法中，三级密钥创建成功后调用 `SubAccountSecretCreateAudit`
- [x] 6.2 在 `cmd/cloud-server/service/application/handlers/sub-account/delete-secret-key/deliver.go` 的 `Deliver` 方法中，三级密钥删除成功后调用 `SubAccountSecretDeleteAudit`
- [x] 6.3 在 `cmd/cloud-server/service/application/handlers/sub-account/update-secret-status/deliver.go` 的 `Deliver` 方法中，三级密钥状态更新成功后调用 `SubAccountSecretUpdateAudit`
