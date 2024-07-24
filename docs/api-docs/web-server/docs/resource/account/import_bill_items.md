### 描述

- 该接口提供版本：v9.9.9。
- 该接口所需权限：
- 该接口功能描述：账单详情导入接口。

### URL

POST /api/v1/account/vendors/{vendor}/bills/items/import

### 输入参数

| 参数名称       | 参数类型   | 必选 | 描述                        |
|------------|--------|----|---------------------------|
| bill_year  | int    | 是  | 账单年份                      |
| bill_month | int    | 是  | 账单月份                      |
| items      | object | 是  | 账单数据,来源自preview接口的items数组 |


### 调用示例
```json
{
  "bill_year": 2024,
  "bill_month": 6,
  "items": [
    {
      "root_account_id": "00000002",
      "main_account_id": "0000000j",
      "vendor": "zenlayer",
      "product_id": 1,
      "bk_biz_id": -1,
      "bill_year": 2024,
      "bill_month": 6,
      "bill_day": 1,
      "version_id": 1,
      "currency": "USD",
      "cost": "48000",
      "res_amount": "0",
      "extension": {
        "bill_id": "24-6-1",
        "zenlayer_order": "Xxx",
        "cid": "1",
        "group_id": "Xxxx",
        "currency": "USD",
        "city": "卡拉奇",
        "pay_content": "Xxxx",
        "type": "专线",
        "acceptance_num": "1",
        "pay_num": "1",
        "unit_price_usd": "48000",
        "total_payable": "48000",
        "billing_period": "202406",
        "contract_period": "Xxxx",
        "remarks": "",
        "business_group": "Xxxx",
        "cpu": null,
        "disk": null,
        "memory": null
      }
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
    "ids": [
      "00000021",
      "00000022",
      "00000023",
      "00000024",
      "00000025",
      "00000026",
      "00000027",
      "00000028",
      "00000029",
      "0000002a",
      "0000002b",
      "0000002c",
      "0000002d",
      "0000002e",
      "0000002f",
      "0000002g",
      "0000002h",
      "0000002i",
      "0000002j",
      "0000002k",
      "0000002l",
      "0000002m",
      "0000002n",
      "0000002o"
    ]
  }
}
```

### 响应参数说明
| 参数名称       | 参数类型   | 描述   |
|------------|--------|------|
| code       | int32  | 状态码  |
| message    | string | 请求信息 |
| data       | object | 返回数据 |

#### data 字段说明
| 参数名称 | 参数类型  | 描述   |
|------|-------|------|
| ids  | array | 账单id |
