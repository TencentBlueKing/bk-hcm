## Context

已批准需求文档 `docs/reqs/独占集群购买查询.md`（TAPD 1069995598138281929）与已定稿接口文档 `docs/api-docs/web-server/docs/biz/load-balancer/list_exclusive_cluster_tags.md`、`list_exclusive_cluster_idle_vips.md` 已经把两个接口的契约定清楚。已核实的关键代码事实：

- **业务视角路由挂载方式**：`cmd/cloud-server/service/load-balancer/load_balancer.go` 的 `InitService` 里，`bizH := rest.NewHandler(); bizH.Path("/bizs/{bk_biz_id}")` 统一挂载业务视角路由；`handler.ListBizAuthRes`（`pkg/tools/hooks/handler/query.go`）是项目标准的业务视角鉴权模式——按路径 `bk_biz_id` 做 `meta.Authorize`，通过后若传入 `Filter` 会自动 AND 上 `bk_biz_id=路径ID` 的过滤条件，`ListBizLoadBalancer` 已是这一模式的现成参照。本变更两个新接口都复用这一模式，天然满足"请求体不接受 `bk_biz_id`、恒以路径参数为准"的既有澄清结论。
- **本地独占集群表已有字段**（`pkg/api/core/cloud/load-balancer/exclusive_cluster.go`）：`BaseExclusiveCluster` 顶层字段含 `BkBizID`、`Region`、`Zone`、`ClusterType`、`ClusterTag`、`Network`、`Isp`、`Egress`、`CloudID`；`TCloudExclusiveClusterExtension.ClustersZone.MasterZone []string` 是嵌套在 `extension` JSON 里的多可用区数组，与顶层 `Zone` 字段是两个不同语义的字段（顶层 `Zone` 是单值主可用区记录，`extension.clusters_zone.master_zone` 才是接口文档要求按其过滤的字段）。data-service 已有的 `POST /load_balancer_exclusive_clusters/list`（`pkg/client/data-service/global/load_balancer_exclusive_cluster.go` 的 `ListExclusiveCluster`）支持对顶层字段做过滤，但通用 `core.ListReq` 过滤器不支持对 JSON 类型的 `extension` 列做嵌套字段过滤——`zone` 过滤只能在 cloud-server 层拿到候选集群后反序列化 `extension` 再过滤。
- **N-05 依赖的云 API 能力已具备但只有底层单页调用**：`load-balancer-exclusive-cluster-purchase` 变更已经新增了 `pkg/adaptor/tcloud/clb.go` 的 `DescribeClusterResources` adaptor 方法（支持 `Idle *bool`、`Limit`/`Offset` 单页查询）以及 hc-service 内部测试端点 `POST /vendors/tcloud/load_balancers/exclusive_clusters/idle_vips/query`（`cmd/hc-service/service/load-balancer/tcloud.go`，注释明确"仅供内部联调/测试使用，不经 cloud-server/web-server 转发"，且不做翻页循环，一次只返回一页）。需求文档要求"翻页取全"，且该测试端点被前一变更设计明确排除在业务调用范围之外，因此本变更需要新增一个专门给 cloud-server 调用的 hc-service 端点，内部完成翻页循环 + `idle=true` 过滤 + 结果合并，而不是复用/改造那个测试端点。
- **归属校验辅助函数已有可参照实现**：`load-balancer-exclusive-cluster-purchase` 变更已在 `cmd/cloud-server/logics/load-balancer/exclusive_cluster_check.go`（提单场景）与 `cmd/hc-service/service/load-balancer/exclusive_cluster_check.go`（下云前复核场景）各实现了一份"按 `cluster_tag`/`cloud_cluster_ids` 查本地表校验业务归属"的逻辑，但两处校验对象都是"多个集群/标签是否都归属同一业务"，返回值是简单的通过/拒绝；本变更 N-05 需要区分"集群本地不存在"（`InvalidParameter`）与"集群存在但归属别的业务"（`PermissionDenied`）两种错误语义，与既有辅助函数的返回粒度不同，需要新增一个专用查询函数，不能直接复用现有函数签名。
- **本次澄清结论（用户已确认）**：① 本变更范围为 N-04+N-05 一起做；② `list_exclusive_cluster_tags.md` 的响应示例与正文说明不一致（正文说 STGW 分组 `clusters` 固定为空数组，示例 JSON 却填充了真实 `stgw-` 集群数据），以**响应示例为准**——STGW 分组也返回真实集群列表，供购买页展示标签下的物理集群（即使购买时用户仍只需选标签、不选具体七层集群）。此结论**推翻**了此前已批准需求文档中的假设 2/R-004（"STGW 分组 clusters 固定为空数组"），本变更落地时需要同步在需求文档与接口文档中标注这一修正。

## Goals / Non-Goals

