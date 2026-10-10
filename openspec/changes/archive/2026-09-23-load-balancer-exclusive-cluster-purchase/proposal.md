## Why

业务用户在「购买负载均衡」时需要选择独占型规格，把实例放到已分配给本业务的四层（TGW）/七层（STGW）独占集群上；当前业务视角申请单与系统提单接口只支持共享型/性能容量型（`sla_type`），没有独占型入参，也无法在提单阶段拦住「选了别人的集群 / 内网独占 / 指定了已占用 VIP」这类错误——审批通过后才会把云侧报错直接抛给用户，可提单成功率为 0%。API 文档（commit `73db73f3`，`--story=138282352`）已经把创建接口的独占字段契约、以及申请单详情/负载均衡详情的 `clusters` 回显契约写清楚，需要据此落地对应的后端实现。

技术方案：《负载均衡-购买支持独占集群-后端方案设计》「七、接口设计 M-01/M-03」「八、购买链路改造 8.1/8.3」；已完成澄清与评估的需求单：业务视角下支持购买（1069995598138281921，`v_status=approved`，size=3，effort=17.5h），完整需求文档见 `docs/reqs/业务视角购买.md`。

## What Changes

- `TCloudLoadBalancerSpec`（`pkg/api/hc-service/load-balancer/tcloud.go`，业务视角申请单/系统提单/资源视角创建/询价共用）新增选填字段 `Exclusive`、`ClusterTag`、`CloudClusterIDs`；`exclusive` 缺省 0，兼容存量共享型/性能容量型请求
- `TCloudLoadBalancerCreateReq.Validate` / `TCloudLoadBalancerSpec.ValidateSpec` 新增结构校验：独占型仅支持公网、`sla_type` 与独占型互斥、`cluster_tag`/`cloud_cluster_ids` 至少一个非空、指定 `vip` 时 `cloud_cluster_ids` 必须恰好一个且 `require_count=1`、未指定四层时不允许传 `vip`、单线运营商必须走共享带宽包
- `ApplicationOfCreateTCloudLB.CheckReq()`（业务视角申请单 + 系统提单共用同一 handler）新增归属校验：`cluster_tag` 命中当前业务已分配的 STGW，`cloud_cluster_ids` 每一个都属于当前业务已分配的 TGW，越权返回 `PermissionDenied`
- hc-service `BatchCreateTCloudClb` 下云前**完整重跑一遍**结构+归属+出口一致性校验（与提单时同一套辅助函数），并额外新增 VIP 闲置复核：指定 `vip` 时复核该 VIP 仍在对应四层集群闲置列表内；计费为 `BANDWIDTH_PACKAGE` 时校验带宽包出口落在「本次可能分配集群」的出口范围内（只四层 `Ep∈E4_set`、只七层 `Ep∈E7_set`、两层都选 `Ep∈E4_set∩E7_set`，出口数据读本地 `exclusive_cluster` 表）；`createOpt` 新增透传 `ClusterIds`/`ClusterTag` 给腾讯云 adaptor（`Exclusive` 为纯校验字段，不下传云侧）
- `pkg/adaptor/tcloud/clb.go` 新增 `DescribeClusterResources` adaptor 方法用于 VIP 闲置查询；hc-service 层额外暴露一个内部测试用 HTTP 端点直接透出该能力，便于联调时独立验证，不经 cloud-server/web-server 对外暴露
- **新增能力**：申请单详情（`GetApplication`/`GetBizApplication`）在 `type=create_load_balancer` 时解析 `content`，按 `cloud_cluster_ids`/`cluster_tag` 关联本地独占集群表拼装 `clusters` 展示数组后再回填 `content`
- **新增能力**：负载均衡详情（资源视角 + 业务视角 `GetLoadBalancer` 两处）`extension` 新增 `exclusive`、`clusters` 字段，数据来自云上 `DescribeLoadBalancers` 返回的 `ClusterTag`/`ClusterIds` 平级字段（公网独占集群关联，腾讯云 SDK 既有字段，无需新增云 API 调用；不是内网专用的 `ExclusiveCluster` 字段），`ClusterIds` 按 ID 反查本地独占集群表以确定真实层级（TGW/STGW，不假设数组元素层级），关联补齐 `cluster_id`/`cluster_name`/`cluster_tag`/`cluster_type`
- 不涉及：询价接口（本单不改）、资源视角创建拒绝独占字段（兄弟单 1069995598138281904）、购买页对外的标签/集群/闲置 VIP 查询接口本身（兄弟单 1069995598138281929；本单在 hc-service 层新增的是内部测试端点，不是该兄弟单负责的对外查询 API）、内网独占集群、非腾讯云 vendor

## Capabilities

### New Capabilities
- `load-balancer-exclusive-cluster-purchase`：业务视角申请单与系统提单接收独占型入参（`exclusive`/`cluster_tag`/`cloud_cluster_ids`）、结构校验、业务归属校验、下云前实时复核（VIP 闲置 + 带宽包出口一致性）、下云透传
- `load-balancer-exclusive-cluster-detail`：申请单详情与负载均衡详情的独占集群信息回显（`clusters` 富化，关联本地独占集群表）

### Modified Capabilities

（无：项目内暂无覆盖负载均衡购买申请单结构校验/详情展示行为的既有 spec，本变更均以新增能力承载）

## Impact

- **pkg/api/hc-service/load-balancer/tcloud.go**：`TCloudLoadBalancerSpec` 新增字段与 `ValidateSpec` 结构校验
- **cmd/cloud-server/service/application/handlers/load_balancer/tcloud/check.go**：`CheckReq` 新增归属校验调用
- **cmd/cloud-server/logics/load-balancer/**：新增独占型归属校验辅助函数（查本地 `exclusive_cluster` 表，按 `bk_biz_id`+`cluster_tag`/`cloud_id` 过滤）
- **cmd/hc-service/service/load-balancer/tcloud.go**：`BatchCreateTCloudClb` 新增下云前完整重跑结构+归属+出口校验、VIP 闲置复核与 `createOpt` 字段透传；新增一个透出 `DescribeClusterResources` 的内部测试端点
- **pkg/adaptor/tcloud/clb.go**：新增 `DescribeClusterResources` adaptor 方法（四层集群闲置 VIP 查询，云上实时查询）
- **cmd/hc-service/logics/**（或同级新增文件）：带宽包出口查询（复用既有 `TCloudListBwPkgOption`/adaptor 能力，按 `bandwidth_package_id` 查 `Egress`）
- **cmd/cloud-server/service/application/get.go / list.go**：`buildApplicationGetResp` 新增 `create_load_balancer` 类型的 `content` 拼装 `clusters` 逻辑
- **pkg/api/core/cloud/load-balancer/tcloud.go**：`TCloudClbExtension` 新增 `Exclusive`、`Clusters` 字段
- **cmd/hc-service/logics/res-sync/tcloud/load_balancer.go**：`convertTCloudExtension` 新增从 `cloud.ClusterTag`/`cloud.ClusterIds` 映射 `Clusters`（不读 `cloud.ExclusiveCluster`），`ClusterIds` 按 ID 反查本地独占集群表确定层级后补齐本地字段
- **docs/api-docs**：核对已存在草稿是否与最终实现一致（尤其 `cluster_tag` 是否必填的措辞，已按已批准需求文档修正为「非必填，与 `cloud_cluster_ids` 至少一个非空」）
- 不涉及 DAO、表结构改动（独占集群表已在既有变更中完成）；不影响 `exclusive=0` 存量购买请求行为
