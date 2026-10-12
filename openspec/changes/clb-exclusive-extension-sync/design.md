## Context

腾讯云 `DescribeLoadBalancers` 已返回 `Exclusive` / `ClusterIds` / `ClusterTag` / `Egress`。现有 CLB 同步只把 `SlaType` 写入 `TCloudClbExtension`。独占集群表由另一条同步链路维护，本期不查该表。

## Goals / Non-Goals

**Goals:**

- 把云上独占字段落进既有 `extension` JSON，纳入 `common.Diff`
- 云上 `Exclusive` 为 null 时原样落库，不按 `0` 处理
- `egress` 比较与写入对齐

**Non-Goals:**

- 列表/详情接口
- 同步时换本地集群主键或派生 `clusters`
- 购买申请单、申请单据详情
- 表结构变更

## Decisions

### D1：只存三个云字段，不在同步里换本地主键

**选择**：`TCloudClbExtension` 只加 `exclusive`、`cluster_ids`、`cluster_tag`。

**理由**：集群名称/本地 ID 有自己的同步生命周期；七层经常没有云上集群 ID。

**备选**：同步时 `cloud_id → cluster_id` 写进 extension —— 放弃。

### D2：字段落在 extension JSON，整体覆写

**选择**：不新增表列。`Exclusive` 直接取云上值，云上为 null 时保持 nil，不写入该键。

**理由**：顶层 `int64` 零值会被 DAL 判 blank 跳过写入。云文档说明 null 表示取不到有效值，按 `0` 落库等于把未知当成共享型；`SlaType` 同样原样落库，口径一致。

### D3：同步补写 `egress`

**选择**：`convertTCloudExtension` 赋值 `Egress`，与 `isLBExtensionChange` 对齐。

**理由**：原先比了不写，每轮 DIFF 都会把未变 CLB 打成 update。

## Risks / Trade-offs

- **[风险] 云上 `Exclusive` 为 null 的实例，库里没有 `extension.exclusive` 键，规格筛选 `json_eq 0/1` 都查不到** → 可接受；null 表示云上未给出有效值，不猜测其类型。

## Migration Plan

1. 发布 hc-service，下一轮 CLB 同步写回三字段和 `egress`。
2. 无 DDL。

**回滚**：回滚二进制。已写入 `extension` 的键对旧代码无害。

## Open Questions

（无）
