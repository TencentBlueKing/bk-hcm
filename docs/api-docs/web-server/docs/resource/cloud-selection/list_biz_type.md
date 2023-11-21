### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源选型-选型推荐。
- 该接口功能描述：查询业务类型。

### URL

POST /api/v1/cloud/selections/biz_types/list

### 输入参数

| 参数名称 | 参数类型   | 必选 | 描述   |
|------|--------|----|------|
| page | object | 是  | 分页设置 |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

### 调用示例

```
{
  "page": {
    "offset": 0,
    "limit": 10
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
    "details": [
      {
        "id": "00000001",
        "biz_type": "biz1",
        "network_latency_tolerance": 180,
        "deployment_architecture": [
          "distributed"
        ]
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

| 参数名称    | 参数类型         | 描述      |
|---------|--------------|---------|
| details | string array | 查询返回的数据 |
| count   | int          | 数量      |

#### details[n]

| 参数名称                      | 参数类型         | 描述                              |
|---------------------------|--------------|---------------------------------|
| id                        | string       | 业务类型id                          |
| biz_type                  | string       | 业务类型名称                          |
| network_latency_tolerance | number       | 网络延迟值                           |
| deployment_architecture   | string array | 部署架构，取值：distributed,centralized |
