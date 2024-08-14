### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下删除腾讯云负载均衡跨域2.0 SNAT IP。

### URL

DELETE /api/v1/cloud/bizs/{bk_biz_id}/vendors/tcloud/load_balancers/{lb_id}/snat_ips

### 输入参数

| 参数名称       | 参数类型         | 必选 | 描述             |
|------------|--------------|----|----------------|
| bk_biz_id  | int64        | 是  | 业务ID           |
| lb_id      | string       | 是  | 负载均衡id         |
| delete_ips | string array | 是  | 待删除SNAT IP地址数组 |


### 调用示例

```json
{
  "delete_ips": [
    "192.168.0.1",
    "192.168.0.2"
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
