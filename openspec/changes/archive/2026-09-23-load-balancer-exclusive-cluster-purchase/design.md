## Context

腾讯云负载均衡的独占型购买链路目前完全没有落地：`TCloudLoadBalancerSpec`（`pkg/api/hc-service/load-balancer/tcloud.go`）只有 `SlaType` 区分共享型/性能容量型，没有 `exclusive`/`cluster_tag`/`cloud_cluster_ids`；该结构体同时被业务视角申请单（`CreateForCreateLB`）、系统提单（`SysCreateForCreateLB`）、资源视角直接创建（`cmd/cloud-server/service/load-balancer/create.go`）与询价接口复用（同一份 `TCloudLoadBalancerCreateReq`/`TCloudLoadBalancerSpec`）。已核实的关键代码事实：

- 申请单创建流程为 `create()`（`cmd/cloud-server/service/application/create.go`）→ `handler.CheckReq()` → `handler.PrepareReq()` → ITSM 审批或 `directDeliver()` → 审批通过后 `handler.Deliver()` → `HCService().TCloud.Clb.BatchCreate()`；`ApplicationOfCreateTCloudLB.CheckReq()` 目前只有 `req.Validate(true)` + 资源账号校验（`cmd/cloud-server/service/application/handlers/load_balancer/tcloud/check.go`），是新增归属校验的天然落点。
- `hc-service` 的 `BatchCreateTCloudClb`（`cmd/hc-service/service/load-balancer/tcloud.go`）组装 `typelb.TCloudCreateClbOption` 后调用 `tcloudAdpt.CreateLoadBalancer`；该 option 结构体已经声明了 `ClusterIds []*string`、`ClusterTag *string`、`ExclusiveCluster *tclb.ExclusiveCluster` 字段（说明底层 adaptor 早已预留透传能力），但当前赋值逻辑完全没有填充这三个字段——这是一处"结构已备好、业务层未接入"的既有缺口，不是需要新增的 adaptor 类型。
- 项目实际使用的是**模块化** SDK 包 `github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317`（`go.mod` 锁定 `v1.0.1096`，`pkg/adaptor/types/load-balancer/tcloud.go` 的 `tclb` 别名即此包；不是单体包 `tencentcloud-sdk-go@v3.x+incompatible`，两者的 `LoadBalancer` 结构体字段不一样，排查时需认准 `go.mod` 里的模块路径）。在这个正确的包版本里，`DescribeLoadBalancers` 响应类型 `LoadBalancer` 上有**三个**独立字段：`ExclusiveCluster *ExclusiveCluster{L4Clusters, L7Clusters, ClassicalCluster}` 明确标注是**内网独占集群**，与本单（仅支持公网）无关；公网独占集群走另外两个平级字段——`ClusterTag *string`（7层独占标签）与 `ClusterIds []*string`（集群ID，注释未区分层级，需要按 ID 反查本地表才能知道某个 ID 是 TGW 还是 STGW）。这意味着"负载均衡详情回显独占集群"不需要新增云 API 调用，只需在既有的 `cmd/hc-service/logics/res-sync/tcloud/load_balancer.go` 的 `convertTCloudExtension` 里多读 `cloud.ClusterTag`/`cloud.ClusterIds` 这两个平级字段（而不是 `cloud.ExclusiveCluster` 内部字段），但没有"闲置 VIP 查询"这类云 API 的适配——`pkg/adaptor/tcloud/clb.go` 目前没有任何 `DescribeClusterResources`/idle 相关方法，F-004 的 VIP 闲置复核需要新增一个 adaptor 方法。
- 本地独占集群表（`pkg/dal/table/cloud/load-balancer/exclusive_cluster.go`）已有 `bk_biz_id`、`cluster_type`（TGW/STGW）、`cluster_tag`、`cloud_id`、`egress` 字段，且已有 `data-service` 的 `POST /load_balancer_exclusive_clusters/list`（`pkg/client/data-service/global/load_balancer_exclusive_cluster.go` 的 `ListExclusiveCluster`）可供 cloud-server/hc-service 直接查询，归属校验与出口比对都基于这张表，不需要新增表或字段。
- 带宽包出口（R-008 的 `Ep`）没有本地落库，是通过既有的 `hc-service` `/bandwidth_packages/list`（`pkg/client/hc-service/tcloud/bandwidth_pacakge.go` 的 `ListBandwidthPackage`）实时查云得到的 `TCloudBandwidthPackage.Egress`，cloud-server 层已经能拿到这个客户端，不需要新增查询通道。
- 已批准需求文档（`docs/reqs/业务视角购买.md`，TAPD 1069995598138281921）与新发布的接口文档草稿（commit `73db73f3`，`--story=138282352`）之间存在两处已澄清的分歧：(1) 接口文档写"`cluster_tag` 必填"，但已批准需求与用户本次澄清维持"`cluster_tag`/`cloud_cluster_ids` 至少一个非空即可"，且 `cluster_tag` 只代表七层、`cloud_cluster_ids` 只代表四层，两者独立校验，不是同一个"资源池标签"；(2) 接口文档新增了申请单详情/负载均衡详情的 `clusters` 富化契约，已批准需求原本明确排除（假设 5、Q-001），用户本次澄清后决定**扩大范围**一并纳入本变更。

