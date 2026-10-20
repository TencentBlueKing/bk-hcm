## Context

bk-hcm 现有的云资源 DAO/data-service 层遵循统一分层约定：`pkg/dal/table` 定义表结构体 + 列描述符 + `InsertValidate()`/`UpdateValidate()`；`pkg/dal/dao` 提供 `Interface + 实现`，仅 data-service 允许直连 DB；`cmd/data-service` 暴露 HTTP CRUD；上层（hc-service/cloud-server）通过 `pkg/client` 调用，不直连 DB。本变更参考项目内最近落地（2026-03~04）的同构模块 `account-secret` / `sub-account-secret`（`pkg/dal/table/cloud/account-secret`、`pkg/dal/dao/cloud/account-secret`）作为代码结构基准，而非更早（2024-02）的 `cert` 模块，因为前者是当前最新的分层写法（`ModifySQLOpts(orm.NewInjectTenantIDOpt(...))` 租户注入、`utils.RearrangeSQLDataWithOption` 动态 SET 表达式等写法均以 account-secret 为准）。

独占集群表需要支撑三个不同调用方的写入语义：
1. 「独占集群同步」子任务：批量创建（新增集群，`bk_biz_id` 恒为 -1）与批量更新云属性（不动 `bk_biz_id`）
2. 「资源视角查询+分配业务」子任务：仅修改 `bk_biz_id`（把未分配集群指给业务）
3. 「CLB 查询接口修改」等子任务：仅读（List）

前两者若共用一个 Update 接口，同步逻辑的请求结构体若意外携带 `bk_biz_id` 字段就可能覆盖分配关系；如何隔离二者是本设计的核心决策之一。

**已核实的项目现有惯例**：负载均衡自身的 DAO（`pkg/dal/dao/cloud/load-balancer/load_balancer.go`）并未为"分配业务"拆出独立的 DAO 方法——`LoadBalancerInterface` 只有一个通用的 `Update(kt, expr, model)`。分配业务（`cmd/data-service/service/cloud/load-balancer/update.go` 的 `BatchUpdateLbBizInfo`）与同步更新云属性（`update.go` 的其它 handler）复用同一个 `Update` 方法，靠两点实现隔离：① data-service 层用专属窄请求体 `BizBatchUpdateReq{IDs, BkBizID}`，物理上不携带其它云属性字段；② `Update` 内部的 `utils.RearrangeSQLDataWithOption` 通过 `isBlank()` 只把**非零值**字段拼进 SQL 的 `SET` 子句——分配业务的 handler 构造的 model 只赋值 `BkBizID`（非零），云属性更新的 handler 构造的 model 不赋值 `BkBizID`（零值），二者天然不会互相覆盖，无需在 DAO 层物理拆分两个方法。本设计采纳该惯例（见决策 2）。

## Goals / Non-Goals

**Goals:**
- 定义 `load_balancer_exclusive_cluster` 表结构、SQL 迁移、表名/枚举/审计类型/云资源类型的注册
- 提供 DAO 层批量 CRUD，作为后续 5 个子任务的稳定契约
- 提供 data-service 层对应的 5 个 HTTP 接口（含独立的 `bk_biz_id` 分配业务接口）与客户端封装，接口形态与字段对齐已发布的 API 文档（`list_exclusive_cluster.md`、`assign_exclusive_cluster_to_biz.md`）

**Non-Goals:**
- 不实现适配层 `DescribeExclusiveClusters` 接入与 res-sync 同步逻辑本身（归属「独占集群资源同步」子任务）
- 不实现 `TCloudClbExtension` 的 `Exclusive`/`ClusterIds`/`ClusterTag` 扩展字段（归属「CLB 同步逻辑调整」子任务）
- 不实现 cloud-server 层的鉴权、"是否已分配"前置校验（V-15）、审计 build 方法与 `cloud_resource_assign_audit.go` 分支注册（归属「资源视角查询+分配业务逻辑」子任务）
- 不实现购买链路改造、CLB 列表/详情接口的 `exclusive`/`sla_type` 展示与筛选（归属其余两个子任务）
- 不定义 `TCloudLBSlaType` 枚举（已在需求单第 3 轮澄清明确排除）

## Decisions

### 1. `clb_resource_count` 采用单列派生值，不落库 `resource_count`/`idle_resource_count`

**选择**：表内只保留 `clb_resource_count`（集群内已绑定的 CLB 实例数），不新增 `resource_count`/`idle_resource_count`/`sync_time` 三列。

