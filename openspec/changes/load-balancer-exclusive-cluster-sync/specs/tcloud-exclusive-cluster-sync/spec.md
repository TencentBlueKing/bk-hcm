## ADDED Requirements

### Requirement: 适配层 ListExclusiveClusters

系统 SHALL 在 `pkg/adaptor/tcloud` 封装腾讯云 `DescribeExclusiveClusters`，方法签名 `ListExclusiveClusters(kt *kit.Kit, opt *typelb.TCloudExclusiveClusterListOption) ([]typelb.TCloudExclusiveCluster, error)`，并在 `tcloud.TCloud` 接口声明。`TCloudExclusiveCluster` SHALL 嵌入 `*tclb.Cluster` 并实现 `GetCloudID()`（返回 `ClusterId`）。`TCloudExclusiveClusterListOption` SHALL 包含 `Region`（必填）、`Page`（`Limit` 默认 20、最大 100）、可选 `CloudIDs`、可选 `Network`、可选 `ClusterTypes`。全量同步调用方传入 `Network=Public` 且 `ClusterTypes=[TGW,STGW]`；条件同步调用方传入 `CloudIDs` 并使用 `cluster-id` 过滤。SHALL NOT 使用 `loadbalancer-id` 做全量同步过滤。`opt` 为 nil 或 Region 为空时 SHALL 返回参数错误。接口变更后 SHALL 重新生成 adaptor mock。

#### Scenario: 分页拉取公网四层七层集群
- **WHEN** 调用 `ListExclusiveClusters`，`Region` 有效，`Network=Public`，`ClusterTypes=["TGW","STGW"]`，`Page.Limit=100`
- **THEN** 系统调用 `DescribeExclusiveClusters`，Filters 含 `network=Public` 与 `cluster-type` Values `TGW,STGW`，Limit=100，返回 `[]TCloudExclusiveCluster`

#### Scenario: Page.Limit 超过 100
- **WHEN** `Page.Limit` 大于 100
- **THEN** 系统返回参数校验错误

#### Scenario: 按集群 ID 拉取
- **WHEN** 调用 `ListExclusiveClusters`，`Region` 有效，`CloudIDs=["tgw-a"]`
- **THEN** 系统调用 `DescribeExclusiveClusters`，Filters 含 `cluster-id=tgw-a`

#### Scenario: opt 非法
- **WHEN** `opt` 为 nil，或 `Region` 为空
- **THEN** 系统返回 InvalidParameter，不调用云 API

### Requirement: res-sync 独占集群三路对比

系统 SHALL 在 `cmd/hc-service/logics/res-sync/tcloud` 实现 `ExclusiveCluster(kt, params *SyncBaseParams, opt *SyncExclusiveClusterOption) (*SyncResult, error)`，使用 `common.Diff` 以 `cloud_id` 为键。云侧 SHALL 按 `params.CloudIDs` 分批调 adaptor；DB 列表 SHALL 使用已有 `Global.ListExclusiveCluster` 按同一批 cloud_ids 查询，将 Raw extension 反序列化为 `TCloudExclusiveClusterExtension` 再 Diff。`clb_resource_count` SHALL 在写入前计算为 `PtrToVal(ResourceCount) - PtrToVal(IdleResourceCount)`，若结果小于 0 则置 0。云 SDK `Tag` SHALL NOT 写入 DB。新增调用 `TCloud.BatchCreateExclusiveCluster`（请求体不含 `bk_biz_id`，由 data-service 写 `-1`）。更新调用 `TCloud.BatchUpdateExclusiveCluster`（请求体 SHALL NOT 含 `bk_biz_id`）。删除调用 `Global.BatchDeleteExclusiveCluster`。`common.CloudResType` SHALL 纳入 `TCloudExclusiveCluster`。`Interface` SHALL 增加 `ExclusiveCluster` 与 `RemoveExclusiveClusterDeleteFromCloud`。

#### Scenario: 云有 DB 无则创建为未分配
- **WHEN** 云上存在 `cloud_id=tgw-a`，DB 无对应记录
- **THEN** 系统调用 `BatchCreateExclusiveCluster`，创建记录 `bk_biz_id=-1`，`clb_resource_count` 为派生差值

#### Scenario: 两侧都有且云属性变化则更新且不改 bk_biz_id
- **WHEN** DB 记录 `bk_biz_id=213`，云上 `egress` 或 `clb_resource_count` 与 DB 不同
- **THEN** 系统调用 `BatchUpdateExclusiveCluster` 更新云属性，该记录 `bk_biz_id` 仍为 213

#### Scenario: 云无 DB 有则删除
- **WHEN** DB 存在 `cloud_id=tgw-b` 且该 ID 不在本次云结果中，且该记录属于本次删除过滤集合
- **THEN** 系统调用 `BatchDeleteExclusiveCluster` 删除该记录

