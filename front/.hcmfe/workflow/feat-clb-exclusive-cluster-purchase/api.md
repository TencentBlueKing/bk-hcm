# API — CLB 独占集群购买、申请单与资源详情

## 1. 契约来源与边界

- 后端文档 PR：`https://github.com/TencentBlueKing/bk-hcm/pull/2078`。
- 原计划指定提交：`81ec1c2e0ee2d753eb5cbfe58418106eff6020e2`；进入 API 阶段时 PR HEAD 为 `dc62b09fc7f2790caef5e73ef85fb6a53e2ca27c`，本次根据后端补充内容再次复核至 `8366ea9d419c5e1202c88168984447ea52fb31bb`。本文以最新已读文档及用户确认的后端口径为准，并记录仍与文档不一致的地方。
- 本迭代仅调用既有/新增接口，不修改后端。接口路径均为相对路径，不固化环境主机。
- 当前前端已有 `ApplyClbModel`、购买侧滑表单、`BandwidthPackageSelector`、申请单 `content` 解析和 CLB 详情请求；在这些入口扩展字段。

## 2. 独占集群标签聚合

`POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/exclusive_clusters/tags/list`

| 请求项 | 类型 | 说明 |
|--------|------|------|
| 路径 `bk_biz_id` | number | 当前业务，接口仅返回已分配给该业务的公网独占集群 |
| `account_id` | string | 云账号 |
| `region` | string | 地域 |
| `isp` | `CMCC`/`CUCC`/`CTCC` 枚举 | 运营商：**仅支持三网直连**；BGP（含自研云 BGP 系取值）等其它类型不支持独占集群，前端据此不展示「独占型」 |
| `zones` | string[] | 当前选中可用区；单可用区传单元素数组 |
| `cluster_type` | `TGW`/`STGW`/空 | 空值同时查询四层与七层 |

响应使用 `data.details[]`，每项包含 `cluster_tag`、`cluster_type`、`clusters[]`。集群项包含 `cloud_cluster_id`、`cluster_id`、`cluster_name`、`egress`、`isp`、`zone`。空 `details` 表示当前条件下不可选择独占型；查询失败时不得提交独占申请。

前端在账号、地域、可用区、运营商变化时重查，异步结果仅可回写到与发起请求时相同的表单条件。四层/七层下拉按 `cluster_type` 分组；未启用的层保留已选值但不参与后续计算。

### 文档差异

PR 的参数表写 `zone: string`，示例与正文写 `zones`。用户已确认继续使用项目现有 `zones` 字段，单可用区传一个元素的数组。PR 的 STGW 响应示例已包含 `clusters[].egress`，但说明表仍写“七层 clusters 固定为空数组”。本迭代按用户确认口径：七层标签的出口取同 `cluster_tag` 的 TGW 集群集合；找不到则有效出口为空，不假定 STGW 的 `clusters` 可稳定提供出口。

## 3. 四层空闲 VIP

`POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/exclusive_clusters/idle_vips/list`

请求体：`account_id`、`region`、选中 TGW 集群的 `cloud_cluster_id`。响应：`data.count` 与 `data.details: string[]`。仅选择具体四层集群时查询；随机集群不查询。改选集群后清空旧 VIP 并重新查询。`count=0` 或 `details` 为空时该具体集群不能成为有效购买配置。指定的 VIP 仍由后端在创建时复核，若已被占用，以后端错误为准提示用户重新选择。

## 4. 购买申请

`POST /api/v1/cloud/vendors/tcloud/applications/types/create_load_balancer`

沿用原购买请求中的业务、账号、地域、公网类型、`zones`、运营商、网络计费等字段。独占规格相关字段如下：

