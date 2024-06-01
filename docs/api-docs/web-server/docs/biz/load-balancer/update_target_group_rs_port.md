### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下批量修改RS端口。

### URL

PATCH /api/v1/cloud/bizs/{bk_biz_id}/target_groups/{target_group_id}/targets/port

### 输入参数

| 参数名称          | 参数类型       | 必选 | 描述                    |
|------------------|--------------|------|------------------------|
| bk_biz_id        | int          | 是   | 业务ID                  |
| target_group_id  | string       | 是   | 目标组ID                 |
| target_ids       | string array | 是   | 目标ID数组，单次最多20个  |
| new_port         | int          | 是   | 新端口                   |

### 调用示例

```json
{
  "target_ids": ["00000001"],
  "new_port": 8081
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "flow_id": "xxxxxxxx"
  }
}
```

### 响应参数说明

| 参数名称  | 参数类型  | 描述    |
|---------|----------|---------|
| code    | int      | 状态码   |
| message | string   | 请求信息 |
| data    | object   | 响应数据 |

#### data

| 参数名称  | 参数类型 | 描述    |
|----------|--------|---------|
| flow_id  | string | 任务id   |

