## Context

项目内已有 3 处结构几乎完全相同的"分配到业务"实现：`cert.AssignCertToBiz`、`load-balancer.AssignLbToBiz`、`network-interface.AssignNetworkInterfaceToBiz`，均遵循同一套分层模式（详见 `docs/reqs/独占集群分配业务.md`「工时预估与规模评分」章节的逐层对照表）：

1. `cloud-server` service 层 handler：解析请求 → `common.ValidateTargetBizID` 校验目标业务在账号 `usage_biz_ids` 范围内 → `ListResBasicInfo` 取资源账号信息 → 按账号维度构造 `meta.ResourceAttribute` 做 `AuthorizeWithPerm` 权限校验 → 调用 `logics` 层执行分配
2. `cloud-server` logics 层：`ValidateBeforeAssign`（按 ID 查询资源，检查是否已分配给"其它"业务）→ 记分配审计（`logicaudit.NewAudit().ResBizAssignAudit`）→ 调 data-service 批量更新 `bk_biz_id`
3. `data-service` audit 层：`cloud_resource_assign_audit.go` 的 `buildAssignAuditInfo` switch 新增一个分支，分发到对应资源的 `XxxAssignAuditBuild` 方法

已核实的项目现有惯例（`load-balancer.ValidateBeforeAssign`，`cmd/cloud-server/logics/load-balancer/assign.go`）：已分配判定条件是"`bk_biz_id != constant.UnassignedBiz && bk_biz_id != 目标bizID`才拒绝"，即已分配给目标业务本身视为幂等放行——这与需求单第 1 轮澄清结论一致，也是本变更采纳的判定条件，而非 TAPD 原始描述/方案设计 V-15 字面的"只要 `bk_biz_id != -1` 就拒绝"。

独占集群与 CLB 本身的关键差异：CLB 分配业务时需要级联分配监听器、目标组等关联资源（`GetLoadBalancerRelateResIDs`/`AssignLoadBalancerRelated`）；独占集群没有子资源，分配粒度是单一表的单一字段，逻辑上更接近 `cert.Assign`（无关联资源级联）而非 `load-balancer.AssignTCloud`（有级联）。

data-service 层的 `PATCH /load_balancer_exclusive_clusters/biz`（B-04 已实现）明确设计为"不做业务前置校验、鉴权、审计，仅供 cloud-server 内部调用"的窄接口，本变更是该接口设计上就预留好的调用方。

## Goals / Non-Goals

**Goals:**
- 新增 cloud-server 分配业务接口，鉴权、前置校验、审计三件套与项目内其它 `XxxAssignToBiz` 接口结构对齐
- 已分配判定条件与 `load-balancer.AssignLbToBiz` 一致（分配给目标业务本身放行，分配给其它业务拒绝），批量整批校验整批执行

**Non-Goals:**
- 不支持解绑（`bk_biz_id` 改回 -1）或重新分配（已分配给其它业务的集群改分配给新业务），本期需求单已明确排除
- 不修改 data-service 层 `PATCH /load_balancer_exclusive_clusters/biz` 接口本身（B-04 已完成且设计上就不做校验，本变更只是新增其调用方）
- 不处理独占集群的关联资源级联分配（独占集群没有子资源，不同于 CLB 需要级联监听器/目标组）

## Decisions

### 1. logics 层不做关联资源级联，直接对齐 `cert.Assign` 而非 `load-balancer.AssignTCloud`

**选择**：新增 `cmd/cloud-server/logics/load-balancer/exclusive_cluster_assign.go`，内部结构为"前置校验 → 记审计 → 批量更新"三步直线流程，不设计 `GetXxxRelateResIDs`/`AssignXxxRelated` 之类的级联分配步骤。

**原因**：独占集群是叶子资源，没有监听器、目标组等下挂子资源，`load-balancer.AssignTCloud` 的级联逻辑对本资源没有对应物；强行套用会引入不必要的空实现步骤。`cert.Assign` 才是结构对等的参照（同为叶子资源、单表更新）。

**备选方案**：完全复制 `load-balancer.AssignTCloud` 的函数骨架，级联步骤留空 → 已否决，会让代码读者误以为未来需要补充级联逻辑，增加认知负担且没有实际收益。

### 2. 已分配判定条件采用 `load-balancer.AssignLbToBiz` 的"仅其它业务才拒绝"语义，覆盖方案设计 V-15 字面表述

**选择**：`ValidateExclusiveClusterBeforeAssign` 查询 `cluster_ids` 当前 `bk_biz_id`，只有 `bk_biz_id != constant.UnassignedBiz && bk_biz_id != 目标bizID` 的记录才计入"已分配"拒绝集合；命中目标业务本身的记录视为幂等，随批次一起放行。批次内只要拒绝集合非空，整批返回 `InvalidParameter`，不做部分成功。

**原因**：需求单（1069995598138281829）第 1 轮澄清已与用户确认此结论，并核实是与 `load-balancer.AssignLbToBiz` 一致的项目既有惯例；方案设计 V-15 原文"分配前查命中 `bk_biz_id != -1` 则拒绝"的字面表述未考虑幂等场景，本设计以澄清结论为准，草稿 API 文档措辞需在实现时同步修正（详见 `docs/reqs/独占集群分配业务.md` 假设 2、Q-002）。

