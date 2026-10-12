### 描述

- 该接口提供版本：v9.9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询业务下可用的负载均衡独占集群标签及标签下的集群

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/exclusive_clusters/tags/list

### 输入参数

| 参数名称         | 参数类型         | 必选 | 描述                                                |
|--------------|--------------|----|---------------------------------------------------|
| bk_biz_id    | int64        | 是  | 业务ID，路径参数                                         |
| account_id   | string       | 是  | 账号ID                                              |
| region       | string       | 是  | 地域                                                |
| isp          | string       | 是  | 运营商（枚举值：CMCC、CUCC、CTCC） |
| zones        | string array | 否  | 主可用区列表，用于按可用区过滤独占集群                         |
| back_zones   | string array | 否  | 备可用区列表，用于按可用区过滤主备独占集群                       |
| cluster_type | string       | 否  | 集群类型（枚举值：TGW、STGW），为空则同时返回四层与七层             |


### 调用示例

查询业务 213 在 ap-guangzhou 地域、ap-guangzhou-3 单可用区下 BGP 线路的独占集群标签。

```json
{
  "account_id": "00000001",
  "region": "ap-guangzhou",
  "isp": "BGP",
  "zones": ["ap-guangzhou-3"],
  "back_zones": [],
  "cluster_type": ""
}
```

查询主备集群时，`zones` 填主可用区，`back_zones` 填备可用区：

```json
{
  "account_id": "00000001",
  "region": "ap-guangzhou",
  "isp": "BGP",
  "zones": ["ap-guangzhou-3"],
  "back_zones": ["ap-guangzhou-4"],
  "cluster_type": ""
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "cluster_tag": "demo_tag",
        "cluster_type": "TGW",
        "clusters": [
          {
            "cloud_cluster_id": "tgw-xxxxxxxx",
            "cluster_id": "00000001",
            "cluster_name": "cluster-l4-1",
            "egress": "center_egress1",
            "isp": "BGP",
            "cluster_zone": {
              "master_zone": ["ap-guangzhou-3"],
              "slave_zone": []
            }
          },
          {
            "cloud_cluster_id": "tgw-yyyyyyyy",
            "cluster_id": "00000003",
            "cluster_name": "cluster-l4-2",
            "egress": "center_egress1",
            "isp": "BGP",
            "cluster_zone": {
              "master_zone": ["ap-guangzhou-3"],
              "slave_zone": ["ap-guangzhou-4"]
            }
          }
        ]
      },
      {
        "cluster_tag": "demo_tag",
        "cluster_type": "STGW",
        "clusters": [
          {
            "cloud_cluster_id": "stgw-xxxxxxxx",
            "cluster_id": "00000001",
            "cluster_name": "cluster-l7-1",
            "egress": "center_egress1",
            "isp": "BGP",
            "cluster_zone": {
              "master_zone": ["ap-guangzhou-3"],
              "slave_zone": []
            }
          },
          {
            "cloud_cluster_id": "stgw-yyyyyyyy",
            "cluster_id": "00000003",
            "cluster_name": "cluster-l7-2",
            "egress": "center_egress1",
            "isp": "BGP",
            "cluster_zone": {
              "master_zone": ["ap-guangzhou-3"],
              "slave_zone": ["ap-guangzhou-4"]
            }
          }
        ]
      }
    ]
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

| 参数名称    | 参数类型  | 描述      |
|---------|-------|---------|
| details | array | 标签聚合结果数组 |

#### data.details[n]

| 参数名称         | 参数类型          | 描述                            |
|--------------|---------------|-------------------------------|
| cluster_tag  | string        | 集群标签   |
| cluster_type | string        | 集群类型（TGW四层、STGW七层），前端映射为四层/七层 |
| clusters     | cluster array | 该标签下的集群列表                     |

##### cluster

| 参数名称                | 参数类型   | 描述                                                   |
|---------------------|--------|------------------------------------------------------|
| cloud_cluster_id    | string        | 集群云上ID               |
| cluster_id          | string        | 集群本地ID                                               |
| cluster_name        | string        | 集群名称                                                 |
| egress              | string        | 网络出口                      |
| isp                 | string        | 运营商                                                  |
| cluster_zone        | object        | 集群主备可用区|

###### cluster_zone

| 参数名称        | 参数类型         | 描述       |
|-------------|--------------|----------|
| master_zone | string array | 集群所在主可用区 |
| slave_zone  | string array | 集群所在备可用区 |
