### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：无。
- 该接口功能描述：查询主机库存以及对应的机型信息。

### URL

POST /api/v1/woa/config/capacity/list_with_device_info

### 输入参数

| 参数名称   | 参数类型   | 必选 | 描述     |
|--------|--------|----|--------|
| filter | object | 否  | 查询过滤条件 |
| page   | object | 否  | 分页设置   |
| fields | array  | 否  | 指定返回的字段列表 |

#### filter

| 参数名称  | 参数类型        | 必选 | 描述                                                              |
|-------|-------------|----|-----------------------------------------------------------------|
| op    | enum string | 是  | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是  | 过滤规则数组。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

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

查询 require_type 是 1 且 region 是 "ap-shanghai" 且 zone 在 "ap-shanghai-1" 或 "ap-shanghai-2" 中的数据。

```json
{
  "op": "and",
  "rules": [
    {
      "field": "require_type",
      "op": "eq",
      "value": 1
    },
    {
      "field": "region",
      "op": "eq",
      "value": "ap-shanghai"
    },
    {
      "field": "zone",
      "op": "in",
      "value": [
        "ap-shanghai-1",
        "ap-shanghai-2"
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

| 参数名称        | 参数类型 | 描述                                                                    |
|-------------|------|-----------------------------------------------------------------------|
| require_type | int  | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤; 6: 滚服项目; 7: 小额绿通                                      |
| region      | string | 地域                                                                    |
| zone        | string | 可用区                                                                  |
| device_type | string | 机型                                                                    |
| device_family | string | 机型族                                                                  |
| cpu_core    | int64 | CPU核数                                                                 |
| memory      | int64 | 内存大小（单位：GB）                                                           |
| capacity    | int64 | 库存容量                                                                  |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 获取详细信息请求参数示例

查询常规项目（require_type=1）在上海地域（ap-shanghai）的库存信息。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "require_type",
        "op": "eq",
        "value": 1
      },
      {
        "field": "region",
        "op": "eq",
        "value": "ap-shanghai"
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

查询常规项目（require_type=1）在上海地域（ap-shanghai）的库存数量。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "require_type",
        "op": "eq",
        "value": 1
      },
      {
        "field": "region",
        "op": "eq",
        "value": "ap-shanghai"
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
    "count": 0,
    "details": [
      {
        "require_type": 1,
        "region": "ap-shanghai",
        "zone": "ap-shanghai-1",
        "device_family": "标准型",
        "device_type": "S3ne.4XLARGE64",
        "cpu_core": 16,
        "memory": 64,
        "capacity": 45,
        "capacity_flag": 3,
        "core_type": "小核心",
        "device_type_class": "CommonType"
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
    "count": 1,
    "details": []
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

| 参数名称        | 参数类型 | 描述                                                            |
|-------------|------|---------------------------------------------------------------|
| require_type | int  | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤                                |
| region      | string | 地域                                                            |
| zone        | string | 可用区                                                           |
| device_family | string | 机型族                                                           |
| device_type | string | 机型                                                            |
| cpu_core    | int64 | CPU核数                                                         |
| memory      | int64 | 内存大小（单位：GB）                                                   |
| capacity    | int64 | 库存容量                                                          |
| capacity_flag | int  | 库存状态标识。1: 无库存（0）; 2: 库存紧张（1~10）; 3: 少量库存（11~50）; 4: 库存充足（51+） |
| core_type     | string | 核心类型(枚举值：大核心、中核心、小核心)                                         |
| device_type_class | string | 通/专用机型，SpecialType(专用)，CommonType(通用)                         |

##### capacity_flag 状态含义：

| 状态值 | 含义   | 容量范围    |
|-----|------|---------|
| 1   | 无库存  | 0       |
| 2   | 库存紧张 | 1~10    |
| 3   | 少量库存 | 11~50   |
| 4   | 库存充足 | 51及以上   |
