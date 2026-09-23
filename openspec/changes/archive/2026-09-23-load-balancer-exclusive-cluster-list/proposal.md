## Why

腾讯云独占集群的表结构、DAO 与 data-service CRUD 接口（`load-balancer-exclusive-cluster-dao` 变更，B-04）已落地，但 cloud-server 层尚无任何对外查询入口：运营无法在"资源接入-负载均衡-独占集群"管理页查看已同步的独占集群清单（含集群基础信息、分配状态、容量等）。这既是运营查看/管理独占集群的直接诉求，也是"独占集群分配业务"（`load-balancer-exclusive-cluster-assign-biz` 变更，N-03/B-09）能够选出候选集群的前置依赖——运营必须先看到列表才能选中集群发起分配。

技术方案：《负载均衡-购买支持独占集群-后端方案设计》「七、接口设计 N-01」「十、权限与审计」「十二、后端任务拆解 B-08」；对应需求单：资源视角下独占集群列表查询接口实现（1069995598138281819）。

## What Changes

- 新增资源视角查询接口 `POST /api/v1/cloud/load_balancers/exclusive_clusters/list`，鉴权 `meta.LoadBalancer` + `meta.Find`
- 采用项目标准 `ListResourceAuthRes` 鉴权模式（与 `cert.ListCert` 一致），按账号维度做资源过滤后转发 data-service 已有的 `POST /load_balancer_exclusive_clusters/list`
- 支持标准 `core.ListReq` 过滤：`bk_biz_id`（`-1` 表示未分配）、`cluster_tag`、`cluster_type`、`zone`、`isp`、`cloud_id`、`region`、`account_id` 等字段，`extension` 以 `tcloud` 强类型返回
- 本变更**仅覆盖资源视角**（需求单原文明确"POST /load_balancers/exclusive_clusters/list（资源视角）"），不包含业务视角查询接口、购买页标签聚合接口（N-04，另有单据）

## Capabilities

### New Capabilities
- `load-balancer-exclusive-cluster-list`：独占集群资源视角列表查询能力，覆盖 cloud-server 层鉴权、过滤转发与 vendor 强类型 extension 序列化

### Modified Capabilities

（无：本变更纯新增接口，不修改任何已发布 spec 的行为）

## Impact

- **cmd/cloud-server/service/load-balancer**：新增 `exclusive_cluster.go`（`ListExclusiveCluster` handler），`load_balancer.go` 的 `InitService` 新增一条路由注册
- **pkg/api/cloud-server/load-balancer**：新增响应类型（`tcloud` extension 强类型序列化后的列表结果），复用 data-service 已有的 `core.ListReq`/`corelb.BaseExclusiveCluster`
- **docs/api-docs**：核对已存在的草稿 `docs/api-docs/web-server/docs/resource/load-balancer/list_exclusive_cluster.md` 与最终实现字段一致（草稿已在 B-04 变更中同步过一轮，本变更只需按实现细节做最终核对）
- 不涉及 data-service、DAO、表结构改动（均已在 B-04 完成）；不影响任何现有接口行为
