### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询子网的IP地址数量信息列表（**注意：仅供前端使用，GCP暂不支持**）。

### URL

POST /api/v1/cloud/subnets/{id}/ips/count

### 输入参数

| 参数名称 | 参数类型   | 必选  | 描述   |
|------|--------|-----|------|
| id   | string | 是   | 子网ID |

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "available_ipv4_count": 1
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

| 参数名称                 | 参数类型   | 描述             |
|----------------------|--------|----------------|
| available_ipv4_count | uint64 | 子网里可用的IPv4地址数量 |
| total_ip_address_count | uint64 | 子网里IPv4地址总量 |
| used_ip_address_count | uint64 | 子网里已经被使用的IPv4地址数量 |
