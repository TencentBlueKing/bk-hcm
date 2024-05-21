### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：给指定的负载均衡，批量关联安全组（仅支持：tcloud）。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/security_groups/associate/load_balancers

### 输入参数

| 参数名称               | 参数类型         | 必选 | 描述                  |
|--------------------|--------------|----|---------------------|
| bk_biz_id          | int64        | 是  | 业务ID                |
| lb_id              | string       | 是  | 负载均衡的ID             |
| security_group_ids | string array | 是  | 安全组的ID数组，最多支持50个安全组 |

### 调用示例

```json
{
  "lb_id": "00001112",
  "security_group_ids": [
    "00000002",
    "00000003"
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
| code    | int    | 状态码  |
| message | string | 请求信息 |
