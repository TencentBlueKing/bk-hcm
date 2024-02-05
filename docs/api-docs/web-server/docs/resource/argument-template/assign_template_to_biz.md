### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源分配。
- 该接口功能描述：分配参数模版到业务下。

### URL

POST /api/v1/cloud/argument_templates/assign/bizs

### 输入参数

| 参数名称      | 参数类型       | 必选  | 描述            |
|--------------|--------------|-------|----------------|
| template_ids | string array | 是    | 参数模版的ID列表 |
| bk_biz_id    | int          | 是    | 业务的ID        |

### 调用示例

```json
{
    "template_ids": [
        "00000001",
        "00000002"
    ],
    "bk_biz_id": 3
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

| 参数名称  | 参数类型 | 描述   |
|---------|---------|--------|
| code    | int     | 状态码  |
| message | string  | 请求信息 |
