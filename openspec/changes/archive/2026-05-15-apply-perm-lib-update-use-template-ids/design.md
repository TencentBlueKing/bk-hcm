## Context

`apply_permission_policy_library_update` 接口当前以 `account_ids` 为入参，在 Biz 申请单流程和 Resource 直接执行路径两条链路上都以账号 ID 作为操作对象。

然而 Update 语义的目标是**刷新已存在的权限模版**——每个权限模版（`permission_template`）已记录了 `account_id`、`cloud_policy_id`（云上 CAM 策略 ID）等全部必要信息。以 `permission_template_id` 为入参，可以直接定位到目标记录，省去了"通过 accountID + libraryID 联合查询模版"的额外 roundtrip，语义也更精确。

两条链路受影响：
1. **Biz 路径**：用户提交含 `permission_template_ids` 的申请单 → ITSM 审批 → Deliver 时按 templateID 更新
2. **Resource 路径**：直接调 apply 接口立即执行，返回每个模版的执行结果

当前 create 和 update 共用 `ApplyPermPolicyLibContent` 作为审批单 DB content 存储结构，两者入参字段已分歧（create 需要 accountID，update 需要 templateID），必须拆分。

## Goals / Non-Goals

**Goals:**
- `apply_permission_policy_library_update` 两条链路（Biz/Resource）均改为以 `permission_template_ids` 为入参
- 响应结果以 `permission_template_id` 标识（新增 `ApplyTemplateResult`）
- 拆分 content 结构体，create/update 各自独立
- Handler 工厂签名改为 sub-account 模式，支持 action-specific 反序列化
- `apply-update` 的 CheckReq 简化（模版存在即已应用，无需重复检查 applied 状态）
- 旧 `ApplyUpdate(accountIDs)` 方法替换为 `ApplyUpdateByTemplateIDs`

**Non-Goals:**
- Create 路径不变（仍使用 `account_ids`）
- 不改动权限模版（permission_template）的数据库结构
- 不改动 ITSM 审批流程本身
- 不改动非 apply 的其他 permission_policy_library 接口

## Decisions

### 决策 1：content 结构体采用三层拆分，工厂函数改为 sub-account 模式

**问题**：create 存 `AccountID`，update 存 `PermissionTemplateID`，无法再共用一个 content struct。

**方案**：
- `ApplyPermPolicyLibBaseContent`：仅含 `Action/Vendor/BkBizID`，用于 `NewHandlerFromApplication` 中第一次反序列化以确定 action
- `ApplyPermPolicyLibCreateContent`：embed base + `PolicyLibraryID` + `AccountID`
- `ApplyPermPolicyLibUpdateContent`：embed base + `PolicyLibraryID` + `PermissionTemplateID`
- 工厂签名改为 `func(opt, base *BaseContent, rawContent string) (handler, error)`，各 handler 内部二次反序列化为 action-specific struct

**理由**：与 sub-account 模式完全对齐，已有实践证明该模式在同类多 action 场景下可维护性好；各 action handler 自持数据，避免共用 struct 带来的字段污染。

**备选**：在原 content struct 上同时保留 `AccountID` 和 `PermissionTemplateID` 字段（其中一个留空）。缺点：字段语义不清，序列化冗余，后续维护混乱，不采用。

### 决策 2：CheckReq 简化——模版存在即已应用

**问题**：原 CheckReq 单独调 `CheckAccountApplied` 验证账号是否已绑定策略库。改为 templateID 后，模版存在本身就意味着已绑定，无需重复查询。

**方案**：CheckReq 改为：
1. 验证 PermissionTemplateID 非空
2. `GetTemplateByID` 查模版（不存在则失败）
3. 验证 `template.PolicyLibraryID == content.PolicyLibraryID`
4. 验证模版对应账号的 biz 在 library 的 biz scope 内

去掉原来的 `CheckAccountApplied` 调用。

### 决策 3：新增 `ApplyTemplateResult`，update 路径响应独立化

**问题**：原 `ApplyAccountResult` 含 `account_id`，update 改为按 templateID 操作后响应字段需对应。

**方案**：新增 `ApplyTemplateResult { permission_template_id, status, reason }` 和 `ApplyPermissionPolicyLibraryUpdateResult { results []ApplyTemplateResult }`，仅在 update 路径使用；create 路径继续使用原 `ApplyAccountResult`。

### 决策 4：`ApplyUpdate` 替换为 `ApplyUpdateByTemplateIDs`

**方案**：在 applier 中新增：
- `GetTemplateByID(kt, templateID)`：按 ID 查权限模版（带 extension）
- `tcloudApplyUpdateForTemplate(kt, library, templateID)`：查模版 → 直接使用模版内的 `cloudID`（cloud policy ID）和 `account_id` → 调用 `TCloudUpdateCAMPolicy` + `TCloudUpdateLocalTemplate`
- `ApplyUpdateByTemplateIDs(kt, vendor, libraryID, templateIDs)`：批量执行

删除旧 `ApplyUpdate(accountIDs)` 方法（无调用方）。

## Risks / Trade-offs

- **BREAKING CHANGE**：两个 API 的请求/响应字段均变更，已有调用方需同步修改 → 变更文档须清晰标注，同步通知前端/调用方
- **DB content 向后兼容**：历史审批单中 content 字段存的是旧结构（含 `account_id`），新工厂函数反序列化时会因字段不匹配而失败 → 这是新功能接口，当前无历史数据，风险可接受；若有数据迁移需求，须另行处理

## Open Questions

无。
