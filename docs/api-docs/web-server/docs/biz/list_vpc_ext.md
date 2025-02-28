### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询VPC列表（带云厂商私有结构）。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/vpcs/list

### 输入参数

| 参数名称      | 参数类型   | 必选   | 描述                               |
|-----------|--------|------|----------------------------------|
| bk_biz_id | int64  | 是    | 业务ID                             |
| vendor    | string | 是    | 供应商（枚举值：tcloud、aws、azure、huawei） |
| filter    | object | 是    | 查询过滤条件                           |
| page      | object | 是    | 分页设置                             |

#### filter

| 参数名称  | 参数类型        | 必选  | 描述                                                              |
|-------|-------------|-----|-----------------------------------------------------------------|
| op    | enum string | 是   | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是   | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### filter.rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选  | 描述                                          |
|-------|-------------|-----|---------------------------------------------|
| field | string      | 是   | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| op    | enum string | 是   | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis） |
| value | 可变类型        | 是   | 查询条件Value值                                  |

##### rules 表达式说明：

##### 1. 操作符

| 操作符 | 描述                                          | 操作符的value支持的数据类型                              |
|-----|---------------------------------------------|-----------------------------------------------|
| eq  | 等于。不能为空字符串                                  | boolean, numeric, string                      |
| neq | 不等。不能为空字符串                                  | boolean, numeric, string                      |
| gt  | 大于                                          | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte | 大于等于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lt  | 小于                                          | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lte | 小于等于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in  | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素    | boolean, numeric, string                      |
| nin | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素   | boolean, numeric, string                      |
| cs  | 模糊查询，区分大小写                                  | string                                        |
| cis | 模糊查询，不区分大小写                                 | string                                        |

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
| count	 | bool	   | 是	  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据 detail，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count |
| start	 | uint32	 | 否	  | 记录开始位置，start 起始值为0                                                                                                                               |
| limit	 | uint32	 | 否	  | 每页限制条数，最大500，不能为0                                                                                                                                |
| sort	  | string	 | 否	  | 排序字段，返回数据将按该字段进行排序                                                                                                                               |
| order	 | string	 | 否	  | 排序顺序（枚举值：ASC、DESC）                                                                                                                               |

#### 查询参数介绍：

| 参数名称        | 参数类型   | 描述                                   |
|-------------|--------|--------------------------------------|
| id          | string | VPC的ID                               |
| vendor      | string | 云厂商（枚举值：tcloud、aws、azure、gcp、huawei） |
| account_id  | string | 云账号ID                                |
| cloud_id    | string | VPC的云ID                              |
| name        | string | VPC名称                                |
| region      | string | 地域                                   |
| category    | string | VPC类别                                |
| memo        | string | 备注                                   |
| bk_biz_id   | int64  | 业务ID，-1表示没有分配到业务                     |
| creator     | string | 创建者                                  |
| reviser     | string | 更新者                                  |
| created_at  | string | 创建时间，标准格式：2006-01-02T15:04:05Z       |
| updated_at  | string | 更新时间，标准格式：2006-01-02T15:04:05Z       |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 获取详细信息请求参数示例

如查询云账号ID为"00000001"的腾讯云VPC列表。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "account_id",
        "op": "eq",
        "value": "00000001"
      },
      {
        "field": "vendor",
        "op": "eq",
        "value": "tcloud"
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

如查询云账号ID为"00000001"的腾讯云VPC数量。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "account_id",
        "op": "eq",
        "value": "00000001"
      },
      {
        "field": "vendor",
        "op": "eq",
        "value": "tcloud"
      }
    ]
  },
  "page": {
    "count": true
  }
}
```

### 响应示例

#### TCloud-获取详细信息示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "detail": [
      {
        "id": "00000001",
        "vendor": "tcloud",
        "account_id": "00000001",
        "cloud_id": "vpc-xxxxxxxx",
        "name": "vpc-default",
        "region": "ap-guangzhou",
        "category": "biz",
        "memo": "default vpc",
        "bk_biz_id": 100,
        "creator": "tom",
        "reviser": "tom",
        "created_at": "2019-07-29 11:57:20",
        "updated_at": "2019-07-29 11:57:20",
        "extension": {
          "cidr": [
            {
              "type": "ipv4",
              "cidr": "127.0.0.0/16",
              "category": "master"
            },
            {
              "type": "ipv6",
              "cidr": "::/56",
              "category": "master"
            }
          ],
          "is_default": true,
          "enable_multicast": false,
          "dns_server_set": [
            "127.0.0.1",
            "127.0.0.2"
          ],
          "domain_name": "aa.bb.cc"
        }
      }
    ]
  }
}
```

