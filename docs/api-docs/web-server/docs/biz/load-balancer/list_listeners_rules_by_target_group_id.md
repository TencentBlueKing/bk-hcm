### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询指定的目标组绑定的监听器/规则 列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/tcloud/target_groups/{target_group_id}/rules/list

### 输入参数

| 参数名称            | 参数类型   | 必选 | 描述     |
|-----------------|--------|----|--------|
| bk_biz_id       | int    | 是  | 业务ID   |
| target_group_id | string | 是  | 目标组ID  |
| filter          | object | 否  | 查询过滤条件 |
| page            | object | 是  | 分页设置   |

#### filter

| 参数名称  | 参数类型        | 必选 | 描述                                                              |
|-------|-------------|----|-----------------------------------------------------------------|
| op    | enum string | 是  | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是  | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选 | 描述                                          |
|-------|-------------|----|---------------------------------------------|
| field | string      | 是  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| op    | enum string | 是  | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis） |
| value | 可变类型        | 是  | 查询条件Value值                                  |

##### rules 表达式说明：

##### 1. 操作符

| 操作符 | 描述                                        | 操作符的value支持的数据类型                              |
|-----|-------------------------------------------|-----------------------------------------------|
| eq  | 等于。不能为空字符串                                | boolean, numeric, string                      |
| neq | 不等。不能为空字符串                                | boolean, numeric, string                      |
| gt  | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte | 大于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lt  | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lte | 小于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in  | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string                      |
| nin | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string                      |
| cs  | 模糊查询，区分大小写                                | string                                        |
| cis | 模糊查询，不区分大小写                               | string                                        |

##### 2. 协议示例

查询 name 是 "Jim" 且 age 大于18小于30 且 servers 类型是 "api" 或者是 "web" 的数据。

```json
{
  "op": "and",
  "rules": [
    {
      "field": "name",
      "op": "eq",
      "value": "Jim"
    },
    {
      "field": "age",
      "op": "gt",
      "value": 18
    },
    {
      "field": "age",
      "op": "lt",
      "value": 30
    },
    {
      "field": "servers",
      "op": "in",
      "value": [
        "api",
        "web"
      ]
    }
  ]
}
```

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint   | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint   | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

#### 查询参数介绍：

| 参数名称         | 参数类型   | 描述                             |
|--------------|--------|--------------------------------|
| id           | string | 资源ID                           |
| cloud_id     | string | 云资源ID                          |
| name         | string | 资源名称                           |
| lbl_id       | string | 监听器ID                          |
| cloud_lbl_id | string | 云监听器ID                         |
| lb_id        | string | 负载均衡ID                         |
| cloud_lb_id  | string | 云负载均衡ID                        |
| url          | string | 关联的URL                         |
| memo         | string | 备注                             |
| creator      | string | 创建者                            |
| reviser      | string | 修改者                            |
| created_at   | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at   | string | 修改时间，标准格式：2006-01-02T15:04:05Z |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 获取详细信息请求参数示例

