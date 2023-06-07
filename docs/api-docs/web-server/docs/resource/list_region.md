### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询地域列表。

### URL

POST /api/v1/cloud/vendors/{vendor}/regions/list

#### 路径参数说明
| 参数名称          | 参数类型                           | 必选 | 描述                                                      |
| ----------------- | ------------------------------- | ---- | -------------------------------------------------------- |
| vendor | string | 是 | 云厂商 |

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

| 参数名称  | 参数类型     | 必选   | 描述                                                                                                                                                  |
|-------|----------|------|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool     | 是    | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32   | 否    | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32   | 否    | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string   | 否    | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string   | 否    | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |                                          

### 调用和响应 示例

### azure req
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

### azure resp
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
            },
        ]
    }
}
```

### huawei req
### ecs 对应 cvm disk networkinterface
### vpc 对应 vpc subnet sg sgRule routetable
### eip 对应 eip
### ims 对应 publicimage
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

### huawei resp
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
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000kh",
                "service": "ims",
                "region_id": "cn-north-1",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000kp",
                "service": "ims",
                "region_id": "cn-southwest-2",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000kr",
                "service": "ims",
                "region_id": "na-mexico-1",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000kt",
                "service": "ims",
                "region_id": "cn-north-9",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000kv",
                "service": "ims",
                "region_id": "ap-southeast-1",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000kw",
                "service": "ims",
                "region_id": "ap-southeast-2",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000kx",
                "service": "ims",
                "region_id": "af-south-1",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000l0",
                "service": "ims",
                "region_id": "cn-east-2",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000lb",
                "service": "ims",
                "region_id": "sa-brazil-1",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000lh",
                "service": "ims",
                "region_id": "ap-southeast-3",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000lj",
                "service": "ims",
                "region_id": "cn-north-2",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000ll",
                "service": "ims",
                "region_id": "cn-north-4",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000lo",
                "service": "ims",
                "region_id": "ap-southeast-4",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000lr",
                "service": "ims",
                "region_id": "cn-south-2",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000lt",
                "service": "ims",
                "region_id": "cn-south-1",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000lu",
                "service": "ims",
                "region_id": "la-south-2",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            },
            {
                "id": "000000m8",
                "service": "ims",
                "region_id": "la-north-2",
                "type": "public",
                "locales_pt_br": "",
                "locales_zh_cn": "",
                "locales_en_us": "",
                "locales_es_us": "",
                "locales_es_es": "",
                "creator": "brookszhang",
                "reviser": "brookszhang",
                "created_at": "2023-03-09T20:11:47Z",
                "updated_at": "2023-03-09T20:11:47Z"
            }
        ]
    }
}
```

### tcloud req (aws、gcp请求参数类似)
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

### tcloud resp (aws、gcp返回参数类似)
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
