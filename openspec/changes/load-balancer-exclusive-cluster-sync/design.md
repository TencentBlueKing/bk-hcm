## Context

Wave 1 DAO 已落地：`load_balancer_exclusive_cluster` 表、data-service 5 个接口、tcloud `BatchCreate`/`BatchUpdate` 与 global `List`/`BatchDelete`/`BatchUpdateBizID`。`LoadBalancerExclusiveClusterCloudResType` 已注册。同步链路尚未接入。

现有全量同步分层：cloud-server `SyncAllResource`（`syncFuncMap` + `getSyncOrder`，遇错 fail-fast）→ hc-service `POST /.../sync` → `handler.ResourceSyncV2`（先拉全量云实例、再按 `allCloudIDMap` 删、再 `Sync`）→ res-sync `common.Diff`（`cloud_id`）→ data-service。CLB / 安全组已走 V2。

云接口：`DescribeExclusiveClusters`，SDK `tclb.Cluster`。本期接全量同步与按云上 ID 的条件同步。

约束：对齐 iWiki 6.1 与 DAO 决策（单列 `clb_resource_count`、更新不写 `bk_biz_id`、无 `sync_time`）；核心同步路径只插缝，不顺带重构。

## Goals / Non-Goals

**Goals:**
- 按账号 + 地域把公网 TGW/STGW 独占集群同步进表（新增未分配、更新云属性、删除云上已无）
- 支持按 `cloud_ids` 条件同步独占集群，复用全量 `/sync` 入口，清理范围限定在指定 ID 内
- 全量同步时独占集群排在 CLB 之前

**Non-Goals:**
- 不改 CLB 同步、不补 `TCloudClbExtension.Exclusive` / `ClusterIds` / `ClusterTag`（与集群表同步无耦合）
- 不改 DAO / 表结构 / data-service CRUD
- 不同步 Private、VPCGW（见决策 3）
- 不接 `DescribeClusterResources`、闲置 VIP、分配业务、购买链路
- 不落库云 `Tag`、不落 `resource_count` / `idle_resource_count` / `sync_time`

## Decisions

### 1. hc-service 用 ResourceSyncV2，Next 串行 Limit=100，并发走 baseHandler 默认

**选择**：handler 实现 `HandlerV2[typeslb.TCloudExclusiveCluster]`，嵌入现有 `baseHandler`。`Next` 每次 `DescribeExclusiveClusters` Limit=100（等于 `constant.CloudResourceSyncMaxLimit` 与云 API 上限）。`SyncConcurrent()` **不覆盖**，沿用 `baseHandler`：请求 `Concurrent` > 0 用请求值，否则读 hc-service `SyncConfig`，至少为 1。`Sync` 不消费预取实例，而是按本批 `cloud_ids` 通过 `cluster-id` 再查云侧数据，res-sync 内按 `constant.TCLBDescribeMax`(20) 切批（对齐 `listLBFromCloud`，因 `cluster-id` 过滤值上限云文档未明确）。

**原因**：集群对象没有第二套详情接口，`ClusterSet[]` 即全量。V2 先收集 `allCloudIDMap` 再删，避免 V1「先删再翻页」窗口。集群数量预计不多，按 ID 再查一次代价可接受，也能复用条件同步路径。并发是否加大等看实际数量，不必在本期写死 1。

**备选**：ResourceSync V1 → 否决：Next 只拿 ID 后 Sync 再拉一遍，对本接口是浪费。本期写死 `SyncConcurrent=1` → 否决：与 `baseHandler` 重复，后续调并发还要改代码。

### 2. adaptor 包一层 `*tclb.Cluster`，落库只映射表列 + 已有 extension

**选择**：`TCloudExclusiveCluster` 嵌入 `*tclb.Cluster`，`GetCloudID()` 返回 `ClusterId`。云 SDK `Tag` 不同步。写入映射：

| 云字段 | 落库 |
|---|---|
| ClusterId / ClusterName / ClusterType / ClusterTag / Zone / Network / Isp / Egress / IPVersion / MaxConn | 表顶层（`max_conn` 保持 `*int64`，STGW 的 null 存 SQL NULL） |
| ResourceCount、IdleResourceCount | 不落两列；`clb_resource_count = PtrToVal(ResourceCount) - PtrToVal(IdleResourceCount)`，结果小于 0 时记 Warn 并置 0 |
| MaxInFlow 等容量、ClustersVersion、DisasterRecoveryType、LoadBalanceDirectorCount、ClustersZone | `TCloudExclusiveClusterExtension`（已定义，nil 标量用 `PtrToVal` → 0） |
| Tag | 丢弃 |

