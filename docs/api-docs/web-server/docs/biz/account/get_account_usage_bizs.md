### 描述

- 该接口提供版本：v1.8.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询指定账号的使用业务。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/accounts/usage_bizs/{account_id}

### 输入参数

| 参数名称       | 参数类型   | 必选 | 描述   |
|------------|--------|----|------|
| bk_biz_id  | int64  | 是  | 业务ID |
| account_id | string | 是  | 账号ID |

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": [123]
}
```

### 响应参数说明

| 参数名称    | 参数类型      | 描述       |
|---------|-----------|----------|
| code    | int32     | 状态码      |
| message | string    | 请求信息     |
| data    | array int | 账号使用业务列表 |
