## 1. 适配层 DescribeExclusiveClusters

- [x] 1.1 在 `pkg/adaptor/types/load-balancer/` 新增 `TCloudExclusiveCluster`（嵌入 `*tclb.Cluster`，实现 `GetCloudID()`）与 `TCloudExclusiveClusterListOption`（`Region` 必填、`Page`、`CloudIDs`、`Network`、`ClusterTypes`）及 `Validate`
- [x] 1.2 实现 `pkg/adaptor/tcloud` 的 `ListExclusiveClusters`：校验 opt → 组装 `DescribeExclusiveClusters`（Limit 默认 20 最大 100）；Filters 按调用方传入的 cluster-id / network / cluster-type
- [x] 1.3 `pkg/adaptor/tcloud/interface.go` 的 `TCloud` 增加 `ListExclusiveClusters` 声明，并 `go generate` 刷新 `pkg/adaptor/mock/tcloud`

## 2. res-sync 独占集群

- [x] 2.1 `cmd/hc-service/logics/res-sync/common/diff.go` 的 `CloudResType` 联合类型加入 `typeslb.TCloudExclusiveCluster`
- [x] 2.2 新增 `cmd/hc-service/logics/res-sync/tcloud/load_balancer_exclusive_cluster.go`：`SyncExclusiveClusterOption`、`isExclusiveClusterChange`、`clb_resource_count` 派生（负值钳 0）、云 `Tag` 不落库
- [x] 2.3 实现 `listExclusiveClusterFromDB`：`Global.ListExclusiveCluster` 按 `account_id+region+vendor` 与可选 `cloud_ids` 查询，Raw extension 反序列化为 `TCloudExclusiveClusterExtension`
- [x] 2.4 实现 create / update / delete：分别调 `TCloud.BatchCreateExclusiveCluster`、`TCloud.BatchUpdateExclusiveCluster`（model 不赋 `BkBizID`）、`Global.BatchDeleteExclusiveCluster`；create 请求不含 `bk_biz_id`
- [x] 2.5 实现 `ExclusiveCluster` 主方法（按 `cloud_ids` 调 adaptor）+ `RemoveExclusiveClusterDeleteFromCloud`（DB 过滤 `account_id+region+vendor=TCloud+network=Public+cluster_type IN (TGW,STGW)`）
- [x] 2.6 `cmd/hc-service/logics/res-sync/tcloud/client.go` 的 `Interface` 声明上述两个方法

## 3. hc-service 全量同步入口

- [x] 3.1 新增 `cmd/hc-service/service/sync/tcloud/load_balancer_exclusive_cluster.go`：嵌入 `baseHandler` 的 `HandlerV2`，不覆盖 `SyncConcurrent()`；`Next` 串行 Limit=100 且带 Public+TGW/STGW；`Sync` 按本批 cloud_ids 同步
- [x] 3.2 `cmd/hc-service/service/sync/tcloud/service.go` 注册 `POST /load_balancer_exclusive_clusters/sync`
- [x] 3.3 `pkg/client/hc-service/tcloud/clb.go` 增加 `SyncExclusiveCluster(kt, *sync.TCloudSyncReq)`，打到上述路径
- [x] 3.4 条件同步复用 `/sync`：`Next` 在 `TCloudSyncReq.CloudIDs` 非空时只查一次，`RemoveDeletedFromCloud` 传入 `CloudIDs` 限定清理范围；不新增 `by_condition/sync` 接口

## 4. cloud-server 全量编排

- [x] 4.1 新增 `cmd/cloud-server/service/sync/tcloud/load_balancer_exclusive_cluster.go`：按 region 调 hc-service，写 `SyncDetail`（对齐 `SyncLoadBalancer`）
- [x] 4.2 `sync_all_resource.go`：`syncFuncMap` 注册 `LoadBalancerExclusiveClusterCloudResType`；`getSyncOrder()` 插入 `CertCloudResType` 与 `LoadBalancerCloudResType` 之间
- [x] 4.3 `cond_sync_resource.go`：新增 `CondSyncExclusiveCluster`（逐 region 调 `SyncExclusiveCluster`，带 `cloud_ids`）并注册进 `condSyncFuncMap`，接通已有 `POST /vendors/{vendor}/accounts/{account_id}/resources/{res}/sync_by_cond`（`res=load_balancer_exclusive_cluster`）；不注册 `condAsyncFuncMap`

## 5. 编译核对

- [x] 5.1 相关包 `go build` / `go test` 能过（adaptor mock、res-sync `CloudResType`、hc-service、cloud-server）