## Goals / Non-Goals

**Goals:**
- 业务视角申请单与系统提单能接收、校验、透传独占型入参，合法组合（只四层/只七层/两层都选）均可提单并在审批通过后下云；非法组合在提单阶段以 `InvalidParameter`/`PermissionDenied` 拒绝
- 提单时（F-002）与下云前（F-004）都完整跑一遍结构+归属+出口一致性校验，下云前额外再加一项 VIP 闲置复核；两处校验规则实现共用同一套辅助函数，不因为"提单"和"下云"是两个不同入口而出现校验口径不一致
- 申请单详情与负载均衡详情能回显独占集群的人类可读信息（集群名、本地 ID、标签、类型），复用已同步的本地独占集群表，不引入新的云 API 依赖（LB 详情）或仅引入一次必要的云 API（VIP 闲置复核）

**Non-Goals:**
- 不改询价接口（询价场景下要不要支持独占型，需求单第 1 轮澄清已确认不纳入本单）
- 不在资源视角创建/询价里拒绝独占字段（兄弟单 1069995598138281904 负责；本单先合入时资源视角可能短暂能透传独占字段，属已知窗口期风险，见需求单 Q-002）
- 不实现购买页用的标签聚合/闲置 VIP 列表对外 API（`list_exclusive_cluster_tags`/`list_exclusive_cluster_idle_vips`，兄弟单 1069995598138281929）；本单在**内部下云前校验**里复用同一条云上能力，并在 hc-service 层额外加一个测试用端点（见决策 5），但不在 cloud-server/web-server 对外暴露、不做面向业务用户的鉴权与参数封装
- 不做内网独占集群、非腾讯云 vendor 支持
- 不按七层标签实时反查云上候选集群再解析成 `ClusterIds`（需求单已否决的方案）

## Decisions

### 1. 独占字段加在共享的 `TCloudLoadBalancerSpec`，校验按"能否脱离 DB 判断"拆成两层

**选择**：`Exclusive *int64`、`ClusterTag *string`、`CloudClusterIDs []string` 加在 `TCloudLoadBalancerSpec`（`ValidateSpec()` 覆盖）；纯结构规则（R-001~R-006、R-010、R-011：独占型仅公网、`sla_type` 互斥、`cluster_tag`/`cloud_cluster_ids` 至少一个非空、`vip` 依附四层且恰好一个、未指定四层不许传 `vip`、单线必须带宽包、指定 `vip` 时 `require_count=1`）放进 `ValidateSpec()`/`TCloudLoadBalancerCreateReq.Validate()`；需要查本地表或调云的规则（R-007 归属、R-008 出口一致性、VIP 闲置）放进 `ApplicationOfCreateTCloudLB.CheckReq()`（提单时全量跑）与 hc-service `BatchCreateTCloudClb`（下云前同样全量重跑归属+出口一致性校验，额外再加 VIP 闲置这一项）。

