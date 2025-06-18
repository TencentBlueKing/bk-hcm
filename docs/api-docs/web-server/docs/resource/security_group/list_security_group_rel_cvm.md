### 描述

- 该接口提供版本：v1.8.0+。
- 该接口所需权限：资源-资源查看。
- 该接口功能描述：查询安全组关联的cvm列表。

### URL

POST /api/v1/cloud/security_groups/{sg_id}/related_resources/cvms/list

### 输入参数

| 参数名称       | 参数类型     | 必选 | 描述                                         |
|------------|----------|----|--------------------------------------------|
| sg_id      | string   | 是  | 安全组ID                                      |
| page       | object   | 是  | 分页设置                                       |
| filter     | object   | 是  | 查询过滤条件                                     |
| fields     | []string | 否  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍 |

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

| 参数名称                | 参数类型   | 描述                                   |
|---------------------|--------|--------------------------------------|
| bk_biz_id           | int64  | 资源的业务ID                              |
| cloud_id            | string | 云资源ID                                |
| name                | string | 名称                                   |
| bk_cloud_id         | int64  | 云区域ID                                |
| account_id          | string | 账号ID                                 |
| region              | string | 地域                                   |
| zone                | string | 可用区                                  |
| cloud_image_id      | string | 云镜像ID                                |
| os_name             | string | 操作系统名称                               |
| memo                | string | 备注                                   |
| status              | string | 状态                                   |
| recycle_status      | string | 回收状态                                 |
| machine_type        | string | 设备类型                                 |
| cloud_created_time  | string | Cvm在云上创建时间，标准格式：2006-01-02T15:04:05Z |
| cloud_launched_time | string | Cvm启动时间，标准格式：2006-01-02T15:04:05Z    |
| cloud_expired_time  | string | Cvm过期时间，标准格式：2006-01-02T15:04:05Z    |
| creator             | string | 创建者                                  |
| reviser             | string | 修改者                                  |
| created_at          | string | 创建时间，标准格式：2006-01-02T15:04:05Z       |
| updated_at          | string | 修改时间，标准格式：2006-01-02T15:04:05Z       |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

```json
{
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "id": "00000001",
        "cloud_id": "cvm-123",
        "name": "cvm-test",
        "vendor": "tcloud",
        "bk_biz_id": -1,
        "region": "ap-hk",
        "zone": "ap-hk-1",
        "cloud_vpc_ids": [
          "vpc-123"
        ],
        "cloud_subnet_ids": [
          "subnet-123"
        ],
        "status": "init",
        "private_ipv4_addresses": [
          "127.0.0.1"
        ],
        "private_ipv6_addresses": [],
        "public_ipv4_addresses": [
          "127.0.0.2"
        ],
        "public_ipv6_addresses": []
      }
    ]
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

| 参数名称                   | 参数类型         | 描述                                   |
|------------------------|--------------|--------------------------------------|
| id                     | string       | 资源ID                                 |
| cloud_id               | string       | 云资源ID                                |
| name                   | string       | 名称                                   |
| vendor                 | string       | 供应商（枚举值：tcloud、aws、azure、gcp、huawei） |
| bk_biz_id              | int64        | 业务ID                                 |
| region                 | string       | 地域                                   |
| zone                   | string       | 可用区                                  |
| cloud_vpc_ids          | string array | 云VpcID列表                             |
| cloud_subnet_ids       | string array | 云子网ID列表                              |
| status                 | string       | 状态                                   |
| private_ipv4_addresses | string array | 内网IPv4地址                             |
| private_ipv6_addresses | string array | 内网IPv6地址                             |
| public_ipv4_addresses  | string array | 公网IPv4地址                             |
| public_ipv6_addresses  | string array | 公网IPv6地址                             |


