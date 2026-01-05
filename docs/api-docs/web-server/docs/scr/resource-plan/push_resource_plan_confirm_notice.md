### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-资源预测。
- 该接口功能描述：手动触发资源预测确认通知。

### URL

POST /api/v1/woa/plans/resources/demands/confirm_notifications/push

### 输入参数

| 参数名称       | 参数类型      | 必选 | 描述                      |
|------------|-----------|----|-------------------------|
| bk_biz_ids | int array | 否  | 业务ID列表。不传或为空时，表示通知所有业务。 |

### 调用示例

```json
{
  "bk_biz_ids": [
    2005000001,
    2005000002
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
      2005000001
    ],
    "failed_ids": [
      2005000002
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

| 参数名称        | 参数类型      | 描述            |
|-------------|-----------|---------------|
| success_ids | int array | 通知发送成功的业务ID列表 |
| failed_ids  | int array | 通知发送失败的业务ID列表 |


