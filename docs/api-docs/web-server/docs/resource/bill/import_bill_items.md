### 描述

- 该接口提供版本：v1.6.0。
- 该接口所需权限：账单创建
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
  "data": null
}
```

### 响应参数说明
| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | null   | 返回数据 |

