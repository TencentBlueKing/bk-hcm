## ADDED Requirements

### Requirement: 带删除保护的列表回 exclusive 与 sla_type
系统 SHALL 在资源/业务视角的带删除保护 CLB 列表响应顶层返回 `exclusive` 与 `sla_type`，取值来自腾讯云 CLB 的 `extension`。非腾讯云条目 MUST 返回 `exclusive=0` 且 `sla_type` 为空字符串。系统 MUST NOT 在该列表返回 `clusters`，也 MUST NOT 合成「实例规格」中文。

#### Scenario: 腾讯云独占型
- **WHEN** 列表命中一条 `vendor=tcloud` 且 `extension.exclusive=1`、`extension.sla_type` 为空的 CLB
- **THEN** 响应该条 MUST 含顶层 `exclusive=1` 与 `sla_type=""`

#### Scenario: 腾讯云性能容量型
- **WHEN** 列表命中一条 `vendor=tcloud` 且 `extension.sla_type` 为非空档位的 CLB
- **THEN** 响应该条 MUST 含顶层 `sla_type` 为该档位原始值

#### Scenario: 腾讯云 Exclusive 取不到有效值
- **WHEN** 列表命中一条 `vendor=tcloud` 且 `extension` 中没有 `exclusive` 的 CLB
- **THEN** 响应该条 MUST NOT 含顶层 `exclusive`，MUST NOT 按 `0` 补齐

#### Scenario: 非腾讯云
- **WHEN** 列表命中非 `tcloud` 的 CLB
- **THEN** 响应该条 MUST 含 `exclusive=0` 与 `sla_type=""`

---

### Requirement: 列表支持按 extension.exclusive 与 extension.sla_type 筛选
系统 SHALL 允许带删除保护的列表（及共用同一 DAO `List` 的 CLB 列表）使用 `json_eq` 过滤 `extension.exclusive`（Numeric）与 `extension.sla_type`（String）。系统 MUST NOT 将这两字段列为排序字段。

#### Scenario: 筛独占型
- **WHEN** 过滤条件为 `extension.exclusive json_eq 1`
- **THEN** 系统 SHALL 只返回 `extension.exclusive` 等于 1 的 CLB，且 MUST NOT 因字段不在白名单而拒绝请求

#### Scenario: 筛共享型档位为空
- **WHEN** 过滤条件为 `extension.exclusive json_eq 0` 且 `extension.sla_type json_eq ""`
- **THEN** 系统 SHALL 只返回非独占且无性能容量档位的 CLB

#### Scenario: 未放行的 JSON 路径
- **WHEN** 过滤字段为未加入白名单的 `extension` 子路径
- **THEN** 系统 SHALL 拒绝该过滤条件

---

### Requirement: 详情派生 extension.clusters
系统 SHALL 在资源/业务视角的腾讯云 CLB 详情中返回 `extension.exclusive`、`extension.cluster_ids`、`extension.cluster_tag`，并派生 `extension.clusters`。`clusters` MUST NOT 落库。`extension.cluster_tag` MUST 仅表示七层（STGW）独占集群标签。

当 `exclusive != 1` 时，`clusters` MUST 为空数组。当 `exclusive=1` 时，系统 SHALL 用去重后的非空 `cluster_ids` 与 `account_id` 查询本地独占集群表：命中则填 `cluster_id`/`cluster_name`/`cluster_tag`/`cluster_type`；未命中则该元素只含 `cloud_cluster_id`。若 `cluster_tag` 非空且反查结果中没有 STGW，系统 SHALL 追加一条仅含 `cluster_tag` 与 `cluster_type=STGW` 的元素。

#### Scenario: 非独占型
- **WHEN** 详情 CLB 的 `extension.exclusive` 不为 1
- **THEN** `extension.clusters` MUST 为空数组，且系统 MUST NOT 查询独占集群表

#### Scenario: 指定四层集群且本地已同步
- **WHEN** `exclusive=1` 且 `cluster_ids` 含 `tgw-xxx`，本地独占集群表存在同账号该 `cloud_id`
- **THEN** `clusters` SHALL 含对应元素，`cloud_cluster_id` 为该云上 ID，`cluster_id`/`cluster_name`/`cluster_tag`/`cluster_type` 取自本地表

#### Scenario: 本地表未同步到该集群
- **WHEN** `exclusive=1` 且某 `cluster_ids` 在本地独占集群表不存在
- **THEN** 该元素 MUST 只返回 `cloud_cluster_id`，`cluster_id` 与 `cluster_name` 为空字符串

#### Scenario: 七层仅有标签
- **WHEN** `exclusive=1`，`cluster_tag` 非空，且反查结果中没有 `cluster_type=STGW` 的记录
- **THEN** `clusters` SHALL 追加一条 `cluster_tag` 为该标签、`cluster_type=STGW`、其余 ID/名称为空的元素