#### Scenario: 无变化不写库
- **WHEN** 云与 DB 映射后的可比较字段全部相同
- **THEN** 系统不调用 BatchCreate / BatchUpdate / BatchDelete

### Requirement: hc-service ResourceSyncV2 全量同步入口

系统 SHALL 提供 `POST /load_balancer_exclusive_clusters/sync`，请求体复用 `TCloudSyncReq`（本期使用 `account_id`、`region`）。handler SHALL 使用 `handler.ResourceSyncV2`，嵌入现有 `baseHandler`，SHALL NOT 覆盖 `SyncConcurrent()`。`Next` SHALL 串行分页，每页 Limit=100，Filters 为 Public+TGW/STGW。`Sync` SHALL 按本批云上实例 ID 调用 res-sync。`RemoveDeletedFromCloud` 的 DB 过滤 MUST 为 `account_id + region + vendor=TCloud + network=Public + cluster_type IN (TGW, STGW)`。`pkg/client/hc-service/tcloud` 的 `ClbClient` SHALL 封装该方法。

#### Scenario: 全量同步按 100 翻页
- **WHEN** 某地域云上有 150 个匹配集群
- **THEN** `Next` 至少两次调用云 API（100+50），随后删除差集并按批同步实例

#### Scenario: 全量删除过滤不含 Private / VPCGW
- **WHEN** 全量同步某地域，DB 若存在 `network=Private` 或 `cluster_type=VPCGW` 的记录
- **THEN** 这些记录不进入本次删除候选

### Requirement: cloud-server 全量同步注册

系统 SHALL 在 `cmd/cloud-server/service/sync/tcloud/sync_all_resource.go` 的 `syncFuncMap` 注册 `LoadBalancerExclusiveClusterCloudResType`。`getSyncOrder()` SHALL 把该类型放在 `CertCloudResType` 之后、`LoadBalancerCloudResType` 之前。全量同步按账号下地域逐个调用 hc-service，并更新 `SyncDetail` 状态。

#### Scenario: 全量顺序先集群后 CLB
- **WHEN** 执行 `SyncAllResource`
- **THEN** 独占集群同步函数在 `SyncLoadBalancer` 之前被调用

#### Scenario: 独占集群失败则本轮不跑 CLB
- **WHEN** 独占集群同步返回错误
- **THEN** `SyncAllResource` 立即返回该资源类型与错误，不调用 `SyncLoadBalancer`

### Requirement: hc-service 条件同步复用全量入口

系统 SHALL NOT 新增独立的 `by_condition/sync` 接口，条件同步复用 `POST /load_balancer_exclusive_clusters/sync`，由 `TCloudSyncReq` 可选的 `cloud_ids`（`max=20`）驱动，与 `security_group` 等资源保持一致。`Next` SHALL 在 `cloud_ids` 非空时只按这批 ID 查询一次，不再分页；`RemoveDeletedFromCloud` SHALL 把 `cloud_ids` 传入 `SyncRemovedParams`，使清理范围限定在指定 ID 内。

cloud-server 侧 SHALL 在 `cond_sync_resource.go` 的 `condSyncFuncMap` 注册 `LoadBalancerExclusiveClusterCloudResType → CondSyncExclusiveCluster`，逐 region 调用 hc-service 并透传 `cloud_ids`，从而接通已有的资源视角入口 `POST /vendors/{vendor}/accounts/{account_id}/resources/{res}/sync_by_cond`（`res=load_balancer_exclusive_cluster`）。SHALL NOT 注册 `condAsyncFuncMap`（不走 CLB 的异步任务流程）。独占集群 SHALL NOT 加入 `allowedSyncAllResTypes`，即条件同步必须指定 `regions`。

#### Scenario: 资源视角条件同步
- **WHEN** 调用 `sync_by_cond`，`res=load_balancer_exclusive_cluster`，body 含 `regions=["ap-guangzhou"]` 与 `cloud_ids=["tgw-a"]`
- **THEN** cloud-server 取账号同步锁后逐 region 调用 hc-service `/load_balancer_exclusive_clusters/sync`，同步返回结果而非异步任务 ID

#### Scenario: 条件同步未传 regions
- **WHEN** 调用 `sync_by_cond`，`res=load_balancer_exclusive_cluster`，body 未含 `regions`
- **THEN** 系统返回参数校验错误

#### Scenario: 条件同步指定集群
- **WHEN** 请求 `cloud_ids=["tgw-a","tgw-b"]` 且云上返回 `tgw-a`
- **THEN** `Next` 只调用云 API 一次，系统同步 `tgw-a`，DB 清理仅在 `tgw-a`、`tgw-b` 这两个 ID 范围内按 Diff 结果处理

#### Scenario: 条件同步未命中云资源
- **WHEN** 请求 `cloud_ids=["tgw-missing"]` 且云上未返回记录
- **THEN** 系统不触及范围外记录，仅在 `tgw-missing` 这个指定 ID 范围内按 Diff 结果处理 DB 记录
