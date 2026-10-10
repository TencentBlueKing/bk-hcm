# load-balancer-exclusive-cluster-detail Specification

## Purpose
TBD - created by archiving change load-balancer-exclusive-cluster-purchase. Update Purpose after archive.
## Requirements
### Requirement: 申请单详情回显独占集群信息
申请单详情接口（资源视角 `GET /api/v1/cloud/applications/{application_id}`、业务视角 `GET /api/v1/cloud/bizs/{bk_biz_id}/applications/{application_id}`）SHALL 在 `type=create_load_balancer` 时，于返回的 `content` 中追加 `clusters` 数组：解析 `content` 中的 `exclusive`/`cluster_tag`/`cloud_cluster_ids`，非独占型（`exclusive` 非 1）时 `clusters` 为空数组；独占型时按 `cluster_tag` 查当前业务已分配的 STGW、按 `cloud_cluster_ids` 查当前业务已分配的 TGW，关联本地独占集群表得到 `cloud_cluster_id`、`cluster_id`（本地 ID）、`cluster_name`、`cluster_tag`、`cluster_type` 后拼入 `clusters`；本地表未同步或已删除的集群，`cluster_id`/`cluster_name` 为空字符串，`cloud_cluster_id` 仍返回原值。该富化为读时计算，不修改申请单落库的原始 `content`。

#### Scenario: 独占型申请单详情包含集群信息
- **WHEN** 查询一个 `type=create_load_balancer` 且 `exclusive=1` 的申请单详情
- **THEN** 返回的 `content` 中包含 `clusters` 数组，元素包含 `cloud_cluster_id`/`cluster_id`/`cluster_name`/`cluster_tag`/`cluster_type`

#### Scenario: 非独占型申请单详情的 clusters 为空
- **WHEN** 查询一个 `type=create_load_balancer` 且 `exclusive` 为 0 或不传的申请单详情
- **THEN** 返回的 `content` 中 `clusters` 为空数组，`cluster_tag`、`cloud_cluster_ids` 为空

#### Scenario: 七层集群在提单时通常没有具体集群 ID
- **WHEN** 独占型申请单只指定了 `cluster_tag`（七层），未指定四层 `cloud_cluster_ids`
- **THEN** `clusters` 中七层元素只有 `cluster_tag`、`cluster_type` 有值，其余字段为空字符串

#### Scenario: 四层随机分配时 clusters 不含具体集群
- **WHEN** 独占型申请单四层选择「随机分配」（`cloud_cluster_ids` 为空但 `exclusive=1` 且仅七层有值，或四层未指定具体落地集群）
- **THEN** 对应 TGW 元素的 `cloud_cluster_id`/`cluster_id`/`cluster_name` 为空字符串，展示需要从负载均衡详情接口读取实际落地值

#### Scenario: 业务视角只能查看归属本业务的申请单
- **WHEN** 业务视角调用方查询一个不属于其 `bk_biz_id` 的申请单
- **THEN** 返回 `RecordNotFound`，不触发 `clusters` 富化逻辑

### Requirement: 负载均衡详情回显独占集群关联
负载均衡详情接口（资源视角 `POST /api/v1/cloud/load_balancers/{id}`、业务视角 `POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/{id}`）的 `extension` SHALL 新增 `exclusive`（是否独占型实例）与 `clusters`（独占集群信息列表）字段。数据来源为已同步的腾讯云 `DescribeLoadBalancers` 返回的 `ClusterTag`（七层独占标签）与 `ClusterIds`（集群 ID 数组）这两个平级字段——**不是** `ExclusiveCluster` 字段（该字段是内网独占集群，与本单公网场景无关）；`ClusterIds` 中的元素**不假设**层级，同步时按 ID 反查本地独占集群表，以本地记录的 `cluster_type` 为准判定该 ID 是 TGW 还是 STGW，并补齐 `cluster_id`（本地 ID）、`cluster_name`、`cluster_tag`；非独占型实例 `clusters` 为空数组。详情页「实例规格」展示由 `exclusive` 与 `sla_type` 合成：`exclusive` 为 1 展示独占型；否则 `sla_type` 非空展示对应档位；否则展示共享型。

#### Scenario: 独占型负载均衡详情包含集群列表
- **WHEN** 查询一个独占型负载均衡的详情
- **THEN** `extension.exclusive` 为 1，`extension.clusters` 包含已关联的四层/七层集群信息

#### Scenario: 非独占型负载均衡的 clusters 为空
- **WHEN** 查询一个共享型或性能容量型负载均衡的详情
- **THEN** `extension.exclusive` 为 0，`extension.clusters` 为空数组

#### Scenario: 仅使用四层或七层独占集群时数组只含对应类型元素
- **WHEN** 负载均衡只使用了四层独占集群（未使用七层）
- **THEN** `extension.clusters` 数组中只出现 `cluster_type=TGW` 的元素

#### Scenario: 云上 ClusterIds 混合了四层与七层落地 ID
- **WHEN** 云上返回的 `ClusterIds` 数组同时包含一个 TGW ID 与一个 STGW ID
- **THEN** 同步逻辑按每个 ID 反查本地独占集群表得到各自真实的 `cluster_type`，不假设数组内元素全部同层级，`extension.clusters` 中两个元素的 `cluster_type` 分别正确标注为 `TGW`/`STGW`

#### Scenario: 七层集群云侧未返回具体集群 ID
- **WHEN** 云上未返回七层 STGW 具体集群 ID（云侧调度决定）
- **THEN** 该元素仅 `cluster_tag`、`cluster_type` 有值，其余字段为空字符串

#### Scenario: 本地表未同步或已删除时仍返回云上 ID
- **WHEN** 集群在本地独占集群表中已被删除或尚未同步
- **THEN** `cluster_id`、`cluster_name` 为空字符串，`cloud_cluster_id` 仍返回云上原值

#### Scenario: 提单入参与云上实际落地可能不完全一致
- **WHEN** 提单时选择了随机分配或标签范围较大的七层集群
- **THEN** 负载均衡详情的 `exclusive`/`clusters` 以云上同步回来的实际落地结果为准，可能与提单入参不完全一致

