### 描述

- 该接口提供版本：v1.6.11+。
- 该接口所需权限：平台管理-滚服管理。
- 该接口功能描述：查询滚服申请记录列表。

### URL

POST /api/v1/woa/rolling_servers/applied_records/list

### 输入参数

| 参数名称           | 参数类型 | 必选    | 描述        |
|-------------------|--------|---------|------------|
| filter            | object | 是      | 查询过滤条件  |
| page              | object | 是      | 分页设置     |

#### filter

| 参数名称  | 参数类型        | 必选  | 描述                                                              |
|-------|-------------|-----|-----------------------------------------------------------------|
| op    | enum string | 是   | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是   | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选  | 描述                                         |
|-------|-------------|-----|--------------------------------------------|
| field | string      | 是   | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍 |
| op    | enum string | 是   | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis）       |
| value | 可变类型        | 是   | 查询条件Value值                                 |

##### rules 表达式说明：

##### 1. 操作符

| 操作符 | 描述                                    | 操作符的value支持的数据类型                                |
|-----|-------------------------------------------|--------------------------------------------------------|
| eq  | 等于。不能为空字符串                         | boolean, numeric, string                                |
| neq | 不等。不能为空字符串                         | boolean, numeric, string                                |
| gt  | 大于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte | 大于等于                                   | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lt  | 小于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lte | 小于等于                                   | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in  | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string           |
| nin | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string          |
| cs  | 模糊查询，区分大小写                         | string                                                   |
| cis | 模糊查询，不区分大小写                       | string                                                    |

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

| 参数名称  | 参数类型 | 必选  | 描述                                                                                                                                                  |
|----------|--------|------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
| count    | bool   | 是   | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start    | uint32 | 否   | 记录开始位置，start 起始值为0                                                                                                                             |
| limit    | uint32 | 否   | 每页限制条数，最大500，不能为0                                                                                                                             |

#### 查询参数介绍：

| 参数名称           | 参数类型 | 描述              |
|-------------------|--------|-------------------|
| applied_type     | string  | 申请类型           |
| bk_biz_id        | int     | 业务ID             |
| order_id         | string  | 主机申请的订单号     |
| suborder_id      | string  | 主机申请的子订单号    |
| year             | int     | 申请时间年份         |
| month            | int     | 申请时间月份         |
| day              | int     | 申请时间天           |
| applied_core     | int     | cpu申请核心数        |
| delivered_core   | int     | cpu交付核心数        |
| creator          | string  | 创建者              |
| created_at       | string  | 创建时间，标准格式：2006-01-02T15:04:05Z |

### 调用示例

#### 获取详细信息请求参数示例

如查询业务ID为1的滚服申请记录列表。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "bk_biz_id",
        "op": "eq",
        "value": 1
      }
    ]
  },
  "page": {
    "count": false,
    "start": 0,
    "limit": 1
  }
}
```

#### 获取数量请求参数示例

如查询业务ID为1的滚服申请记录数量。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "bk_biz_id",
        "op": "eq",
        "value": 1
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

如查询申请类型为普通申请的列表响应示例。

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "details": [
      {
        "id": 1,
        "applied_type": "normal",
        "bk_biz_id": 1,
        "order_id": "1001",
        "suborder_id": "1001-1",
        "year": 2024,
        "month": 10,
        "day": 1,
        "applied_core": 0,
        "delivered_core": 0,
        "exempted_returned_core": 0,
        "creator": "xxxx",
        "created_at": "2024-10-01T00:00:00Z"
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

| 参数名称  | 参数类型   | 描述   |
|---------|-----------|--------|
| code    | int       | 状态码  |
| message | string    | 请求信息 |
| data    | object    | 响应数据 |

#### data

| 参数名称  | 参数类型 | 描述                    |
|---------|---------|-------------------------|
| count   | int     | 当前规则能匹配到的总记录条数 |
| details | array   | 查询返回的数据            |

#### data.details[n]

| 参数名称                   | 参数类型   | 描述                                                             |
|------------------------|--------|----------------------------------------------------------------|
| id                     | string | 滚服申请表的主键ID                                                     |
| applied_type           | string | 申请类型(枚举值：normal-普通申请、resource_pool-资源池申请、cvm_product-管理员cvm生产) |
| bk_biz_id              | string | 业务ID                                                           |
| order_id               | string | 主机申请的订单号                                                       |
| suborder_id            | string | 主机申请的子订单号                                                      |
| year                   | string | 申请时间年份                                                         |
| month                  | string | 申请时间月份                                                         |
| day                    | string | 申请时间天                                                          |
| applied_core           | string | 申请资源数                                                          |
| delivered_core         | string | 已交付资源数                                                         |
| exempted_returned_core | string | 减免退还核心数                                                        |
| creator                | string | 创建者                                                            |
| created_at             | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                 |