**原因**：`TCloudLoadBalancerSpec` 已经是业务视角/系统提单/资源视角创建/询价四个入口共用的结构体，新增字段不拆分新结构体可以让四个入口自动获得字段能力，资源视角的"应该拒绝"由兄弟单在其 `CheckReq` 里做黑名单式拦截，不需要本单为它单独建一份不含独占字段的结构体。`ValidateSpec()` 是纯函数（不依赖 kit/DB），只能承载结构规则；`CheckReq()` 本来就是"能拿到 `Cts.Kit`、`Client.DataService()`"的地方，是归属/出口这类需要查表规则的既有落点（对照 `logicsaccount.IsResourceAccount` 的现有用法）。

**备选方案**：为业务视角购买单独定义 `TCloudExclusiveLoadBalancerCreateReq` → 已否决，会导致系统提单、资源视角创建两个复用同一 handler/adaptor 链路的入口需要做结构体转换胶水代码，且与"两个提单入口共用同一套校验"的既有需求（F-001）矛盾。

### 2. `cluster_tag`（七层）与 `cloud_cluster_ids`（四层）各自独立做归属校验，不引入"资源池标签统一限定"语义

**选择**：`cluster_tag` 非空时，查本地表 `cluster_type=STGW AND bk_biz_id=当前业务 AND cluster_tag=请求值`，命中至少一条即通过；`cloud_cluster_ids` 中每个 ID 都必须能在本地表 `cluster_type=TGW AND bk_biz_id=当前业务 AND cloud_id=该ID` 命中。两者是"与"关系的两个独立子校验，不要求 `cloud_cluster_ids` 对应的 TGW 必须挂在 `cluster_tag` 下。

**原因**：本次澄清用户明确"购买的时候 `cloud_cluster_ids` 代表四层独占集群的购买，`cluster_tag` 则代表七层独占集群"，即两者是并列的两种维度而非"tag 统领四层候选范围"。接口文档 `list_exclusive_cluster_tags.md` 里同一个 `cluster_tag` 同时聚合 TGW/STGW 分组，只是购买页做候选展示用的分组方式，不代表创建接口的校验语义必须绑死这个分组——已按用户澄清结论以已批准需求文档（假设 1）为准，接口文档"`cluster_tag` 必填"的措辞判定为文档遗留问题，本次实现不采纳，需在实现验收时同步给文档作者反馈修正。
**因此 R-008 出口计算维持既有的 `E4_set`/`E7_set` 独立集合与交集运算，不做简化。**

**备选方案**：`cloud_cluster_ids` 必须属于 `cluster_tag` 关联的 TGW（tag 统一限定候选） → 已否决，与用户本次澄清结论直接冲突。

### 3. F-002（提单）与 F-004（下云前）都跑全量校验，F-004 额外再加 VIP 闲置复核

**选择**：F-002（`CheckReq`，位于 cloud-server）跑全部结构 + 归属 + 出口一致性校验（R-001~R-008、R-010、R-011）；F-004（`Deliver`/`BatchCreateTCloudClb`，位于 hc-service，下云前）**同样完整重跑一遍**结构 + 归属 + 出口一致性校验（不是只挑两项），再额外加一项「指定 VIP 是否仍闲置」。两处校验判定逻辑保持一致（同样的表、同样的过滤条件、同样的比对规则），但因 cloud-server 与 hc-service 是两个独立服务、不能跨服务直接调用对方包内函数，hc-service 侧需要基于 `Client.DataService().Global.*` 新增一份等价实现，不是字面意义上共享同一个 Go 函数。

**原因**：本次澄清用户明确要求提单与下云前"都要做校验"，不按"是否会随时间变化"做取舍——业务的独占集群分配关系（归属）、带宽包/集群出口在 ITSM 审批等待期间同样可能被运营改变（例如集群被重新分配给其它业务、带宽包被删除或改配置），只在下云前复核 VIP 闲置和出口是不够的，归属也需要兜底重新校验，否则会出现"提单时归属校验通过，审批通过后集群已被转移给别的业务，仍然创建成功"的漏洞。VIP 闲置检查因为在提单时查询没有意义（审批期间几乎必然变化），所以只在下云前做（不是提单时跳过的唯一理由从"可以偷懒"变成"提单时做没有意义"）。

**备选方案**：F-002 全量、F-004 只挑 VIP 闲置 + 出口两项（原设计） → 已否决，与用户本次澄清结论冲突，且遗漏了"归属关系在审批期间被改变"这一真实风险。

