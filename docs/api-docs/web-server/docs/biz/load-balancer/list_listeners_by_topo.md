### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：根据负载均衡拓扑条件查询监听器信息。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/listeners/by_topo/list

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
  "target_ports": [80],
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
  "target_ports": [80],
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
        "id": "00000001",
        "cloud_id": "lbl-123",
        "name": "listener-test",
        "vendor": "tcloud",
        "account_id": "0000001",
        "bk_biz_id": -1,
        "lb_id": "xxxx",
        "cloud_lb_id": "xxxx",
        "protocol": "https",
        "port": 80,
        "end_port": 0,
        "scheduler": "WRR", 
        "default_domain": "www.qq.com",
        "region": "ap-xxx",
        "zones": [
          "ap-xxx-1"
        ],
        "sni_switch": 0,
        "memo": "cvm test",
        "lb_vips": ["127.0.0.1"],
        "lb_domain": "www.xxx.com",
        "lb_region": "ap-xxx",
        "lb_network_type": "OPEN",
        "rule_domain_count": 1,
        "url_count": 1,
        "target_count": 1,
        "non_zero_weight_target_count": 1,
        "target_group_id": "0000001",
        "creator": "Jim",
        "reviser": "Jim",
        "created_at": "2023-02-12T14:47:39Z",
        "updated_at": "2023-02-12T14:55:40Z"
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

| 参数名称                         | 参数类型         | 描述                                               |
|------------------------------|--------------|--------------------------------------------------|
| id                           | string       | 资源ID                                             |
| cloud_id                     | string       | 云资源ID                                            |
| name                         | string       | 名称                                               |
| vendor                       | string       | 供应商                                              |
| account_id                   | string       | 账号ID                                             |
| bk_biz_id                    | int          | 业务ID                                             |
| lb_id                        | string       | 负载均衡ID                                           |
| cloud_lb_id                  | string       | 云负载均衡ID                                          |
| protocol                     | string       | 协议                                               |
| port                         | int          | 端口                                               |
| end_port                     | int          | 如果该字段不为0，代表监听器端口为端口段，该值为端口段末尾值                   |
| scheduler                    | string       | 均衡方式(WRR:按权重轮询 LEAST_CONN:最小连接数、IP_HASH:IP Hash) |
| default_domain               | string       | 默认域名                                             |
| region                       | string       | 地域                                               |
| zones                        | string array | 可用区数组                                            |
| sni_switch                   | int          | 是否开启SNI特性(0:关闭 1:开启)，当协议为HTTPS时必传                |
| memo                         | string       | 备注                                               |
| lb_vips                      | string array | 负载均衡VIP                                          |
| lb_domain                    | string       | 负载均衡域名                                           |
| lb_region                    | string       | 负载均衡地域                                           |
| lb_network_type              | string       | 负载均衡网络类型列表，"OPEN"(公网)，"INTERNAL"(内网)             |
| rule_domain_count            | int          | 规则域名数量                                           |
| url_count                    | int          | 规则url数量                                          |
| target_count                 | int          | rs数量                                             |
| non_zero_weight_target_count | int          | 权重不为0的rs数量                                       |
| target_group_id              | string       | 目标组ID                                            |
| creator                      | string       | 创建者                                              |
| reviser                      | string       | 修改者                                              |
| created_at                   | string       | 创建时间，标准格式：2006-01-02T15:04:05Z                   |
| updated_at                   | string       | 修改时间，标准格式：2006-01-02T15:04:05Z                   |
