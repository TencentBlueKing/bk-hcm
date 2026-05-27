## Context

权限策略库（permission_policy_library）可以通过 Apply 操作将策略同步到多个二级账号，每次同步会在本地创建一条 permission_template 记录。前端需要一个接口来展示某个策略库当前可更新的账号列表（即已应用该策略库、且账号仍在策略库业务范围内的权限模版）。

当前 `PolicyLibraryApplier` 中已有：
- `GetPolicyLibraryDetail()`：查询策略库详情（含 bk_biz_ids）
- `listAllInScopeAccountIDs()`（私有）：查询当前业务范围内的所有账号ID
- `listAllAppliedAccountIDs()`（私有）：查询已应用的账号ID

## Goals / Non-Goals

**Goals:**
- 返回指定策略库下已应用且当前在业务范围（bk_biz_ids）内的权限模版，全量不分页
- 保证返回的模版所对应的账号，能被 Apply Update 成功处理（不出现因业务范围变化导致的误导）

**Non-Goals:**
- 不返回业务范围外的历史模版（即使曾经应用过）
- 不支持分页

## Decisions

### 决策1：按业务范围过滤，而非全量返回

接口用于"构建可更新账号列表"，Apply Update 本身会校验 `CheckAccountsBizInScope`，若不过滤会导致用户看到"可选"账号但操作失败的 UX 问题。

**结论**：仅返回账号仍在策略库 `bk_biz_ids` 范围内的模版。

### 决策2：方案 B —— 扫描模版后内存过滤，而非 batch IN

**备选方案 A**：先算 `applied ∩ in_scope` 集合，再按 account_id 分批（500）IN 查 permission_template 表。
- 缺点：同一张表扫两遍（第一遍取 account_id，第二遍取完整对象）

**选定方案 B**：
1. 预先获取 `in_scope_set`（`map[string]struct{}`）
2. 分页扫 permission_template（`policy_library_id = {id}`）
3. 每条模版在内存中判断 `AccountID ∈ in_scope_set`

优点：只扫一遍模版表；分页由模版表自然控制（每页500），无需额外 batch IN。

### 决策3：在 applier 上封装 `ListTemplatesInScope` 公有方法

`listAllInScopeAccountIDs` 为私有方法，handler 无法直接调用。在 `PolicyLibraryApplier` 上新增公有方法 `ListTemplatesInScope`，封装"获取业务范围内模版"的完整逻辑，handler 通过 applier 调用。

公有方法签名为 `ListTemplatesInScope(kt *kit.Kit, vendor enumor.Vendor, libraryID string) (any, error)`，内部通过 vendor switch 分发到各云厂商私有实现（如 `tcloudListTemplatesInScope`），返回 `any` 以支持多云扩展。响应类型 `PermissionPolicyLibraryPermTmplResult` 的 `Details` 字段同样使用 `any` 类型，在序列化时包含各厂商差异字段（如 TCloud 的 `extension.cloud_type`）。

## Risks / Trade-offs

- **内存开销**：`in_scope_set` 需加载所有业务范围内的账号ID到内存。正常情况下账号数量可控（百到千量级），风险低。
- **数据一致性**：业务范围（bk_biz_ids）可能在查询过程中变更，但属于可接受的最终一致。
