## Context

cloud-server 层已有 Create / List / Update 三个接口，底层 data-service 的 BatchDelete（DAO → handler → Global Client）也已就绪。本次只需在 cloud-server 层补充 Delete handler 和路由注册。

参考实现：`update.go`（同目录，相同的鉴权 + 审计模式）、`account/update.go`（先查记录再校验 vendor 的模式）。

## Goals / Non-Goals

**Goals:**
- 在 cloud-server 层实现单条删除接口 `DELETE /vendors/{vendor}/permission_policy_libraries/{id}`
- 删除前校验 vendor 与记录实际 vendor 一致
- 删除前记录审计日志
- 预留云权限模板关联检查位（TODO）

**Non-Goals:**
- 云权限模板关联检查的实际实现（该功能尚不存在）
- 软删除 / 回收站
- 批量删除（接口只接受单个 id）

## Decisions

### 1. 删除前查询记录校验 vendor

通过 `Global.PermissionPolicyLibrary.ListPermissionPolicyLibrary` 查询 `filter: id = {id}`，一次请求同时完成：记录是否存在 + vendor 是否匹配。

**替代方案**：不校验 vendor，直接用 `And(id={id}, vendor={vendor})` 作为删除条件 —— 拒绝，因为这样无法区分"记录不存在"和"vendor 不匹配"两种错误场景，用户体验差。

### 2. 通过 Global Client 删除

根据现有架构设计决策（参考 `archive/2026-03-23-add-permission-policy-library/design.md`），读/删操作走 Global Client。因此不需要按 vendor switch 分发，直接调用 `Global.PermissionPolicyLibrary.BatchDelete`。

### 3. 审计放在 cloud-server 层

与 update.go 保持一致，在 CS 层调用 `svc.audit.ResDeleteAudit`，在实际删除之前记录。

### 4. 关联检查暂留 TODO

云权限模板功能尚未实现，删除时预留 TODO 注释，后续补充实际检查逻辑。

## Risks / Trade-offs

- [关联检查缺失] 在云权限模板实现之前，已关联的策略库可被删除 → 当前该关联不存在，无实际风险；后续实现模板功能时需同步补充检查
