### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源-资源查看。
- 该接口功能描述：查询安全组关联的云上资源数量。

### URL

POST /api/v1/cloud/security_groups/related_resources/query_count

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
      "cvm": 10,
      "clb": 2,
      "db": 0,
      "container": 0
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

| 参数名称      | 参数类型   | 描述          |
|-----------|--------|-------------|
| id        | string | 安全组ID       |
| cvm       | int    | 安全组关联的CVM数量 |
| clb       | int    | 安全组关联的CLB数量 |
| db        | int    | 安全组关联的DB数量  |
| container | int    | 安全组关联的容器数量  |
