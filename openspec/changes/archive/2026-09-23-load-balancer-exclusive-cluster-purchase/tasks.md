## 1. 独占字段与结构校验（`pkg/api/hc-service/load-balancer/tcloud.go`）

- [x] 1.1 `TCloudLoadBalancerSpec` 新增 `Exclusive *int64`、`ClusterTag *string`、`CloudClusterIDs []string` 字段（`json` tag 对齐接口文档：`exclusive`/`cluster_tag`/`cloud_cluster_ids`）
- [x] 1.2 `ValidateSpec()` 新增结构校验：`exclusive=1` 时 `sla_type` 必须为空、`cluster_tag`/`cloud_cluster_ids` 不能都为空、`load_balancer_type` 必须为 `OPEN`；`exclusive=0`/不传时 `cluster_tag`/`cloud_cluster_ids` 必须都为空
- [x] 1.3 `ValidateSpec()` 新增 VIP 相关结构校验：未传 `cloud_cluster_ids` 时不允许传 `vip`；传 `vip` 时 `cloud_cluster_ids` 必须恰好一个；传 `vip` 时 `require_count` 必须为 1
- [x] 1.4 `ValidateSpec()` 新增计费结构校验：单线运营商（`vip_isp` 为 CMCC/CUCC/CTCC）必须使用 `BANDWIDTH_PACKAGE`；`internet_charge_type=BANDWIDTH_PACKAGE` 时 `bandwidth_package_id` 必填（若现网已有等价校验，核实后跳过重复实现）
- [x] 1.5 `cloud_cluster_ids` 数量上限校验（不超过 100 个，超过返回 `InvalidParameter`）

## 2. 业务归属校验（cloud-server 层）

- [x] 2.1 在 `cmd/cloud-server/logics/load-balancer/` 新增归属校验辅助函数（如 `exclusive_cluster_check.go`）：按 `bk_biz_id + cluster_type=STGW + cluster_tag` 查本地表校验 `cluster_tag` 归属；按 `bk_biz_id + cluster_type=TGW + cloud_id in cloud_cluster_ids` 查本地表校验每个四层 ID 归属，任一不命中返回 `PermissionDenied`
- [x] 2.2 `ApplicationOfCreateTCloudLB.CheckReq()`（`cmd/cloud-server/service/application/handlers/load_balancer/tcloud/check.go`）在 `req.Validate(true)` 通过后调用 2.1 的归属校验（仅 `exclusive=1` 时触发）
- [x] 2.3 归属校验响应确保不泄露其它业务的集群清单（`PermissionDenied` 错误信息不回显查询到的其它业务集群数据）

## 3. 带宽包出口一致性校验（提单时）

- [x] 3.1 新增出口比对辅助函数：给定 `cloud_cluster_ids`（查 TGW 本地表 `egress` 去重集合 `E4_set`）与 `cluster_tag`（查当前业务已分配 STGW 本地表 `egress` 去重集合 `E7_set`），按「只四层/只七层/两层交集」规则返回允许出口集合
- [x] 3.2 `CheckReq()` 中，`internet_charge_type=BANDWIDTH_PACKAGE` 且 `exclusive=1` 时，调用 `Client.HCService().TCloud.BandwidthPackage.ListBandwidthPackage` 按 `bandwidth_package_id` 实时查 `Egress`，与 3.1 的允许集合比对，不通过返回 `InvalidParameter`
- [x] 3.3 查不到带宽包信息（返回 0 条）时按 `InvalidParameter` 处理

## 4. 下云透传与下云前完整复核（hc-service）

- [x] 4.1 `BatchCreateTCloudClb`（`cmd/hc-service/service/load-balancer/tcloud.go`）组装 `createOpt` 时新增：`createOpt.ClusterIds = req.CloudClusterIDs` 转换、`createOpt.ClusterTag = req.ClusterTag`；确认 `Exclusive` 不写入 `createOpt`（不下传云侧）
- [x] 4.2 在 `pkg/adaptor/tcloud/clb.go` 新增 `DescribeClusterResources` adaptor 方法（映射腾讯云 CLB `DescribeClusterResources` API），入参含集群云上 ID，出参含闲置 VIP 列表；对应在 `pkg/adaptor/types/load-balancer/tcloud.go` 新增 `TCloudDescribeClusterResourcesOption`/结果类型
- [x] 4.3 在 `cmd/hc-service/service/load-balancer/tcloud.go` 新增一个内部测试用 HTTP 端点（如 `POST /vendors/tcloud/load_balancers/exclusive_clusters/idle_vips/query`），薄封装调用 4.2 的 adaptor 方法并注册路由；仅供联调/测试直接验证闲置 VIP 数据，不经 cloud-server/web-server 转发，不需要额外鉴权改造
- [x] 4.4 `BatchCreateTCloudClb` 调用云创建前，若 `req.Exclusive=1`：**重新执行归属校验**（判定逻辑与 2.1 一致，按 `cluster_tag`/`cloud_cluster_ids` 再查一次本地表；因 hc-service 与 cloud-server 是不同服务、不能直接跨包调用 2.1 函数，需在 hc-service 层新增一份等价实现，通过 `Client.DataService().Global.LoadBalancer.ListExclusiveCluster` 查询，与既有 `cs.DataService().Global.*` 用法一致），不通过返回 `PermissionDenied`、不调用创建——覆盖「提单后审批等待期间集群被重新分配」的场景
- [x] 4.5 `BatchCreateTCloudClb` 调用云创建前，若 `req.Exclusive=1` 且指定了 `vip`：调用 4.2 的方法按 `cluster-id` + `vip` + `idle=True` 过滤，复核该 VIP 仍在 `CloudClusterIDs[0]` 内闲置（云上 idle 过滤值为 `True`/`False`），无结果则返回 `InvalidParameter`、不调用创建
- [x] 4.6 `BatchCreateTCloudClb` 调用云创建前，若 `req.Exclusive=1` 且计费为 `BANDWIDTH_PACKAGE`：重新执行第 3 节的出口一致性校验（复用 3.1/3.2 的辅助函数），不通过则返回 `InvalidParameter`、不调用创建
- [x] 4.7 确认审批流程中 `content`（`GenerateApplicationContent`）已经通过 `TCloudLoadBalancerCreateReq` 内联结构自动带上新增字段，无需额外改动（`cmd/cloud-server/service/application/handlers/load_balancer/tcloud/prepare.go`）

