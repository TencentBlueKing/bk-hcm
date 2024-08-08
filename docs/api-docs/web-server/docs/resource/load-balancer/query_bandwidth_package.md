### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：负载均衡创建。
- 该接口功能描述：查询带宽包。

### URL

POST /api/v1/cloud/bandwidth_packages/query

### 输入参数

#### tcloud

| 参数名称            | 参数类型         | 必选 | 描述                                    |
|-----------------|--------------|----|---------------------------------------|
| account_id      | string       | 是  | 账号ID                                  |
| region          | string       | 是  | 地域                                    |
| page            | Page         | 是  | 分页信息                                  |
| pkg_cloud_ids   | string array | 否  | 带宽包云id 过滤                             |
| pkg_names       | string array | 否  | 带宽包名称过滤                               |
| network_types   | string array | 否  | 带宽包网络类型过滤                             |
| charge_types    | string array | 否  | 带宽包的计费类型过滤                            |
| resource_types  | string array | 否  | 按带宽包资源类型过滤, 支持`Address`和`LoadBalance` |
| resource_ids    | string array | 否  | 按带宽包资源ID过滤                            |
| res_address_ips | string array | 否  | 按带宽包资源IP过滤                            |

##### Page

| 参数名称   | 参数类型 | 必选 | 描述                      |
|--------|------|----|-------------------------|
| offset | uint | 否  | 查询带宽包偏移量，默认为0。          |
| limit  | uint | 是  | 查询带宽包返回数量，默认为20，最大值为100 |

#### network_type 带宽包的网络类型取值范围：

- `BGP`  普通BGP共享带宽包
- `HIGH_QUALITY_BGP`  精品BGP共享带宽包
- `SINGLEISP_CMCC`  中国移动共享带宽包
- `SINGLEISP_CTCC`  中国电信共享带宽包
- `SINGLEISP_CUCC`  中国联通共享带宽包

#### network_type 带宽包的计费类型取值范围：

- `TOP5_POSTPAID_BY_MONTH` 按月后付费TOP5计费
- `PERCENT95_POSTPAID_BY_MONTH` 按月后付费月95计费
- `ENHANCED95_POSTPAID_BY_MONTH` 按月后付费增强型95计费
- `FIXED_PREPAID_BY_MONTH` 包月预付费计费
- `PEAK_BANDWIDTH_POSTPAID_BY_DAY`  后付费日结按带宽计费

### 调用示例

#### tcloud

```json
{
  "account_id": "0000001",
  "region": "ap-guangzhou",
  "page": {
    "limit": 10
  },
  "network_type": [
    "BGP"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "total_count": 1,
    "packages": [
      {
        "id": "bwp-1234556",
        "name": "name",
        "network_type": "BGP",
        "charge_type": "PRIMARY_TRAFFIC_POSTPAID_BY_HOUR",
        "status": "CREATED",
        "bandwidth": 999,
        "egress": "egress123",
        "create_time": "2024-05-20T11:19:21Z",
        "deadline": "",
        "resource_set": []
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

#### data[tcloud]

| 参数名称        | 参数类型                    | 描述    |
|-------------|-------------------------|-------|
| total_count | int                     | 总带宽包数 |
| packages    | bandwidth_package array | 带宽包数组 |

#### bandwidth_package

| 参数名称         | 参数类型           | 描述                                                |
|--------------|----------------|---------------------------------------------------|
| id           | string         | 后续计价单元，HOUR、GB                                    |
| name         | string         | 折扣 ，如20.0代表2折                                     |
| network_type | string         | 带宽包类型                                             |
| charge_type  | string         | 带宽包的计费类型                                          |
| status       | string         | 带宽包状态，包括'CREATING','CREATED','DELETING','DELETED' |
| bandwidth    | int64          | 带宽包限速大小。单位：Mbps，-1表示不限速。                          |
| egress       | string         | 网络出口                                              |
| create_time  | string         | 创建时间                                              |
| deadline     | string         | 预付费带宽包到期时间                                        |
| resource_set | resource array | 带宽包资源信息                                           |

##### resource 带宽包资源信息

| 参数名称          | 参数类型   | 描述                                |
|---------------|--------|-----------------------------------|
| resource_type | string | 带宽包资源类型，包括'Address'和'LoadBalance' |
| resource_id   | string | 带宽包资源Id，形如'eip-xxxx', 'lb-xxxx'   |
| address_ip    | string | 带宽包资源Ip                           |
