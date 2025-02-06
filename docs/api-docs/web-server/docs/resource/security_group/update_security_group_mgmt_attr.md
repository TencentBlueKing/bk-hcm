### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源-IaaS资源操作。
- 该接口功能描述：更新安全组管理属性。

### URL

PATCH /api/v1/cloud/security_groups/{id}/mgmt_attrs

### 输入参数

| 参数名称          | 参数类型      | 必选 | 描述                                |
|---------------|-----------|----|-----------------------------------|
| id            | string    | 是  | 安全组ID                             |
| mgmt_type     | string    | 否  | 管理类型，枚举值：biz（业务管理）、platform（平台管理） |
| manager       | string    | 否  | 负责人                               |
| bak_manager   | string    | 否  | 备份负责人                             |
| usage_biz_ids | int array | 否  | 使用业务列表，-1代表全部业务可使用                |
| mgmt_biz_id   | int       | 否  | 管理业务。当管理类型为platform时，该字段必须为空      |

### 调用示例

```json
{
  "mgmt_type": "biz",
  "manager": "lihua",
  "bak_manager": "hanmeimei",
  "usage_biz_ids": [
    123,
    234
  ],
  "mgmt_biz_id": 123
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
