### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下给指定目标组批量移除RS

### URL

DELETE /api/v1/cloud/bizs/{bk_biz_id}/target_groups/{target_group_id}/rs/batch

### 输入参数

| 参数名称          | 参数类型       | 必选 | 描述                   |
|------------------|--------------|------|-----------------------|
| bk_biz_id        | int          | 是   | 业务ID                 |
| target_group_id  | string       | 是   | 目标组ID                |
| target_ids       | string array | 是   | RS的ID列表，单次最多100个 |

### 调用示例

```json
{
  "target_ids": [
    "00000001",
    "00000002"
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