**原因**：需求单第 2 轮澄清已核实 `res-sync` 的 `common.Diff` 机制支持"先计算再比较"的写法（参考 `isLBExtensionChange` 对 `BandWidth` 的处理），因此同步逻辑可以在写入前用云上 `ResourceCount - IdleResourceCount` 算出差值再落库，无需额外两列。DAO 层只需保证 `clb_resource_count` 可正常读写，具体计算逻辑属于「独占集群资源同步」子任务。

**已知偏差**：现有 API 文档 `list_exclusive_cluster.md` 的"查询参数介绍"章节仍列出 `resource_count`/`idle_resource_count`/`sync_time` 三个可查询字段（文档未完全同步最终方案，对应需求单 Q-001），本设计以 DAO 层落地的最终表结构为准，后续需求单负责人需同步更新该文档，本变更不修正该文档内容。

**备选方案**：落库 `resource_count` + `idle_resource_count` 两列 → 已否决，需求单已明确选择方案 A（仅 `clb_resource_count` 一列）。

### 2. DAO 层保留单一通用 `BatchUpdateWithTx`，`bk_biz_id` 隔离下沉到请求体层 + 零值跳过机制

**选择**：DAO 层只提供一个通用的 `BatchUpdateWithTx(kt, tx, models []Table) error`（逐条 `UpdateValidate()` + `utils.RearrangeSQLDataWithOption` 生成动态 SET 表达式，与 `account-secret`/`sub-account-secret` 的 `BatchUpdate` 写法一致），不额外定义 `BatchUpdateBizIDWithTx`。协议隔离改为在**更上层**实现：
- data-service 层保留两个独立 HTTP 接口与两个独立、字段收窄的请求体——`PATCH /vendors/{vendor}/load_balancer_exclusive_clusters`（云属性更新，`BatchUpdateReq` 不含 `bk_biz_id` 字段）与 `PATCH /load_balancer_exclusive_clusters/biz`（分配业务，`BatchUpdateBizIDReq{ClusterIDs, BkBizID}` 只含这两个字段），物理上不给调用方夹带其它字段的机会；
- 两个 handler 都调用同一个 `BatchUpdateWithTx`，只是构造的 `Table` model 字段不同：分配业务的 handler 只赋值 `ID`/`BkBizID`/`Reviser`（其余字段留零值），云属性更新的 handler 赋值云属性字段但不赋值 `BkBizID`（保持零值）；`RearrangeSQLDataWithOption` 内部的 `isBlank()` 只把非零值字段拼进 SQL 的 `SET` 子句，零值字段自动跳过，天然不会互相覆盖。

**原因**：这是负载均衡自身现有代码的真实写法（`LoadBalancerDao.Update` 被 `BatchUpdateLbBizInfo` 与云属性更新 handler 共用，见 Context 小节引用），本设计对齐该惯例，不额外新增一层"物理隔离方法"的 DAO 契约，减少与既有代码风格的偏离，也少维护一个方法。项目内还存在一个更通用的跨资源方案 `CloudDao.AssignResourceToBiz(kt, tx, resType, expr, bizID)`（按 `resType.ConvTableName()` 动态解析表名，一个方法服务 security-group/vpc/subnet/eip/cvm/disk/route-table 等多种资源），但 `LoadBalancerCloudResType` 本身未注册进其 `assignResAuditTypeMap`，即 CLB 也没有走这条通用入口；本变更同样不接入该通用机制（会引入跨模块耦合且超出 DAO 层职责边界），维持"data-service 专属窄接口 + 通用 DAO 方法"的两层结构。

**代价**：相比"编译期就不存在 `bk_biz_id` 字段"的物理隔离，零值跳过机制存在一个理论边界情况——如果未来误传 `bk_biz_id` 恰好等于其零值语义（本表零值是 `0`，而合法业务 ID 与「未分配」标记 `-1` 均不会自然产生 `0`），该字段会被跳过而不会被误写为 `0`；`BatchUpdateBizIDReq.BkBizID` 的校验（`validate:"gt=0"`）进一步保证不会传入 `0`，风险可控。

**备选方案 A**：DAO 层拆两个物理隔离的方法（`BatchUpdateWithTx` + `BatchUpdateBizIDWithTx`）→ 已否决，与负载均衡现有 DAO 写法不一致，多维护一个方法却没有换来负载均衡自己都没采用的额外安全收益。

**备选方案 B**：单一 `BatchUpdateWithTx` 接口 + 请求体含可选 `bk_biz_id` 指针字段，靠调用方"不传即不改" → 已否决，请求体层面不做隔离，任何调用方疏忽传入非零值都会破坏分配关系，且不同调用场景（同步 vs 分配）混在同一个请求体类型里语义不清晰。

### 3. Extension 使用强类型 Go 结构体，接口层与 DB 层分别用结构体和 `types.JsonField` 表达