**备选方案**：严格按 V-15 字面/`cert.Assign` 的"只要已分配（不区分是否目标业务）就拒绝" → 已否决，与需求单澄清结论及项目内负载均衡自身的分配接口行为不一致，会导致运营对同一批集群重复提交分配请求时出现不必要的报错。

### 3. 审计构建方法挂载在已有的 `loadbalancer.LoadBalancer` 审计结构体上，新增文件不新增类型

**选择**：在 `cmd/data-service/service/audit/cloud/load-balancer/` 包下新增 `exclusive_cluster.go` 文件，为已有的 `*LoadBalancer` 类型（`base.go` 定义，`dao dao.Set` 字段）新增方法 `LoadBalancerExclusiveClusterAssignAuditBuild(kt, assigns)`，并新增配套的 `ListExclusiveCluster(kt, dao, ids)` 查询辅助函数（与 `ListLoadBalancer`/`ListTargetGroup`/`ListListener` 同一模式：按 ID 批量查表，转 `map[string]tablelb.LoadBalancerExclusiveClusterTable`）。`cloud_resource_assign_audit.go` 的 `buildAssignAuditInfo` switch 新增 `case enumor.LoadBalancerExclusiveClusterAuditResType: audits, err = ad.loadBalancer.LoadBalancerExclusiveClusterAssignAuditBuild(kt, assigns)`。

**原因**：`ad.loadBalancer`（`*loadbalancer.LoadBalancer`）已经是 `cloud_resource_assign_audit.go` 里 `LoadBalancerAuditResType` 分支的接收者，独占集群是 CLB 的附属资源，复用同一个审计结构体（而非新建 `ExclusiveClusterAudit` 类型）符合"审计构建方法按资源大类聚合、不按每个子资源类型建一个结构体"的项目现状（`LoadBalancer` 结构体已经承载 CLB 自身、监听器、URL 规则等多种子资源的审计构建方法）。

**备选方案**：新建独立的 `ExclusiveClusterAudit` 结构体与 `dao.Set` 字段 → 已否决，`cloud_audit.go` 需要多注册一个字段与初始化代码，而独占集群作为 CLB 附属资源的审计逻辑足够简单（无子资源级联），不足以独立成一个新审计结构体。

### 4. `AssignExclusiveClusterToBizReq` 复用 `AssignLbToBizReq` 的字段设计与命名风格

**选择**：`pkg/api/cloud-server/load-balancer` 新增 `AssignExclusiveClusterToBizReq{ClusterIDs []string, BkBizID int64}`，字段命名、`Validate()` 校验逻辑（非空、`BatchOperationMaxLimit` 上限、`BkBizID > 0`）与 `AssignLbToBizReq` 保持一致的风格，而不是直接复用 data-service 层的 `ExclusiveClusterBatchUpdateBizIDReq`。

**原因**：cloud-server 层请求体与 data-service 层请求体在项目内历来是分别定义的两个类型（即便字段相同），职责边界清晰（cloud-server 面向外部 API 契约，data-service 面向内部服务间协议），且 cloud-server 层的 `Validate()` 需要返回 `errf.InvalidParameter` 包装的错误类型，与 data-service 层的裸 `error` 不同。

**备选方案**：直接复用 `dataproto.ExclusiveClusterBatchUpdateBizIDReq` 作为 cloud-server 对外请求体 → 已否决，让 cloud-server 的 API 契约直接耦合到 data-service 内部协议类型，未来任一层调整字段都会影响另一层，且错误包装类型不匹配。

## Risks / Trade-offs

- **[Risk] `ValidateExclusiveClusterBeforeAssign` 与 `common.ValidateTargetBizID` 之间存在两次查询独占集群基础信息（分别为账号信息与 `bk_biz_id` 状态）** → **Mitigation**：与 `load-balancer.AssignLbToBiz` 现状完全一致（`ValidateTargetBizID` 查账号、`ValidateBeforeAssign` 单独查 `bk_biz_id`），非本变更引入的新增开销，属于项目现有模式的一致代价，不在本变更中优化合并
- **[Risk] 分配前置校验（cloud-server 层）与批量更新（data-service 层）非同一事务，理论上存在极小的竞态窗口（两次请求并发分配同一集群到不同业务）** → **Mitigation**：与 `load-balancer`/`cert` 等既有分配接口面临的风险等级一致，非本变更独有；数据库唯一约束与实际业务场景（运营低频操作）使该竞态发生概率极低，不在本变更范围内引入分布式锁等额外机制
- **[Trade-off] 审计构建方法复用 `LoadBalancer` 结构体而非独立类型，未来若独占集群审计逻辑变复杂（如支持解绑/重新分配的审计），`base.go`/`exclusive_cluster.go` 文件可能超过单文件行数建议上限（800 行）** → 当前 `exclusive_cluster.go` 预计新增约 40-50 行，远低于阈值；若未来复杂度上升，可在届时再拆分独立类型，不提前过度设计

## Migration Plan

纯新增接口与新增审计分支，不涉及表结构或已有接口行为变更，无需数据迁移。部署顺序：随版本发布即可生效，依赖 B-04（已发布）；回滚只需下线该路由与还原审计 switch 分支，不影响任何存量数据（分配关系数据本身由 B-04 的表管理，本变更不涉及该表结构）。
