## ADDED Requirements

### Requirement: 独占型入参接收与规格互斥校验
业务视角申请单（`POST /api/v1/cloud/vendors/tcloud/applications/types/create_load_balancer`）与系统提单（`POST /api/v1/cloud/vendors/tcloud/system/applications/types/create_load_balancer`）SHALL 接收 `exclusive`（0/1，缺省 0）、`cluster_tag`（七层独占集群标签）、`cloud_cluster_ids`（四层独占集群云上 ID 数组）三个新增入参，并对以下互斥关系做结构校验，不合法时返回 `InvalidParameter` 且不创建申请单：
- `exclusive=1` 时 `sla_type` 必须为空，且 `cluster_tag`/`cloud_cluster_ids` 不能都为空
- `exclusive=0` 或不传时，`cluster_tag`/`cloud_cluster_ids` 必须都为空
- `exclusive=1` 时 `load_balancer_type` 必须为 `OPEN`（公网），内网独占型不支持

#### Scenario: 只传四层集群的独占型请求通过结构校验
- **WHEN** 业务已分配四层集群 `tgw-38feq8c6`，提单 `exclusive=1`、`cloud_cluster_ids=["tgw-38feq8c6"]`、`cluster_tag` 不传、`sla_type` 为空、`load_balancer_type=OPEN`
- **THEN** 结构校验通过，进入归属校验环节

#### Scenario: 只传七层标签的独占型请求通过结构校验
- **WHEN** 业务已分配七层标签 `ziyan-serven`，提单 `exclusive=1`、`cluster_tag="ziyan-serven"`、`cloud_cluster_ids` 不传
- **THEN** 结构校验通过，进入归属校验环节

#### Scenario: 独占型但两个集群标识都为空
- **WHEN** 提单 `exclusive=1`，`cluster_tag` 与 `cloud_cluster_ids` 都为空
- **THEN** 返回 `InvalidParameter`，不创建申请单

#### Scenario: 非独占型却传了集群标识
- **WHEN** 提单 `exclusive=0` 或不传，同时传了 `cluster_tag` 或 `cloud_cluster_ids`
- **THEN** 返回 `InvalidParameter`，不创建申请单

#### Scenario: 独占型请求为内网负载均衡
- **WHEN** 提单 `exclusive=1` 且 `load_balancer_type=INTERNAL`
- **THEN** 返回 `InvalidParameter`

#### Scenario: 独占型请求同时指定性能容量档位
- **WHEN** 提单 `exclusive=1` 且 `sla_type="clb.c2.medium"`
- **THEN** 返回 `InvalidParameter`

#### Scenario: 存量共享型请求不受影响
- **WHEN** 提单不传 `exclusive`、不传独占字段（存量请求格式）
- **THEN** 申请单创建成功，行为与本变更上线前一致

### Requirement: 指定 VIP 的结构约束
指定 `vip` 时，SHALL 要求：`cloud_cluster_ids` 已传且必须恰好一个（不允许零个或多个）；未传 `cloud_cluster_ids`（仅七层场景）时不允许传 `vip`；指定 `vip` 时 `require_count` 必须为 1。不满足时返回 `InvalidParameter`。

#### Scenario: 仅七层场景指定 VIP
- **WHEN** 只传七层 `cluster_tag`、不传 `cloud_cluster_ids`，同时传了 `vip`
- **THEN** 返回 `InvalidParameter`

#### Scenario: 指定 VIP 但四层集群 ID 不唯一
- **WHEN** 传 `vip="1.1.1.1"`，`cloud_cluster_ids` 为 0 个或 2 个以上
- **THEN** 返回 `InvalidParameter`

#### Scenario: 指定 VIP 但购买数量大于 1
- **WHEN** 传 `vip` 且恰好一个 `cloud_cluster_ids`，但 `require_count=2`
- **THEN** 返回 `InvalidParameter`

#### Scenario: 多个四层集群 ID 表示随机分配
- **WHEN** 业务已分配两个四层集群，提单 `exclusive=1`、`cloud_cluster_ids` 传两个 ID、不传 `vip`
- **THEN** 申请单创建成功（云侧在列表内随机挑选，本次不做出口比对因为未使用共享带宽包）

### Requirement: 计费方式结构校验
独占型请求 SHALL 校验：运营商为单线（`vip_isp` 为 CMCC/CUCC/CTCC）时 `internet_charge_type` 必须为 `BANDWIDTH_PACKAGE`；`internet_charge_type=BANDWIDTH_PACKAGE` 时 `bandwidth_package_id` 必填。不满足时返回 `InvalidParameter`。

