## Why

独占集群 DAO / data-service 已就绪（`load-balancer-exclusive-cluster-dao`），但云上 `DescribeExclusiveClusters` 尚未接入，定时全量同步不会写入 `load_balancer_exclusive_cluster`。运营看不到真实集群清单，后续分配业务、购买页候选集群都没有数据源。

技术方案：负载均衡-购买支持独占集群-后端方案设计「六、同步逻辑」6.1 与 Wave 2a B-05 / B-06。本期打通全量同步链路，并补充按云上 ID 的条件同步入口。

## What Changes

- 适配层封装腾讯云 `DescribeExclusiveClusters`：分页 Limit=100；全量同步由调用方带 `network=Public` + `cluster-type=TGW,STGW`（iWiki 6.1 产品范围，不是云 API 限制）
- res-sync 用 `common.Diff` 按 `cloud_id` 三路对比：新增 `bk_biz_id=-1`、更新只写云属性（不碰 `bk_biz_id`）、删除云上已不存在的记录；`clb_resource_count = ResourceCount - IdleResourceCount`
- hc-service 全量同步走 `handler.ResourceSyncV2`（先删后写），条件同步复用同一入口的可选 `cloud_ids` 且清理范围限定在指定 ID 内；cloud-server `getSyncOrder()` 把独占集群排在 `LoadBalancerCloudResType` 之前
- **不包含**：CLB `TCloudClbExtension` 独占字段、cloud-server 查询/分配业务、购买链路、`DescribeClusterResources` / 闲置 VIP、DAO 表结构变更

## Capabilities

### New Capabilities

- `tcloud-exclusive-cluster-sync`：适配层拉独占集群、res-sync 三路对比、hc-service / cloud-server 全量同步编排（排在 CLB 之前），条件同步复用全量入口

### Modified Capabilities

（无）

## Impact

- **pkg/adaptor**：`tcloud.TCloud` 新增 `ListExclusiveClusters`；types 新增 `TCloudExclusiveCluster`（包 `*tclb.Cluster` + `GetCloudID()`）；需 `go generate` 刷新 mock
- **cmd/hc-service/logics/res-sync/tcloud**：新增独占集群同步；`Interface` 增加方法；`common.CloudResType` 并入 `TCloudExclusiveCluster`
- **cmd/hc-service/service/sync/tcloud**：`ResourceSyncV2` handler + `POST /load_balancer_exclusive_clusters/sync`（可选 `cloud_ids` 即条件同步）
- **pkg/client/hc-service/tcloud**：Clb client 增加 `SyncExclusiveCluster`
- **cmd/cloud-server/service/sync/tcloud**：`syncFuncMap` / `getSyncOrder`（全量）；`condSyncFuncMap` 注册 `CondSyncExclusiveCluster`（条件同步，走已有 `sync_by_cond` 通用入口）
- 复用已有 data-service 客户端：`TCloud.BatchCreateExclusiveCluster` / `BatchUpdateExclusiveCluster`、`Global.ListExclusiveCluster` / `BatchDeleteExclusiveCluster`
- 无表结构变更、无新依赖、不改 CLB 同步
