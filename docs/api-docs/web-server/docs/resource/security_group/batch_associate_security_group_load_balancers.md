### 描述

- 该接口提供版本：v1.8.9+。
- 该接口所需权限：资源-IaaS资源操作。
- 该接口功能描述：安全组批量绑定CLB

### URL

POST /api/v1/cloud/security_groups/associate/load_balancers/batch

### 输入参数

| 参数名称           | 参数类型       | 必选 | 描述                                  |
|-------------------|--------------|------|--------------------------------------|
| security_group_id | string       | 是   | 安全组ID                               |
| load_balancer_ids | string array | 是   | 负载均衡ID列表，最大20个                 |

### 调用示例

```json
{
  "security_group_id": "00001111",
  "load_balancer_ids": ["00000001"]
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

| 参数名称  | 参数类型  | 描述    |
|---------|----------|---------|
| code    | int      | 状态码   |
| message | string   | 请求信息 |