### 4. 带宽包出口（`Ep`）查询复用 `hc-service` 既有 `/bandwidth_packages/list`，不新增本地表

**选择**：`CheckReq()` 与 `BatchCreateTCloudClb` 都通过 `Client.HCService().TCloud.BandwidthPackage.ListBandwidthPackage(kt, &ListTCloudBwPkgOption{PkgCloudIds: [bandwidth_package_id]})` 实时查云拿 `Egress`，查不到（返回 0 条）按 `InvalidParameter` 处理。

**原因**：项目内带宽包数据本来就没有本地落库（`pkg/dal` 下无 `bandwidth_package` 表），既有的带宽包选择流程也是通过这一实时查云接口获取数据；引入本地缓存表会带来额外的同步一致性成本，且与兄弟单 M-06（带宽包接口增加出口过滤条件，本单不实现）职责重叠。

**备选方案**：新增本地带宽包缓存表，同步 `Egress` 字段 → 已否决，超出本单范围且与 M-06 职责重叠，本单只是"读"这个字段，不需要建立同步管线。

### 5. VIP 闲置复核新增一个 adaptor 方法；hc-service 层额外暴露一个 HTTP 端点用于测试，不在 cloud-server/web-server 对外暴露

**选择**：在 `pkg/adaptor/tcloud/clb.go` 新增 `DescribeClusterResources(kt, opt)` 方法（映射腾讯云 `DescribeClusterResources` API）；`hc-service` 的 `BatchCreateTCloudClb` 在指定了 `vip` 时调用它确认该 VIP 仍在 `cloud_cluster_ids[0]` 对应集群的闲置列表内。**同时**在 `cmd/hc-service/service/load-balancer/tcloud.go` 新增一个薄封装的 HTTP 端点（如 `POST /vendors/tcloud/load_balancers/exclusive_clusters/idle_vips/query`，命名与路径需与兄弟单 1069995598138281929 未来对外的 `list_exclusive_cluster_idle_vips` 区分，避免路由/语义混淆），直接透出 `DescribeClusterResources` 的查询能力，供联调/测试直接调用 hc-service 验证闲置 VIP 数据，不需要每次都跑完整购买链路才能验证这一步。该端点只注册在 hc-service（服务间调用层），不经 cloud-server/web-server 对外暴露，不需要额外鉴权改造（hc-service 现状是内部服务间调用，无终端用户直接访问）。

**原因**：F-004/AC-013/AC-017 的功能需求只要求"下云前内部复核"，但本次用户明确要求 hc-service 侧要有一个可独立调用的 HTTP 端点用于测试——不依赖完整跑一遍 ITSM 审批+下云流程就能验证 `DescribeClusterResources` adaptor 方法与闲置 VIP 判断逻辑是否正确，这对单测之外的集成联调阶段是必要的调试入口。放在 hc-service 而不是 cloud-server/web-server，是因为对外的"购买页闲置 VIP 查询"属于兄弟单 1069995598138281929 的范围（面向业务用户的鉴权、参数校验、响应格式都由该单负责），本单只需要一个内部测试入口，两者路径与鉴权层级不同，不会产生同一功能重复对外暴露的问题。

**备选方案**：只在 adaptor 层新增方法，不暴露任何 HTTP 端点（原设计） → 已否决，与用户本次澄清结论冲突，且会导致该能力在集成测试阶段只能通过完整购买链路间接验证，联调效率低。

### 6. 申请单详情的 `clusters` 富化在读路径按需解析，不改写落库的 `content`

**选择**：`buildApplicationGetResp`（`cmd/cloud-server/service/application/get.go`）在 `application.Type == enumor.CreateLoadBalancer` 时，对 `RemoveSenseField` 后的 `content` 做一次反序列化，取出 `exclusive`/`cluster_tag`/`cloud_cluster_ids`，非独占型（`exclusive` 非 1）直接跳过；独占型则查本地独占集群表（`cluster_tag` 查 STGW、`cloud_cluster_ids` 查 TGW），拼出 `clusters` 数组注入返回的 `content` JSON 字符串后再返回给调用方；数据库里持久化的 `content` 保持申请时的原始请求体不变（不做写时富化）。

