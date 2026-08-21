### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询地域列表。

### URL

POST /api/v1/cloud/vendors/{vendor}/regions/list

#### 路径参数说明

| 参数名称   | 参数类型   | 必选  | 描述  |
|--------|--------|-----|-----|
| vendor | string | 是   | 云厂商 |

### 输入参数

| 参数名称   | 参数类型   | 必选  | 描述     |
|--------|--------|-----|--------|
| filter | object | 是   | 查询过滤条件 |
| page   | object | 是   | 分页设置   |

#### filter

| 参数名称  | 参数类型        | 必选  | 描述                                                              |
|-------|-------------|-----|-----------------------------------------------------------------|
| op    | enum string | 是   | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是   | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选  | 描述                                          |
|-------|-------------|-----|---------------------------------------------|
| field | string      | 是   | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| op    | enum string | 是   | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis） |
| value | 可变类型        | 是   | 查询条件Value值                                  |

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

| 参数名称  | 参数类型   | 必选  | 描述                                                                                                                                                  |
|-------|--------|-----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是   | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否   | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否   | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否   | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否   | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |                                          

### 调用和响应 示例

#### 通用返回响应说明

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

#### aws 请求示例

AWS 每个账号可以独立禁用region，因此获取AWS的region信息需要加入`account_id`过滤条件

```json
{
  "page": {
    "count": false,
    "limit": 10
  },
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "account_id",
        "op": "eq",
        "value": "00000002"
      }
    ]
  }
}
```

##### aws 请求参数说明

| 参数名称       | 参数类型   | 描述    |
|------------|--------|-------|
| account_id | string | 云账号ID |

#### aws

##### aws查询响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 0,
    "details": [
      {
        "id": "0000000e",
        "vendor": "aws",
        "account_id": "00000002",
        "region_id": "ap-south-1",
        "region_name": "ap-south-1",
        "status": "opt-in-not-required",
        "endpoint": "ec2.ap-south-1.amazonaws.com",
        "creator": "hcm-backend-sync",
        "reviser": "hcm-backend-sync",
        "created_at": "2023-08-03T13:43:41Z",
        "updated_at": "2023-08-03T13:43:41Z"
      }
    ]
  }
}

```

##### aws 响应参数说明


###### data.detail[n]

| 参数名称        | 参数类型   | 描述                                            |
|-------------|--------|-----------------------------------------------|
| id          | string | 地域的数据库ID                                      |
| vendor      | string | 云厂商（aws）                                      |
| account_id  | string | 云账号ID                                         |
| region_id   | string | 地域ID（唯一标识）                                    |
| region_name | string | 地域名称                                          |
| status      | string | 状态（opt-in-not-required,opted-in,not-opted-in） |
| endpoint    | string | 服务端点                                          |
| creator     | string | 创建者                                           |
| reviser     | string | 更新者                                           |
| created_at  | string | 创建时间，标准格式：2006-01-02T15:04:05Z                |
| updated_at  | string | 更新时间，标准格式：2006-01-02T15:04:05Z                |

#### azure

##### azure 请求示例

```json 
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "type",
        "op": "eq",
        "value": "Region"
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

##### azure 响应示例

```json 
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "id": "000000ki",
        "cloud_id": "/subscriptions/7C99B444-456A-4EEF-A083-40F6AB39CEAA/locations/southcentralus",
        "name": "southcentralus",
        "type": "Region",
        "display_name": "South Central US",
        "region_display_name": "(US) South Central US",
        "geography_group": "",
        "latitude": "",
        "longitude": "",
        "physical_location": "",
        "region_type": "Physical",
        "paired_region_name": "",
        "paired_region_id": "",
        "creator": "sync-timing-admin",
        "reviser": "sync-timing-admin",
        "created_at": "2023-02-21T15:27:16Z",
        "updated_at": "2023-02-21T15:27:16Z"
      }
    ]
  }
}
```

##### azure 响应参数说明

###### data.detail[n]