#### Scenario: 单线运营商未使用共享带宽包计费
- **WHEN** `vip_isp=CMCC` 且 `internet_charge_type` 不是 `BANDWIDTH_PACKAGE`
- **THEN** 返回 `InvalidParameter`

#### Scenario: 共享带宽包计费但未传带宽包 ID
- **WHEN** `internet_charge_type=BANDWIDTH_PACKAGE` 但不传 `bandwidth_package_id`
- **THEN** 返回 `InvalidParameter`

### Requirement: 独占集群业务归属校验
提单时 SHALL 校验 `cluster_tag` 与 `cloud_cluster_ids` 是否属于当前 `bk_biz_id` 已分配的独占集群：`cluster_tag` 非空时必须能在当前业务已分配的七层（STGW）集群中命中该标签；`cloud_cluster_ids` 中每一个都必须是当前业务已分配的四层（TGW）集群云上 ID。命中失败 SHALL 返回 `PermissionDenied`，且响应不包含其它业务的集群清单。

#### Scenario: 集群标签或 ID 不属于当前业务
- **WHEN** `cluster_tag` 或某个 `cloud_cluster_ids` 不属于当前 `bk_biz_id`
- **THEN** 返回 `PermissionDenied`，申请单不创建

#### Scenario: 越权访问不泄露其它业务集群信息
- **WHEN** 业务 A 的申请人传入业务 B 已分配集群的 `cloud_cluster_ids` 或 `cluster_tag`
- **THEN** 返回 `PermissionDenied`，响应中不包含业务 B 的集群列表

### Requirement: 共享带宽包出口一致性校验
计费方式为 `BANDWIDTH_PACKAGE` 时，SHALL 校验所选带宽包出口（`Ep`，实时查云得到）落在「本次可能分配到的集群」出口范围内，出口数据一律读本地独占集群表：
- 只选四层：`Ep` 必须属于 `cloud_cluster_ids` 对应 TGW 的出口去重集合（`E4_set`）
- 只选七层：`Ep` 必须属于当前业务已分配、该标签下全部 STGW 的出口去重集合（`E7_set`）
- 四层+七层都选：`Ep` 必须属于 `E4_set ∩ E7_set`（交集为空则任何带宽包都不通过）
不满足时返回 `InvalidParameter`；前端已做的过滤不能替代该后端校验；`cloud_cluster_ids` 不按出口裁剪，原样透传。

#### Scenario: 只四层且带宽包出口不匹配
- **WHEN** 计费为共享带宽包，只传四层 `cloud_cluster_ids`，带宽包出口不在这些 TGW 的出口集合内
- **THEN** 返回 `InvalidParameter`，且不从列表中删除任何 ID 后重试

#### Scenario: 只四层且带宽包出口匹配多出口集合中的一个
- **WHEN** 计费为共享带宽包，只传多个四层 ID，`E4_set={center_egress1, center_egress2}`，带宽包出口为 `center_egress1`
- **THEN** 申请单创建成功

#### Scenario: 只七层且带宽包出口属于标签出口集合
- **WHEN** 计费为共享带宽包，只传七层 `cluster_tag`，该标签下已分配 STGW 有出口 `{center_egress1, center_egress2}`，带宽包出口为 `center_egress1`
- **THEN** 申请单创建成功

#### Scenario: 只七层且带宽包出口不在标签出口集合内
- **WHEN** 计费为共享带宽包，只传七层 `cluster_tag`，带宽包出口不在该标签下已分配 STGW 的出口集合内
- **THEN** 返回 `InvalidParameter`

#### Scenario: 四层七层都选且带宽包出口落在交集内
- **WHEN** 计费为共享带宽包，同时传四层与七层，`E4_set={center_egress1, center_egress2}`，`E7_set={center_egress1, center_egress3}`，带宽包出口为 `center_egress1`
- **THEN** 申请单创建成功，且下云时 `cloud_cluster_ids` 仍全部下传、不按出口裁剪

#### Scenario: 四层七层都选但带宽包出口只属于四层出口集合
- **WHEN** 计费为共享带宽包，同时传四层与七层，带宽包出口属于 `E4_set` 但不属于 `E7_set`
- **THEN** 返回 `InvalidParameter`，不创建申请单

#### Scenario: 四层七层都选但带宽包出口只属于七层出口集合
- **WHEN** 计费为共享带宽包，同时传四层与七层，带宽包出口属于 `E7_set` 但不属于 `E4_set`
- **THEN** 返回 `InvalidParameter`

#### Scenario: 四层七层出口集合交集为空
- **WHEN** 计费为共享带宽包，同时传四层与七层，`E4_set ∩ E7_set` 为空
- **THEN** 返回 `InvalidParameter`

