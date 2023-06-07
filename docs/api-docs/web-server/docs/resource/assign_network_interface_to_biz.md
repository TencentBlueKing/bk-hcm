### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源分配。
- 该接口功能描述：分配网络接口到业务下。

### URL

POST /api/v1/cloud/network_interfaces/assign/bizs

### 输入参数

| 参数名称      | 参数类型         | 必选  | 描述       |
|-----------|--------------|-----|--------------------|
| network_interface_ids | string array | 是   | 网络接口的ID列表 |
| bk_biz_id | int64     | 是   | 业务的ID    |

### 调用示例

```json
{
  "network_interface_ids": [
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

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
