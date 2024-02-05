### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：IaaS资源删除。
- 该接口功能描述：业务下删除参数模版。

### URL

DELETE /api/v1/cloud/bizs/{bk_biz_id}/argument_templates/batch

### 输入参数

| 参数名称   | 参数类型       | 必选  | 描述            |
|-----------|--------------|-------|----------------|
| bk_biz_id | int64        | 是    | 业务ID          |
| ids       | string array | 是    | 参数模版的ID列表  |

### 调用示例s

```json
{
  "ids": [
    "1",
    "2",
    "3"
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

| 参数名称  | 参数类型 | 描述    |
|---------|---------|---------|
| code    | int     | 状态码   |
| message | string  | 请求信息 |