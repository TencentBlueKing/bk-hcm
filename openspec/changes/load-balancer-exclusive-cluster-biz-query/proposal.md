## Why

业务用户在「购买负载均衡」选择独占型规格时，需要先看到本业务已分配的四层（TGW）/七层（STGW）独占集群标签、标签下的具体集群，并在选定具体四层集群后查看该集群当前闲置的 VIP，才能完成规格选择与提单。当前 HCM 完全没有面向业务视角的独占集群查询能力——已有的独占集群列表接口（`load-balancer-exclusive-cluster-list` 变更产出）是资源视角（管理员纳管用途），不做"仅返回已分配给当前业务"的过滤，也不做标签聚合，无法直接服务购买页；上线前购买页可购买覆盖率为 0%（即便集群已分配给业务，用户也看不到任何独占型规格信息）。

已批准需求文档（`docs/reqs/独占集群购买查询.md`，TAPD [1069995598138281929](https://tapd.woa.com/tapd_fe/69995598/story/detail/1069995598138281929)）与已定稿接口文档（`docs/api-docs/web-server/docs/biz/load-balancer/list_exclusive_cluster_tags.md`、`list_exclusive_cluster_idle_vips.md`）已经明确了两个接口的契约，需要据此落地后端实现。

## What Changes

- 新增业务视角标签聚合查询接口（N-04）：`POST /bizs/{bk_biz_id}/load_balancers/exclusive_clusters/tags/list`，按 `(cluster_tag, cluster_type)` 分组返回当前业务已分配的公网独占集群标签及标签下的集群清单（TGW、STGW 两类分组均返回真实集群列表，与已定稿接口文档的响应示例一致）；聚合逻辑在 cloud-server 层完成，复用已有的 data-service 通用列表接口，不新增 data-service 接口
- 新增业务视角闲置 VIP 查询接口（N-05）：`POST /bizs/{bk_biz_id}/load_balancers/exclusive_clusters/idle_vips/list`，先校验目标集群归属（已分配给路径业务、且为 TGW 四层），再实时调云查询闲置 VIP（不落库，每次实时查）
- hc-service 层新增一个业务可调用的内部服务间接口，封装「按 `idle=true` 过滤 + 翻页取全 + Vip 去重」的完整查询逻辑（复用 `load-balancer-exclusive-cluster-purchase` 变更已新增的 `DescribeClusterResources` adaptor 方法），与该变更中"仅供内部联调/测试、不对外转发"的 `idle_vips/query` 端点区分职责，本变更新增的端点是业务闲置 VIP 查询链路的正式后端支撑，供 cloud-server 转发调用
- `pkg/client/hc-service` 新增对应客户端方法，供 cloud-server 调用新增的 hc-service 端点

## Capabilities

### New Capabilities

- `load-balancer-exclusive-cluster-tags-query`：业务视角独占集群标签聚合查询能力，覆盖 cloud-server 层鉴权、业务归属过滤、按标签+类型分组聚合、可用区过滤
- `load-balancer-exclusive-cluster-idle-vip-query`：业务视角独占集群闲置 VIP 查询能力，覆盖 cloud-server 层归属校验、hc-service 层翻页聚合查询、adaptor 层实时查云

### Modified Capabilities

（无：项目内暂无覆盖业务视角独占集群查询行为的既有 spec，本变更均以新增能力承载；不修改 `load-balancer-exclusive-cluster-list`（资源视角列表，已发布 spec）与 `load-balancer-exclusive-cluster-purchase`（购买链路，已发布 spec）的既有需求行为）

## Impact

- **cmd/cloud-server/service/load-balancer**：新增业务视角标签聚合查询、闲置VIP查询两个 handler；`load_balancer.go` 的 `bizService` 新增两条路由注册
- **cmd/cloud-server/logics/load-balancer**：新增归属校验辅助函数（复用/对齐 `load-balancer-exclusive-cluster-purchase` 变更已有的 `exclusive_cluster_check.go` 归属查询逻辑）
- **pkg/api/cloud-server/load-balancer**：新增两个接口的请求/响应类型
- **cmd/hc-service/service/load-balancer**：新增一个业务可调用的服务间接口（翻页取全查询闲置 VIP），与既有测试端点 `idle_vips/query` 并存、职责区分
- **pkg/client/hc-service/tcloud**：新增对应客户端方法
- **不涉及**：data-service、DAO、表结构改动（独占集群表已发布）；不修改 `DescribeClusterResources` adaptor 方法签名（`load-balancer-exclusive-cluster-purchase` 变更已发布，本变更仅新增调用方）；不涉及购买创建/提单校验逻辑本身（`load-balancer-exclusive-cluster-purchase` 变更负责）