#### Aws-获取详细信息示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "detail": [
      {
        "id": "00000001",
        "vendor": "aws",
        "account_id": "00000001",
        "cloud_id": "vpc-xxxxxxxx",
        "name": "vpc-default",
        "region": "ap-guangzhou",
        "category": "biz",
        "memo": "default vpc",
        "bk_biz_id": 100,
        "creator": "tom",
        "reviser": "tom",
        "created_at": "2019-07-29 11:57:20",
        "updated_at": "2019-07-29 11:57:20",
        "extension": {
          "cidr": [
            {
              "type": "ipv4",
              "cidr": "127.0.0.0/24",
              "address_pool": "",
              "state": "associated"
            },
            {
              "type": "ipv6",
              "cidr": "::/56",
              "address_pool": "Amazon",
              "state": "associated"
            }
          ],
          "state": "available",
          "instance_tenancy": "default",
          "is_default": false,
          "enable_dns_support": false,
          "enable_dns_hostnames": false
        }
      }
    ]
  }
}
```

#### Gcp-获取详细信息示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "detail": [
      {
        "id": "00000001",
        "vendor": "gcp",
        "account_id": "00000001",
        "cloud_id": "vpc-xxxxxxxx",
        "name": "vpc-default",
        "region": "ap-guangzhou",
        "category": "biz",
        "memo": "default vpc",
        "bk_biz_id": 100,
        "creator": "tom",
        "reviser": "tom",
        "created_at": "2019-07-29 11:57:20",
        "updated_at": "2019-07-29 11:57:20",
        "extension": {
          "self_link": "https://www.googleapis.com/compute/v1/projects/xxx/global/networks/test",
          "auto_create_subnetworks": false,
          "enable_ula_internal_ipv6": true,
          "internal_ipv6_range": "::/48",
          "mtu": 1460,
          "routing_mode": "REGIONAL"
        }
      }
    ]
  }
}
```

