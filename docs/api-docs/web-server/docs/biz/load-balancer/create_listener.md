### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：监听器创建。
- 该接口功能描述：业务下创建监听器。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/{lb_id}/listeners/create

### 输入参数

| 参数名称          | 参数类型       | 必选 | 描述              |
|------------------|--------------|------|------------------|
| bk_biz_id        | int          | 是   | 业务ID            |
| account_id       | string       | 是   | 账号ID            |
| lb_id            | string       | 是   | 负载均衡ID         |
| name             | string       | 是   | 名称              |
| protocol         | string       | 是   | 协议              |
| port             | int          | 是   | 端口              |
| scheduler        | string       | 是   | 均衡方式(WRR:按权重轮询 LEAST_CONN:最小连接数、IP_HASH:IP Hash) |
| session_type     | string       | 是   | 会话保持类型(NORMAL表示默认会话保持类型。QUIC_CID表示根据Quic Connection ID做会话保持) |
| session_expire   | int          | 是   | 会话保持时间，最小值30秒 |
| target_group_id  | string       | 是   | 目标组ID           |
| domain           | string       | 否   | 默认域名，当协议为HTTP、HTTPS时必传               |
| url              | string       | 否   | URL路径，当协议为HTTP、HTTPS时必传                |
| sni_switch       | int          | 否   | 是否开启SNI特性(0:关闭 1:开启)，当协议为HTTPS时必传 |
| certificate      | object       | 否   | 证书信息，当协议为HTTPS时必传                     |

### certificate

| 参数名称          | 参数类型       | 描述                                   |
|------------------|--------------|--------------------------------------|
| ssl_mode         | string       | 认证类型，UNIDIRECTIONAL：单向认证，MUTUAL：双向认证 |
| ca_cloud_id      | string       | CA证书的ID                              |
| cert_cloud_ids | string array | 服务端证书的ID数组                           |

### 调用示例

```json
{
  "account_id": "0000001",
  "name": "xxx",
  "protocol": "TCP",
  "port": 22,
  "scheduler": "WRR",
  "session_type": "NORMAL",
  "session_expire": 30,
  "target_group_id": "00000001",
  "domain": "www.xxxx.com",
  "url": "/api/url"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000001"
  }
}
```

### 响应参数说明

| 参数名称  | 参数类型  | 描述    |
|---------|----------|---------|
| code    | int      | 状态码   |
| message | string   | 请求信息 |
| data    | object   | 响应数据 |

#### data

| 参数名称  | 参数类型 | 描述    |
|----------|--------|---------|
| id       | string | 监听器id |
