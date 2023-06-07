### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询Azure安全组规则列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/azure/security_groups/{security_group_id}/rules/list

### 输入参数

| 参数名称              | 参数类型   | 必选      | 描述     |
|-------------------|--------|---------|--------|
| bk_biz_id         | int64  | 是       | 业务ID   |
| security_group_id | string | 是       | 安全组ID  |
| filter            | object | 是       | 查询过滤条件 |
| page              | object | 是       | 分页设置   |

#### filter

| 参数名称  | 参数类型        | 必选  | 描述                                                              |
|-------|-------------|-----|-----------------------------------------------------------------|
| op    | enum string | 是   | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是   | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选  | 描述                                         |
|-------|-------------|-----|--------------------------------------------|
| field | string      | 是   | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍 |
| op    | enum string | 是   | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis）       |
| value | 可变类型        | 是   | 查询条件Value值                                 |

##### rules 表达式说明：

##### 1. 操作符

| 操作符 | 描述                                        | 操作符的value支持的数据类型                             |
|-----|-------------------------------------------|----------------------------------------------|
| eq  | 等于。不能为空字符串                                | boolean, numeric, string                     |
| neq | 不等。不能为空字符串                                | boolean, numeric, string                     |
| gt  | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte | 大于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lt  | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lte | 小于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in  | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string                     |
| nin | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string                     |
| cs  | 模糊查询，区分大小写                                | string                                       |
| cis | 模糊查询，不区分大小写                               | string                                       |

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

| 参数名称  | 参数类型   | 必选  | 描述                                                                                                                                                  |
|-------|--------|-----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是   | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否   | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否   | 每页限制条数，最大500，不能为0                                                                                                                                   |

#### 查询参数介绍：

| 参数名称 | 参数类型   | 描述              |
|------|--------|-----------------|
| type | string | 规则类型。（枚举值：egress、ingress）   |

### 调用示例

#### 获取详细信息请求参数示例

如查询Azure安全组ID为1的安全组规则列表。

```json
{
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
  }
}
```

#### 获取数量请求参数示例

如查询Azure安全组ID为1的安全组规则数量。

```json
{
  "page": {
    "count": true
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

如查询Azure安全组ID为1的安全组规则列表响应示例。

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "details": [
      {
        "id": 1,
        "cloud_id": "/subscriptions/eaa/resourceGroups/xxx/providers/Microsoft.Network/networkSecurityGroups/nsg/securityRules/HTTP",
        "etag": "W/\"ab62e9a371b494c\"",
        "name": "HTTP",
        "memo": "security_group_rule",
        "destination_address_prefix": "*",
        "destination_address_prefixes": [],
        "cloud_destination_app_security_group_ids": [],
        "destination_port_range": "80",
        "destination_port_ranges": [],
        "protocol": "TCP",
        "provisioning_state": "Succeeded",
        "source_address_prefix": "*",
        "source_address_prefixes": [],
        "cloud_source_app_security_group_ids": [],
        "source_port_range": "*",
        "source_port_ranges": [],
        "priority": 100,
        "type": "egress",
        "access": "Allow",
        "cloud_security_group_id": "sg-xxxxxx",
        "account_id": "1",
        "security_group_id": "1",
        "creator": "tom",
        "reviser": "tom",
        "created_at": "2019-07-29 11:57:20",
        "updated_at": "2019-07-29 11:57:20"
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

| 参数名称    | 参数类型   | 描述             |
|---------|--------|----------------|
| count   | uint64 | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据        |

#### data.details[n]

| 参数名称                                     | 参数类型   | 描述                                                                                                           |
|------------------------------------------|--------|--------------------------------------------------------------------------------------------------------------|
| id                                       | string | 安全组规则ID                                                                                                      |
| cloud_id                                 | string | 安全组规则云ID。                                                                                                    |
| etag                                     | string | 在更新资源时更改的唯一只读字符串。                                                                                            |
| name                                     | string | 资源组中唯一的资源名称。此名称可用于访问资源。                                                                                      |
| memo                                     | string | 备注。                                                                                                          |
| destination_address_prefix               | string | 目的地址前缀。CIDR或目标IP范围。星号‘*’也可用于匹配所有源IP。也可以使用‘VirtualNetwork’、‘AzureLoadBalancer’和‘Internet’等默认标签。               |
| destination_address_prefixes             | string | 目的地址带有前缀。CIDR或目标IP范围。                                                                                        |
| cloud_destination_app_security_group_ids | string | 目标应用安全组云ID列表。                                                                                                |
| destination_port_range                   | string | 目标端口或范围。介于0和65535之间的整数或范围。星号‘*’也可用于匹配所有端口。                                                                   |
| destination_port_ranges                  | string | 目的端口范围。                                                                                                      |
| protocol                                 | string | 网络协议。（枚举值：*、Ah、Esp、Icmp、Tcp、Udp）                                                                             |
| provisioning_state                       | string | 调度状态。（枚举值：Deleting、Failed、Succeeded、Updating）                                                                |
| source_address_prefix                    | string | CIDR或来源IP范围。星号‘*’也可用于匹配所有源IP。也可以使用‘VirtualNetwork’、‘AzureLoadBalancer’和‘Internet’等默认标签。如果这是入口规则，则指定网络流量源自何处。 |
| source_address_prefixes                  | string | CIDR或来源IP范围。                                                                                                 |
| cloud_source_app_security_group_ids      | string | 源安全组云ID列表。                                                                                                   |
| source_port_range                        | string | 源端口或范围。介于0和65535之间的整数或范围。星号‘*’也可用于匹配所有端口。                                                                    |
| source_port_ranges                       | string | 源端口范围。                                                                                                       |
| priority                                 | uint32 | 规则的优先级。该值可以介于100和4096之间。对于集合中的每个规则，优先级编号必须是唯一的。优先级数字越小，规则的优先级越高。                                             |
| type                                     | string | 规则类型。（枚举值：egress、ingress）                                                                                    |
| access                                   | string | 允许或拒绝网络流量。（枚举值：Allow、Deny）                                                                                   |
| cloud_security_group_id                  | string | 安全组规则所属安全组云ID。                                                                                               |
| account_id                               | string | 账号ID                                                                                                         |
| security_group_id                        | string | 规则所属安全组ID                                                                                                    |
| creator                                  | string | 创建者                                                                                                          |
| reviser                                  | string | 最后一次修改的修改者                                                                                                   |
| created_at                               | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                                                                                         |
| updated_at                               | string | 最后一次修改时间，标准格式：2006-01-02T15:04:05Z                                                                                                     |