#### Azure-获取详细信息示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "detail": [
      {
        "id": "00000001",
        "vendor": "azure",
        "account_id": "00000001",
        "cloud_id": "vpc-xxxxxxxx",
        "name": "vpc-default",
        "region": "ap-guangzhou",
        "category": "biz",
        "memo": "default vpc",
        "bk_biz_id": 100,
        "creator": "tom",
        "reviser": "tom",
        "created_at": "2019-07-29 11:57:20",
        "updated_at": "2019-07-29 11:57:20",
        "extension": {
          "resource_group": "test",
          "cidr": [
            {
              "cidr": "127.0.0.0/16",
              "type": "ipv4"
            }
          ],
          "dns_servers": [
            "127.0.0.1"
          ]
        }
      }
    ]
  }
}
```

#### HuaWei-获取详细信息示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "detail": [
      {
        "id": "00000001",
        "vendor": "huawei",
        "account_id": "00000001",
        "cloud_id": "vpc-xxxxxxxx",
        "name": "vpc-default",
        "region": "ap-guangzhou",
        "category": "biz",
        "memo": "default vpc",
        "bk_biz_id": 100,
        "creator": "tom",
        "reviser": "tom",
        "created_at": "2019-07-29 11:57:20",
        "updated_at": "2019-07-29 11:57:20",
        "extension": {
          "cidr": [
            {
              "cidr": "127.0.0.0/8",
              "type": "ipv4"
            }
          ],
          "status": "ACTIVE",
          "enterprise_project_id": "0"
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

| 参数名称        | 参数类型   | 描述                                   |
|-------------|--------|--------------------------------------|
| id          | string | VPC的ID                               |
| vendor      | string | 云厂商（枚举值：tcloud、aws、azure、gcp、huawei） |
| account_id  | string | 云账号ID                                |
| cloud_id    | string | VPC的云ID                              |
| name        | string | VPC名称                                |
| region      | string | 地域                                   |
| category    | string | VPC类别                                |
| memo        | string | 备注                                   |
| bk_biz_id   | int64  | 业务ID，-1表示没有分配到业务                     |
| creator     | string | 创建者                                  |
| reviser     | string | 更新者                                  |
| created_at  | string | 创建时间，标准格式：2006-01-02T15:04:05Z       |
| updated_at  | string | 更新时间，标准格式：2006-01-02T15:04:05Z       |

#### data.extension(tcloud)

| 参数名称             | 参数类型         | 描述         |
|------------------|--------------|------------|
| cidr             | object array | CIDR列表     |
| is_default       | boolean      | 是否是默认VPC   |
| enable_multicast | boolean      | 是否开启组播     |
| dns_server_set   | string array | DNS服务器列表   |
| domain_name      | string       | DHCP域名选项值。 |

#### data.extension(tcloud).cidr

| 参数名称     | 参数类型   | 描述                                            |
|----------|--------|-----------------------------------------------|
| type     | string | 地址类型（枚举值：ipv4、ipv6）                           |
| cidr     | string | CIDR                                          |
| category | string | 类别（枚举值：master【主】、assistant【辅助】、container【容器】） |

#### data.extension(aws)

| 参数名称                 | 参数类型         | 描述                               |
|----------------------|--------------|----------------------------------|
| cidr                 | object array | CIDR信息                           |
| state                | string       | 状态（枚举值：pending、available）        |
| instance_tenancy     | string       | 实例租期（枚举值：default、dedicated、host） |
| is_default           | boolean      | DNS服务器列表                         |
| enable_dns_hostnames | boolean      | 是否启用 DNS 主机名                     |
| enable_dns_support   | boolean      | 是否启用 DNS 解析                      |

#### data.extension(aws).cidr

| 参数名称         | 参数类型   | 描述                                                                         |
|--------------|--------|----------------------------------------------------------------------------|
| type         | string | 地址类型（枚举值：ipv4、ipv6）                                                        |
| cidr         | string | CIDR                                                                       |
| address_pool | string | 地址池                                                                        |
| state        | string | 状态（枚举值：associating、associated、disassociating、disassociated、failing、failed） |

#### data.extension(gcp)

| 参数名称                     | 参数类型    | 描述                          |
|--------------------------|---------|-----------------------------|
| self_link                | string  | 资源URL                       |
| auto_create_subnetworks  | boolean | 是否默认创建子网                    |
| enable_ula_internal_ipv6 | boolean | 是否启用 VPC 网络 ULA 内部 IPv6 范围  |
| internal_ipv6_range      | string  | VPC 网络 ULA 内部 IPv6 范围       |
| mtu                      | int64   | 最大传输单元                      |
| routing_mode             | string  | 动态路由模式（枚举值：REGIONAL、GLOBAL） |

#### data.extension(azure)

| 参数名称           | 参数类型         | 描述       |
|----------------|--------------|----------|
| resource_group | string       | 资源组      |
| dns_servers    | string array | DNS服务器列表 |
| cidr           | object array | CIDR列表   |

#### data.extension(azure).cidr

| 参数名称 | 参数类型   | 描述                  |
|------|--------|---------------------|
| type | string | 地址类型（枚举值：ipv4、ipv6） |
| cidr | string | CIDR                |

#### data.extension(huawei)

| 参数名称                  | 参数类型         | 描述                     |
|-----------------------|--------------|------------------------|
| cidr                  | object array | CIDR列表                 |
| status                | string       | 状态（枚举值：PENDING、ACTIVE） |
| enterprise_project_id | string       | 企业项目ID                 |

#### data.extension(huawei).cidr

| 参数名称 | 参数类型   | 描述                  |
|------|--------|---------------------|
| type | string | 地址类型（枚举值：ipv4、ipv6） |
| cidr | string | CIDR                |
