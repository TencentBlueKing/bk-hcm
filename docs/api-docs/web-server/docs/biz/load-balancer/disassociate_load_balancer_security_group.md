### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限： 负载均衡操作。
- 该接口功能描述：给指定的负载均衡，取消与安全组的关联（仅支持：tcloud）。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/security_groups/disassociate/load_balancers

### 输入参数

| 参数名称              | 参数类型   | 必选 | 描述     |
|-------------------|--------|----|--------|
| bk_biz_id         | int64  | 是  | 业务ID   |
| lb_id             | string | 是  | 负载均衡ID |
| security_group_id | string | 是  | 安全组ID  |

### 调用示例

```json
{
  "lb_id": "00001112",
  "security_group_id": "00001111"
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
