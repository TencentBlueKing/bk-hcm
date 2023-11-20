### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：。
- 该接口功能描述：添加收藏。

### URL

POST /api/v1/cloud/collections/create

### 输入参数

| 参数名称     | 参数类型   | 必选 | 描述                                      |
|----------|--------|----|-----------------------------------------|
| res_type | string | 是  | 资源类型。（枚举值: cloud_selection_scheme：选型方案） |
| res_id   | string | 是  | 收藏的资源ID。                                |

### 调用示例

```json
{
  "res_type": "cloud_selection_scheme",
  "res_id": "00000001"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": ""
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
