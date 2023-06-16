### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：。
- 该接口功能描述：查询全部Cmdb管控区域接口。

### URL

POST /api/v1/web/all/cloud_areas/list

### 输入参数

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "cound": 1,
    "info": [
      {
        "id": "00000001",
        "name": "tcloud"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称  | 参数类型         | 描述      |
|-------|--------------|---------|
| count | uint64       | 总记录条数   |
| info  | object array | 查询返回的数据 |

#### data.detail[n]

| 参数名称 | 参数类型   | 描述     |
|------|--------|--------|
| id   | string | VPC的ID |
| name | string | 管控区域名称 |
