### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：。
- 该接口功能描述：查询某类资源的收藏列表。

### URL

GET /api/v1/cloud/collections/{res_type}/list

### 输入参数

| 参数名称     | 参数类型   | 必选 | 描述   |
|----------|--------|----|------|
| res_type | string | 是  | 资源类型 |

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "id": "00000001",
      "user": "Jim",
      "res_type": "cloud_selection_scheme",
      "res_id": "00000001",
      "creator": "Jim",
      "created_at": "2023-02-05T15:29:15Z"
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

| 参数名称       | 参数类型   | 描述                             |
|------------|--------|--------------------------------|
| id         | uint64 | 收藏记录ID                         |
| user       | string | 收藏的用户                          |
| res_type   | string | 收藏的资源类型                        |
| res_id     | string | 收藏的资源ID                        |
| creator    | string | 创建者                            |
| created_at | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
