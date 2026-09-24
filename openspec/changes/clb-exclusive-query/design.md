## Context

列表文档落在带删除保护的双视角接口（已走 `ListLoadBalancerRaw`）。详情文档要求 `extension.clusters` 与独占集群表联查，后端只回原始值。同步侧已把三个云字段写入 `extension`。申请单详情、购买入参不在本期。

## Goals / Non-Goals

**Goals:**

- 列表顶层回 `exclusive`、`sla_type`，并支持 `json_eq` 筛选
- 详情用云上集群 ID 反查本地独占集群表，派生 `clusters`

**Non-Goals:**

- CLB 同步与 DIFF
- 购买申请单、申请单据详情
- 普通 `POST /load_balancers/list`（不带删除保护）
- 按 `exclusive` / `sla_type` 排序

## Decisions

### D1：列表走已有 Raw 接口，筛选靠 DAO 白名单

**选择**：改 `listLoadBalancerWithDeleteProtect`，顶层填 `exclusive`/`sla_type`。`columnTypes` 增加 `extension.exclusive`（Numeric）和 `extension.sla_type`（String）。

**理由**：文档字段已经写在带删除保护的列表上。`json_eq` 算子已有，缺的是带点号路径白名单。

**备选**：改普通 `list_load_balancer` —— 文档没改，且不取 Raw。

### D2：`clusters` 只在详情 GET 派生，不落库

**选择**：`TCloudClbExtension.Clusters` 仅响应使用。`exclusive != 1` 回空数组；按 `cluster_ids` + `account_id` 查独占集群表；未命中只回云上 ID；`cluster_tag` 非空且没有 STGW 时补一条七层元素。

**理由**：`cluster_id`/`cluster_name` 随集群同步变，不该冻在 CLB 行里。`cluster_tag` 只表示七层。

## Risks / Trade-offs

- **[风险] 独占集群未同步到本地** → 详情 `cluster_id`/`cluster_name` 为空，`cloud_cluster_id` 仍回。
- **[风险] 云上七层不给 STGW 集群 ID** → 只回 `cluster_tag` + `cluster_type=STGW`。
- **[Trade-off] 筛的是 `extension` JSON 路径** → 前端必须传 `extension.exclusive` / `extension.sla_type` + `json_eq`。

## Migration Plan

1. 发布 cloud-server / data-service；未同步前 `exclusive=0`、`clusters=[]`。
2. 无 DDL。

**回滚**：回滚二进制。筛选白名单回滚后 `json_eq` 再次被拒。

## Open Questions

（无）
