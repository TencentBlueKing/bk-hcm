### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：二级账号操作。
- 该接口功能描述：删除账号密钥。

### URL

DELETE /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/account_secrets/batch

### 输入参数

| 参数名称      | 参数类型         | 必选 | 描述              |
|-----------|--------------|----|-----------------|
| bk_biz_id | int64        | 是  | 业务ID            |
| vendor    | string       | 是  | 云厂商, 枚举值：tcloud |
| ids       | string array | 是  | 密钥ID列表，长度限制100  |

### 调用示例

```json
{
  "ids": [
    "00000001",
    "00000002"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": ""
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述   |
|------|------|------|
| code | int32 | 状态码  |
| message | string | 请求信息 |
