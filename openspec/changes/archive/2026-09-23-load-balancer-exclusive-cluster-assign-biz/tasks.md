## 1. cloud-server 请求体

- [x] 1.1 `pkg/api/cloud-server/load-balancer/exclusive_cluster.go`（若 N-01 变更已创建该文件则追加）新增 `AssignExclusiveClusterToBizReq{ClusterIDs []string \`json:"cluster_ids"\`, BkBizID int64 \`json:"bk_biz_id"\`}`
- [x] 1.2 实现 `Validate()`：`BkBizID <= 0` 报错；`len(ClusterIDs) == 0` 报错；`len(ClusterIDs) > constant.BatchOperationMaxLimit`（100）报错；风格对齐 `AssignLbToBizReq.Validate()`

## 2. cloud-server logics 层

- [x] 2.1 新增 `cmd/cloud-server/logics/load-balancer/exclusive_cluster_assign.go`
- [x] 2.2 实现 `ValidateExclusiveClusterBeforeAssign(kt *kit.Kit, cli *dataservice.Client, ids []string, bizID int64) error`：`ListReq{Fields: []string{"id","bk_biz_id"}, Filter: tools.ContainersExpression("id", ids)}` 查 `cli.Global.ListExclusiveCluster`（实测该方法挂在 `Global` 顶层而非 `Global.LoadBalancer` 下，已按实际签名调用）；筛出 `bk_biz_id != constant.UnassignedBiz && bk_biz_id != bizID` 的记录，非空则返回整批错误（含具体 ID 列表），与 `lblogic.ValidateBeforeAssign` 写法一致
- [x] 2.3 实现 `AssignExclusiveClusterToBiz(kt *kit.Kit, cli *dataservice.Client, ids []string, bizID int64) error`：调用 2.2 前置校验 → `logicaudit.NewAudit(cli).ResBizAssignAudit(kt, enumor.LoadBalancerExclusiveClusterAuditResType, ids, bizID)` 记审计 → 调用 `cli.Global.BatchUpdateExclusiveClusterBizID(kt, &dataproto.ExclusiveClusterBatchUpdateBizIDReq{ClusterIDs: ids, BkBizID: bizID})`（同样挂在 `Global` 顶层）完成批量更新；不做关联资源级联（独占集群无子资源，参照 `cert.Assign` 而非 `lblogic.AssignTCloud`）

## 3. cloud-server service 层 handler

- [x] 3.1 新增 `cmd/cloud-server/service/load-balancer/exclusive_cluster_assign.go`
- [x] 3.2 实现 `AssignExclusiveClusterToBiz(cts *rest.Contexts) (any, error)`：解析 `cslb.AssignExclusiveClusterToBizReq` 并 `Validate()`
- [x] 3.3 调用 `common.ValidateTargetBizID(cts.Kit, svc.client.DataService(), enumor.LoadBalancerExclusiveClusterCloudResType, req.ClusterIDs, req.BkBizID)` 做目标业务范围校验
- [x] 3.4 调用 `svc.client.DataService().Global.Cloud.ListResBasicInfo` 以 `dataproto.ListResourceBasicInfoReq{ResourceType: enumor.LoadBalancerExclusiveClusterCloudResType, IDs: req.ClusterIDs}` 取各集群的 `account_id`
- [x] 3.5 按账号维度构造 `[]meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.LoadBalancer, Action: meta.Assign, ResourceID: info.AccountID}, BizID: req.BkBizID}`，调用 `svc.authorizer.AuthorizeWithPerm` 做权限校验
- [x] 3.6 权限通过后调用 `lblogic.AssignExclusiveClusterToBiz(cts.Kit, svc.client.DataService(), req.ClusterIDs, req.BkBizID)`（2.3 实现），返回其错误或 `nil`

## 4. 路由注册

- [x] 4.1 `cmd/cloud-server/service/load-balancer/load_balancer.go` 的资源视角路由注册处新增 `h.Add("AssignExclusiveClusterToBiz", http.MethodPost, "/load_balancers/exclusive_clusters/assign/bizs", svc.AssignExclusiveClusterToBiz)`

## 5. data-service 审计构建

- [x] 5.1 新增 `cmd/data-service/service/audit/cloud/load-balancer/exclusive_cluster.go`
- [x] 5.2 实现 `ListExclusiveCluster(kt *kit.Kit, dao dao.Set, ids []string) (map[string]tablelb.LoadBalancerExclusiveClusterTable, error)`：与 `base.go` 的 `ListLoadBalancer`/`ListTargetGroup`/`ListListener` 同一写法（`tools.ContainersExpression("id", ids)` + `dao.LoadBalancerExclusiveCluster().List`）
- [x] 5.3 为 `*LoadBalancer` 类型新增方法 `LoadBalancerExclusiveClusterAssignAuditBuild(kt *kit.Kit, assigns []protoaudit.CloudResourceAssignInfo) ([]*tableaudit.AuditTable, error)`：按 5.2 查询集群信息，逐条构造 `tableaudit.AuditTable{ResID, CloudResID, ResName, ResType: enumor.LoadBalancerExclusiveClusterAuditResType, Action: enumor.Assign, BkBizID, Vendor, AccountID, Operator, Source, Rid, AppCode, Detail: &tableaudit.BasicDetail{Changed: map[string]interface{}{"bk_biz_id": one.AssignedResID}}}`，与 `LoadBalancerAssignAuditBuild` 写法一致（不含 `switch one.AssignedResType` 的 `Deliver` 分支，本变更只支持 `BizAuditAssignedResType`，出现其它类型直接报错）

## 6. data-service 审计分发接入

- [x] 6.1 `cmd/data-service/service/audit/cloud/cloud_resource_assign_audit.go` 的 `buildAssignAuditInfo` switch 新增 `case enumor.LoadBalancerExclusiveClusterAuditResType: audits, err = ad.loadBalancer.LoadBalancerExclusiveClusterAssignAuditBuild(kt, assigns)`

## 7. 文档核对

- [x] 7.1 修正 `docs/api-docs/web-server/docs/resource/load-balancer/assign_exclusive_cluster_to_biz.md` 的"整批拒绝条件"措辞：将"仅未分配（`bk_biz_id` 为 -1）的集群可被分配，列表中存在已分配集群时整批拒绝"改为"已分配给其它业务的集群会导致整批拒绝；已分配给目标业务本身的集群视为幂等，随批次一起返回成功"

## 8. 自测（用户执行，不在本次实现范围内）

- [ ] 8.1 本地起 cloud-server + data-service，验证 `AssignExclusiveClusterToBiz` 的鉴权、目标业务校验、幂等放行、整批拒绝与审计记录行为
