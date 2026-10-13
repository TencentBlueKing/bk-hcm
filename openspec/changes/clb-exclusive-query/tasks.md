## 1. 带删除保护的列表

- [x] 1.1 `LoadBalancerWithDeleteProtect` 顶层增加 `Exclusive`、`SlaType`
- [x] 1.2 `listLoadBalancerWithDeleteProtect` 从 `TCloudClbExtension` 取出两字段；非 tcloud 保持 0 / 空串
- [x] 1.3 DAO `List` 的 `columnTypes` 放行 `extension.exclusive`（Numeric）与 `extension.sla_type`（String）

## 2. CLB 详情派生 clusters

- [x] 2.1 `TCloudClbExtension` 增加不落库的 `Clusters` 与 `TCloudClbExclusiveCluster`
- [x] 2.2 `getLoadBalancer` 的 TCloud 分支在 Get 之后调用填充函数
- [x] 2.3 实现 `fillTCloudClbExclusiveClusters`：非独占回空数组；按 `cluster_ids` + `account_id` 反查；未命中只回云上 ID；无 STGW 时用七层 `cluster_tag` 补一条
