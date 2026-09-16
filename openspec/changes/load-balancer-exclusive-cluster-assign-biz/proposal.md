## Why

腾讯云独占集群需要指定归属业务后，才能在购买负载均衡链路中做归属校验（方案设计 V-07）；独占集群的表结构、DAO 与 data-service CRUD（`load-balancer-exclusive-cluster-dao` 变更，B-04，已发布）已具备批量更新 `bk_biz_id` 的能力（`PATCH /load_balancer_exclusive_clusters/biz`），但该接口刻意不做业务前置校验、鉴权与审计（仅供内部调用），cloud-server 层目前完全没有暴露"分配业务"的对外入口。运营因此无法在"独占集群管理页"把未分配集群指定给业务，购买链路的标签归属校验（N-04 接口 V-07）也就没有数据基础。

技术方案：[负载均衡-购买支持独占集群-后端方案设计](https://iwiki.woa.com/p/4040244998)「七、接口设计 N-03」「八、购买链路改造 8.3 D 组 V-15」「十、权限与审计」「十二、后端任务拆解 B-09」；已完成澄清与评估的需求单：[独占集群分配业务接口](https://tapd.woa.com/tapd_fe/69995598/story/detail/1069995598138281829)（1069995598138281829，`v_status=approved`，size=3，effort=16h），完整需求文档见 `docs/reqs/独占集群分配业务.md`。

## What Changes

- 新增 cloud-server 接口 `POST /api/v1/cloud/load_balancers/exclusive_clusters/assign/bizs`，鉴权 `meta.LoadBalancer` + `meta.Assign`
- 前置校验：目标业务范围校验（复用 `common.ValidateTargetBizID`，与项目内所有 `XxxAssignToBiz` 接口一致）+ 已分配状态校验（`ValidateBeforeAssign`：仅"已分配给其它业务"才拒绝，"已分配给目标业务本身"视为幂等放行，与 `load-balancer.AssignLbToBiz` 语义一致，批量语义为整批拒绝不支持部分成功）
- 分配成功后记录审计（`LoadBalancerExclusiveClusterAuditResType`），需在 `cmd/data-service/service/audit/cloud/cloud_resource_assign_audit.go` 的 `buildAssignAuditInfo` 分发 switch 中新增分支，并新增对应的 `LoadBalancerExclusiveClusterAssignAuditBuild` 审计构建方法
- 调用 data-service 已有的 `PATCH /load_balancer_exclusive_clusters/biz`（B-04 已实现，本身不做校验/鉴权/审计）完成批量更新
- **不支持**（本期明确排除）：解绑（`bk_biz_id` 改回 -1）、重新分配（已分配给其它业务的集群改分配给新业务）

## Capabilities

### New Capabilities
- `load-balancer-exclusive-cluster-assign-biz`：独占集群批量分配业务能力，覆盖 cloud-server 层鉴权、目标业务范围校验、已分配状态前置校验、审计记录与批量更新编排

### Modified Capabilities

（无：本变更纯新增接口与新增审计分支，不修改任何已发布 spec 的行为；`cloud_resource_assign_audit.go` 的 switch 新增 case 属于新增能力接入，不改变已有分支的行为）

## Impact

- **cmd/cloud-server/service/load-balancer**：新增 `exclusive_cluster_assign.go`（`AssignExclusiveClusterToBiz` handler），`load_balancer.go` 的 `InitService` 新增一条路由注册
- **cmd/cloud-server/logics/load-balancer**：新增 `exclusive_cluster_assign.go`（`AssignExclusiveClusterToBiz` 逻辑函数 + `ValidateExclusiveClusterBeforeAssign` 前置校验）
- **pkg/api/cloud-server/load-balancer**：新增 `AssignExclusiveClusterToBizReq` 请求体（`ClusterIDs []string` + `BkBizID int64`）
- **cmd/data-service/service/audit/cloud/load-balancer**：新增 `exclusive_cluster.go`（`LoadBalancerExclusiveClusterAssignAuditBuild` 方法 + `ListExclusiveCluster` 查询辅助函数，挂在已有的 `LoadBalancer` 审计结构体上）
- **cmd/data-service/service/audit/cloud/cloud_resource_assign_audit.go**：`buildAssignAuditInfo` 的 switch 新增 `case enumor.LoadBalancerExclusiveClusterAuditResType` 分支
- **docs/api-docs**：核对已存在的草稿 `docs/api-docs/web-server/docs/resource/load-balancer/assign_exclusive_cluster_to_biz.md`，修正"整批拒绝条件"措辞（草稿现为"仅未分配集群可被分配"，需按已澄清结论改为"已分配给其它业务才拒绝，分配给目标业务本身放行"）
- 不涉及 DAO、表结构改动（均已在 B-04 完成）；不影响任何现有接口行为
