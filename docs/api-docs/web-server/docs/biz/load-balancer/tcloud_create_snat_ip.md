### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下新增腾讯云负载均衡跨域2.0 SNAT IP。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/tcloud/load_balancers/{lb_id}/snat_ips/create

### 输入参数

| 参数名称      | 参数类型          | 必选 | 描述            |
|-----------|---------------|----|---------------|
| bk_biz_id | int64         | 是  | 业务ID          |
| lb_id     | string        | 是  | 负载均衡id        |
| snat_ips  | snat_ip array | 是  | 待创建SNAT IP 数组 |

### snat_ip

| 参数名称      | 参数类型   | 描述                   |
|-----------|--------|----------------------|
| subnet_id | string | SNAT IP 所在子网cloud_id |
| ip        | string | 指定IP，留空自动生成          |

### 调用示例

```json
{
  "snat_ips": [
    {
      "subnet_id": "subnet-abcd"
    },
    {
      "subnet_id": "subnet-1234"
    }
  ]
}
```

### 响应示例


```json
{
  "code": 0,
  "message": "",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |
