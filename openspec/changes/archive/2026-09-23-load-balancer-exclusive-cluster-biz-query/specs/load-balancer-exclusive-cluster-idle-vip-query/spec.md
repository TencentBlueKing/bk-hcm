## ADDED Requirements

### Requirement: 业务视角独占集群闲置VIP查询
系统 SHALL 提供接口 `POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/exclusive_clusters/idle_vips/list`，输入 `account_id`（必填）、`region`（必填）、`cloud_cluster_id`（必填）。接口 SHALL 先查本地独占集群表校验：目标集群必须存在、类型必须为 TGW（四层）、且 `bk_biz_id` 必须等于路径参数；三项均通过后再实时查询腾讯云获取该集群当前闲置的 VIP 列表（不落库，每次实时查询，翻页取全后去重返回）。

#### Scenario: 正常查询已分配给当前业务的四层集群闲置VIP
- **WHEN** TGW 集群 `tgw-38feq8c6` 已分配给业务 213，且当前闲置 VIP 为 `1.1.1.1`、`1.1.1.2`
- **AND** 业务 213 调用闲置VIP查询，传入 `cloud_cluster_id=tgw-38feq8c6`
- **THEN** 返回 `count=2`，`details=["1.1.1.1","1.1.1.2"]`

#### Scenario: 目标集群不归属当前业务返回权限拒绝
- **WHEN** TGW 集群 `tgw-xxxx` 已分配给业务 214（非当前请求业务 213）
- **AND** 业务 213 调用闲置VIP查询并传入该 `cloud_cluster_id`
- **THEN** 返回 `PermissionDenied`，且不发起腾讯云实时查询调用，响应中不包含该集群归属或详情信息

#### Scenario: 集群本地不存在
- **WHEN** `cloud_cluster_id` 在本地独占集群表中查不到任何记录
- **THEN** 返回 `InvalidParameter`，不发起腾讯云实时查询调用

#### Scenario: 目标集群为七层类型不支持
- **WHEN** `cloud_cluster_id` 对应的本地记录 `cluster_type` 为 STGW（七层）
- **THEN** 返回 `InvalidParameter`，不发起腾讯云实时查询调用

#### Scenario: 集群当前无闲置VIP
- **WHEN** 目标集群已分配给当前业务，但当前没有闲置 VIP
- **THEN** 返回 `count=0`，`details=[]`

#### Scenario: 闲置VIP数量超过单页容量时自动翻页取全
- **WHEN** 目标集群当前闲置 VIP 数量超过单次云 API 调用的单页返回上限
- **THEN** 接口内部自动翻页查询直至取完全部数据，返回的 `count`/`details` 反映去重后的完整闲置 VIP 集合，调用方无需感知翻页过程

#### Scenario: 请求体不接受 bk_biz_id 覆盖
- **WHEN** 请求体中携带 `bk_biz_id` 字段
- **THEN** 接口忽略该字段，归属校验始终以路径参数 `bk_biz_id` 为唯一依据
