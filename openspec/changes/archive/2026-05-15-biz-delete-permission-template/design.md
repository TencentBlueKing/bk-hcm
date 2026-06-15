## Context

系统已有"创建权限模板"和"更新权限模板"两种申请单类型，均通过 ITSM 审批流程操作云资源。它们的 application handler 都继承自 `ApplicationBasePermissionTemplate`，而该 base 嵌入了 `PolicyLibraryApplier`（处理策略库应用逻辑）。

"删除权限模板"操作的本质是删除云上 CAM Policy + 删除本地记录，与策略库（PolicyLibrary）无关，不需要 `PolicyLibraryApplier` 的任何方法。当前 hc-service 也缺少 `DeleteCAMPolicy` 接口，adaptor 层缺少 `DeletePolicy` 方法，需要补充。

## Goals / Non-Goals

**Goals:**
- 补全删除云权限模板的申请单流程（CheckReq → ITSM 审批 → Deliver）
- 审批通过后删除云上 CAM Policy 并清理本地记录
- 校验约束：仅允许删除自定义策略（CloudType == TCloudCustomPolicy），且关联三级账号数为 0
- 支持 vendor 扩展，当前仅实现 TCloud

**Non-Goals:**
- 不支持批量删除（单次只删一个模板）
- 不实现 AWS/GCP 等其他 vendor（留 switch-case 框架）
- 不修改 PolicyLibraryApplier

## Decisions

### 决策 1：delete handler 嵌入 ApplicationBasePermissionTemplate

**选择**：`ApplicationOfDeletePermTemplate` 嵌入 `permissiontemplate.ApplicationBasePermissionTemplate`（而非直接嵌入 `handlers.BaseApplicationHandler`），复用 base 已提供的 `BkBizID`、`GetBkBizIDs`、`PrepareReq`、`PrepareReqFromContent`、`GetItsmApproverByTemplateID` 等方法，自身覆写 `GetItsmApprover` 委托给 `GetItsmApproverByTemplateID(kt, content.ID)`。

**理由**：delete handler 虽与策略库操作无关，但 `ApplicationBasePermissionTemplate` 同时提供了 handler 生命周期所需的所有公共方法（bkBizID 封装、ITSM approver 查询等），直接复用可避免在 delete 包中重复实现多个桩方法，降低维护成本。`PolicyLibraryApplier` 的方法不会被调用，嵌入不带来运行时开销。

**原备选方案**（曾被选取后废弃）：直接嵌入 `handlers.BaseApplicationHandler + bkBizID`——需在 delete/init.go 中手动复现 `BkBizID`、`GetBkBizIDs`、`PrepareReq`、`PrepareReqFromContent`、`GetItsmApprover` 等全部桩方法，代码冗余，与 create/update handler 结构不对称。

### 决策 2：deliver 逻辑内联，不新增 applier 方法

**选择**：`Deliver()` 中直接调用 `a.Client.HCService()` 和 `a.Client.DataService()`，不在 `PolicyLibraryApplier` 里添加 `ApplyDelete` 或类似方法。

**理由**：删除逻辑只有两步（cloud delete + local delete），内联比抽象更清晰；同时避免将非"apply"语义的操作混入 `PolicyLibraryApplier`。

### 决策 3：check.go / deliver.go 各自做 vendor switch

**选择**：公开方法（`CheckReq`、`Deliver`）内部 `switch a.Vendor()`，调用私有的 vendor 具体函数（`checkTCloud`、`deleteTCloud`）。

**理由**：与现有 `CheckPermTmplUpdatability` → `checkTCloudPermTmplUpdatability` 模式对称，便于后续新增 vendor 时只需扩展 switch。

### 决策 4：hc-service 新增 DeleteCAMPolicy 接口

**选择**：在 `cmd/hc-service/service/permission-template/cam_policy.go` 新增 `TCloudDeleteCAMPolicy`，路由 `DELETE /permission_templates/cam/delete_policy`，adaptor 层新增 `TCloudImpl.DeletePolicy`。

**理由**：与已有 `CreateCAMPolicy` / `UpdateCAMPolicy` 保持接口风格一致，cloud-server 不直接调用云 SDK，所有 cloud API 调用都经由 hc-service。

## Risks / Trade-offs

- **风险：删除后无法回滚** → cloud 侧 CAM Policy 一旦删除不可恢复；通过 ITSM 审批流程作为保障，audit 记录完整操作历史
- **风险：并发申请单** → 若两个申请单同时审批通过删除同一模板，第二次 hc-service 调用可能报"Policy 不存在"；当前容忍此情形（deliver 失败进入 DeliverError 状态），不额外加锁
- **trade-off：sub-account 关联检查时序** → check 阶段与 deliver 阶段之间可能有新 sub-account 绑定，届时 deliver 可能失败；接受此 trade-off，校验在创建申请单时做即可

## Migration Plan

无存量数据迁移。纯新增功能，部署时先发 hc-service（新增 DeleteCAMPolicy 路由），再发 cloud-server（使用该路由）。
