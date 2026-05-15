## Context

系统已实现权限策略库的"应用（创建）"和"应用（更新）"接口，前端在调用这两个接口时需要提供目标二级账号ID列表。但目前缺乏一个查询接口来告知前端"哪些账号还没有应用过此策略库"，导致前端无法自动筛选候选账号。

现有可复用能力：
- `applier.GetPolicyLibraryDetail`：获取策略库详情（含 BkBizIDs）
- `applier.CheckAccountApplied`：逐账号检查是否已应用（不适合批量场景）
- `data-service` 的账号列表接口支持 `vendor` 和 `bk_biz_id` 过滤
- `data-service` 的 `ListPermissionTemplateExt` 支持按 `policy_library_id` 过滤

## Goals / Non-Goals

**Goals:**
- 实现 `GET .../unapplied_account_ids` 只读查询接口
- 在 applier 中封装可复用的批量查询辅助方法
- 全量返回，不分页

**Non-Goals:**
- 不实现 `account_ids`（已应用账号）接口
- 不做权限管控（仅 Find 鉴权）
- 不涉及任何写操作

## Decisions

### Decision 1：新增 3 个 applier 辅助方法而非在 handler 直接实现

**选择**：在 `PolicyLibraryApplier` 中新增 `ListUnappliedAccountIDs`（公开入口）、`listAllInScopeAccountIDs`（私有）、`listAllAppliedAccountIDs`（私有）三个方法。

**理由**：与现有 `ApplyCreate`/`ApplyUpdate` 的封装模式一致；辅助方法可供未来 `account_ids` 接口复用；handler 只做校验和鉴权，业务逻辑集中在 applier 中。

### Decision 2：使用 `slice.NotIn` 计算差集

**选择**：先获取全量候选账号ID（inScopeAccountIDs），再获取已应用账号ID（appliedAccountIDs），调用 `slice.NotIn(appliedAccountIDs, inScopeAccountIDs)` 得到未应用列表。

**理由**：`slice.NotIn` 语义清晰，时间复杂度 O(n+m)；避免逐账号查询 permission_template 表（N+1 问题）。

### Decision 3：查候选账号时加 vendor 过滤

**选择**：`listAllInScopeAccountIDs` 查账号表时过滤 `vendor = vendor AND bk_biz_id IN library.BkBizIDs`，使用 `tools.ExpressionAnd(tools.RuleEqual("vendor", vendor), tools.RuleIn("bk_biz_id", batch))` 构建过滤条件。

**理由**：路由带了 vendor 参数，语义上策略库为特定云厂商服务，候选账号也应属于相同厂商；与 `ListPermissionPolicyLibrary` 的 vendor 过滤模式一致。

### Decision 4：分页全量扫描策略

**选择**：`listAllInScopeAccountIDs` 先用 `slice.Split` 将 `bizIDs` 按 `DefaultMaxPageLimit`（500）分批，每批对账号表发起分页查询（内层 for 循环，`req.Page.Start` 递增），直到返回数量 < Limit 为止。`listAllAppliedAccountIDs` 同理（无外层批次，直接 `start` 递增）。

**理由**：`RuleIn` 条件支持的值列表长度有上限，对 bizIDs 分批可避免超限；账号数量和模板数量可能超过单页上限（500），内层分页确保完整覆盖。

### Decision 5：handler 合并到 `list.go`

**选择**：`ListPermissionPolicyLibraryUnappliedAccountIDs` handler 追加到已有的 `list.go` 文件，不单独新建文件。

**理由**：该 handler 属于查询类操作，与 `ListPermissionPolicyLibrary` 同属 list 语义，合入 `list.go` 保持文件分工一致；减少不必要的文件碎片化。

## Risks / Trade-offs

- **[风险] 读取期间数据变化**：扫描多页期间若有新账号加入或新模板创建，可能导致结果轻微不一致。→ 这是只读查询接口的固有特性，与现有 List 接口一致，可接受。
- **[风险] 候选账号量大时性能**：若某个业务下账号数千，多页扫描耗时较长。→ 接口定位为管理操作（非高频），可接受；后续可加缓存优化。
- **[权衡] 不分页全量返回**：接口文档明确要求全量返回，前端用于渲染选择列表。单次响应可能较大，但账号数量有限（通常百级别）。
