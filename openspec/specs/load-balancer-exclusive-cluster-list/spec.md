# load-balancer-exclusive-cluster-list Specification

## Purpose
TBD - created by archiving change load-balancer-exclusive-cluster-list. Update Purpose after archive.
## Requirements
### Requirement: 资源视角独占集群列表查询接口
系统 SHALL 提供资源视角接口 `POST /api/v1/cloud/load_balancers/exclusive_clusters/list`，供已认证且具备 `meta.LoadBalancer`+`meta.Find` 权限的调用方按标准 `core.ListReq`（`filter`+`page`）查询独占集群列表，返回结果的 `extension` 字段必须是按 `vendor` 反序列化后的结构化对象（当前仅支持 `tcloud`）。

#### Scenario: 具备权限时按过滤条件返回列表
- **WHEN** 调用方具备 `meta.LoadBalancer`+`meta.Find` 权限，提交 `filter={cluster_type=TGW, region=ap-guangzhou}`、`page={count:false, limit:500}`
- **THEN** 接口返回 `code=0`，`data.details` 中每条记录的顶层字段（`id`/`cloud_id`/`name`/`vendor`/`account_id`/`bk_biz_id`/`region`/`zone`/`cluster_type`/`cluster_tag`/`network`/`isp`/`egress`/`ip_version`/`max_conn`/`clb_resource_count`/`memo`/`creator`/`reviser`/`created_at`/`updated_at`）与 data-service 存储值一致，且仅包含匹配过滤条件的记录

#### Scenario: 未分配集群可按 bk_biz_id=-1 过滤
- **WHEN** 调用方提交 `filter={bk_biz_id eq -1}`
- **THEN** 接口只返回 `bk_biz_id=-1`（未分配）的集群记录

#### Scenario: count 模式只返回总数
- **WHEN** 调用方提交 `page={count:true}`
- **THEN** 接口返回 `data.count` 为匹配过滤条件的记录总数，`data.details` 为空数组

#### Scenario: 无权限时返回空列表而非报错
- **WHEN** 调用方不具备任何账号下 `meta.LoadBalancer`+`meta.Find` 权限
- **THEN** 接口返回 `code=0`，`data={count:0, details:[]}`，不返回 `PermissionDenied` 错误（与 `cert.ListCert` 等资源视角列表接口的无权限行为一致）

### Requirement: extension 按 tcloud 强类型反序列化
系统 SHALL 将 data-service 返回的每条集群记录的 `extension`（原始 JSON）反序列化为 `TCloudExclusiveClusterExtension` 强类型结构（12 个标量字段 + `clusters_zone` 嵌套对象）后再返回给调用方；任一条记录反序列化失败，接口 MUST 返回错误而非静默跳过该条记录或返回部分结果。

#### Scenario: extension 正常解析
- **WHEN** 某条集群记录的 `extension` 是合法的 tcloud 扩展字段 JSON（如 `{"max_in_flow":10240,...,"clusters_zone":{"master_zone":["ap-guangzhou-3"],"slave_zone":[]}}`）
- **THEN** 响应中该条记录的 `extension` 字段是结构化对象，字段值与原始 JSON 完全对应

#### Scenario: extension 解析失败时整体报错
- **WHEN** 某条集群记录的 `extension` 因数据损坏无法解析为 `TCloudExclusiveClusterExtension`
- **THEN** 接口返回非 0 错误码，错误信息中包含该记录的 `id`，不返回任何记录（不做部分成功）