| 参数名称                | 参数类型   | 描述                                            |
|---------------------|--------|-----------------------------------------------|
| id                  | string | 地域的数据库ID                                      |
| cloud_id            | string | 云厂商上的id                                       |
| type                | string | 地域类型（Region,EdgeZone）                         |
| name                | string | 地域名（唯一标识）                                     |
| display_name        | string | 地域的友好名称                                       |
| region_display_name | string | 带地域所在区域的友好名称                                  |
| status              | string | 状态（opt-in-not-required,opted-in,not-opted-in） |
| geography_group     | string | 地域所在地理组                                       |
| latitude            | string | 地域所在维度                                        |
| longitude           | string | 地域所在经度                                        |
| physical_location   | string | 地域物理位置                                        |
| region_type         | string | 地域类型 (Physical,Logical）                       |
| paired_region_name  | string | 和该地域配对的地域名                                    |
| paired_region_id    | string | 和该地域配对的地域云厂商ID                                |
| creator             | string | 创建者                                           |
| reviser             | string | 更新者                                           |
| created_at          | string | 创建时间，标准格式：2006-01-02T15:04:05Z                |
| updated_at          | string | 更新时间，标准格式：2006-01-02T15:04:05Z                |

#### GCP

###### GCP 请求示例

```json
{
  "page": {
    "count": false,
    "limit": 10
  },
  "filter": {}
}
```

##### GCP 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 0,
    "details": [
      {
        "id": "0000000a",
        "vendor": "gcp",
        "region_id": "australia-southeast1",
        "region_name": "australia-southeast1",
        "status": "UP",
        "creator": "hcm-backend-sync",
        "reviser": "hcm-backend-sync",
        "created_at": "2023-06-29T09:57:14Z",
        "updated_at": "2023-06-29T09:57:14Z"
      }
    ]
  }
}
```

###### GCP 响应参数说明

###### data.detail[n]

| 参数名称        | 参数类型   | 描述                             |
|-------------|--------|--------------------------------|
| id          | string | 地域的数据库ID                       |
| vendor      | string | 云厂商（gcp）                       |
| region_id   | string | 地域ID（唯一标识）                     |
| region_name | string | 地域名称                           |
| status      | string | 状态（UP,DOWN）                    |
| creator     | string | 创建者                            |
| reviser     | string | 更新者                            |
| created_at  | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at  | string | 更新时间，标准格式：2006-01-02T15:04:05Z |

#### huawei 请求说明

##### 华为云请求示例

华为云不同类型的服务的可用地域不同，服务类型于对应的服务关系：

| 服务类型 | 对应服务                            |
|------|---------------------------------|
| ecs  | cvm、disk、networkinterface       |
| vpc  | vpc subnet sg sgRule routetable |
| eip  | eip                             |
| ims  | publicimage                     |

```json 
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "type",
        "op": "eq",
        "value": "public"
      },
      {
        "field": "service",
        "op": "eq",
        "value": "ims"
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

##### 华为云响应示例

```json 
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "id": "000000kc",
        "service": "ims",
        "region_id": "cn-east-3",
        "type": "public",
        "locales_pt_br": "",
        "locales_zh_cn": "",
        "locales_en_us": "",
        "locales_es_us": "",
        "locales_es_es": "",
        "creator": "jim",
        "reviser": "jim",
        "created_at": "2023-03-09T20:11:47Z",
        "updated_at": "2023-03-09T20:11:47Z"
      }
    ]
  }
}
```

###### 华为云 响应参数说明


###### data.detail[n]

| 参数名称          | 参数类型   | 描述                             |
|---------------|--------|--------------------------------|
| id            | string | 地域的数据库ID                       |
| service       | string | 服务类型，参见上方表格                    |
| region_id     | string | 地域ID（唯一标识）                     |
| type          | string | 地域类型                           |
| locales_zh_cn | string | 地域的中文名称                        |
| locales_pt_br | string | 地域的葡萄牙语名称                      |
| locales_en_us | string | 地域的英文名称                        |
| locales_es_us | string | 地域的美国西班牙语名称                    |
| locales_es_es | string | 地域的西班牙语名称                      |
| creator       | string | 创建者                            |
| reviser       | string | 更新者                            |
| created_at    | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at    | string | 更新时间，标准格式：2006-01-02T15:04:05Z |

#### TCloud

##### TCloud 请求示例

```json 
{
  "filter": {
    "op": "and",
    "rules": [
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

### TCloud 响应示例

```json 
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "id": "00000024",
        "vendor": "tcloud",
        "region_id": "na-toronto",
        "region_name": "北美地区(多伦多)",
        "status": "AVAILABLE",
        "creator": "sync-timing-admin",
        "reviser": "sync-timing-admin",
        "created_at": "2023-02-25T18:01:57Z",
        "updated_at": "2023-02-25T18:01:57Z"
      }
    ]
  }
}
```

###### TCloud 响应参数说明

###### data.detail[n]

| 参数名称        | 参数类型   | 描述                             |
|-------------|--------|--------------------------------|
| id          | string | 地域的数据库ID                       |
| vendor      | string | 云厂商（tcloud）                    |
| region_id   | string | 地域ID（唯一标识）如ap-guangzhou        |
| region_name | string | 地域描述，例如，华南地区(广州)               |
| status      | string | 状态（AVAILABLE）                  |
| creator     | string | 创建者                            |
| reviser     | string | 更新者                            |
| created_at  | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at  | string | 更新时间，标准格式：2006-01-02T15:04:05Z |