## 5. 申请单详情独占集群回显

- [x] 5.1 在 `cmd/cloud-server/service/application/` 新增 `content` 富化辅助函数：解析 `content` JSON，`type=create_load_balancer` 且 `exclusive=1` 时，按 `cluster_tag`/`cloud_cluster_ids` 查本地独占集群表（复用 2.1 的查询能力或新增只读查询），拼出 `clusters` 数组注入 `content`
- [x] 5.2 `buildApplicationGetResp`（`cmd/cloud-server/service/application/get.go`）在 `RemoveSenseField` 之后调用 5.1，非独占型或非 `create_load_balancer` 类型直接跳过
- [x] 5.3 处理本地表未同步/已删除的集群：`cluster_id`/`cluster_name` 填空字符串，`cloud_cluster_id` 保留原值；七层未指定具体集群或四层随机分配时对应字段同样置空

## 6. 负载均衡详情独占集群回显

- [x] 6.1 `pkg/api/core/cloud/load-balancer/tcloud.go` 的 `TCloudClbExtension` 新增 `Exclusive *bool`、`Clusters []TCloudExtensionCluster`（新增该类型，字段：`CloudClusterID`/`ClusterID`/`ClusterName`/`ClusterTag`/`ClusterType`）
- [x] 6.2 `cmd/hc-service/logics/res-sync/tcloud/load_balancer.go` 的 `convertTCloudExtension` 新增：读取 `cloud.ClusterTag`（7层独占标签）与 `cloud.ClusterIds`（集群 ID 数组，**不是** `cloud.ExclusiveCluster`——该字段是内网独占集群，与本单公网场景无关），任一非空时设置 `Extension.Exclusive=true` 并收集 `cloud.ClusterIds` 里的云上集群 ID 待反查
- [x] 6.3 在同步批次层面（`ListLoadBalancer` 同步入口）新增一次批量查询本地独占集群表（按去重后的云上集群 ID 集合），构建 `cloud_id -> 本地记录` 映射；本地记录的 `cluster_type` 才是该 ID 的真实层级（**不假设** `cloud.ClusterIds` 里的元素都是 TGW，同一数组可能混有 TGW/STGW 的 ID），传给 6.2 用于填充 `cluster_id`/`cluster_name`/`cluster_tag`/`cluster_type`；若 `cloud.ClusterTag` 非空但反查结果中没有对应 STGW 记录（云侧只回传标签、未回传具体落地 ID），额外补一条只有 `cluster_tag`/`cluster_type=STGW` 有值的元素
- [x] 6.4 资源视角 `GetLoadBalancer`（`cmd/cloud-server/service/load-balancer/query.go`）与业务视角 `GetLoadBalancer` 确认 `extension` 序列化自动带上新增字段，无需额外改动；核实响应结构体（如有独立的 `LoadBalancerGetResp` 未直接复用 core 类型）需要补充字段透传

## 7. 单元测试

- [x] 7.1 `TCloudLoadBalancerSpec.ValidateSpec` 结构校验单测（覆盖 AC-002~AC-009、AC-018、AC-019）
- [x] 7.2 归属校验单测（覆盖 AC-010、AC-S01）
- [x] 7.3 出口一致性校验单测（覆盖 AC-014、AC-020~AC-028，含边界：交集为空、多出口）
- [x] 7.4 下云透传与下云前复核单测（覆盖 AC-012、AC-013、AC-017），VIP 闲置查询走 mock
- [x] 7.5 存量兼容性回归单测（覆盖 AC-001、AC-015、AC-016）
- [x] 7.6 申请单详情 `clusters` 富化单测（独占型/非独占型/本地表缺失场景）
- [x] 7.7 负载均衡详情同步 `extension.exclusive`/`clusters` 单测（含本地表关联缺失场景）

## 8. 文档与需求单同步

- [x] 8.1 修正 `create_application_for_create_tcloud_load_balancer.md`/`sys_create_application_for_create_tcloud_load_balancer.md` 中「`cluster_tag` 必填」的措辞，改为与已批准需求文档一致的「`cluster_tag`/`cloud_cluster_ids` 至少一个非空」
- [x] 8.2 更新 `docs/reqs/业务视角购买.md` 的「本期不包含」「假设 5」「Q-001」为「本期已包含」，并按扩大后的范围重新核对工时/规模评分（详情回显部分未计入原 17.5h/size=3 评估；已补充第 5 轮澄清记录，重估为 24h/size=5）
- [ ] 8.3 性能验收测量：AC-P01（提单结构+归属校验 P99<200ms）、AC-P02（下云前 VIP 闲置+出口比对额外耗时 P99<3s）在测试环境执行并记录结果 —— **阻塞**：需实际测试环境（真实 DB + 已同步的独占集群分配数据 + 可用腾讯云测试账号）才能压测，当前开发环境无法执行，需人工在测试环境补测