**Goals:**
- 业务视角能一次调用获取本业务已分配的公网独占集群标签聚合结果（TGW、STGW 分组均含真实集群列表），支撑购买页标签下拉、集群下拉与"独占型"选项可见性判断
- 业务视角能对指定四层集群做实时闲置 VIP 查询（翻页取全、不落库），支撑购买页"指定 IP"下拉
- 两个接口都严格按路径 `bk_biz_id` 做归属过滤/校验，不接受请求体 `bk_biz_id` 覆盖，跨业务访问一律拒绝（前者静默过滤、后者返回 `PermissionDenied`）

**Non-Goals:**
- 不改动 data-service 层与本地表结构（复用已发布的通用列表接口与已发布的独占集群表）
- 不改动 `DescribeClusterResources` adaptor 方法签名（`load-balancer-exclusive-cluster-purchase` 变更已发布，本变更只新增一个更高层的 hc-service 调用封装）
- 不实现购买创建/提单校验逻辑本身（`load-balancer-exclusive-cluster-purchase` 变更已完成，本变更只提供查询数据源）
- 不修改已有的资源视角独占集群列表接口（`load-balancer-exclusive-cluster-list` 变更）
- 不做共享带宽包按出口过滤（M-06，另有变更）

## Decisions

### 1. N-04 聚合逻辑放在 cloud-server 层，不新增 data-service 接口；`zone` 过滤在应用层做

**选择**：cloud-server 新增 handler 调用已有的 `Client.DataService().Global.LoadBalancer.ListExclusiveCluster`（复用现成客户端方法，非新增），过滤条件为 `bk_biz_id=路径ID AND network=Public AND account_id=req.AccountID AND region=req.Region AND isp=req.Isp [AND cluster_type=req.ClusterType]`（`cluster_type` 为空时不加此条件，取两类）；拿到候选集合后在内存里：① 过滤 `cluster_tag=""` 的记录；② 反序列化每条记录的 `extension`（`corelb.TCloudExclusiveClusterExtension`），若 `req.Zone` 非空则判断 `extension.clusters_zone.master_zone` 是否包含该值，不包含则整条记录被过滤掉；③ 按 `(cluster_tag, cluster_type)` 分组，组内按 `cloud_id` 顺序拼出 `clusters` 数组（`cloud_cluster_id`/`cluster_id`/`cluster_name`/`egress`/`isp`/`zone`，`zone` 取顶层 `Zone` 字段，不是 `master_zone` 数组）；某标签下某类型分组过滤后为空集群列表时，该分组条目本身也不放入 `details`（与已批准需求文档 R-005 一致）。

**原因**：聚合所需的候选数据量级是"单业务在单个账号/地域/运营商组合下已分配的独占集群数"，参照已有资源同规模数据（分配业务变更里同类查询），预期在几十条以内，内存分组开销可忽略，不值得为此新增一个 data-service 专用聚合接口；`zone` 过滤依赖 `extension` 嵌套字段，data-service 的通用 `core.ListReq` 过滤器本身不支持 JSON 路径过滤，在 cloud-server 层反序列化后过滤是最小改动路径。

**备选方案**：在 data-service 新增一个"按标签聚合"的专用查询接口，SQL 层用 JSON 函数过滤 `zone` → 已否决，需要新增 data-service 路由 + SQL JSON 查询语法（不同 DB 方言写法不同，增加维护成本），且当前数据量级不需要下沉到 SQL 层聚合。

### 2. STGW 分组返回真实集群列表（采纳响应示例口径，推翻此前假设 2/R-004）

**选择**：TGW、STGW 两类分组的 `clusters` 字段处理逻辑完全一致——都是"过滤后按 `cluster_tag` 分组内的真实集群清单"，不对 STGW 做特殊的"固定返回空数组"处理。

**原因**：本次用户澄清已明确以已定稿接口文档的响应示例为准；采纳该口径后 N-04 的聚合代码不需要按 `cluster_type` 分支处理，两类分组走同一份聚合逻辑，实现更简单、一致性更好。购买时用户对七层仍然只传 `cluster_tag`（不传具体七层 `cloud_cluster_id`），`clusters` 里的七层集群条目只是给购买页展示"该标签下有哪些物理集群"的参考信息，不影响提单参数结构。

**备选方案**：维持假设 2/R-004，STGW 分组 `clusters` 固定空数组 → 已否决，与本次用户澄清结论冲突；需要在落地后同步反馈修正已批准需求文档与本次结论保持一致的说明。

### 3. N-05 新增一个 hc-service 业务查询端点，与既有内部测试端点并存、职责区分

