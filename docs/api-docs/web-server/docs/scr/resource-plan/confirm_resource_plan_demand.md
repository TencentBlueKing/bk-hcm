### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-资源预测。
- 该接口功能描述：确认资源预测需求。

### URL

POST /api/v1/woa/plans/resources/demands/confirm

### 输入参数

| 参数名称       | 参数类型         | 必选 | 描述                      |
|------------|--------------|----|-------------------------|
| bk_biz_id  | int          | 是  | 业务ID。                   |
| demand_ids | string array | 是  | 预测需求ID列表，至少需要1个，数量最大100 |

### 调用示例

```json
{
  "bk_biz_id": 2005000001,
  "demand_ids": [
    "0000001z",
    "0000002z"
  ]
}
```

### 响应示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "success_ids": [
      "0000001z"
    ],
    "failed_ids": [
      "0000002z"
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                          |
|---------|--------|-----------------------------|
| result  | bool   | 请求成功与否。true:请求成功；false 请求失败 |
| code    | int    | 错误编码。0 表示 success，>0 表示失败错误 |
| message | string | 请求失败返回的错误信息                 |
| data    | object | 响应数据                        |

#### data

| 参数名称        | 参数类型         | 描述               |
|-------------|--------------|------------------|
| success_ids | string array | 成功确认的资源预测需求ID列表。 |
| failed_ids  | string array | 失败的资源预测需求ID列表。   |