查询创建者是Jim的监听器列表。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "created_at",
        "op": "eq",
        "value": "Jim"
      }
    ]
  },
  "page": {
    "count": false,
    "start": 0,
    "limit": 10
  }
}
```

#### 获取数量请求参数示例

查询创建者是Jim的监听器数量。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "created_at",
        "op": "eq",
        "value": "Jim"
      }
    ]
  },
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
        "cloud_id": "loc-123",
        "name": "loc-test",
        "lbl_id": "xxxx",
        "lbl_name": "xxxx",
        "cloud_lbl_id": "lbl-xxxx",
        "lb_id": "xxxx",
        "lb_name": "xxxx",
        "cloud_lb_id": "lb-xxxx",
        "target_group_id": "xxxx",
        "cloud_target_group_id": "cloud-tg-xxxx",
        "private_ipv4_addresses": [
          "127.0.0.1"
        ],
        "private_ipv6_addresses": [],
        "public_ipv4_addresses": [
          "127.0.0.1"
        ],
        "public_ipv6_addresses": [],
        "protocol": "https",
        "port": 80,
        "domain": "www.qq.com",
        "url": "/",
        "scheduler": "WRR",
        "sni_switch": 0,
        "session_type": "NORMAL",
        "session_expire": 0,
        "health_check": {
          "health_switch": 1,
          "time_out": 2,
          "interval_time": 5,
          "health_num": 3,
          "un_health_num": 3,
          "check_port": 80,
          "check_type": "HTTP",
          "http_version": "HTTP/1.0",
          "http_check_path": "/",
          "http_check_domain": "www.weixin.com",
          "http_check_method": "GET",
          "source_ip_type": 1
        },
        "certificate": {
          "ssl_mode": "MUTUAL",
          "ca_cloud_id": "ca-001",
          "cert_cloud_ids": [
            "cert-001"
          ]
        },
        "inst_type": "cvm",
        "vpc_id": "vpc-123",
        "vpc_name": "vpc-name",
        "cloud_vpc_id": "cloud-vpc-123",
        "memo": "listener test",
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
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称    | 参数类型  | 描述             |
|---------|-------|----------------|
| count   | int   | 当前规则能匹配到的总记录条数 |
| details | array | 查询返回的数据        |

#### data.details[n]

| 参数名称                   | 参数类型         | 描述                             |
|------------------------|--------------|--------------------------------|
| id                     | string       | 资源ID                           |
| cloud_id               | string       | 云资源ID                          |
| name                   | string       | 资源名称                           |
| rule_type              | string       | URL规则类型                        |
| lbl_id                 | string       | 监听器ID                          |
| lbl_name               | string       | 监听器名称                          |
| cloud_lbl_id           | string       | 云监听器ID                         |
| lb_id                  | string       | 负载均衡ID                         |
| lb_name                | string       | 云负载均衡名称                        |
| cloud_lb_id            | string       | 云负载均衡ID                        |
| target_group_id        | string       | 目标组ID                          |
| cloud_target_group_id  | string       | 云目标组ID                         |
| private_ipv4_addresses | string array | 负载均衡的内网IPv4地址                  |
| private_ipv6_addresses | string array | 负载均衡的内网IPv6地址                  |
| public_ipv4_addresses  | string array | 负载均衡的外网IPv4地址                  |
| public_ipv6_addresses  | string array | 负载均衡的外网IPv6地址                  |
| protocol               | string       | 协议                             |
| port                   | string       | 端口                             |
| domain                 | string       | 关联的域名                          |
| url                    | string       | 关联的URL                         |
| scheduler              | string       | 均衡方式                           |
| sni_switch             | int          | 是否开启SNI特性，此参数仅适用于HTTPS监听器      |
| session_type           | string       | 会话保持类型                         |
| session_expire         | int          | 会话保持时间，0为关闭                    |
| health_check           | object       | 健康检查                           |
| certificate            | object       | 证书信息                           |
| inst_type              | string       | 资源实例类型                         |
| vpc_id                 | string       | VPCID                          |
| vpc_name               | string       | VPC名称                          |
| cloud_vpc_id           | string       | 云VPCID                         |
| memo                   | string       | 备注                             |
| creator                | string       | 创建者                            |
| reviser                | string       | 修改者                            |
| created_at             | string       | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at             | string       | 修改时间，标准格式：2006-01-02T15:04:05Z |

### health_check

| 参数名称           | 参数类型 | 描述                                                                    |
|-------------------|--------|-------------------------------------------------------------------------|
| health_switch     | int    | 是否开启健康检查：1（开启）、0（关闭）                                        |
| time_out          | int    | 健康检查的响应超时时间，可选值：2~60，单位：秒                                 |
| interval_time     | int    | 健康检查探测间隔时间                                                        |
| health_num        | int    | 健康阈值                                                                  |
| un_health_num     | int    | 不健康阈值                                                                |
| check_port        | int    | 自定义探测相关参数。健康检查端口，默认为后端服务的端口                            |
| check_type        | string | 健康检查使用的协议。取值 TCP/HTTP/HTTPS/GRPC/PING/CUSTOM                     |
| http_code         | string | 健康检查类型                                                               |
| http_version      | string | HTTP版本                                                                  |
| http_check_path   | string | 健康检查路径（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）            |
| http_check_domain | string | 健康检查域名                                                                |
| http_check_method | string | 健康检查方法（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式），默认值：HEAD，可选值HEAD或GET |
| source_ip_type    | string | 健康检查源IP类型：0（使用LB的VIP作为源IP），1（使用100.64网段IP作为源IP）         |
| context_type      | string | 健康检查的输入格式，可取值：HEX或TEXT；                                        |

### http_code 取值说明

| 值  | 说明                |
|----|-------------------|
| 1  | 表示探测后返回值 1xx 代表健康 |
| 2  | 表示返回 2xx 代表健康     |
| 4  | 表示返回 3xx 代表健康     |
| 8  | 表示返回 4xx 代表健康，    |
| 16 | 表示返回 5xx 代表健康。    |

若希望多种返回码都可代表健康，则将相应的值相加。

### certificate

| 参数名称          | 参数类型       | 描述                                            |
|------------------|--------------|-------------------------------------------------|
| ssl_mode         | string       | 认证类型，UNIDIRECTIONAL：单向认证，MUTUAL：双向认证 |
| ca_cloud_id      | string       | CA证书的云ID                                     |
| cert_cloud_ids   | string array | 服务端证书的云ID                                  |
