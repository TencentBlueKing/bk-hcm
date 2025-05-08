### 描述

- 该接口提供版本：v1.8.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询安全组关联的云上资源数量。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/security_groups/related_resources/query_count

### 输入参数

| 参数名称      | 参数类型         | 必选 | 描述            |
|-----------|--------------|----|---------------|
| bk_biz_id | int64        | 是  | 业务ID          |
| ids       | string array | 是  | 安全组ID列表，最大100 |

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

#### 查询成功
```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": "00000001",
      "resources": [
        {
          "res_name": "cvm",
          "count": 10
        },
        {
          "res_name": "load_balancer",
          "count": 2
        },
        {
          "res_name": "db",
          "count": 0
        },
        {
          "res_name": "container",
          "count": 0
        }
      ]
    }
  ]
}
```

#### 部分失败
```json
{
  "code": 0,
  "message": "ok",
  "data": [
    
    {
      "id": "00000002",
      "error": "resource not found"
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

| 参数名称      | 参数类型         | 描述         |
|-----------|--------------|------------|
| id        | string       | 安全组ID      |
| error     | string       | 云上查询查询错误信息 |
| resources | object array | 安全组关联的资源数量 |

#### resource[n]

| 参数名称     | 参数类型   | 描述      |
|----------|--------|---------|
| res_name | string | 关联的资源名称 |
| count    | int    | 关联的资源数量 |
