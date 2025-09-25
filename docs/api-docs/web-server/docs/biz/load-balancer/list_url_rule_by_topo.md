### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述： 根据负载均衡拓扑条件查询URL规则信息

### URL

POST  /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/url_rules/by_topo/list

### 输入参数

| 参数名称             | 参数类型         | 必选 | 描述                                                       |
|------------------|--------------|----|----------------------------------------------------------|
| bk_biz_id        | string       | 是  | 业务id                                                     |
| vendor           | string       | 是  | 云厂商                                                      |
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
| page             |  object      | 是  | 分页设置                                                     |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint   | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint   | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

### 调用示例

#### 获取详细信息请求参数示例

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
  "target_ports": [8080],
  "page": {
    "count": false,
    "start": 0,
    "limit": 10
  }
}
```

#### 获取数量请求参数示例

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
  "target_ports": [8080],
  "page": {
    "count": true
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "id": "00000007",
        "lb_vips": ["127.0.0.1"],
        "lbl_protocol": "HTTP",
        "lbl_port": 8080,
        "rule_url": "/xxx",
        "rule_domain": "www.xxx.com",
        "target_count": 1,
        "lb_id": "00000001",
        "cloud_lbl_id": "lbl-00000001"
      }
    ]
  }
}
```

#### 获取数量返回结果示例

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
| details | array | 查询返回的数据        |

#### data.details[n]

| 参数名称         | 参数类型         | 描述       |
|--------------|--------------|----------|
| id           | string       | URL规则ID  |
| lb_vips      | string array | 负载均衡VIP  |
| lbl_protocol | string       | 监听器协议    |
| lbl_port     | int          | 监听器端口    |
| rule_url     | string       | 监听器的URL  |
| rule_domain  | string array | 监听器的域名   | 
| target_count | int          | 监听器的RS数量 |
| lb_id        | string       | 负载均衡ID   |
| cloud_lbl_id | string       | 云监听器ID   |