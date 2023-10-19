### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：IaaS资源删除。
- 该接口功能描述：删除Azure安全组规则。

### URL

DELETE /api/v1/cloud/vendors/azure/security_groups/{security_group_id}/rules/{id}

### 输入参数

| 参数名称              | 参数类型   | 必选  | 描述      |
|-------------------|--------|-----|---------|
| security_group_id | string | 是   | 安全组ID   |
| id                | string | 是   | 安全组规则ID |

### 调用示例

```json

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
