### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：IaaS资源访问。
- 该接口功能描述：查询虚拟机关联资源。(仅供前端使用)

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/cvms/rel_res/batch

### 输入参数

| 参数名称 | 参数类型         | 必选 | 描述      |
|------|--------------|----|---------|
| ids  | string array | 是  | 虚拟机ID列表 |

### 调用示例

```json
{
  "ids": [
    "000000001"
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
      "id": "000000001",
      "disk_count": 1,
      "eip_count": 0,
      "eip": [
        "127.0.0.1"
      ]
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data[n]

| 参数名称       | 参数类型         | 描述          |
|------------|--------------|-------------|
| id         | string       | 虚拟机ID       |
| disk_count | int          | 磁盘数量（包含系统盘） |
| eip_count  | int          | EIP数量       |
| eip        | string array | EIP列表       |
