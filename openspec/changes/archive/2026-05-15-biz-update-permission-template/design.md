## Context

系统已有「创建云权限模板」审批流（commit `10efbd74`），采用 `OperatePermissionTemplate` 申请类型 + action 分发模式（`ActionHandlerRegistry`）。

当前 `PolicyLibraryApplier.ApplyUpdate` 方法服务于「策略库更新后同步到已绑定账号」的场景，通过 `(policy_library_id, account_id)` 查找模板。

本变更要实现的「更新权限模板」场景不同：面向的是**自定义模板**（`policy_library_id=nil`, TCloud `cloud_type=1`），用户主动选择一个策略库来更新模板内容，是单模板 by ID 的操作。

## Goals / Non-Goals

**Goals:**
- 新增 `update_permission_template` 审批流接口，复用现有 action 分发框架
- 在 `applier.go` 中新增 `ApplyUpdateTemplate` 方法处理 by-template-ID 的更新场景
- 审批通过后更新云端 CAM Policy + 本地模板记录（含 `policy_library_id` 绑定）

**Non-Goals:**
- 不支持更新非自定义模板（策略库来源模板的更新走 `apply_permission_policy_library_update`）
- 不支持修改模板名称（`name` 字段不在本次 update 范围内）

## Decisions

### Decision 1：新增 `ApplyUpdateWithTmplInfo` 而非复用 `ApplyUpdate`

`ApplyUpdate` 通过 `(policy_library_id, account_id)` 查模板，而自定义模板无 `policy_library_id`，根本找不到。且 `ApplyUpdate` 面向多账号批量场景，本次是单模板精确更新。

**决策**：在 `applier.go` 中新增 `ApplyUpdateWithTmplInfo(kt, vendor, libraryID string, templateIDs []string, tmplInfo UpdateTmplBaseInfo)`，接受显式 templateIDs，调用方（deliver.go）传入 `[]string{templateID}` 实现单模板更新。`ApplyUpdate` 内部调用时传 `UpdateTmplBaseInfo{Memo: library.Memo}` 保持向后兼容。

**替代方案**：修改 `ApplyUpdate` 增加 by-ID 分支 → 会使方法职责模糊，拒绝。

---

### Decision 2：新增 `UpdateTmplBaseInfo` 结构体，通过 caller 设置默认值

`updateTCloudCAMPolicy` 的 `Description` 当前硬编码为 `library.Memo`，`updateTCloudLocalTemplate` 不写 `policy_library_id`。

**决策**：新增 `UpdateTmplBaseInfo` 结构体（含 `Memo *string`），将两个私有方法签名改为接受 `tmplInfo UpdateTmplBaseInfo`，移除对 `library.Memo` 的直接引用：
- `updateTCloudCAMPolicy(... tmplInfo UpdateTmplBaseInfo)`：使用 `tmplInfo.Memo` 作为 CAM Policy description
- `updateTCloudLocalTemplate(... tmplInfo UpdateTmplBaseInfo)`：使用 `tmplInfo.Memo` 作为 memo 字段，并始终写入 `policy_library_id`

向后兼容由调用方负责：现有 `ApplyUpdate` 传 `UpdateTmplBaseInfo{Memo: library.Memo}`，行为不变。新的 `ApplyUpdateWithTmplInfo` 则由调用方（deliver.go）传入用户 memo。

**替代方案**：在方法内加 `descOverride *string` 参数做 nil-fallback → 私有方法不需要如此防御，且职责应在 caller，拒绝。

---

### Decision 3：复用 `getTCloudTemplateByIDs` 而非新增 `GetTemplateByID`

`ApplyUpdateWithTmplInfo` 需要按 templateID 获取完整模板（含 extension，用于校验 `cloud_type`）。

**决策**：`applyTCloudUpdateForTemplate` 内部复用已有的私有方法 `getTCloudTemplateByIDs`，无需额外暴露 `GetTemplateByID`。校验逻辑（cloud_type、policy_library_id）由 `CheckPermTmplUpdatability` 在 CheckReq 阶段完成，Deliver 阶段不重复校验。

**替代方案**：新增公开 `GetTemplateByID` 方法 → 引入不必要的公开 API，拒绝。

---

### Decision 4：`CheckPermTmplUpdatability` 同时支持自定义模板和已绑同一策略库的模板

**条件**（满足任一即允许）：
1. `policy_library_id IS NULL` AND `extension.cloud_type == TCloudCustomPolicy(1)`：自定义模板，允许绑定任意策略库
2. `policy_library_id != ""` AND `policy_library_id == targetPolicyLibraryID`：已绑定同一策略库，允许重新应用（幂等更新）

**决策**：`CheckPermTmplUpdatability` 中两个条件均视为合法，其他组合一律拒绝。

## Risks / Trade-offs

- **`TCloudUpdateLocalTemplate` 修改影响**：改用 `UpdateTmplBaseInfo` 传参并始终写入 `policy_library_id`。现有 `ApplyUpdate` 调用路径传 `UpdateTmplBaseInfo{Memo: library.Memo}`，行为不变；写入相同的 `policy_library_id` 值是幂等操作，安全。
- **自定义模板更新后变为策略库绑定模板**：更新后 `policy_library_id` 从 nil 变为有值，语义上模板从「自定义」变为「策略库关联」。后续该模板仍可通过 `ApplyUpdateWithTmplInfo` 用同一策略库再次更新（条件 2），但不允许更换为其他策略库。这是预期行为。
