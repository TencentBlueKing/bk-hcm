### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源-资源分配。
- 该接口功能描述：预览安全组分配到业务的结果，是否可分配。

### URL

POST /api/v1/cloud/security_groups/assign/bizs/preview

### 输入参数

| 参数名称 | 参数类型         | 必选 | 描述            |
|------|--------------|----|---------------|
| ids  | string array | 是  | 安全组ID列表，最大100 |

### 调用示例

```json
{
  "ids": [
    "00000001",
    "00000002"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": "00000001",
      "assignable": false,
      "reason": "无一级业务标签",
      "assigned_biz_id": 0
    },
    {
      "id": "00000002",
      "assignable": true,
      "reason": "",
      "assigned_biz_id": 213
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述   |
|---------|--------------|------|
| code    | int32        | 状态码  |
| message | string       | 请求信息 |
| data    | object array | 响应数据 |

#### data[n]

| 参数名称            | 参数类型   | 描述                |
|-----------------|--------|-------------------|
| id              | string | 安全组ID             |
| assignable      | bool   | 是否可分配             |
| reason          | string | 不可分配原因（不可分配时）     |
| assigned_biz_id | int    | 将要被分配到的业务ID（可分配时） |
