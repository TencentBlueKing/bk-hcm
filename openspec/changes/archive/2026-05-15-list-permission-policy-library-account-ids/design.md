## Context

权限策略库（Permission Policy Library）模块已有 CRUD 及 Apply 操作，数据模型上策略库通过 `permission_template` 表与二级账号（account）建立关联关系（`policy_library_id` 字段）。

当前已有相关基础设施：
- `applier.listAllAppliedAccountIDs`：分页扫描 `permission_template` 表，返回指定策略库的全部已应用账号 ID（去重）
- `applier.listAllInScopeAccountIDs`：根据 vendor + bizIDs 查询账号表，返回范围内账号 ID
- `buildLibraryAccountCountMap`：统计策略库关联账号数量

本次变更在 cloud-server 层新增两个 GET 接口，利用上述已有方法实现，无需改动 data-service 层。

## Goals / Non-Goals

**Goals:**
- 提供 Resource 接口：全量返回策略库已关联的二级账号 ID（去重，不分页）
- 提供 Biz 接口：返回策略库已关联且管理业务为指定 bk_biz_id 的账号 ID（去重，不分页）
- 两个接口均按职责分别做权限校验

**Non-Goals:**
- 不新增 data-service 层接口
- 不支持分页
- 不返回账号详细信息，仅返回 ID

## Decisions

### 决策 1：Biz 接口的过滤方式

**选择**：先调用 `listAllAppliedAccountIDs` 获取全部已应用账号 ID，再批量查询账号表，过滤出 `bk_biz_id` 等于路径参数的账号。

**理由**：
- 代码最大化复用已有方法，无需新增 DS 接口
- 账号数量通常不大，两步查询性能可接受
- 备选方案（直接查账号表求交集）逻辑更复杂，且 `listAllInScopeAccountIDs` 要求传入 bizIDs 切片，不适合单 biz 场景

### 决策 2：权限校验

- **Resource 接口**：`meta.PermissionPolicyLibrary + meta.Find + ResourceID=id`，与现有 `ListPermissionPolicyLibraryUnappliedAccountIDs` 保持一致
- **Biz 接口**：`meta.Biz + meta.Access + BizID=bk_biz_id`，业务访问标准模式，与 account-secret、load-balancer 等接口一致

### 决策 3：是否校验 bk_biz_id 在策略库 BkBizIDs 中

**选择**：Biz 接口 **需要** 校验，若 bk_biz_id 不在策略库的 BkBizIDs 中，返回 400 参数错误，提前拒绝无效请求。

**理由**：
- 明确的业务语义：biz 接口本身表达"业务视角"，若该业务与策略库无关联，应快速失败而非静默返回空列表，避免调用方误以为"查询成功但无数据"
- 安全性：防止任意业务探测策略库内容
- 实现成本低：`applier.GetPolicyLibraryDetail` 已有，获取 library 后做 BkBizIDs 包含检查即可
- 备选（不校验）：行为模糊，调用方难以区分"无关联账号"与"业务无权关联此库"两种语义

## Risks / Trade-offs

- [风险] 账号数量极大时，两步查询（permission_template → account）有性能压力 → 缓解：两步均有分页，单次上限 500 条，可接受
- [风险] `listAllAppliedAccountIDs` 当前是 `applier` 的私有方法 → 缓解：直接在 Handler 中内联类似逻辑，或在 `applier.go` 新增公开辅助方法
