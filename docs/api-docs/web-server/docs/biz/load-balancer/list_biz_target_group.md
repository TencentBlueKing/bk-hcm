### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：业务下查询目标组。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/target_groups/list

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述     |
|-----------|--------|----|--------|
| bk_biz_id | int    | 是  | 业务ID   |
| filter    | object | 是  | 查询过滤条件 |
| page      | object | 是  | 分页设置   |

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
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

#### 查询参数介绍：

| 参数名称              | 参数类型         | 描述                             |
|-------------------|--------------|--------------------------------|
| id                | string       | 资源ID                           |
| cloud_id          | string       | 云资源ID                          |
| name              | string       | 名称                             |
| vendor            | string       | 供应商（枚举值：tcloud）                |
| account_id        | string       | 账号ID                           |
| target_group_type | string       | 目标组类型                          |
| region            | string       | 地域                             |
| protocol          | string       | 协议                             |
| port              | int          | 端口                             |
| weight            | int          | 权重                             |
| cloud_vpc_id      | string       | 云vpcID                         |
| vpc_id            | string       | vpcID                          |
| memo              | string       | 备注                             |
| creator           | string       | 创建者                            |
| reviser           | string       | 修改者                            |
| created_at        | string       | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at        | string       | 修改时间，标准格式：2006-01-02T15:04:05Z |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 获取详细信息请求参数示例

查询创建者是Jim的目标组列表。

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
    "limit": 500
  }
}
```

#### 获取数量请求参数示例

查询创建者是Jim的目标组数量。

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
        "cloud_id": "clb-123",
        "name": "clb-test",
        "vendor": "tcloud",
        "account_id": "0000001",
        "bk_biz_id": -1,
        "target_group_type": "local",
        "region": "ap-hk",
        "protocol": "TCP",
        "port": 22,
        "weight": 22,
        "memo": "memo-test",
        "creator": "Jim",
        "reviser": "Jim",
        "created_at": "2023-02-12T14:47:39Z",
        "updated_at": "2023-02-12T14:55:40Z",
        "lb_id": "xxxx",
        "lb_name": "xxxx",
        "cloud_lb_id": "lb-xxxx",
        "private_ipv4_addresses": [
          "127.0.0.1"
        ],
        "private_ipv6_addresses": [],
        "public_ipv4_addresses": [
          "127.0.0.1"
        ],
        "public_ipv6_addresses": [],
        "listener_num": 10,
        "health_check": {
          "health_switch": 1,
          "time_out": 2,
          "interval_time": 5,
          "health_num": 3,
          "un_health_num": 3,
          "http_code": null,
          "check_type": "TCP",
          "source_ip_type": 0
        }
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
| name                   | string       | 名称                             |
| vendor                 | string       | 供应商（枚举值：tcloud）                |
| account_id             | string       | 账号ID                           |
| bk_biz_id              | int          | 业务ID                           |
| target_group_type      | string       | 目标组类型                          |
| region                 | string       | 地域                             |
| protocol               | string       | 协议                             |
| port                   | int          | 端口                             |
| weight                 | int          | 权重                             |
| vpc_id                 | string array | vpcID数组                        |
| memo                   | string       | 备注                             |
| creator                | string       | 创建者                            |
| reviser                | string       | 修改者                            |
| created_at             | string       | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at             | string       | 修改时间，标准格式：2006-01-02T15:04:05Z |
| lb_id                  | string       | 负载均衡ID                         |
| lb_name                | string       | 云负载均衡名称                        |
| cloud_lb_id            | string       | 云负载均衡ID                        |
| health_check           | health_check | 健康检查                           |
| private_ipv4_addresses | string array | 负载均衡的内网IPv4地址                  |
| private_ipv6_addresses | string array | 负载均衡的内网IPv6地址                  |
| public_ipv4_addresses  | string array | 负载均衡的外网IPv4地址                  |
| public_ipv6_addresses  | string array | 负载均衡的外网IPv6地址                  |
| listener_num           | int          | 绑定的监听器数量                       |

### health_check

| 参数名称              | 参数类型   | 描述                                                                |
|-------------------|--------|-------------------------------------------------------------------|
| health_switch     | int    | 是否开启健康检查：1（开启）、0（关闭）                                              |
| time_out          | int    | 健康检查的响应超时时间，可选值：2~60，单位：秒                                         |
| interval_time     | int    | 健康检查探测间隔时间                                                        |
| health_num        | int    | 健康阈值                                                              |
| un_health_num     | int    | 不健康阈值                                                             |
| check_port        | int    | 自定义探测相关参数。健康检查端口，默认为后端服务的端口                                       |
| check_type        | string | 健康检查使用的协议。取值 TCP/HTTP/HTTPS/GRPC/PING/CUSTOM                      |
| http_version      | string | HTTP版本                                                            |
| http_check_path   | string | 健康检查路径（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）                      |
| http_check_domain | string | 健康检查域名                                                            |
| http_check_method | string | 健康检查方法（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式），默认值：HEAD，可选值HEAD或GET |
| source_ip_type    | string | 健康检查源IP类型：0（使用LB的VIP作为源IP），1（使用100.64网段IP作为源IP）                   |
