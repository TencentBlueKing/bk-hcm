### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询网络接口列表，支持是否绑定EIP、是否绑定主机的查询。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/network_interfaces/associate/list

### 输入参数

| 参数名称         | 参数类型     | 必选    | 描述                                         |
|--------------|----------|-------|--------------------------------------------|
| bk_biz_id    | int64    | 是     | 业务ID                                       |
| filter       | object   | 是     | 查询过滤条件                                     |
| page         | object   | 是     | 分页设置                                       |
| is_associate | bool     | 否     | 是否有关联数据(true:获取已关联的列表 false:获取未关联的列表，默认不传) |

#### filter

| 参数名称  | 参数类型        | 必选  | 描述                                                              |
|-------|-------------|-----|-----------------------------------------------------------------|
| op    | enum string | 是   | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是   | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### filter.rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选  | 描述                                         |
|-------|-------------|-----|--------------------------------------------|
| field | string      | 是   | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍 |
| op    | enum string | 是   | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis）       |
| value | 可变类型     | 是   | 查询条件Value值                                 |

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

| 参数名称   | 参数类型    | 必选  | 描述                                                                                                                                               |
|--------|---------|-----|--------------------------------------------------------------------------------------------------------------------------------------------------|
| count	 | bool	   | 是	 | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据 detail，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count |
| start	 | uint32  | 否	 | 记录开始位置，start 起始值为0                                                                                                                               |
| limit	 | uint32  | 否	 | 每页限制条数，最大500，不能为0                                                                                                                                |
| sort	 | string  | 否	 | 排序字段，返回数据将按该字段进行排序                                                                                                                               |
| order	 | string  | 否	 | 排序顺序（枚举值：ASC、DESC）                                                                                                                               |

#### 查询参数介绍：

| 参数名称        | 参数类型   | 描述                             |
|-------------|--------|-------------------------------------- |
| id          | string | 主键ID                                |
| vendor      | string | 云厂商（枚举值：azure、huawei、gcp）      |
| name        | string | 网络接口名称                            |
| account_id  | string | 云资源的账号ID                          |
| region      | string | 地区ID                                 |
| zone        | string | 可用区                                 |
| vpc_id      | string | VPC的ID                               |
| cloud_vpc_id | string | 云VPC的ID                             |
| subnet_id   | string | 子网ID                                 |
| cloud_subnet_id | string | 云子网ID                           |
| private_ipv4  | string array | 内网IPv4                       |
| private_ipv6  | string array | 内网IPv6                       |
| public_ipv4   | string array | 公网IPv4                       |
| public_ipv6   | string array | 公网IPv6                       |
| bk_biz_id   | int     | 业务ID                                |
| instance_id | string  | 关联的实例ID                           |
| creator     | string | 创建者                                  |
| reviser     | string | 更新者                                  |
| created_at  | string | 创建时间，标准格式：2006-01-02T15:04:05Z    |
| updated_at  | string | 更新时间，标准格式：2006-01-02T15:04:05Z    |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 获取详细信息请求参数示例

如查询账号ID为"00000024"，已绑定主机的Azure网络接口列表。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "account_id",
        "op": "eq",
        "value": "00000024"
      },
      {
        "field": "vendor",
        "op": "eq",
        "value": "azure"
      }
    ]
  },
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
  },
  "is_associate": true
}
```

#### 获取数量请求参数示例

如查询账号ID为"00000024"的Azure网络接口数量。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "account_id",
        "op": "eq",
        "value": "00000024"
      },
      {
        "field": "vendor",
        "op": "eq",
        "value": "azure"
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
  "message": "ok",
  "data": {
    "detail": [
      {
        "id": "00000001",
        "vendor": "azure",
        "name": "resource_group_test",
        "account_id": "00000024",
        "region": "eastasia",
        "zone": "us-central1-a",
        "vpc_id": "000001",
        "cloud_vpc_id": "/subscriptions/1001-0000-0000-0000-111111111111/resourceGroups/rsg1001/providers/Microsoft.Network/virtualNetworks/test-vnet",
        "subnet_id": "000002",
        "cloud_subnet_id": "/subscriptions/1001-0000-0000-0000-111111111111/resourceGroups/rsg1001/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/default",
        "private_ipv4": ["127.0.0.1"],
        "private_ipv6": ["xx:xx:xx:xx:xx"],
        "public_ipv4": ["127.0.0.2"],
        "public_ipv6": ["xx:xx:xx:xx:xx"],
        "bk_biz_id": 10010,
        "instance_id": "1001-0000-0000-0000-xxxxxxxxx",
        "creator": "tom",
        "reviser": "tom",
        "created_at": "2019-07-29 11:57:20",
        "updated_at": "2019-07-29 11:57:20",
        "cvm_id": "0000000x",
        "rel_creator": "xxxx",
        "rel_created_at": "2019-07-29 11:57:20"
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
    "count": 0
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

| 参数名称   | 参数类型   | 描述                                       |
|--------|--------|------------------------------------------|
| count  | uint64 | 当前规则能匹配到的总记录条数，仅在 count 查询参数设置为 true 时返回 |
| detail | array  | 查询返回的数据，仅在 count 查询参数设置为 false 时返回       |

#### data.detail[n]

| 参数名称        | 参数类型   | 描述                            |
|-------------|--------|--------------------------------------|
| id          | string | 主键ID                                |
| vendor      | string | 云厂商（枚举值：azure、huawei、gcp）      |
| name        | string | 网络接口名称                            |
| account_id  | string | 云资源的账号ID                          |
| region      | string | 地区ID                                 |
| zone        | string | 可用区                                 |
| vpc_id      | string | VPC的ID                               |
| cloud_vpc_id | string | 云VPC的ID                             |
| subnet_id   | string | 子网ID                                 |
| cloud_subnet_id | string | 云子网ID，格式：半角逗号分割           |
| private_ipv4  | string array | 内网IPv4                       |
| private_ipv6  | string array | 内网IPv6                       |
| public_ipv4   | string array | 公网IPv4                       |
| public_ipv6   | string array | 公网IPv6                       |
| bk_biz_id   | int     | 业务ID                                |
| instance_id | string  | 关联的实例ID                           |
| creator     | string | 创建者                                  |
| reviser     | string | 更新者                                  |
| created_at  | string | 创建时间，标准格式：2006-01-02T15:04:05Z    |
| updated_at  | string | 更新时间，标准格式：2006-01-02T15:04:05Z    |
| cvm_id      | string | 主机ID                                  |
| rel_creator | string | 绑定主机时的创建者                        |
| rel_created_at | string | 绑定主机时的创建时间，标准格式：2006-01-02T15:04:05Z |
