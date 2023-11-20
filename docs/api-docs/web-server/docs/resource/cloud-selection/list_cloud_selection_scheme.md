### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源选型-方案查看。
- 该接口功能描述：查询云选型方案列表。

### URL

POST /api/v1/cloud/selections/schemes/list

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

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

#### 查询参数介绍：

| 参数名称                      | 参数类型         | 描述                             |
|---------------------------|--------------|--------------------------------|
| id                        | uint64       | 审计ID                           |
| bk_biz_id                 | int          | 业务ID                           |
| name                      | string       | 方案名                            |
| biz_type                  | string       | 业务类型                           |
| deployment_architecture   | string       | 部署架构                           |
| network_latency_tolerance | double       | 用户容忍网络延迟                       |
| composite_score           | double       | 综合评分                           |
| net_score                 | double       | 网络评分                           |
| cost_score                | double       | 成本评分                           |
| cover_rate                | double       | 覆盖率                            |
| result_idc_ids            | string array | 推荐结果机房ID列表                     |
| creator                   | string       | 创建者                            |
| reviser                   | string       | 更新者                            |
| created_at                | string       | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at                | string       | 更新时间，标准格式：2006-01-02T15:04:05Z |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "operator",
        "op": "eq",
        "value": "Jim"
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

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "operator",
        "op": "eq",
        "value": "Jim"
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
        "bk_biz_id": 310,
        "name": "scheme1",
        "biz_type": "游戏",
        "vendors": [
          "aws",
          "tcloud"
        ],
        "deployment_architecture": [
          "xxxx"
        ],
        "cover_ping": 120,
        "composite_score": 80,
        "net_score": 70,
        "cost_score": 90,
        "cover_rate": 80,
        "user_distribution": [
          {
            "name": "country_1",
            "children": [
              {
                "name": "province_1_1",
                "value": 0.1
              },
              {
                "name": "province_1_2",
                "value": 0.1
              }
            ]
          },
          {
            "country": "country_2",
            "children": [
              {
                "name": "province_2_1",
                "value": 0.1
              },
              {
                "name": "province_2_2",
                "value": 0.1
              }
            ]
          }
        ],
        "result_idc_ids": [
          "00000002",
          "00000007"
        ],
        "creator": "Jim",
        "reviser": "Jim",
        "created_at": "2023-02-05T15:29:15Z",
        "updated_at": "2023-02-05T15:29:15Z"
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

| 参数名称                    | 参数类型                   | 描述                                   |
|-------------------------|------------------------|--------------------------------------|
| id                      | uint64                 | 审计ID                                 |
| bk_biz_id               | int                    | 业务ID                                 |
| name                    | string                 | 方案名                                  |
| biz_type                | string                 | 业务类型                                 |
| vendors                 | string array           | 供应商（枚举值：tcloud、aws、azure、gcp、huawei） |
| deployment_architecture | string array           | 部署架构                                 |
| cover_ping              | double                 | 用户容忍网络延迟                             |
| composite_score         | double                 | 综合评分                                 |
| net_score               | double                 | 网络评分                                 |
| cost_score              | double                 | 成本评分                                 |
| cover_rate              | double                 | 覆盖率                                  |
| user_distribution       | UserDistribution array | 用户占比                                 |
| result_idc_ids          | string array           | 推荐结果机房ID列表                           |
| creator                 | string                 | 创建者                                  |
| reviser                 | string                 | 更新者                                  |
| created_at              | string                 | 创建时间，标准格式：2006-01-02T15:04:05Z       |
| updated_at              | string                 | 更新时间，标准格式：2006-01-02T15:04:05Z       |

#### UserDistribution

| 参数名称     | 参数类型                   | 描述          |
|----------|------------------------|-------------|
| name     | string                 | 地理标签，国家/州省等 |
| value    | double                 | 分布权重        |
| children | UserDistribution array | 下级地理标签      |
