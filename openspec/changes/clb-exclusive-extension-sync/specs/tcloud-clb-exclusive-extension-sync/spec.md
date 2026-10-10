## ADDED Requirements

### Requirement: CLB extension 同步独占云字段
系统 SHALL 在腾讯云 CLB 同步时，将云上 `Exclusive`、`ClusterIds`、`ClusterTag`、`Egress` 写入 `TCloudClbExtension`，并随既有 `extension` JSON 整段落库。云上 `Exclusive` 为 null 时 MUST 原样落库（不写入该键），MUST NOT 按 `0` 处理。系统 MUST NOT 在同步路径查询独占集群表，也 MUST NOT 写入集群本地主键或 `clusters`。

#### Scenario: 云上返回独占型
- **WHEN** `DescribeLoadBalancers` 返回 `Exclusive=1`，并带 `ClusterIds` 与 `ClusterTag`
- **THEN** 系统 SHALL 将 `extension.exclusive=1`、`extension.cluster_ids`、`extension.cluster_tag`、`extension.egress` 写入该 CLB

#### Scenario: 云上 Exclusive 为 null
- **WHEN** 云上 `Exclusive` 为 null
- **THEN** 系统 SHALL NOT 写入 `extension.exclusive`，保持该字段为空

#### Scenario: 同步不派生 clusters
- **WHEN** 系统执行 CLB 创建或更新同步
- **THEN** 写入库的 `extension` MUST NOT 包含由本地表拼出的 `clusters`

---

### Requirement: 独占字段纳入 CLB extension DIFF
系统 SHALL 在 `isLBExtensionChange` 中比较 `exclusive`、`cluster_ids`、`cluster_tag` 与云上对应字段；`exclusive` 按指针比较，两侧均为 null 视为一致，null 与具体值视为变更。`egress` 的比较与写入 MUST 同时存在。

#### Scenario: 七层标签变更触发更新
- **WHEN** 库中 `extension.cluster_tag` 与云上 `ClusterTag` 不一致
- **THEN** 系统 SHALL 将该 CLB 判定为需要更新

#### Scenario: exclusive 与 egress 均无变化
- **WHEN** 库中 `exclusive` 及 `egress` 与云上一致，且其余 extension 字段也一致
- **THEN** 系统 SHALL NOT 仅因这两个字段对 CLB 发起空更新