**原因**：`clusters` 是"当前本地独占集群表状态"的展示快照，读时计算能保证展示的集群名称/本地 ID 始终反映最新的本地表（例如集群改名后旧申请单详情也能看到新名字），而写时固化会导致集群改名/重新分配后历史申请单详情展示过期信息；申请单本身是一次性审批记录，不需要为了省这一次查询而牺牲展示准确性。`GetApplication`/`GetBizApplication` 目前都是低频、单条查询场景（详情页），多一次本地表查询对性能无实质影响。

**备选方案**：审批通过下云时把 `clusters` 一次性写回 `content`（写时固化）→ 已否决，写路径要改申请单更新逻辑（目前 `content` 创建后不再更新），且拒审、待审状态的申请单详情页仍需要展示 `clusters`（此时还没下云，没有下云结果可写）。

### 7. 负载均衡详情 `extension.clusters` 消费云上 `DescribeLoadBalancers` 的 `ClusterTag`/`ClusterIds` 平级字段（不是 `ExclusiveCluster`），`ClusterIds` 按 ID 反查本地表判定层级

**选择**：`convertTCloudExtension`（`cmd/hc-service/logics/res-sync/tcloud/load_balancer.go`）新增：读取 `cloud.ClusterTag`（7层独占标签）与 `cloud.ClusterIds`（集群 ID 数组，**不假设**其中元素都是 TGW——同一个数组可能混有公网 TGW 与 STGW 的 ID，具体层级必须按 ID 查本地表才能确定），两者任一非空即设置 `Extension.Exclusive = true`；**不**读取 `cloud.ExclusiveCluster`（该字段是内网独占集群，与本单公网场景无关，`L4Clusters`/`L7Clusters` 语义不适用于本单）。同步批次里把所有 LB 的 `ClusterIds` 去重后一次性查本地独占集群表（`cloud_id in (...)`），本地表返回的 `cluster_type` 才是该 ID 的真实层级（TGW/STGW），据此关联出 `cluster_id`/`cluster_name`/`cluster_tag`/`cluster_type` 填入 `Extension.Clusters`；若 `cloud.ClusterTag` 非空但 `ClusterIds` 反查结果中没有对应的 STGW 记录（云侧未回传具体 STGW 落地 ID，只回传了标签），额外补一条只有 `cluster_tag`/`cluster_type=STGW` 有值、其余字段为空字符串的元素。本地表查不到的 ID（尚未同步或已删除）同样填空字符串，`cloud_cluster_id` 保留云上原值。

**原因**：已核实项目实际使用的模块化 SDK（`go.mod` 锁定 `tencentcloud-sdk-go/tencentcloud/clb v1.0.1096`）中，`LoadBalancer`（`DescribeLoadBalancers` 响应）的 `ExclusiveCluster` 字段明确注释为"内网独占集群"，与公网独占型（本单唯一支持的场景，R-003）无关；公网独占集群的关联信息是另外两个平级字段 `ClusterTag`/`ClusterIds`。`ClusterIds` 的注释未区分层级，且创建请求侧 `cloud_cluster_ids`（四层）与 `cluster_tag`（七层）是分开传的两个字段，不能想当然认为回读时 `ClusterIds` 里全是 TGW——必须以本地独占集群表的 `cluster_type` 为准，而不是以"这个字段名字听起来像哪一层"做假设。

**备选方案**：假定 `ClusterIds` 全部是 TGW、直接按创建时的字段语义反推层级，不查本地表确认 → 已否决，与本次用户澄清结论冲突，且云侧行为一旦包含非纯 TGW 的 ID（用户已提示"可能包含了四层、七层"）就会导致 `cluster_type` 展示错误。

## Risks / Trade-offs

