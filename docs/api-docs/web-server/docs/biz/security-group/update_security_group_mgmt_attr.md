### 描述

- 该接口提供版本：v1.8.0+。
- 该接口所需权限：业务-IaaS资源操作。
- 该接口功能描述：更新安全组管理属性

### URL

PATCH /api/v1/cloud/bizs/{bk_biz_id}/security_groups/mgmt_attrs/batch

### 输入参数

| 参数名称          | 参数类型      | 必选 | 描述             |
|---------------|-----------|----|----------------|
| bk_biz_id     | int64     | 是  | 业务ID           |
| id            | string    | 是  | 安全组ID          |
| manager       | string    | 否  | 负责人            |
| bak_manager   | string    | 否  | 备份负责人          |
| usage_biz_ids | int array | 否  | 使用业务列表，不支持改为-1 |

### 调用示例

```json
{
  "id": "0000001",
  "manager": "lihua",
  "bak_manager": "hanmeimei",
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
