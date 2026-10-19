# API — 独占集群管理页面

> 本轮只覆盖列表 + 分配。契约**已由后端确认**：[bk-hcm#2078](https://github.com/TencentBlueKing/bk-hcm/pull/2078)（`docs/api-docs/web-server/docs/resource/load-balancer/list_exclusive_cluster.md`、`assign_exclusive_cluster_to_biz.md`）。下表已按该文档更新，前端不再使用占位数据。
> 资源路径挂在负载均衡下：`load_balancers/exclusive_clusters`。未分配业务：`bk_biz_id = -1`（与现网资源一致）。

## 1. 查询独占集群列表

`POST /api/v1/cloud/load_balancers/exclusive_clusters/list`

### 请求

与现网 `certs/list` 相同：必传 `filter` + `page`。

| 参数 | 类型 | 必选 | 说明 |
|------|------|------|------|
| filter.op | `and` / `or` | 是 | 多规则关系 |
| filter.rules | array | 是 | 最多 5 条；空数组表示不过滤 |
| page.start | int | 是 | 偏移 |
| page.limit | int | 是 | 页大小 |
| page.count | bool | 否 | `true` 时只计数 |

#### 筛选字段（对应 PRD）

| field | 操作符 | 说明 |
|-------|--------|------|
| account_id | eq | 左侧云账号树当前账号 |
| bk_biz_id | eq / in | 业务；未分配用 `-1` |
| cluster_tag | eq / in | 集群标签 |
| cluster_type | eq / in | 本页按 PRD 只筛 `TGW` / `STGW` |
| zone | eq / in | 可用区 |
| isp | eq / in | 运营商 |
| id 或 cloud_id | eq / in / cs | 集群 ID |
| name | cs / cis | 集群名称 |

### 请求示例

```json
{
  "filter": {
    "op": "and",
    "rules": [
      { "field": "account_id", "op": "eq", "value": "<ACCOUNT_ID>" },
      { "field": "cluster_type", "op": "in", "value": ["TGW"] }
    ]
  },
  "page": { "start": 0, "limit": 20, "count": false }
}
```

### 列表项字段（对应表格列）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | HCM 资源 ID |
| cloud_id | string | 云上集群 ID（列「集群 ID」） |
| name | string | 集群名称 |
| cluster_type | string | 本页展示/筛选只认 `TGW` / `STGW`（PRD） |
| cluster_tag | string | 集群标签；空串表示未打标签 |
| zone | string | 可用区 |
| isp | string | 运营商（`BGP` / `CMCC` / `CUCC` / `CTCC` / `INTERNAL`） |
| max_conn | int \| null | 最大连接数；STGW 云上可能无值 → 前端展示 `--`（后端文档写 `—`，前端按 HCM 空值规范用 `--`） |
| clb_resource_count | int | 集群内已绑定 CLB 数量，对应表格列「集群内实例数」 |
| bk_biz_id | int | 业务；`-1` 未分配。响应**不含**业务名，前端用业务映射展示 |
| account_id | string | 所属账号 |
| region / network / egress / ip_version / max_conn / extension | - | 后端返回但本轮表格不取（PRD 只有 9 列） |

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "count": 2,
    "details": [
      {
        "id": "00000001",
        "cloud_id": "cluster-xxx",
        "name": "example-tgw",
        "cluster_type": "TGW",
        "cluster_tag": "tag-a",
        "zone": "ap-guangzhou-1",
        "isp": "BGP",
        "max_conn": 100000,
        "clb_resource_count": 3,
        "bk_biz_id": -1,
        "account_id": "<ACCOUNT_ID>"
      }
    ]
  }
}
```

## 2. 分配到业务（单条 / 批量同一接口）

`POST /api/v1/cloud/load_balancers/exclusive_clusters/assign/bizs`

| 参数 | 类型 | 必选 | 说明 |
|------|------|------|------|
| cluster_ids | string[] | 是 | 未分配集群的本地 ID，最多 100 个 |
| bk_biz_id | int | 是 | 目标业务，必须大于 0 |

### 请求示例

```json
{
  "cluster_ids": ["00000001", "00000002"],
  "bk_biz_id": 3
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok"
}
```

### 业务规则（前端依赖）

- 已分配（`bk_biz_id != -1`）的行不可再提交；列表中混有已分配集群时后端**整批拒绝**并返回 `InvalidParameter`。
- 失败时列表不乐观更新。
- 分配会记录审计。
- 不提供解绑/再次分配接口（本轮不做）。

## 3. 复用现网接口（无新契约）

| 用途 | 现网 |
|------|------|
| 左侧账号树 | 现网资源接入账号树，无新接口 |
| 分配弹窗业务单选 | 现网账号可用业务列表（与证书/安全组分配相同） |

## 4. 错误码（期望）

| 场景 | 前端处理 |
|------|----------|
| `code != 0` 列表失败 | http 全局提示；入口把 list 置空，表格保持默认空态。不要用页面级 `bk-exception` 盖住表格 |
| 分配失败（含已分配再分配 `InvalidParameter`） | 弹窗提示后端 `message`，列表保持原值 |

## 5. 缺口

- 后端接口实现是否已部署到联调环境待确认；契约本身已固定，前端直连真实接口。
- 不在本轮：购买独占型（`list_exclusive_cluster_tags` / `list_exclusive_cluster_idle_vips` 属兄弟单）、单据/CLB 列表/详情字段所用的 CLB 实例规格查询。