`id` / `vendor` / `account_id` / `bk_biz_id` / `memo` / 审计字段由本地填写。购买时闲置数仍实时查云，不读本表。

**备选**：adaptor 自建精简结构体、不嵌 SDK → 与 `TCloudClb` 包 `*tclb.LoadBalancer` 不一致。全量持久化 SDK 含 Tag → 否决，表和 extension 都没有该列。

### 3. 全量拉数固定 Public + TGW/STGW；删除 DB 过滤必须同集合

**来源**：不是云 API 限制。`DescribeExclusiveClusters` 的 `network` / `cluster-type` 均可不传，也能拉 Private 与 VPCGW。该范围来自 iWiki「负载均衡-购买支持独占集群-后端方案设计」：

- 「二、关键名词」：`VPCGW` 标明「内网（Private）集群类型，**本次不涉及**」
- 「6.1 独占集群同步」主流程：`network=Public`、`cluster_type=['TGW','STGW']`、`Limit=100` 分页拉全量
- 购买段只用公网四层/七层（集群标签仅 TGW/STGW 有）

**选择**：全量 `Filters`：`network=Public`，`cluster-type` Values=`TGW,STGW`（同一 Filter 多值，云侧 OR）。`RemoveDeletedFromCloud` 的 DB 条件为 `account_id + region + network=Public + cluster_type IN (TGW,STGW)`，再与 `allCloudIDMap` 差集删除。`loadbalancer-id` 是「CLB 落在哪台集群」的反查，**全量同步不用**。

**备选**：不带 filter 拉账号下全部集群 → 超出 iWiki 本期范围，且删除若仍按 account+region 会把将来可能存在的 Private/VPCGW 一并清掉。按类型拆两次请求 → 否决，一次 Filter 多值即可。

### 4. 三路对比走已有 data-service 协议隔离，不新增 DS 接口

**选择**：res-sync `ExclusiveCluster(kt, params, opt)`：云侧按 `params.CloudIDs` 分批调用 `ListExclusiveClusters`；DB 用 `Global.ListExclusiveCluster`（Raw JSON）按同一批 cloud_ids 反序列化 extension 再 `common.Diff`。新增 → `TCloud.BatchCreateExclusiveCluster`（服务端写 `bk_biz_id=-1`）；变更 → `TCloud.BatchUpdateExclusiveCluster`（请求体无 `bk_biz_id`）；删除 → `Global.BatchDeleteExclusiveCluster`。不新增 ListExt。`CloudResType` 联合类型加入 `TCloudExclusiveCluster`。

**备选**：同步 Update 结构体带可选 `BkBizID` → 否决，DAO 变更已禁止。新建 vendor ListExt → 否决，List Raw 足够 Diff。

### 5. getSyncOrder 插在 Cert 与 LoadBalancer 之间

**选择**：`PermissionTemplate → … → Cert → LoadBalancerExclusiveCluster → LoadBalancer → …`。cloud-server 按 region 调 hc-service，与 `SyncLoadBalancer` 同形。hc-service client 挂在现有 `ClbClient`：`POST /load_balancer_exclusive_clusters/sync`。条件同步不另开接口，走同一路由的可选 `cloud_ids`（与 `security_group` 等 15 个资源一致），不复用 CLB 的 `by_condition` 异步适配流程。

**原因**：`SyncAllResource` fail-fast；集群在 CLB 前，CLB 成功时集群已入库。挂 ClbClient 因为独占集群是 CLB 域资源，不必新 client 类型。

## Risks / Trade-offs

- **[Risk] Public+TGW/STGW 漏掉账号下其它集群** → **Mitigation**：与 iWiki 6.1 产品范围一致；删除过滤与拉数同集合，避免误删。后续若要 Private/VPCGW，单独开变更扩 filter。
- **[Risk] `IdleResourceCount`/`ResourceCount` 为 nil 或差值异常** → **Mitigation**：`PtrToVal` 后相减，负值钳 0 并 Warn；购买页仍实时查闲置数。
- **[Risk] 独占集群同步失败导致整轮 `SyncAllResource` 在 CLB 前失败** → **Mitigation**：这是排序的刻意代价；监控同步详情状态即可，不改为 continue。
- **[Trade-off] adaptor 嵌 SDK Cluster 会带上不同步的 Tag 指针** → 仅内存，不写库。
- **[Trade-off] 不落 idle 数** → 表无法直接算剩余 VIP；购买链路必须调 `DescribeClusterResources`（本变更不做）。

## Migration Plan

无 DDL。发布 hc-service / cloud-server 后，下一轮定时同步自动灌数；新集群 `bk_biz_id=-1`，已分配记录只更新云属性。回滚：停同步即可，表数据保留，不影响现有 CLB。

## Open Questions

- 无。
