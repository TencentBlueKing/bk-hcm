### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：二级账号操作。
- 该接口功能描述：更新账号密钥。

### URL

PATCH /api/v1/cloud/bizs/{bk_biz_id}/account_secrets/{id}

### 输入参数

| 参数名称    | 参数类型   | 必选 | 描述                          |
|---------|--------|----|-----------------------------|
| bk_biz_id | int64  | 是  | 业务ID                        |
| id      | string | 是  | 密钥ID                        |
| type    | string | 否  | 密钥类型, 枚举值：resource(资源管理)、security(安全管理) |
| extension | object | 否  | 云厂商差异扩展字段                 |

#### extension[tcloud]

| 参数名称               | 参数类型   | 必选 | 描述        |
|--------------------|--------|----|-----------|
| cloud_secret_id    | string | 是  | 云密钥id     |
| cloud_secret_key   | string | 是  | 云密钥key    |

### 调用示例

```json
{
  "type": "resource",
  "extension": {
    "cloud_secret_id": "xxxx",
    "cloud_secret_key": "xxxx"
  }
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

| 参数名称 | 参数类型 | 描述   |
|------|------|------|
| code | int32 | 状态码  |
| message | string | 请求信息 |
| data | object | 响应数据 |
