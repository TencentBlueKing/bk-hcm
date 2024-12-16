### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：账单管理。
- 该接口功能描述：查询汇率

### URL

POST /api/v1/account/bills/exchange_rates/list

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

#### 查询参数介绍：

| 参数名称          | 参数类型   | 描述   |
|---------------|--------|------|
| year          | int    | 所属年份 |
| month         | int    | 所属月份 |
| from_currency | string | 原币种  |
| to_currency   | string | 目标币种 |

### 调用示例

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "year",
        "op": "eq",
        "value": 2024
      }
    ]
  },
  "page": {
    "limit": 10
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 0,
    "details": [
      {
        "id": "1",
        "year": 2024,
        "month": 6,
        "from_currency": "USD",
        "to_currency": "CNY",
        "exchange_rate": "7.1",
        "creator": "admin",
        "reviser": "admin",
        "created_at": "2024-07-01T01:31:48Z",
        "updated_at": "2024-07-01T01:31:48Z"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称          | 参数类型   | 描述                             |
|---------------|--------|--------------------------------|
| id            | string | 调账id                           |
| year          | int    | 所属年份                           |
| month         | int    | 所属月份                           |
| from_currency | string | 原币种                            |
| to_currency   | string | 目标币种                           |
| exchange_rate | string | 汇率，字符串形式小数                     |
| created_at    | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at    | string | 修改时间，标准格式：2006-01-02T15:04:05Z |
