### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询Azure路由列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/azure/route_tables/{route_table_id}/routes/list

### 输入参数

| 参数名称           | 参数类型   | 必选 | 描述     |
|----------------|--------|----|--------|
| bk_biz_id      | int64  | 是  | 业务ID   |
| route_table_id | string | 是  | 路由表ID  |
| filter         | object | 是  | 查询过滤条件 |
| page           | object | 是  | 分页设置   |

#### filter

| 参数名称  | 参数类型        | 必选  | 描述                                                              |
|-------|-------------|-----|-----------------------------------------------------------------|
| op    | enum string | 是   | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是   | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### filter.rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选  | 描述                                          |
|-------|-------------|-----|---------------------------------------------|
| field | string      | 是   | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| op    | enum string | 是   | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis） |
| value | 可变类型        | 是   | 查询条件Value值                                  |

##### rules 表达式说明：

##### 1. 操作符

| 操作符 | 描述                                        | 操作符的value支持的数据类型                             |
|-----|-------------------------------------------|----------------------------------------------|
| eq  | 等于。不能为空字符串                                | boolean, numeric, string                     |
| neq | 不等。不能为空字符串                                | boolean, numeric, string                     |
| gt  | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte | 大于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lt  | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lte | 小于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in  | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string                     |
| nin | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string                     |
| cs  | 模糊查询，区分大小写                                | string                                       |
| cis | 模糊查询，不区分大小写                               | string                                       |

##### 2. 协议示例

查询 name 是 "Jim" 且 age 大于18小于30 且 servers 类型是 "api" 或者是 "web" 的数据。

```json
{
  "op": "and",
  "rules": [
    {
      "field": "name",
      "op": "eq",
      "value": "Jim"
    },
    {
      "field": "age",
      "op": "gt",
      "value": 18
    },
    {
      "field": "age",
      "op": "lt",
      "value": 30
    },
    {
      "field": "servers",
      "op": "in",
      "value": [
        "api",
        "web"
      ]
    }
  ]
}
```

#### page

| 参数名称   | 参数类型    | 必选  | 描述                                                                                                                                               |
|--------|---------|-----|--------------------------------------------------------------------------------------------------------------------------------------------------|
| count	 | bool	   | 是	  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据 detail，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count |
| start	 | uint32	 | 否	  | 记录开始位置，start 起始值为0                                                                                                                               |
| limit	 | uint32	 | 否	  | 每页限制条数，最大500，不能为0                                                                                                                                |
| sort	  | string	 | 否	  | 排序字段，返回数据将按该字段进行排序                                                                                                                               |
| order	 | string	 | 否	  | 排序顺序（枚举值：ASC、DESC）                                                                                                                               |

#### 查询参数介绍：

| 参数名称                 | 参数类型   | 描述                                                                        |
|----------------------|--------|---------------------------------------------------------------------------|
| id                   | string | 路由ID                                                                      |
| cloud_id             | string | 路由的云上ID                                                                   |
| route_table_id       | string | 路由表ID                                                                     |
| cloud_route_table_id | string | 路由表的云上ID                                                                  |
| name                 | string | 路由名称                                                                      |
| address_prefix       | string | 目的网段，格式为：CIDR                                                             |
| next_hop_type        | string | 下一跳类型（枚举值：Internet、None、VirtualAppliance、VirtualNetworkGateway、VnetLocal） |
| next_hop_ip_address  | string | 下一跳地址                                                                     |
| provisioning_state   | string | 路由类型（枚举值：Deleting、Failed、Succeeded、Updating）                              |
| creator              | string | 创建者                                                                       |
| reviser              | string | 更新者                                                                       |
| created_at           | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                             |
| updated_at           | string | 更新时间，标准格式：2006-01-02T15:04:05Z                                             |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 获取详细信息请求参数示例

如查询ID为"00000001"的路由列表。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "id",
        "op": "eq",
        "value": "00000001"
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

如查询ID为"00000001"的路由数量。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "id",
        "op": "eq",
        "value": "00000001"
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
  "message": "ok",
  "data": {
    "details": [
      {
        "id": "0000000a",
        "cloud_id": "/subscriptions/xxx/resourceGroups/xxx/providers/Microsoft.Network/routeTables/test/routes/test",
        "route_table_id": "0000006e",
        "cloud_route_table_id": "/subscriptions/xxx/resourceGroups/xxx/providers/Microsoft.Network/routeTables/test",
        "name": "test",
        "address_prefix": "ApiManagement.AustraliaEast",
        "next_hop_type": "VirtualAppliance",
        "next_hop_ip_address": "127.0.0.1",
        "provisioning_state": "Succeeded",
        "creator": "tom",
        "reviser": "tom",
        "created_at": "2019-07-29 11:57:20",
        "updated_at": "2019-07-29 11:57:20"
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

| 参数名称   | 参数类型   | 描述                                       |
|--------|--------|------------------------------------------|
| count  | uint64 | 当前规则能匹配到的总记录条数，仅在 count 查询参数设置为 true 时返回 |
| detail | array  | 查询返回的数据，仅在 count 查询参数设置为 false 时返回       |

#### data.detail[n]

| 参数名称                 | 参数类型   | 描述                                                                        |
|----------------------|--------|---------------------------------------------------------------------------|
| id                   | string | 路由ID                                                                      |
| cloud_id             | string | 路由的云上ID                                                                   |
| route_table_id       | string | 路由表ID                                                                     |
| cloud_route_table_id | string | 路由表的云上ID                                                                  |
| name                 | string | 路由名称                                                                      |
| address_prefix       | string | 目的网段，格式为：CIDR                                                             |
| next_hop_type        | string | 下一跳类型（枚举值：Internet、None、VirtualAppliance、VirtualNetworkGateway、VnetLocal） |
| next_hop_ip_address  | string | 下一跳地址                                                                     |
| provisioning_state   | string | 路由类型（枚举值：Deleting、Failed、Succeeded、Updating）                              |
| creator              | string | 创建者                                                                       |
| reviser              | string | 更新者                                                                       |
| created_at           | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                             |
| updated_at           | string | 更新时间，标准格式：2006-01-02T15:04:05Z                                             |