| 字段 | 类型 | 本迭代取值 |
|------|------|------------|
| `exclusive` | 0/1 | 独占型传 1；共享/性能容量型传 0 或不传 |
| `sla_type` | string | 独占型、共享型传空字符串；性能容量型传现有档位值 |
| `cloud_cluster_ids` | string[] | 启用四层且指定集群时传单个云上 ID；随机集群时传该标签下全部候选云上 ID；未启用四层不传 |
| `cluster_tag` | string | 启用七层时传所选七层标签；未启用七层不传 |
| `vip` | string | 四层指定 IP 时传该 IP；随机 IP 传空字符串；未启用四层不传或传空字符串 |
| `bandwidth_package_id` | string | 仅选择共享带宽包计费且已选有效包时传 |

业务规则：

1. 独占型仅在业务视角、公网且至少启用四层或七层一层时提交。
2. `exclusive=1` 时 `cloud_cluster_ids` 与 `cluster_tag` 至少一个非空，可同时传；`sla_type` 必须为空。
3. 未勾选层即使保留了前端状态，也不得输出该层请求字段。
4. 四层随机集群传候选 ID 数组；用户已确认具体集群 + 随机 IP 时传 `vip: ""`。
5. 非独占型不得携带 `cluster_tag`、`cloud_cluster_ids` 等独占参数。
6. 带宽包出口不匹配、标签未分配给业务、参数组合非法或指定 VIP 已占用时，后端可能返回 `InvalidParameter` / `PermissionDenied`；前端保持表单并展示错误，不假设创建成功。

### 文档差异

PR 的文字写 `exclusive=1` 时 `cluster_tag` 必填，但同 PR 的四层独占示例不传 `cluster_tag`。用户已确认最终业务规则是 `cluster_tag` 与 `cloud_cluster_ids` 至少一个非空：四层仅购买不传七层标签。PR 空闲 VIP 文档又写随机集群“不传 cloud_cluster_ids”，与购买文档及用户确认的“传标签下全部候选 ID”不一致；本迭代按购买申请示例及用户确认规则实施，并在联调验证随机四层场景。

## 5. 共享带宽包查询及出口过滤

沿用现有 `/api/v1/cloud/{业务前缀}bandwidth_packages/query`，业务视角实际路径为 `POST /api/v1/cloud/bizs/{bk_biz_id}/bandwidth_packages/query`。请求仍含 `account_id`、`region`、现有 `network_types` 及 `page`；响应为 `data.total_count`、`data.packages[]`，每个带宽包包含 `egress`、`status` 等字段。

选择任意四层/七层标签后，按用户确定方案全量拉取所有分页，再在前端以有效出口集合过滤；不依赖下拉滚动才能看到后续页。项目已有 `@blueking/roll-request`，本接口分页键为 `offset` / `limit`、总数字段为 `total_count`、列表字段为 `packages`，因此使用对应配置和 getter 完成全量读取。保留现有账号、地域、运营商服务端条件，以及带宽包状态、上海/南昌可用区限制。

有效出口集合：

- 四层指定集群：该集群 `egress` 单元素集合。
- 四层随机集群：同标签全部 TGW 候选集群的 `egress` 去重集合。
- 仅七层：同七层标签的 TGW 集群 `egress` 去重集合；没有可解析出口时为空集合。
- 四层和七层同时启用：以上两组集合的交集。
- 未勾选层完全不参与计算。

勾选状态、标签、集群或出口集合变化后立即清空旧 `bandwidth_package_id`，全量请求重新开始；仅当前条件对应的最新请求可更新列表。出口集合为空、查询失败、无匹配包或未选有效包时，共享带宽包计费模式不得提交。切回非独占型则恢复原选择器行为。

PR 原指定提交曾为带宽包查询增加服务端 `egresses` 过滤说明；最新 PR 文件清单已不包含该文档变更。故本次不以服务端 `egresses` 为必要条件，按用户明确要求全量拉取并在前端过滤。

## 6. 申请单详情

