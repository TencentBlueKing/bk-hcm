### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源-资源分配。
- 该接口功能描述：批量分配安全组到业务。

### URL

POST /api/v1/cloud/security_groups/assign/bizs/batch

### 输入参数

| 参数名称            | 参数类型         | 必选 | 描述       |
|-----------------|--------------|----|----------|
| security_groups | object array | 是  | 分配的安全组列表 |

#### security_groups[n]

| 参数名称              | 参数类型   | 必选 | 描述    |
|-------------------|--------|----|-------|
| security_group_id | string | 是  | 安全组ID |
| bk_biz_id         | int64  | 是  | 业务ID  |

### 调用示例

```json
{
  "security_groups": [
    {
      "security_group_ids": "00000001",
      "bk_biz_id": 100
    }
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
