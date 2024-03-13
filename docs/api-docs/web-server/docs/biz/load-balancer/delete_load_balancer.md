### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：负载均衡删除。
- 该接口功能描述：业务下删除负载均衡。

### URL

DELETE /api/v1/bizs/{bk_biz_id}/cloud/load_balancers/{id}

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述    |
|-----------|--------|----|-------|
| bk_biz_id | int64  | 是  | 业务ID  |
| id        | string | 是  | lb ID |

### 调用示例

ID 是 00000002

/api/v1/cloud/load_balancers/00000002

```json
{}
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