业务视角：`GET /api/v1/cloud/bizs/{bk_biz_id}/applications/{application_id}`；资源视角：`GET /api/v1/cloud/applications/{application_id}`。两份 PR 文档均说明：当申请类型为 `create_load_balancer` 时，响应中的 `content` 是 JSON 字符串，解析后在原购买字段之外新增 `exclusive`、`cluster_tag`、`cloud_cluster_ids`、`clusters`；`clusters[]` 的元素形态与负载均衡详情 `extension.clusters[]` 相同。

| 回显项 | `content` 来源 | 规则 |
|--------|----------------|------|
| 独占区块是否显示 | `exclusive` | 为 1 时显示；否则隐藏整个区块 |
| 四层集群标签 | `clusters[]` 中 `cluster_type=TGW` 的 `cluster_tag` | 缺失时为空值 |
| 四层集群名称 | 同一 TGW 元素的 `cluster_name` | 随机分配或本地未同步时可能为空值 |
| 四层集群 IP | `vip` | 指定 IP 回显原申请值；随机或空字符串时为空值 |
| 七层集群标签 | `clusters[]` 中 `cluster_type=STGW` 的 `cluster_tag` | 缺失时为空值；`cluster_tag` 是申请原始七层标签，可作为缺少 STGW 元素时的兼容回显 |

四层随机分配时，文档说明 TGW 元素的 `cloud_cluster_id`、`cluster_id`、`cluster_name` 以及 `vip` 可能为空；单据详情保持“申请参数”口径，不查询交付后资源，也不从当前标签列表推断历史申请值。`content` 缺字段或解析失败时不得阻断其它详情内容。资源视角文档的通用响应示例仍采用 `data.details[]`，业务视角示例为 `data` 对象；沿用项目现有单据详情请求层解析各视角响应，不在展示组件中自行假定两种视角有同一外层包装。

## 7. CLB 资源详情

沿用业务或资源视角 `load_balancers/{id}` 详情接口。腾讯云 `data.extension` 新增：

| 字段 | 类型 | 用途 |
|------|------|------|
| `exclusive` | int | 1 为独占型，0 为非独占型 |
| `sla_type` | string | 非独占时，非空表示性能容量档位，空值表示共享型 |
| `clusters` | array | 独占集群实际关联结果 |

`clusters[]` 中通过 `cluster_type=TGW/STGW` 区分四层、七层，并读取 `cluster_tag`、`cluster_name` 等字段；本地未同步时 `cluster_id`、`cluster_name` 可能为空。实例规格优先级为 `exclusive=1` → 独占型，否则 `sla_type` 非空 → 对应性能容量档位，否则共享型。详情中的实际结果可能不同于申请参数，不能用申请单 `content` 覆盖。

经用户与后端确认，CLB 资源详情中的“四层集群 IP”按该 CLB 的负载均衡 VIP 回显，直接复用现有 `getInstVip(props.detail)`；不调用“闲置 VIP 列表”反查已创建资源。该工具函数按 `public_ipv4_addresses` → `public_ipv6_addresses` → `private_ipv4_addresses` → `private_ipv6_addresses` 顺序取第一个非空地址组并用逗号连接，均为空时沿用现有 `--` 占位。可用局部 computed 复用结果，但不重复实现一套 IP 优先级。

## 8. 联调重点

1. 核对 `zones` 数组在标签接口的实际接收行为，以及 TGW/STGW 返回的集群及出口集合。
2. 核对仅四层且 `cluster_tag` 不传、仅七层且 `cloud_cluster_ids` 不传、四七层同时传三种申请组合。
3. 核对四层随机集群传全部候选 ID 与随机 VIP 空字符串。
4. 核对空闲 VIP 被并发占用、标签无权限、带宽包出口不匹配的错误返回与页面提示。
5. 核对带宽包跨多页全量读取、出口交集和旧请求失效。
6. 核对业务/资源视角申请单详情的外层响应形态、`content.clusters` 的 TGW/STGW 数据，以及随机分配时 `vip` 和集群名称为空的回显。
7. 核对 CLB 详情“四层集群 IP”与页面已有负载均衡 VIP 一致，覆盖 IPv4、IPv6、仅内网地址和全空场景。
