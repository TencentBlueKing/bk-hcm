## 1. 枚举与基础类型

- [x] 1.1 在 `pkg/criteria/enumor/application.go` 中新增 `ApplyPermissionPolicyLibrary ApplicationType = "apply_permission_policy_library"` 并加入 `Validate()` switch
- [x] 1.2 新建 `pkg/criteria/enumor/permission_policy_library.go`，定义 `PermPolicyLibAction` 枚举（`apply_create` / `apply_update`）及其 `Validate()` 方法

## 2. API 请求/内容结构体

- [x] 2.1 新建 `pkg/api/cloud-server/application/permission_policy_library.go`，定义 `BizApplyPermissionPolicyLibraryCreateReq`（`policy_library_id`、`account_ids`）及 `Validate()` 方法
- [x] 2.2 在同文件定义 `ApplyPermPolicyLibContent`（`Action`、`Vendor`、`BkBizID`、`PolicyLibraryID`、`AccountID`）

## 3. Application Handler — Base

- [x] 3.1 新建 `cmd/cloud-server/service/application/handlers/permission-policy-library/base.go`，参照三级账号的 `subaccount/base.go` 实现：
  - `ActionHandlerFactory` 类型
  - `RegisterActionHandler` / `actionHandlerRegistry`（key 类型为 `enumor.PermPolicyLibAction`）
  - `NewHandlerFromApplication` 分发函数
  - `ApplicationBasePermissionPolicyLibrary` 基础 handler（嵌入 `BaseApplicationHandler` 和 `PolicyLibraryApplier`，实现 `PrepareReq`、`PrepareReqFromContent`、`GetBkBizIDs`、`GetItsmApprover`）

## 4. Application Handler — apply-create

- [x] 4.1 新建 `handlers/permission-policy-library/apply-create/init.go`：`ApplicationOfApplyPermPolicyLibCreate` 结构体、构造器 `NewApplicationOfApplyPermPolicyLibCreate`、`BuildContent` 辅助函数、`init()` 注册
- [x] 4.2 新建 `handlers/permission-policy-library/apply-create/check.go`：实现 `CheckReq()`（校验 policy_library_id 和 account_id 非空、账号 bk_biz_id 与请求 bk_biz_id 一致、账号在库的 biz 范围内、账号未重复应用）
- [x] 4.3 新建 `handlers/permission-policy-library/apply-create/create_itsm_ticket.go`：实现 `RenderItsmTitle()`、`RenderItsmForm()`（展示策略库名称、账号ID、bk_biz_id、云厂商）
- [x] 4.4 新建 `handlers/permission-policy-library/apply-create/deliver.go`：实现 `GenerateApplicationContent()` 和 `Deliver()`（调用 `a.ApplyCreate`，通过 `resp.Results[0].Status` 判断成功/失败）

## 5. Application Service — 创建方法与路由

- [x] 5.1 在 `cmd/cloud-server/service/application/create.go` 新增 `CreateBizForApplyPermissionPolicyLibraryCreate` 方法（biz 鉴权、PermissionPolicyLibrary Apply 鉴权、vendor 校验、解码并校验请求，调用 `createBizForApplyPermPolicyLibCreate`）
- [x] 5.2 在同文件新增私有方法 `createBizForApplyPermPolicyLibCreate`（循环调用 `BuildContent` + `NewApplicationOfApplyPermPolicyLibCreate` + `a.create()`，收集 IDs，返回 `*core.BatchCreateResult`）

## 6. Application Service — 审批分发

- [x] 6.1 在 `cmd/cloud-server/service/application/approve.go` 的 `getHandlerByApplication` 中新增 `case enumor.ApplyPermissionPolicyLibrary`，调用 `permissionpolicylibrary.NewHandlerFromApplication`

## 7. 路由注册

- [x] 7.1 在 `cmd/cloud-server/service/application/init.go` 的 `bizService` 中注册路由：`POST /bizs/{bk_biz_id}/vendors/{vendor}/applications/types/apply_permission_policy_library_create` → `CreateBizForApplyPermissionPolicyLibraryCreate`

## 8. 编译与 lint 验证

- [x] 8.1 执行 `go build ./...` 确认无编译错误
- [x] 8.2 执行 `golangci-lint run` 确认无 lint 错误（重点关注函数长度 ≤ 80 行）
