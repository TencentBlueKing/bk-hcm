### 描述

- 该接口提供版本：v9.9.9.9+。
- 该接口所需权限：资源-资源查看。
- 该接口功能描述：查询负载均衡独占集群列表。

### URL

POST /api/v1/cloud/load_balancers/exclusive_clusters/list

### 输入参数

| 参数名称   | 参数类型   | 必选 | 描述     |
|--------|--------|----|--------|
| filter | object | 是  | 查询过滤条件 |
| page   | object | 是  | 分页设置   |

#### filter

| 参数名称  | 参数类型        | 必选 | 描述                                                              |
|-------|-------------|----|-----------------------------------------------------------------|
| op    | enum string | 是  | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是  | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选 | 描述                                          |
|-------|-------------|----|---------------------------------------------|
| field | string      | 是  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| op    | enum string | 是  | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis） |
| value | 可变类型        | 是  | 查询条件Value值                                  |

##### rules 表达式说明：

##### 1. 操作符

| 操作符 | 描述                                        | 操作符的value支持的数据类型                              |
|-----|-------------------------------------------|-----------------------------------------------|
| eq  | 等于。不能为空字符串                                | boolean, numeric, string                      |
| neq | 不等。不能为空字符串                                | boolean, numeric, string                      |
| gt  | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte | 大于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lt  | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lte | 小于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in  | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string                      |
| nin | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string                      |
| cs  | 模糊查询，区分大小写                                | string                                        |
| cis | 模糊查询，不区分大小写                               | string                                        |

##### 2. 协议示例

查询 cluster_type 是 "TGW" 且 region 是 "ap-guangzhou" 的数据。

```json
{
  "op": "and",
  "rules": [
    {
      "field": "cluster_type",
      "op": "eq",
      "value": "TGW"
    },
    {
      "field": "region",
      "op": "eq",
      "value": "ap-guangzhou"
    }
  ]
}
```

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

#### 查询参数介绍：

| 参数名称                | 参数类型   | 描述                                     |
|---------------------|--------|----------------------------------------|
| id                  | string | 集群本地ID  |
| cloud_id            | string | 集群云上ID |
| name                | string | 集群名称                                   |
| vendor              | string | 供应商（枚举值：tcloud）                        |
| account_id          | string | 账号ID                                   |
| bk_biz_id           | int64  | 业务ID，-1表示未分配                           |
| region              | string | 地域                                     |
| zone                | string | 可用区                                    |
| cluster_type        | string | 集群类型（枚举值：TGW、STGW）               |
| cluster_tag         | string | 集群标签，空字符串表示未打标签                        |
| network             | string | 网络类型（枚举值：Public、Private）               |
| isp                 | string | 运营商（枚举值：BGP、CMCC、CUCC、CTCC、INTERNAL）   |
| egress              | string | 网络出口，如center_egress1                   |
| ip_version          | string | IP版本                                   |
| max_conn            | int64  | 最大连接数，STGW集群云上可能无值                     |
| clb_resource_count  | int64  | 集群内已绑定的CLB数量                           |
| memo                | string | 备注                                     |
| creator             | string | 创建者                                    |
| reviser             | string | 修改者                                    |
| created_at          | string | 创建时间，标准格式：2006-01-02T15:04:05Z         |
| updated_at          | string | 修改时间，标准格式：2006-01-02T15:04:05Z         |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 获取详细信息请求参数示例