#### Scenario: 端到端合法组合可提单并下云
- **WHEN** 计费为共享带宽包，同时传四层 `cloud_cluster_ids` 与七层 `cluster_tag`，带宽包出口属于 `E4_set ∩ E7_set`，指定 VIP 仍闲置
- **THEN** 提单成功，审批通过后下云创建调用发出

### Requirement: hc-service 内部测试端点透出集群闲置资源查询
hc-service SHALL 新增一个内部 HTTP 端点，薄封装调用新增的 `DescribeClusterResources` adaptor 方法，用于联调/测试阶段独立验证闲置 VIP 判断逻辑，不依赖跑完整的提单+审批+下云链路。该端点仅注册在 hc-service（服务间调用层），不经 cloud-server/web-server 转发对外暴露，不做面向业务用户的鉴权与参数封装；对外的购买页闲置 VIP 列表查询（面向业务用户）由兄弟单 1069995598138281929 负责，与本端点是不同的路径与职责层级。

#### Scenario: 联调时直接调用 hc-service 端点验证闲置 VIP
- **WHEN** 测试人员直接调用 hc-service 新增的内部端点，传入某四层集群云上 ID
- **THEN** 返回该集群当前的闲置 VIP 列表，不需要先创建、提交、审批任何申请单

### Requirement: 系统提单与业务视角申请单校验一致
系统提单接口 SHALL 与业务视角申请单使用同一套结构、归属、出口校验规则，唯一差异是申请人由请求体中的 `applicant` 指定而非当前登录用户。

#### Scenario: 系统提单复用相同校验规则
- **WHEN** 系统提单接口传入与业务视角申请单相同的合法独占字段组合及合法 `applicant`
- **THEN** 申请单创建成功，校验结果与业务视角申请单一致

### Requirement: 审批通过后下云透传
审批通过后交付时，SHALL 将 `cloud_cluster_ids` 原样作为云侧四层 `ClusterIds`、`cluster_tag` 原样作为云侧七层 `ClusterTag`、`vip`（如有值）透传给腾讯云 `CreateLoadBalancer`；`exclusive` 字段 SHALL NOT 下传云侧，仅用于本地区分独占型与共享/性能容量型。

#### Scenario: 下云请求包含集群参数但不含 exclusive
- **WHEN** 申请单审批通过，`content` 含 `cloud_cluster_ids` 与 `cluster_tag`
- **THEN** 调用云创建时带上对应四层 ID 列表与七层标签，且请求体不含 `exclusive` 字段

### Requirement: 下云前完整复核
下云前（真正调用云创建之前，`exclusive=1` 时）SHALL 完整重跑一遍提单时的校验，并额外增加一项只在下云前才有意义的检查；任一失败则交付失败（`InvalidParameter`/`PermissionDenied`），不调用云创建：
- 重新校验 `cluster_tag`/`cloud_cluster_ids` 是否仍属于当前 `bk_biz_id` 已分配的独占集群（规则同「独占集群业务归属校验」），覆盖「提单后审批等待期间集群被重新分配给其它业务」的场景，不通过返回 `PermissionDenied`
- 计费方式为 `BANDWIDTH_PACKAGE` 时，重新校验带宽包出口与本次可能分配集群出口的一致性（规则同「共享带宽包出口一致性校验」）
- 指定了 `vip` 时，复核该 VIP 在对应四层集群的闲置列表内仍然闲置（这一项在提单阶段不做，因为审批等待期间闲置状态几乎必然变化，提单时查询没有意义）
未使用共享带宽包时不做出口比对；未指定 `vip` 时不做闲置复核。

#### Scenario: 归属在审批等待期间被改变
- **WHEN** 提单时 `cloud_cluster_ids` 归属当前业务校验通过，审批通过后下云前该集群已被重新分配给其它业务
- **THEN** 不调用云创建，申请单进入交付失败状态而非静默成功，返回 `PermissionDenied`

#### Scenario: 指定 VIP 在提单时闲置、下云前已被占用
- **WHEN** 指定 VIP 在提单时闲置、交付时已被占用
- **THEN** 不调用云创建，申请单进入交付失败状态而非静默成功，返回 `InvalidParameter`

#### Scenario: 指定 VIP 复核仍闲置可正常下云
- **WHEN** `exclusive=1`、恰好一个已分配四层 ID、该集群下 VIP 闲置、`require_count=1`，提单并指定该 VIP
- **THEN** 申请单创建成功，审批通过后下云前复核归属仍有效、仍闲置，创建调用发出