**选择**：`pkg/api/core/cloud/load-balancer/exclusive_cluster.go` 定义 `TCloudExclusiveClusterExtension` 结构体（12 个标量字段 + `ClustersZone` 嵌套对象），DAO/Table 层沿用项目惯例用 `types.JsonField`（原始 JSON 字符串）存储；data-service handler 负责在两者间序列化/反序列化。字段定义以已发布的 API 文档 `list_exclusive_cluster.md` 的 `extension[tcloud]` 章节（12 字段 + `clusters_zone`）为准，而非方案设计 iWiki 5.1 节仅举例的 4 个字段。

**原因**：项目内同类模式（`permission-template-crud`）已验证"Table 层 JsonField + Core API 层强类型 + handler 序列化"的分层写法；强类型可以在 API 层获得字段校验与 IDE 补全，同时不改变 DB 层通用的 JSON 存储方式，避免为每个 vendor 扩展字段单独加表列。

**待办（不阻塞本变更）**：这 12 个字段与腾讯云 SDK `ExclusiveCluster` 结构体的逐字段对应关系，留给「独占集群资源同步」子任务在实现同步逻辑时核对（需求单 Q-001）。

### 4. 唯一键不含 vendor 维度

**选择**：`UNIQUE KEY (cloud_id, account_id, tenant_id)`，不加 `vendor`。

**原因**：需求单已澄清本期只支持 tcloud，且同一 `account_id` 不会跨 vendor 复用；与需求单第 1 轮澄清结论一致。若未来支持多云，可在该唯一键上追加 `vendor` 列并做数据迁移，不影响本变更的既有数据。

### 5. `max_conn` 使用 `*int64`（nullable），不用 `0` 表达"无此值"

**选择**：`max_conn` 落库为可空列（`db:"max_conn" json:"max_conn"`，Go 侧用指针类型），云上 STGW 集群该字段返回 `null` 时也存 SQL `NULL`，不归一化为 `0`。

**原因**：`0` 是合法的业务取值空间之外的"无意义默认值"，会与"确实是 0 连接数"混淆；nullable 语义与需求单 AC 描述及前端"展示 `—`"的约定一致。

## Risks / Trade-offs

- **[Risk] Extension 12 字段与云 SDK 字段名/类型不完全对应** → **Mitigation**：本变更只按 API 文档定义结构体，不接入真实同步逻辑；字段核对显式转交下游「独占集群资源同步」子任务处理，避免本变更被云 SDK 细节阻塞。
- **[Risk] `clb_resource_count` 派生计算逻辑在同步侧实现，DAO 层无法保证语义正确性（如集群刚创建时该列为 0 是否符合预期）** → **Mitigation**：DAO 层的 `InsertValidate()` 不对该列做业务语义校验，只做类型/长度校验；正确性验证放在下游子任务的集成测试中。
- **[Risk] `PATCH /load_balancer_exclusive_clusters/biz` 在本变更中不做"是否已分配"业务校验（V-15），单独调用可能被误用为绕过前置校验的后门** → **Mitigation**：需求单已明确该校验属于 cloud-server 层职责，DAO/data-service 层保持"按传入参数正确写库"的单一职责；接口注释与 API 文档需显著标注"仅供 cloud-server 内部调用，不做业务前置校验"。
- **[Risk] DAO 层不再有物理隔离方法，若未来有新 handler 误用 `BatchUpdateWithTx` 且不慎给 model 的 `BkBizID` 赋了非零值** → **Mitigation**：`BkBizID` 只应由分配业务 handler 赋值，其它 handler（同步更新云属性）代码审查时需明确检查 model 构造代码不触碰该字段；这与负载均衡现有代码面临的风险等级一致，非本变更独有。
- **[Trade-off] 批量创建审计（Create）在 DAO 事务内直接写入，但 Update/Delete 审计的 build 方法留给下游子任务** → 与需求单范围一致，`enumor.LoadBalancerExclusiveClusterAuditResType` 常量在本变更注册，但 `cloud_resource_assign_audit.go` 的分支接入不在本变更内，短期内"分配业务"接口调用后不会产生审计记录，直到下游子任务补齐。

## Migration Plan

纯新增表和新增接口，不影响任何现有表结构或接口行为，无需数据迁移。

- 部署顺序：SQL 迁移脚本随发布流程重命名为下一个序号（当前最新 `0053`）后随版本发布执行，`START TRANSACTION ... COMMIT` 保证建表 + `id_generator` 种子的原子性
- 回滚策略：新表无历史数据依赖，如需回滚可直接 `DROP TABLE load_balancer_exclusive_cluster` 并删除 `id_generator` 对应记录；由于所有下游子任务均在本变更之后才开始写入/读取该表，回滚不会影响其它现有功能
