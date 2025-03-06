
### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：账单管理。
- 该接口功能描述：编辑调账明细，已确定的调账明细不能编辑，该接口不能确认调账明细

### URL

PATCH /api/v1/account/bills/adjustment_items/{id}

### 输入参数

| 参数名称            | 参数类型   | 必选 | 描述                          |
|-----------------|--------|----|-----------------------------|
| main_account_id | string | 否  | 所属主账号id                     |
| product_id      | int    | 否  | 运营产品id                      |
| bk_biz_id       | int    | 否  | 业务id                        |
| type            | string | 否  | 调账类型 枚举值（increase、decrease） |
| currency        | string | 否  | 币种                          |
| cost            | string | 否  | 金额                          |
| memo            | string | 否  | 备注信息                        |


### 调用示例

```json
{
  "count": 0,
  "details": [
    {
      "id": "0000000a",
      "main_account_id": "00000001",
      "product_id": 1234,
      "bk_biz_id": 5678,
      "type": "increase",
      "memo": "",
      "currency": "RMB",
      "cost": "42.67512105"
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data":null
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
