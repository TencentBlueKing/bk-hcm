### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：无。
- 该接口功能描述：查询性能保障规格参数。腾讯云代理接口 DescribeSlaCapacity

### URL

POST /api/v1/cloud/vendors/{vendor}/load_balancers/sla/capacity/describe

### 输入参数

| 参数名称   | 参数类型     | 必选 | 描述                                   |
|------------|--------------|------|----------------------------------------|
| vendor     | string       | 是   | 云厂商（枚举值：tcloud，当前只支持tcloud） |
| account_id | string       | 是   | 云账户id                               |
| region     | string       | 是   | 地域                                   |
| sla_types  | string array | 否   | 规格类型                               |

### 调用示例

```json
{
  "account_id": "0000001",
  "region": "ap-guangzhou",
  "sla_types": ["clb.c1.small"]
}
```

### 响应示例

```json
{
  "result": true,
  "code": 0,
  "message": "",
  "data": {
    "SlaSet": [
      {
        "SlaType": "clb.c1.small",
        "SlaName": "简约型",
        "MaxConn": 10000,
        "MaxCps": 1000,
        "MaxOutBits": 100,
        "MaxInBits": 100,
        "MaxQps": 1000
      }
    ],
    "RequestId": "2023aed2-bb5c-403c-8ef2-289f363f1c46"
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
|----------|----------|----------|
| code     | int32    | 状态码   |
| message  | string   | 请求信息 |
| data     | Object   | 响应数据 |

#### data

| 参数名称  | 参数类型     | 描述                         |
|-----------|--------------|------------------------------|
| SlaSet    | object array | 机型规格数组                 |
| RequestId | string       | 请求腾讯云唯一id,用于定位问题 |

#### SlaSet[0]

| 参数名称   | 参数类型 | 描述           |
|------------|----------|----------------|
| SlaType    | string   | 性能容量型规格 |
| SlaName    | string   | 规格名称       |
| MaxConn    | int      | 并发连接数上限 |
| MaxCps     | int      | 新建连接数上限 |
| MaxOutBits | int      | 最大出口流量   |
| MaxInBits  | int      | 最大入口流量   |
| MaxQps     | int      | 最大qps        |
