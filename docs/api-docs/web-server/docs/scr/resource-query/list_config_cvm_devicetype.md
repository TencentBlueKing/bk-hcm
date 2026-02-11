### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：无。
- 该接口功能描述：机型配置信息查询。

### URL

POST /api/v1/woa/config/findmany/config/cvm/devicetype

### 输入参数

| 参数名称   | 参数类型   | 必选 | 描述     |
|--------|--------|----|--------|
| filter | object | 是  | 查询过滤条件 |
| page   | object | 是  | 分页参数     |
| fields | array  | 否  | 指定返回的字段列表 |

#### filter

| 参数名称 | 参数类型        | 必选 | 描述                                                              |
|------|-------------|----|-----------------------------------------------------------------|
| op   | enum string | 是  | 逻辑操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules| array       | 是  | 过滤规则数组，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rule[n] （详情请看 rules 表达式说明）

| 参数名称 | 参数类型        | 必选 | 描述                                          |
|------|-------------|----|---------------------------------------------|
| field| string      | 是  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| op   | enum string | 是  | 操作符（枚举值：equal、not_equal、gt、gte、lt、lte、in、nin、contains） |
| value| 可变类型        | 是  | 查询条件Value值                                  |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                                                              |
|------|------|----|-----------------------------------------------------------------|
| count| bool | 否  | 是否只返回总数。如果为true，则只返回count，不返回详情列表，此时start和limit必须为0 |
| start| int  | 否  | 起始位置，从0开始。仅在count为false时生效                                |
| limit| int  | 否  | 每页返回的记录数。仅在count为false时生效                                |
| sort | string | 否 | 排序字段。仅在count为false时生效                                        |
| order| enum string | 否 | 排序方向（枚举值：asc、desc）。仅在count为false且sort不为空时生效              |

##### rule 表达式说明：

##### 1. 操作符

| 操作符      | 描述                                        | 操作符的value支持的数据类型                              |
|---------|-------------------------------------------|-----------------------------------------------|
| equal   | 等于。不能为空字符串                                | boolean, numeric, string                      |
| not_equal | 不等。不能为空字符串                                | boolean, numeric, string                      |
| gt      | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte     | 大于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lt      | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lte     | 小于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in      | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string                      |
| nin     | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string                      |
| contains| 模糊查询，区分大小写                                | string                                        |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "zone",
        "op": "equal",
        "value": "ap-shanghai-2"
      },
      {
        "field": "cpu_core",
        "op": "equal",
        "value": 4
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

#### 获取总数请求参数示例

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "zone",
        "op": "equal",
        "value": "ap-shanghai-2"
      }
    ]
  },
  "page": {
    "count": true,
    "start": 0,
    "limit": 0
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "count": 2,
    "details": [
      {
        "id": "6001",
        "vendor": "tcloud_ziyan",
        "device_type": "S3.MEDIUM4",
        "device_type_class": "CommonType",
        "device_class": "标准型",
        "device_family": "标准型",
        "core_type": "中核心",
        "cpu_core": 2,
        "memory": 16,
        "technical_class": "",
        "disable": false,
        "source": "sync"
      },
      {
        "id": "6002",
        "vendor": "tcloud_ziyan",
        "device_type": "S3.LARGE8",
        "device_type_class": "CommonType",
        "device_class": "标准型",
        "device_family": "标准型",
        "core_type": "中核心",
        "cpu_core": 4,
        "memory": 32,
        "technical_class": "",
        "disable": false,
        "source": "sync"
      }
    ]
  }
}
```

#### 获取总数返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "count": 2
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述                         |
|---------|--------------|----------------------------|
| result  | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code    | int          | 错误编码。 0表示success，>0表示失败错误  |
| message | string       | 请求失败返回的错误信息                |
| data	   | object array | 响应数据                       |

#### data

| 参数名称   | 参数类型         | 描述             |
|--------|--------------|----------------|
| count  | int          | 当前规则能匹配到的总记录条数 |
| details| object array | 机型列表（仅在page.count为false时返回）           |

#### details[0]

| 参数名称              | 参数类型   | 描述                                    |
|-------------------|--------|---------------------------------------|
| id                | string | 机型ID                                  |
| vendor            | string | 云厂商                                   |
| device_type       | string | 机型                                    |
| device_type_class | string | 通/专用机型，SpecialType(专用)，CommonType(通用) |
| device_class      | string | 机型分类                                  |
| device_family     | string | 机型族                                   |
| core_type         | string | 核心类型，枚举值：小核心、中核心、大核心                  |
| cpu_core          | int    | CPU核数                                 |
| memory            | int    | 内存容量，单位：GB                            |
| technical_class   | string | 技术分类                                  |
| disable           | bool   | 是否禁用                                  |
| source            | string | 机型来源：枚举值：sync(同步)、manually(手动添加)      |
