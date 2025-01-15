
### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：账单管理。
- 该接口功能描述：查询调账明细

### URL

POST /api/v1/account/bills/adjustment_items/list

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

| 参数名称            | 参数类型   | 描述                             |
|-----------------|--------|--------------------------------|
| id              | string | 调账id                           |
| root_account_id | string | 所属根账号id                        |
| main_account_id | string | 所属主账号id                        |
| product_id      | int    | 运营产品id                         |
| bk_biz_id       | int    | 业务id                           |
| bill_year       | int    | 所属年份                           |
| bill_month      | int    | 所属月份                           |
| bill_day        | int    | 所属日期                           |
| type            | string | 调账类型 increase/decrease         |
| memo            | string | 备注                             |
| operator        | string | 操作人                            |
| currency        | string | 币种 RMB/USD                     |
| cost            | string | 原币种消费（元）                       |
| rmb_cost        | string | 人民币消费                          |
| state           | string | 未确定、已确定                        |
| creator         | string | 创建者                            |
| created_at      | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at      | string | 修改时间，标准格式：2006-01-02T15:04:05Z |

### 调用示例

```json
{
  "count": 0,
  "details": [
    {
      "id": "0000000b",
      "root_account_id": "00000001",
      "main_account_id": "00000001",
      "product_id": 3043,
      "bk_biz_id": 100857,
      "bill_year": 2024,
      "bill_month": 6,
      "bill_day": 18,
      "type": "increase",
      "memo": "",
      "operator": "ryanjrchen",
      "currency": "RMB",
      "cost": "42.67512105",
      "rmb_cost": "42.67512105",
      "state": "已确定",
      "creator": "ryanjrchen",
      "created_at": "2024-06-18T13:09:34Z",
      "updated_at": "2024-06-18T13:09:34Z"
    }
  ]
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
        "id": "00000001",
        "root_account_id": "00000001",
        "main_account_id": "00000001",
        "product_id": 3456,
        "bk_biz_id": 1,
        "bill_year": 2024,
        "bill_month": 6,
        "bill_day": 18,
        "type": "increase",
        "memo": "",
        "operator": "admin",
        "currency": "RMB",
        "cost": "42.67512105",
        "rmb_cost": "42.67512105",
        "state": "confirmed",
        "creator": "admin",
        "created_at": "2024-06-18T13:09:34Z",
        "updated_at": "2024-06-18T13:09:34Z"
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

| 参数名称            | 参数类型   | 描述                             |
|-----------------|--------|--------------------------------|
| id              | string | 调账id                           |
| root_account_id | string | 所属根账号id                        |
| main_account_id | string | 所属主账号id                        |
| product_id      | int    | 运营产品id                         |
| bk_biz_id       | int    | 业务id                           |
| bill_year       | int    | 所属年份                           |
| bill_month      | int    | 所属月份                           |
| bill_day        | int    | 所属日期                           |
| type            | string | 调账类型 increase/decrease         |
| memo            | string | 备注                             |
| operator        | string | 操作人                            |
| currency        | string | 币种 RMB/USD                     |
| cost            | string | 原币种消费（元）                       |
| rmb_cost        | string | 人民币消费                          |
| state           | string | 未确定、已确定                        |
| creator         | string | 创建者                            |
| created_at      | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at      | string | 修改时间，标准格式：2006-01-02T15:04:05Z |