**选择**：hc-service 新增 `POST /vendors/tcloud/load_balancers/exclusive_clusters/idle_vips/describe`（命名参照既有的 `TCloudDescribeResources` 走 `resources/describe` 的动词式命名，区别于已有测试端点的 `idle_vips/query`），内部实现：以 `Limit=100` 循环调用 `DescribeClusterResources`（`Idle=true` 过滤），累加 `Resources` 直到 `Offset >= TotalCount`，对 `Vip` 去重后返回 `{count, details}`；`pkg/client/hc-service/tcloud` 新增对应客户端方法；cloud-server 新的 N-05 handler 通过该客户端方法调用，不直接触达 adaptor（分层不变）。

**原因**：既有测试端点被前一变更明确设计为"不对外转发、仅供联调"，且不做翻页（单页 `Limit`/`Offset` 由调用方指定，调一次只返回一页），若直接复用会导致 cloud-server 需要自己感知翻页协议并发起多次跨服务 HTTP 调用（每页一次网络往返），增加时延、不利于满足 AC-P02（P99<3s）；把翻页循环下沉到 hc-service 内部（hc-service 到腾讯云的调用本身也是网络往返，但同一进程内循环比跨服务多次往返更快），cloud-server 只需一次调用即可拿到完整闲置 VIP 列表。

**备选方案**：直接复用测试端点 `idle_vips/query`，翻页循环放在 cloud-server 层 → 已否决，增加跨服务网络往返次数、且需要修改前一变更"不对外转发"的既有设计约束（该端点已发布，修改其语义定位需要重新评估影响面）。

### 4. N-05 归属校验区分"集群不存在"与"集群存在但不归属当前业务"两种错误语义

**选择**：新增专用查询函数，按 `cloud_cluster_id`（不带 `bk_biz_id` 条件）查本地独占集群表：查不到 → `InvalidParameter`；查到但 `cluster_type != TGW` → `InvalidParameter`；查到且为 TGW 但 `bk_biz_id != 路径业务ID` → `PermissionDenied`；三项都通过后才发起 hc-service 调用。

**原因**：已批准需求文档假设 1/假设 2 已经明确这两种场景要返回不同错误码（分别对应 AC-010、AC-011 与 AC-009），`load-balancer-exclusive-cluster-purchase` 变更里现有的归属校验函数（`checkExclusiveClusterOwnership`）是"批量校验一组 ID/标签是否都归属"的粗粒度语义，返回值统一是 `PermissionDenied`，不满足本变更"区分不存在与不归属"的精细语义要求，因此新增一个专用函数而不是复用/修改现有函数（避免影响已发布行为）。

**备选方案**：复用现有粗粒度校验函数，查不到时也返回 `PermissionDenied` → 已否决，与已批准需求文档假设 1 的错误码要求冲突。

## Risks / Trade-offs

- **[Risk] N-04 在 cloud-server 层反序列化每条候选记录的 `extension` 做 `zone` 过滤，若单业务独占集群规模远超预期（例如未来扩展到数百条），内存聚合的耗时会上升** → **Mitigation**：先按 `account_id+region+isp[+cluster_type]` 缩小候选集合再反序列化（已经不是全表扫描），当前业务规模（独占集群是稀缺资源，通常个位数到几十）不构成实际性能问题；若未来规模显著增长，可再评估是否需要下沉到 data-service/SQL 层
- **[Risk] N-05 新增的 hc-service 端点内部循环调用腾讯云 `DescribeClusterResources`，若某集群闲置 VIP 数量很大导致翻页次数多，会增加下云前端点的响应时间** → **Mitigation**：与 AC-P02（P99<3s）已考虑翻页场景一致，`Limit=100` 单页步长在正常独占集群容量下（`max_conn`/容量字段通常是数百到数千量级的连接数，不是 VIP 数量，实际闲置 VIP 数predict远小于 100）预期 1~2 页即可取全；仅在"指定具体四层集群"这一较少见路径触发，不影响"随机分配"主路径
- **[Trade-off] STGW 分组按响应示例返回真实集群列表（决策 2）与此前已批准需求文档的假设 2/R-004 冲突，需要在本变更落地后同步修正需求文档措辞，避免后续读者以为两处结论矛盾** → 已在本设计文档"Context"与"决策 2"中显式记录推翻理由，落地后的文档修正作为本变更的收尾任务之一（见 tasks.md）
- **[Trade-off] N-05 新增的 hc-service 端点与既有测试端点在功能上有重叠（都基于同一个 adaptor 方法），存在两个语义相近的路由** → 已通过命名区分（`describe` 业务用、`query` 测试用）并在代码注释中标注职责边界；后续如确认测试端点已无实际使用场景，可在单独的清理变更中评估下线

## Migration Plan

纯新增接口，不改动任何已发布接口的行为、不改表结构。部署无强制前置依赖，但依赖独占集群已同步且已分配到业务（`load-balancer-exclusive-cluster-dao`/`-assign-biz` 已发布）才能返回非空结果，否则返回空 `details`（符合预期，不是故障）。回滚只需下线本变更新增的两个路由与 hc-service 新增端点，不影响任何存量数据。
