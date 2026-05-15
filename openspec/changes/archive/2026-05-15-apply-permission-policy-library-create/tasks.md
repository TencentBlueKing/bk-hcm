## 1. TCloud CAM CreatePolicy Adaptor

- [x] 1.1 在 `pkg/adaptor/types/account/` 新增 `TCloudCreatePolicyOption`、`TCloudCreatePolicyResult` 类型定义
- [x] 1.2 在 `pkg/adaptor/tcloud/` 实现 `CreatePolicy` 方法，封装 `cam.NewCreatePolicyRequest` 调用
- [x] 1.3 在 `pkg/adaptor/tcloud/interface.go` 的 `TCloud` 接口中新增 `CreatePolicy` 方法声明

## 2. hc-service CAM 策略创建接口

- [x] 2.1 在 `pkg/api/hc-service/` 定义 CAM CreatePolicy 请求/响应模型（`CreateCAMPolicyReq`、`CreateCAMPolicyResult`）
- [x] 2.2 在 `cmd/hc-service/service/` 新增 permission-template 相关 handler，实现 CAM 策略创建接口（`POST /vendors/tcloud/permission_templates/cam/create_policy`）
- [x] 2.3 在 hc-service 路由注册中挂载新接口

## 3. hc-service Client 封装

- [x] 3.1 在 `pkg/client/hc-service/tcloud/` 新增 `CreateCAMPolicy` client 方法
- [x] 3.2 确认 client 方法能正确路由到 hc-service 新接口

## 4. 审计枚举与审计 Build

- [x] 4.1 在 `pkg/criteria/enumor/audit.go` 新增 `Apply AuditAction = "apply"` 枚举；在 `pkg/api/data-service/audit/audit.go` 新增 `ApplyOp OperationAction = "apply"` 常量及 ConvAuditAction 映射
- [x] 4.2 在 `cmd/data-service/service/audit/cloud/permission_policy_library.go` 新增 `permissionPolicyLibraryApplyAuditBuild` 方法，批量查询策略库详情和关联账号信息，构建包含 `AssociatedOperationAudit`（账号 ID + 名称）的审计记录
- [x] 4.3 在审计分发入口中新增 Apply 操作的处理分支
- [x] 4.4 在 `cmd/cloud-server/logics/audit/audit.go` 中确认现有 `ResOperationAudit` 可支持 Apply 审计调用（传入 AssociatedResType/AssociatedResID）

## 5. cloud-server API Model

- [x] 5.1 在 `pkg/api/cloud-server/permission_policy_library.go` 新增 `ApplyPermissionPolicyLibraryReq`、`ApplyPermissionPolicyLibraryResult`、`ApplyAccountResult` 模型定义及 Validate 方法

## 6. cloud-server applier 公共逻辑

- [x] 6.1 在 `cmd/cloud-server/service/permission-policy-library/` 新增 `applier.go`，定义 `PolicyLibraryApplier` 结构体及 `ApplyCreate` 入口方法（按 vendor 分派）
- [x] 6.2 实现 `GetPolicyLibraryDetail(kt, id)` 方法
- [x] 6.3 实现 `CheckAccountsBizInScope(kt, allowedBkBizIDs, accountIDs)` 方法，校验所有账号 bk_biz_id 均在策略库允许范围内
- [x] 6.4 实现 `CheckAccountApplied(kt, libraryID, accountID)` 方法
- [x] 6.5 实现 `TCloudCreateCAMPolicy(kt, library, accountID)` 方法
- [x] 6.6 实现 `TCloudCreateLocalTemplate(kt, library, accountID, cloudPolicyID)` 方法
- [x] 6.7 实现 `RecordApplyAudit(kt, libraryID, accountID)` 方法，传入 `AssociatedResType = AccountAuditResType`、`AssociatedResID = accountID`

## 7. cloud-server Apply Create Handler

- [x] 7.1 在 `cmd/cloud-server/service/permission-policy-library/` 新增 `apply.go`，实现 `ApplyPermissionPolicyLibraryCreate` handler
- [x] 7.2 handler 中实现完整的逐账号执行流程：校验未应用 → 调 hc-service 创建 CAM 策略 → 调 data-service 创建 permission_template → 记录审计（含关联账号）→ 收集结果
- [x] 7.3 在 `service.go` 路由注册中新增 `POST /vendors/{vendor}/permission_policy_libraries/{id}/apply` 路由

## 8. 验证与编译

- [x] 8.1 执行 `go build ./...` 确认编译通过
- [x] 8.2 检查 linter 无新增错误
