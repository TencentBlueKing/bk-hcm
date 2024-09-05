### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下增加标准运维中指定的负载均衡规则。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/sops/rule/online

### 输入参数

| 参数名称          | 参数类型       | 必选 | 描述      |
|------------------|--------------|------|---------|
| bk_biz_id        | int          | 是   | 业务ID    |
| account_id       | string       | 是   | 账号ID    |
| bind_rs_records  | object array | 是   | 规则具体的信息 |

#### TODO bind_rs_records（clb excel导入的部分，需clb excel导入修改完后同步修改到这里）

| 参数名称        | 参数类型      | 必选 | 描述                                        |
|----------------|-----------|----|-------------------------------------------|
| action         | string    | 是  | 批处理操作类型                               |
| name           | string    | 是  | 监听器名称                                    |
| protocol       | string    | 是  | 协议类型                                      |
| ip_domain_type | string    | 是  | IP或域名类型                                   |
| vip            | string    | 是  | 负载均衡ip地址                                     |
| vports         | []int     | 是  | 负载均衡监听器端口列表                                   |
| have_end_port  | bool      | 是  | 是否是端口端                                   |
| domain         | string    | 是  | 域名                                         |
| url            | string    | 是  | URL路径                                      |
| cert_cloud_ids | []string  | 是  | 服务器证书云ID列表 |
| ca_cloud_id    | string    | 是  | 客户端证书云ID                                  |
| inst_type      | string    | 是  | 后端实例类型（CVM、ENI）                         |
| rs_ips         | []string  | 是  | 后端实例IP列表                                  |
| rs_ports       | []int     | 是  | 后端实例端口列表                                 |
| weight         | []int     | 是  | 权重列表                                       |
| scheduler      | string    | 是  | 负载均衡调度算法                                 |
| session_expired| int64     | 是  | 会话保持时间（单位：秒）                          |
| health_check   | bool      | 是  | 是否开启健康检查                                  |

### 调用示例

```json
{
  "account_id": "xxx",
  "bind_rs_records": [
    {
      "action": "listener_url_rs",
      "name": "test_listener",
      "protocol": "HTTP",
      "ip_domain_type": "IPv4",
      "vip": "xxx.xxx.xxx.xxx",
      "vports": [8080],
      "have_end_port": false,
      "domain": "www.xxx.com",
      "url": "/path",
      "cert_cloud_ids": ["xxx", "xxx"],
      "ca_cloud_id": "xxx",
      "inst_type": "CVM",
      "rs_ips": ["xxx.xxx.xxx.xxx"],
      "rs_ports": [8080],
      "weight": [10, 20],
      "scheduler": "WRR",
      "session_expired": 3600,
      "health_check": true
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "batch_operation_id": "xxxxxxxx"
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

| 参数名称               | 参数类型 | 描述     |
|--------------------|--------|--------|
| batch_operation_id | string | 批量操作id |