- **[Risk] 资源视角创建/询价复用同一个 `TCloudLoadBalancerSpec`，若兄弟单 1069995598138281904（资源视角拒绝独占字段）晚于本单合入，会有窗口期允许资源视角直接创建独占型 LB（绕过业务归属校验）** → **Mitigation**：与需求单 Q-002 一致，已知风险不在本单解决；建议与该兄弟单同批发布，若无法同批，至少确保资源视角创建路径的现状（不传独占字段）在文档中标注为临时限制
- **[Risk] `DescribeClusterResources`（VIP 闲置复核）是本单新增的 adaptor 方法，若腾讯云该接口有限流或响应慢，会直接影响下云耗时（性能验收 AC-P02 要求 P99<3s）** → **Mitigation**：仅在"指定了 `vip`"这一较少见路径上调用，不影响"随机分配"这一主路径；调用失败按需求文档"视为校验失败，不继续下云"处理，不做重试阻塞
- **[Risk] 申请单详情读时富化（决策 6）在本地独占集群表被大量删除/未同步的边界情况下，会让历史申请单详情出现大量空字符串字段** → **Mitigation**：这是需求文档与接口文档都已经显式约定的行为（"集群在本地表被删除或尚未同步时，`cluster_id`/`cluster_name` 为空字符串"），非本设计引入的新风险，前端按此约定处理展示即可
- **[Trade-off] LB 详情富化（决策 7）依赖同步任务已经跑过一次，新创建的独占型 LB 在第一次同步完成前，详情接口的 `extension.clusters` 会是空数组** → 与现有其它 `extension` 字段（如 `vip_isp`）的一致性行为相同，不是本变更特有的时序问题，不额外处理
- **[Trade-off] R-008 出口比对与归属校验都在 `CheckReq()` 里同步执行多次本地表/云查询（歸属 1~2 次 + 带宽包 1 次），提单接口 RT 会比现有共享型/性能容量型请求略高** → 已计入性能验收 AC-P01（P99<200ms，不含审批），本地表查询走索引（`bk_biz_id`+`cluster_type`+`cluster_tag`/`cloud_id`），量级可控；带宽包查询是既有链路已有的开销模式，非本单新增的性能负担类别
- **[Trade-off] 决策 3 改为 F-004 也全量重跑结构+归属+出口校验后，`BatchCreateTCloudClb` 下云前的查询次数从"2 次"增加到"5~6 次"（结构校验不查库，归属 1~2 次 + 出口 1 次 + VIP 闲置 1 次），下云耗时会比原设计略增** → 已计入性能验收 AC-P02（P99<3s），下云本身是异步/低频操作（相对提单接口的高频调用），多几次索引查询的绝对耗时可控；换来的是审批期间归属被改变这一真实漏洞被兜底，权衡后判定值得
- **[Trade-off] 决策 5 新增的 hc-service 测试端点在 `DescribeClusterResources` 云 API 侧没有独立的限流保护，若被误用于高频轮询会消耗与正式下云链路共享的云 API 配额** → 该端点定位是联调/测试用途，不面向终端用户，不做限流是可接受的初始范围；若后续发现被滥用，可在该端点单独加限流，不影响本设计

## Migration Plan

纯新增字段与新增校验分支，`exclusive` 缺省 0 时行为与上线前完全一致（存量共享型/性能容量型请求不受影响）；不涉及表结构变更（复用已发布的独占集群表）。部署顺序：无强制前置依赖上线要求，但依赖独占集群已同步且已分配到业务（`load-balancer-exclusive-cluster-dao`/`-assign-biz` 已发布）才能实际使用独占型购买，否则归属校验会全部拒绝（符合预期，不是故障）。回滚只需下线本次新增的校验分支与 `createOpt` 字段透传，不影响任何存量数据；申请单详情/LB 详情的 `clusters` 富化是纯读路径新增字段，回滚不影响已展示的历史数据。

## Open Questions

- 接口文档「`cluster_tag` 必填」与「`仅指定四层集群」示例缺失 `cluster_tag`」的自相矛盾措辞，需要在实现验收阶段反馈给文档作者按本设计决策 2 的结论统一修正（含 `create_application_for_create_tcloud_load_balancer.md` 与 `sys_create_application_for_create_tcloud_load_balancer.md` 两处）
- `docs/reqs/业务视角购买.md` 的「本期不包含」「假设 5」「Q-001」需要在本变更落地后同步更新为"本期已包含"，并重新核对该需求单的工时/规模评分是否需要因扩大范围而上调（当前 17.5h/size=3 未计入 detail 富化的工作量）
