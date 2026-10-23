## 1. N-04 标签聚合查询接口（cloud-server）

- [x] 1.1 `pkg/api/cloud-server/load-balancer/exclusive_cluster.go`（或新建同类文件）新增请求/响应类型：`ListExclusiveClusterTagsReq`（`AccountID`/`Region`/`Isp` 必填，`Zone`/`ClusterType` 选填）、`ListExclusiveClusterTagsResult`（`Details []ExclusiveClusterTagGroup`）、`ExclusiveClusterTagGroup`（`ClusterTag`/`ClusterType`/`Clusters`）、`ExclusiveClusterTagItem`（`CloudClusterID`/`ClusterID`/`ClusterName`/`Egress`/`Isp`/`Zone`），字段 `json` tag 对齐 `docs/api-docs/web-server/docs/biz/load-balancer/list_exclusive_cluster_tags.md`
- [x] 1.2 `cmd/cloud-server/logics/load-balancer/` 新增聚合辅助函数：按 `bk_biz_id+network=Public+account_id+region+isp[+cluster_type]` 调用 `Client.DataService().Global.LoadBalancer.ListExclusiveCluster` 取候选集合，过滤 `cluster_tag=""` 的记录，反序列化 `extension` 按 `zone`（若传）过滤 `clusters_zone.master_zone`，按 `(cluster_tag, cluster_type)` 分组拼出结果
- [x] 1.3 `cmd/cloud-server/service/load-balancer/` 新增 handler（如 `exclusive_cluster_tags.go`）：解码校验请求 → `handler.ListBizAuthRes` 鉴权（`meta.LoadBalancer`+`meta.Find`）→ 调用 1.2 的聚合函数 → 返回结果
- [x] 1.4 `cmd/cloud-server/service/load-balancer/load_balancer.go` 的 `bizService` 新增路由注册：`POST /load_balancers/exclusive_clusters/tags/list`

## 2. N-05 闲置VIP查询接口（hc-service 新增服务间端点）

- [x] 2.1 `pkg/api/hc-service/load-balancer/tcloud.go` 新增请求/响应类型：`TCloudDescribeClusterIdleVipsReq`（`AccountID`/`Region`/`ClusterID` 必填）、`TCloudDescribeClusterIdleVipsResult`（`Count uint64`/`Details []string`）
- [x] 2.2 `cmd/hc-service/service/load-balancer/tcloud.go` 新增 handler：内部以 `Limit=100` 循环调用 `tcloudAdpt.DescribeClusterResources`（`Idle=true`），累加直到 `Offset>=TotalCount`，对 `Vip` 去重后返回 `{count, details}`；新增路由 `POST /vendors/tcloud/load_balancers/exclusive_clusters/idle_vips/describe`
- [x] 2.3 `pkg/client/hc-service/tcloud/clb.go` 新增客户端方法封装 2.2 的路由调用

## 3. N-05 闲置VIP查询接口（cloud-server）

- [x] 3.1 `pkg/api/cloud-server/load-balancer/exclusive_cluster.go` 新增请求/响应类型：`ListExclusiveClusterIdleVipsReq`（`AccountID`/`Region`/`CloudClusterID` 必填）、`ListExclusiveClusterIdleVipsResult`（`Count uint64`/`Details []string`）
- [x] 3.2 `cmd/cloud-server/logics/load-balancer/` 新增专用查询函数：按 `cloud_id=req.CloudClusterID` 查本地独占集群表（不带 `bk_biz_id` 条件），区分「查不到」（返回 `InvalidParameter`）、「类型非 TGW」（返回 `InvalidParameter`）、「`bk_biz_id` 与路径不符」（返回 `PermissionDenied`）三种结果
- [x] 3.3 `cmd/cloud-server/service/load-balancer/` 新增 handler：解码校验请求 → `handler.ListBizAuthRes` 鉴权 → 调用 3.2 的归属校验 → 通过后调用 2.3 的 hc-service 客户端方法 → 返回结果
- [x] 3.4 `load_balancer.go` 的 `bizService` 新增路由注册：`POST /load_balancers/exclusive_clusters/idle_vips/list`

## 4. 单元测试

- [x] 4.1 N-04 聚合辅助函数单测（覆盖 spec 中全部 Scenario：正常分组、空结果、空标签过滤、跨业务不可见、zone 过滤、`cluster_type` 过滤，含 mock data-service 客户端）
- [x] 4.2 N-04 handler 单测（覆盖必填参数校验失败、鉴权失败、请求体 `bk_biz_id` 被忽略）
- [x] 4.3 hc-service 翻页聚合 handler 单测（覆盖单页取全、多页翻页、`total_count=0`、adaptor 调用失败透传错误），云调用走 mock
- [x] 4.4 N-05 归属校验函数单测（覆盖查不到/类型非TGW/归属不符/正常通过四种场景）
- [x] 4.5 N-05 handler 单测（覆盖正常查询、越权拒绝且不发起云调用、参数校验失败），hc-service 客户端调用走 mock

## 5. 文档与需求单同步

- [x] 5.1 修正 `docs/reqs/独占集群购买查询.md` 中假设 2/规则 R-004 的措辞：STGW 分组 `clusters` 不再固定为空数组，改为与 TGW 一致返回真实集群列表；同步更新受影响的验收标准描述
- [x] 5.2 核对 `docs/api-docs/web-server/docs/biz/load-balancer/list_exclusive_cluster_tags.md` 的正文说明（"七层 clusters 固定为空数组"）与响应示例的矛盾，反馈文档作者按响应示例口径统一修正正文措辞；同时核对调用示例中的 `zones` 字段名应为 `zone`（与参数表一致）
- [ ] 5.3 性能验收测量：N-04 本地查询场景 P99<200ms（20 次合法请求）、N-05 含翻页场景 P99<3s（5 次合法请求，含至少 1 次触发翻页）在测试环境执行并记录结果