查询 ap-guangzhou 地域下未分配的四层独占集群列表。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "region",
        "op": "eq",
        "value": "ap-guangzhou"
      },
      {
        "field": "cluster_type",
        "op": "eq",
        "value": "TGW"
      },
      {
        "field": "bk_biz_id",
        "op": "eq",
        "value": -1
      }
    ]
  },
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
  }
}
```

#### 获取数量请求参数示例

查询 ap-guangzhou 地域下的独占集群数量。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "region",
        "op": "eq",
        "value": "ap-guangzhou"
      }
    ]
  },
  "page": {
    "count": true
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 0,
    "details": [
      {
        "id": "00000001",
        "cloud_id": "tgw-xxxxxxxx",
        "name": "cluster-l4-1",
        "vendor": "tcloud",
        "account_id": "00000001",
        "bk_biz_id": -1,
        "region": "ap-guangzhou",
        "zone": "ap-guangzhou-3",
        "cluster_type": "TGW",
        "cluster_tag": "demo_tag",
        "network": "Public",
        "isp": "BGP",
        "egress": "center_egress1",
        "ip_version": "IPv4",
        "max_conn": 2000000,
        "clb_resource_count": 20,
        "extension": {
          "max_in_flow": 10240,
          "max_out_flow": 10240,
          "max_in_pkg": 2000000,
          "max_out_pkg": 2000000,
          "max_new_conn": 100000,
          "http_max_new_conn": 50000,
          "https_max_new_conn": 30000,
          "http_qps": 300000,
          "https_qps": 200000,
          "load_balance_director_count": 4,
          "clusters_version": "v1",
          "disaster_recovery_type": "SINGLE-ZONE",
          "clusters_zone": {
            "master_zone": [
              "ap-guangzhou-3"
            ],
            "slave_zone": []
          }
        },
        "memo": "",
        "creator": "admin",
        "reviser": "admin",
        "created_at": "2026-09-15T09:03:06Z",
        "updated_at": "2026-09-15T09:03:06Z"
      }
    ]
  }
}
```

#### 获取数量返回结果示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "count": 1
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称    | 参数类型   | 描述             |
|---------|--------|----------------|
| count   | uint64 | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据        |

#### data.details[n]

| 参数名称                | 参数类型   | 描述                                     |
|---------------------|--------|----------------------------------------|
| id                  | string | 集群本地ID            |
| cloud_id            | string | 集群云上ID      |
| name                | string | 集群名称                                   |
| vendor              | string | 供应商（枚举值：tcloud）                        |
| account_id          | string | 账号ID                                   |
| bk_biz_id           | int64  | 业务ID，-1表示未分配                           |
| region              | string | 地域                                     |
| zone                | string | 可用区                                    |
| cluster_type        | string | 集群类型（TGW四层、STGW七层）             |
| cluster_tag         | string | 集群标签，空字符串表示未打标签                        |
| network             | string | 网络类型（Public公网、Private内网）               |
| isp                 | string | 运营商（BGP、CMCC、CUCC、CTCC、INTERNAL）       |
| egress              | string | 网络出口                                   |
| ip_version          | string | IP版本                                   |
| max_conn            | int64  | 最大连接数，为null时表示云上无此值（STGW集群），前端展示为 `—`  |
| clb_resource_count  | int64  | 集群内已绑定的CLB数量                             |
| extension           | object | 云上扩展字段                                 |
| memo                | string | 备注                                     |
| creator             | string | 创建者                                    |
| reviser             | string | 修改者                                    |
| created_at          | string | 创建时间，标准格式：2006-01-02T15:04:05Z         |
| updated_at          | string | 修改时间，标准格式：2006-01-02T15:04:05Z         |

#### data.details[n].extension[tcloud]

| 参数名称                        | 参数类型   | 描述                                                      |
|-----------------------------|--------|---------------------------------------------------------|
| max_in_flow                 | int64  | 最大入带宽，单位Mbps                                            |
| max_out_flow                | int64  | 最大出带宽，单位Mbps                                            |
| max_in_pkg                  | int64  | 最大入包量，个/秒                                               |
| max_out_pkg                 | int64  | 最大出包量，个/秒                                               |
| max_new_conn                | int64  | 最大新建连接数，个/秒                                             |
| http_max_new_conn           | int64  | http最大新建连接数，个/秒                                         |
| https_max_new_conn          | int64  | https最大新建连接数，个/秒                                        |
| http_qps                    | int64  | http QPS                                                |
| https_qps                   | int64  | https QPS                                               |
| load_balance_director_count | int64  | 集群内转发机数目                                                |
| clusters_version            | string | 集群版本                                                    |
| disaster_recovery_type      | string | 集群容灾类型（SINGLE-ZONE、DISASTER-RECOVERY、MUTUAL-DISASTER-RECOVERY） |
| clusters_zone               | object | 集群所在可用区                                                 |

##### clusters_zone

| 参数名称        | 参数类型         | 描述      |
|-------------|--------------|---------|
| master_zone | string array | 集群所在主可用区 |
| slave_zone  | string array | 集群所在备可用区 |
