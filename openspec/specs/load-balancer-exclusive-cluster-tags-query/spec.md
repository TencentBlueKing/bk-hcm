# load-balancer-exclusive-cluster-tags-query Specification

## Purpose
TBD - created by archiving change load-balancer-exclusive-cluster-biz-query. Update Purpose after archive.
## Requirements
### Requirement: 业务视角独占集群标签聚合查询
系统 SHALL 提供接口 `POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/exclusive_clusters/tags/list`，按 `account_id`（必填）、`region`（必填）、`isp`（必填，枚举 BGP/CMCC/CUCC/CTCC）、`zone`（选填）、`cluster_type`（选填，枚举 TGW/STGW）过滤，仅查询 `bk_biz_id` 等于路径参数、`network=Public` 的独占集群，按 `(cluster_tag, cluster_type)` 分组返回 `details` 数组；服务端强制使用路径 `bk_biz_id` 作为归属过滤条件，不接受请求体中出现 `bk_biz_id` 字段。

#### Scenario: 业务已分配标签下的集群正常返回
- **WHEN** 业务 213 已分配标签 `ziyan_chiji` 下的 TGW 集群 `tgw-38feq8c6`（BGP，位于 `ap-guangzhou-3`）与同标签的 STGW 集群
- **AND** 业务 213 调用标签聚合查询，传入 `account_id=00000001, region=ap-guangzhou, isp=BGP`，`cluster_type` 为空
- **THEN** `details` 包含两条记录：`cluster_type=TGW` 的记录 `clusters` 含 `cloud_cluster_id=tgw-38feq8c6`；`cluster_type=STGW` 的记录 `clusters` 也包含该标签下的真实 STGW 集群信息（不固定为空数组）

#### Scenario: 未分配任何标签的业务返回空结果
- **WHEN** 业务未被分配任何独占集群标签
- **THEN** 接口返回 `details` 为空数组，购买页应据此隐藏「独占型」规格选项

#### Scenario: 集群标签为空的记录不参与聚合
- **WHEN** 已同步的独占集群中存在 `cluster_tag` 为空字符串的记录
- **THEN** 该记录不出现在任何 `details` 分组中

#### Scenario: 跨业务集群不可见
- **WHEN** 标签 `ziyan_qq` 已分配给业务 214（非当前请求业务 213）
- **THEN** 业务 213 调用标签聚合查询时，`details` 中不包含 `ziyan_qq` 标签及其集群信息

#### Scenario: 按可用区过滤
- **WHEN** 标签 `ziyan_chiji` 下集群 A 位于可用区 `ap-guangzhou-3`（`extension.clusters_zone.master_zone` 包含该值）、集群 B 位于 `ap-guangzhou-4`
- **AND** 请求传入 `zone=ap-guangzhou-3`
- **THEN** 该标签对应分组的 `clusters` 只包含集群 A；若某标签下所有集群的 `master_zone` 都不包含请求的 `zone`，该标签对应的分组不出现在 `details` 中

#### Scenario: 必填参数缺失被拒绝
- **WHEN** 请求未传 `isp`、`account_id` 或 `region`
- **THEN** 接口返回 `InvalidParameter`

#### Scenario: 按集群类型过滤
- **WHEN** 请求传入 `cluster_type=TGW`
- **THEN** `details` 只包含 `cluster_type=TGW` 的分组，不返回 STGW 分组

#### Scenario: 请求体不接受 bk_biz_id 覆盖
- **WHEN** 请求体中携带 `bk_biz_id` 字段（与路径参数不同或相同）
- **THEN** 接口忽略请求体中的该字段，始终以路径参数 `bk_biz_id` 作为唯一归属过滤条件

