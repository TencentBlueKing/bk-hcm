## Why

独占集群表已能同步（`load-balancer-exclusive-cluster-sync`），但 CLB 实例同步仍丢掉云上的 `Exclusive` / `ClusterIds` / `ClusterTag`。列表和详情没有实例侧独占标记可读。本期只改同步：把三个字段写进既有 `extension` JSON，并纳入 DIFF。

## What Changes

- `TCloudClbExtension` 新增 `exclusive`、`cluster_ids`、`cluster_tag`
- `convertTCloudExtension` 赋值；云上 `Exclusive` 为 null（取不到有效值）时原样落库，不按 `0` 处理
- `isLBExtensionChange` 比较上述三字段
- 补写已比较但未赋值的 `egress`，避免空更新

**不包含**：列表/详情查询、购买申请单、申请单据详情、独占集群表同步、表结构变更。不同步写入 `clusters` 或集群本地主键。

## Capabilities

### New Capabilities

- `tcloud-clb-exclusive-extension-sync`：CLB 同步把独占相关云字段写入 `TCloudClbExtension`，并纳入 DIFF

### Modified Capabilities

（无）

## Impact

- `pkg/api/core/cloud/load-balancer/tcloud.go`：`TCloudClbExtension` 增三个落库字段
- `cmd/hc-service/logics/res-sync/tcloud/load_balancer.go`：`convertTCloudExtension` / `isLBExtensionChange`
- 无新表、无新依赖
