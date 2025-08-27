### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：根据负载均衡拓扑条件查询rs数量。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/targets/by_topo/count

### 输入参数

| 参数名称             | 参数类型         | 必选 | 描述                                                       |
|------------------|--------------|----|----------------------------------------------------------|
| bk_biz_id        | string       | 是  | 业务id                                                     |
| vendor           | string       | 是  | 云厂商(枚举值：tcloud)                                          |
| account_id       | string       | 是  | 云账号ID                                                    |
| lb_regions       | string array | 否  | CLB所在地域列表，长度限制500                                        |
| lb_network_types | string array | 否  | 负载均衡网络类型列表，"OPEN"(公网)，"INTERNAL"(内网)                     |
| lb_ip_versions   | string array | 否  | 负载均衡IP版本列表，如"ipv4"、"ipv6"、"ipv6_nat64"、"ipv6_dual_stack" |
| cloud_lb_ids     | string array | 否  | 云负载均衡ID列表，长度限制500                                        |
| lb_vips          | string array | 否  | 负载均衡VIP列表，长度限制500                                        |
| lb_domains       | string array | 否  | 负载均衡域名列表，长度限制500                                         |
| lbl_protocols    | string array | 否  | 监听器协议列表，"HTTP"、"HTTPS"、"TCP"、"UDP"、"TCP_SSL"、"QUIC"      |
| lbl_ports        | int array    | 否  | 监听器端口列表，长度限制1000                                         |
| rule_domains     | string array | 否  | 规则域名列表，长度限制500                                           |
| rule_urls        | string array | 否  | 规则url列表，长度限制500                                          |
| target_ips       | string array | 否  | rs ip列表，长度限制5000                                         |
| target_ports     | int array    | 否  | rs port列表，长度限制500                                        |

### 调用示例

```json
{
  "account_id": "0000001",
  "lb_regions": ["ap-guangzhou"],
  "lb_network_types": ["OPEN"],
  "lb_ip_versions": ["ipv4"],
  "cloud_lb_ids": ["lb-0000001"],
  "lb_vips": ["127.0.0.1"],
  "lb_domains": ["www.xxx.com"],
  "lbl_protocols": ["HTTP"],
  "lbl_ports": [8080],
  "rule_domains": ["www.xxx.com"],
  "rule_urls": ["/xxx"],
  "target_ips": ["127.0.0.1"],
  "target_ports": [80]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "count": 1
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

| 参数名称    | 参数类型  | 描述             |
|---------|-------|----------------|
| count   | int   | 当前规则能匹配到的总记录条数 |
