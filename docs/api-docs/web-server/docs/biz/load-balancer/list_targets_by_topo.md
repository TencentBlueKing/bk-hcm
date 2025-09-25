### 描述

- 该接口提供版本：v1.8.7+。
- 该接口所需权限：业务访问。
- 该接口功能描述：根据负载均衡拓扑条件查询rs信息。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/targets/by_topo/list

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
| page             | object       | 是  | 分页设置                                                     |

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
        "inst_id": "00000007",
        "inst_type": "CVM",
        "inst_name": "inst-xxxx",
        "ip": "127.0.0.1",
        "target_count": 1,
        "zone": "ap-guangzhou-6",
        "cloud_vpc_ids": ["vpc-xxxx"],
        "targets": [
          {
            "id": "tg-xxxx",
            "account_id": "0000001",
            "inst_type": "cvm",
            "ip": "127.0.0.1",
            "inst_id": "0000001",
            "cloud_inst_id": "cloud-inst-0000001",
            "inst_name": "inst-xxxx",
            "target_group_region": "ap-xxxx",
            "target_group_id": "0000001",
            "cloud_target_group_id": "cloud-tg-0000001",
            "port": 80,
            "weight": 80,
            "private_ip_address": [],
            "public_ip_address": [],
            "cloud_vpc_ids": ["vpc-123"],
            "zone": "",
            "memo": "",
            "target_group_name": "xxxx",
            "lb_id": "0000001",
            "cloud_lb_id": "lb-0000001",
            "lb_vips": ["127.0.0.1"],
            "lb_domain": "www.xxx.com",
            "lb_region": "ap-xxx",
            "lb_network_type": "OPEN",
            "lbl_id": "0000001",
            "lbl_port": 80,
            "lbl_end_port": 0,
            "lbl_name": "xxx",
            "lbl_protocol": "HTTP",
            "rule_id": "0000001",
            "rule_url": "/xxx",
            "rule_domain": "www.xxx.com",
            "creator": "Jim",
            "reviser": "Jim",
            "created_at": "2023-02-12T14:47:39Z",
            "updated_at": "2023-02-12T14:55:40Z"
          }
        ]
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

| 参数名称          | 参数类型         | 描述                                      |
|---------------|--------------|-----------------------------------------|
| inst_id       | string       | 实例id                                    |
| inst_type     | string       | 实例类型，"CVM"(云服务器)、"ENI"(弹性网卡)、"CCN"(云联网) |
| inst_name     | string       | 实例名称                                    |
| ip            | string       | ip                                      |
| zone          | string       | 可用区                                     |
| cloud_vpc_ids | string array | 云VpcID列表                                | 
| targets       | object array | rs列表                                    |


#### data.details[n].targets[n]

| 参数名称                 | 参数类型         | 描述                                                |
|----------------------|--------------|---------------------------------------------------|
| id                   | string       | rs id                                             |
| account_id           | string       | 账号ID                                              |
| inst_type            | string       | 实例类型                                              |
| ip                   | string       | rs绑定的ip                                           |
| inst_id              | string       | 实例ID                                              |
| cloud_inst_id        | string       | 云实例ID                                             |
| inst_name            | string       | 实例名称                                              |
| target_group_region  | string       | 目标组地域                                             |
| target_group_id      | string       | 目标组id                                             |
| port                 | int          | 端口                                                |
| weight               | int          | 权重                                                |
| private_ip_addresses | string array | 内网IP地址                                            |
| public_ip_addresses  | string array | 外网IP地址                                            |
| cloud_vpc_ids        | string array | 云VpcID列表                                          |
| zone                 | string       | 可用区                                               |
| memo                 | string       | 备注                                                |
| target_group_name    | string       | 目标组名称                                             |
| lb_id                | string       | 负载均衡ID                                            |
| cloud_lb_id          | string       | 云负载均衡ID                                           |
| lb_vips              | string array | 负载均衡VIP列表                                         |
| lb_domain            | string       | 负载均衡域名                                            |
| lb_region            | string       | 负载均衡地域                                            |
| lb_network_type      | string       | 负载均衡网络类型列表，"OPEN"(公网)，"INTERNAL"(内网)              |
| lbl_id               | string       | 监听器ID                                             |
| lbl_port             | int          | 监听器端口                                             |
| lbl_end_port         | int          | 如果该字段不为0，代表监听器端口为端口段，该值为端口段末尾值                    |
| lbl_name             | string       | 监听器名称                                             |
| lbl_protocol         | string       | 监听器协议，"HTTP"、"HTTPS"、"TCP"、"UDP"、"TCP_SSL"、"QUIC" |
| rule_id              | string       | 规则ID                                              |
| rule_url             | string       | 规则url                                             |
| rule_domain          | string       | 规则域名                                              |
| creator              | string       | 创建者                                               |
| reviser              | string       | 修改者                                               |
| created_at           | string       | 创建时间，标准格式：2006-01-02T15:04:05Z                    |
| updated_at           | string       | 修改时间，标准格式：2006-01-02T15:04:05Z                    |
