### 描述

- 该接口提供版本：v1.8.0+。
- 该接口所需权限：业务-IaaS资源操作。
- 该接口功能描述：安全组批量解绑主机

### URL

POST /api/v1/cloud/security_groups/disassociate/cvms/batch

### 输入参数

| 参数名称              | 参数类型         | 必选 | 描述     |
|-------------------|--------------|----|--------|
| security_group_id | string       | 是  | 安全组ID  |
| cvm_ids           | string array | 是  | 主机ID列表 |

### 调用示例

```json
{
  "security_group_id": "00001111",
  "cvm_ids": ["ins-xxxx"]
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
