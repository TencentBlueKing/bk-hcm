### 描述

- 该接口提供版本：v1.8.0+。
- 该接口所需权限：资源-IaaS资源操作。
- 该接口功能描述：批量更新安全组管理属性，仅当所有管理属性均不存在时才允许编辑，所有管理属性都要提供。
    - 注意：通过该接口更新的安全组会被默认设置为业务管理类型，不可再更改为平台管理类型

### URL

PATCH /api/v1/cloud/security_groups/mgmt_attrs/batch

### 输入参数

| 参数名称            | 参数类型         | 必选 | 描述             |
|-----------------|--------------|----|----------------|
| security_groups | object array | 是  | 更新的安全组列表，最大100 |

#### security_groups[n]

| 参数名称        | 参数类型   | 必选 | 描述    |
|-------------|--------|----|-------|
| id          | string | 是  | 安全组ID |
| manager     | string | 是  | 负责人   |
| bak_manager | string | 是  | 备份负责人 |
| mgmt_biz_id | int    | 是  | 管理业务  |

### 调用示例

```json
{
  "security_groups": [
    {
      "id": "00000001",
      "manager": "lihua",
      "bak_manager": "hanmeimei",
      "mgmt_biz_id": 123
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
