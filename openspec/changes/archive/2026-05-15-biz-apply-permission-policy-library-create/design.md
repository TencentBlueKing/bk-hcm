## Context

### 背景

系统已有两个非业务接口（`ApplyPermissionPolicyLibraryCreate` / `ApplyPermissionPolicyLibraryUpdate`），
直接调用云 API 创建/更新 CAM 策略，没有审批流程。

业务侧需要 ITSM 审批流程版本：用户发起申请 → ITSM 审批 → 审批通过后自动执行。
返回格式为 `{ids: [...]}` 数组，每个账号对应一个审批单 ID。

### 现有相关代码

- `PolicyLibraryApplier`（`applier.go`）：封装了 TCloud CAM 策略创建/更新的核心逻辑，全部为公共方法，可直接复用
- Application 框架（`application/create.go`、`approve.go`、`init.go`）：标准的审批单创建→ITSM→交付流程
- 参考三级账号模式：多操作共用一个 `ApplicationType`，通过 `Action` 字段 + 注册工厂 (`init()`) 分发

## Goals / Non-Goals

**Goals:**
- 实现 `POST /bizs/{bk_biz_id}/vendors/{vendor}/applications/types/apply_permission_policy_library_create` 业务接口
- 每个 `account_id` 创建一个独立的 ITSM 审批单，返回所有审批单 ID 数组
- 审批通过后复用 `PolicyLibraryApplier` 公共方法完成 CAM 策略创建
- 参考三级账号的模式设计，保持架构一致性

**Non-Goals:**
- 不实现 `apply_permission_policy_library_update` 业务接口（后续需求，但架构上预留 Action 字段）
- 不修改非业务的直接执行接口

## Decisions

### 决策 1：ApplicationType 粒度 — 共享 type + Action 字段

**选择**：使用一个 `ApplicationType = "apply_permission_policy_library"` + `Action` 字段（`apply_create` / `apply_update`），
枚举类型为 `PermPolicyLibAction`，参考三级账号的 `OperateSubAccount` 模式。

**理由**：
- 后续 `apply_permission_policy_library_update` 接口可复用同一个框架，只需新增 Action handler
- `approve.go` 的 `switch` 只增加一个 case，通过 Registry 分发
- 与三级账号代码模式一致，降低维护成本

**备选**：两个独立 ApplicationType (`apply_permission_policy_library_create` / `update`)
→ 放弃，因为会在 `approve.go` 加两个 case 且无法复用 base 逻辑

---

### 决策 2：批量创建模式 — 循环调用 `a.create()`，任一失败则整体报错

**选择**：参考三级账号 的 `batchCreateBizForAddSubAccount` 模式：
```go
for _, accountID := range req.AccountIDs {
    content := applycreate.BuildContent(bizID, vendor, req, accountID)
    handler := applycreate.NewApplicationOfApplyPermPolicyLibCreate(opt, content)
    result, err := a.create(cts, &proto.CreateCommonReq{}, handler)
    if err != nil {
        return nil, errf.NewFromErr(errf.Aborted, ...)
    }
    ids = append(ids, result.(*core.CreateResult).ID)
}
return &core.BatchCreateResult{IDs: ids}, nil
```

**理由**：ITSM 单据创建是原子操作（创建单据 + 写 DB），单据已创建无法回滚。
任一失败时整体报错可避免产生孤立单据，调用方需重新提交。

---

### 决策 3：Handler 目录结构 — `handlers/permission-policy-library/`

**选择**：
```
handlers/permission-policy-library/
  base.go                          # 共享 base 结构体（嵌入 PolicyLibraryApplier）+ Registry
  apply-create/
    init.go                        # Handler 结构体 + BuildContent 辅助函数 + init() 注册
    check.go                       # CheckReq（含 bk_biz_id 一致性校验）
    create_itsm_ticket.go          # RenderItsmTitle / RenderItsmForm
    deliver.go                     # Deliver + GenerateApplicationContent
```

**理由**：与 `handlers/sub-account/` 一一对应，便于后续扩展 `apply-update/`

---

### 决策 4：Deliver 逻辑 — 调用 PolicyLibraryApplier.ApplyCreate 统一方法

**选择**：在 `deliver.go` 中调用 `ApplyCreate` 统一方法：
```go
resp, err := a.ApplyCreate(kit, vendor, policyLibraryID, []string{accountID})
// 检查 resp.Results[0].Status 是否成功
```

**理由**：`PolicyLibraryApplier.ApplyCreate` 已封装了完整的创建流程（`TCloudCreateCAMPolicy` + `TCloudCreateLocalTemplate` + `RecordApplyAudit`），
通过调用该统一方法避免了在 Deliver 中手动编排多个步骤，逻辑更简洁，结果通过 `ApplyResult.Results[0].Status` 判断是否成功。

## Risks / Trade-offs

- **审批期间库被删除** → Deliver 时重新 `GetPolicyLibraryDetail`，找不到则 DeliverError 状态
- **审批期间账号已人工应用** → `Deliver` 调用 `ApplyCreate`，内部检测重复应用时返回 `ApplyStatusFailed`，Deliver 记录 `DeliverError` 状态
- **批量失败不回滚** → 已提交的 ITSM 单据不撤销，前端需提示用户手动处理。这是现有框架限制。
- **ITSM 流程配置依赖** → 需要在 `approval_process` 表中为 `apply_permission_policy_library` 类型配置服务ID，部署时需手动录入
