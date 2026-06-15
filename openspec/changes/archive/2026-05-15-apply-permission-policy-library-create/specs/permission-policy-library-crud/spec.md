## MODIFIED Requirements

### Requirement: 审计 Action 枚举扩展

系统 SHALL 在 `pkg/criteria/enumor/audit.go` 的 `AuditAction` 常量组中新增 `Apply AuditAction = "apply"`，用于权限策略库"应用"操作的审计记录。

同时 SHALL 在 `pkg/api/data-service/audit/audit.go` 的 `OperationAction` 常量组中新增 `ApplyOp OperationAction = "apply"`，用于 cloud-server 触发审计时指定操作类型。

#### Scenario: Apply 枚举可用

- **WHEN** 代码中引用 `enumor.Apply`
- **THEN** 其值 SHALL 为字符串 `"apply"`

#### Scenario: ApplyOp 转换为 Apply

- **WHEN** 调用 `ApplyOp.ConvAuditAction()`
- **THEN** 系统 SHALL 返回 `enumor.Apply`

#### Scenario: 现有枚举不受影响

- **WHEN** 代码中引用已有的 `enumor.Create`、`enumor.Update`、`enumor.Delete` 等
- **THEN** 其值 SHALL 保持不变

### Requirement: 权限策略库 Apply 审计 Build

系统 SHALL 在 `cmd/data-service/service/audit/cloud/permission_policy_library.go` 中实现 `permissionPolicyLibraryApplyAuditBuild` 方法，构建权限策略库"应用"操作的审计记录。

该方法 SHALL：
1. 批量查询所有 operation 中涉及的策略库详情（`listPermissionPolicyLibrary`）
2. 收集所有非空的 `AssociatedResID`（即被应用的账号 ID），批量查询账号信息（`listAccount`）
3. 为每个 operation 构建 `AuditTable`，`Action = Apply`，`ResType = PermissionPolicyLibraryAuditResType`
4. 若 operation 携带 `AssociatedResID` 且账号查询命中，则 `Detail.Data` 使用 `AssociatedOperationAudit`，包含 `AssResType = AccountAuditResType`、`AssResID`（账号 ID）、`AssResName`（账号名称）；否则 `Detail.Data` 直接为策略库数据

同时 SHALL 在审计分发入口（`buildOperationAuditInfo`）中新增对 `PermissionPolicyLibraryAuditResType` 的处理分支，路由到 `permissionPolicyLibraryApplyAuditBuild`。

#### Scenario: 应用审计记录构建（含关联账号）

- **WHEN** 触发权限策略库的 Apply 审计，operation 中携带 `AssociatedResType = AccountAuditResType` 和有效的 `AssociatedResID`
- **THEN** 系统 SHALL 构建审计记录，`Detail.Data` 为 `AssociatedOperationAudit{AssResType: AccountAuditResType, AssResID: accountID, AssResName: accountName}`

#### Scenario: 应用审计记录构建（无关联账号）

- **WHEN** 触发权限策略库的 Apply 审计，operation 中 `AssociatedResID` 为空
- **THEN** 系统 SHALL 构建审计记录，`Detail.Data` 直接为策略库数据

#### Scenario: 审计分发路由

- **WHEN** data-service 收到 `PermissionPolicyLibraryAuditResType` 类型的操作审计请求
- **THEN** 系统 SHALL 路由到 `permissionPolicyLibraryApplyAuditBuild` 方法
