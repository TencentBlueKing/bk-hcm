## Why

CLB 同步已把 `exclusive` / `cluster_ids` / `cluster_tag` 写入 `extension`，但列表无法按独占型或规格档位筛选，详情也无法拼出业务可读的集群信息。本期只改查询：带删除保护的列表暴露并筛选这两个字段，详情派生 `clusters`。

## What Changes

- 带删除保护的 CLB 列表（资源/业务）：响应顶层新增 `exclusive`、`sla_type`；DAO 白名单放开 `extension.exclusive` / `extension.sla_type` 的 `json_eq`
- CLB 详情（资源/业务）：GET 时用 `extension.cluster_ids` 反查本地独占集群表，派生 `extension.clusters`（不落库）
- `cluster_tag` 仅表示七层（STGW）标签；后端只回原始值，「实例规格」中文由前端合成

**不包含**：CLB 同步、购买申请单、申请单据详情、独占集群表同步、普通不带删除保护的 list、按这两字段排序。

## Capabilities

### New Capabilities

- `tcloud-clb-exclusive-query`：列表顶层回 `exclusive`/`sla_type` 并支持 JSON 筛选；详情派生 `clusters`

### Modified Capabilities

（无）

## Impact

- `pkg/api/core/cloud/load-balancer/tcloud.go`：新增 `Clusters` 与 `TCloudClbExclusiveCluster`（不落库）
- `pkg/api/core/cloud/load-balancer/load_balancer.go`：`LoadBalancerWithDeleteProtect` 顶层字段
- `pkg/dal/dao/cloud/load-balancer/load_balancer.go`：List 筛选白名单
- `cmd/cloud-server/service/load-balancer/query.go`：列表取值、详情填充
