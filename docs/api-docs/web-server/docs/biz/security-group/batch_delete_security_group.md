### 描述

- 该接口提供版本：v1.0.0+。
- 所需权限：业务-IaaS资源删除。
- 功能描述：批量删除安全组。

### URL

DELETE /api/v1/cloud/bizs/{bk_biz_id}/security_groups/batch

### 输入参数

| 参数名称        | 参数类型         | 必选    | 描述       |
|-------------|--------------|-------|----------|
| bk_biz_id   | int64        | 是     | 业务ID     |
| ids         | string array | 是     | 安全组ID列表  |

### 调用示例

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

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
