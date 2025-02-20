### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：账单管理。
- 该接口功能描述：批量创建调账明细

### URL

POST /api/v1/account/vendors/{vendor}/bills/adjustment_items/create

### 输入参数

| 参数名称            | 参数类型                  | 必选 | 描述                    |
|-----------------|-----------------------|----|-----------------------|
| root_account_id | string                | 是  | 所属根账号id               |
| vendor          | string                | 是  | 所属厂商                  |
| items           | adjustment_item array | 是  | 调账明细列表, min=1,max=100 |

### adjustment_item

| 参数名称            | 参数类型   | 必选 | 描述                          |
|-----------------|--------|----|-----------------------------|
| root_account_id | string | 否  | 所属根账号id                     |
| main_account_id | string | 是  | 所属主账号id                     |
| product_id      | int    | 否  | 运营产品id                      |
| bk_biz_id       | int    | 否  | 业务id                        |
| bill_year       | int    | 否  | 所属年份                        |
| bill_month      | int    | 否  | 所属月份                        |
| bill_day        | int    | 是  | 所属日期                        |
| type            | string | 是  | 调账类型 枚举值（increase、decrease） |
| currency        | string | 是  | 币种                          |
| cost            | string | 是  | 金额                          |
| memo            | string | 否  | 备注信息                        |

### 调用示例

```json
{
  "root_account_id": "00000001",
  "vendor": "huawei",
  "items": [
    {
      "root_account_id": "00000001",
      "main_account_id": "00000001",
      "product_id": 6667,
      "bk_biz_id": 1234,
      "bill_year": 2024,
      "bill_month": 6,
      "type": "increase",
      "memo": "",
      "currency": "RMB",
      "cost": "123",
      "rmb_cost": "123"
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
    "id": "00000001"
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

| 参数名称 | 参数类型   | 描述     |
|------|--------|--------|
| id   | string | 调账明细id |

