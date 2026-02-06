### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：无。
- 该接口功能描述：IDCPM下的物理机机型信息查询。

### URL

POST /api/v1/woa/config/findmany/config/idcpm/devicetype

### 输入参数

| 参数名称   | 参数类型 | 必选 | 描述      |
|-----------|--------|-----|-----------|
| filter    | object | 是  | 查询过滤条件 |
| page      | object | 是  | 分页设置    |

#### filter

| 参数名称  | 参数类型        | 必选  | 描述                                                              |
|-------|-------------|-----|-----------------------------------------------------------------|
| op    | enum string | 是   | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是   | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rule[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型   | 必选 | 描述                                           |
|-------|--------|----|----------------------------------------------|
| field | string | 是  | 查询条件 Field 名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍 |
| op    | string | 是  | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin）         |
| value | any    | 是  | 查询条件 Value 值                                 |

##### rule 表达式说明：

##### 1. 操作符

| 操作符 | 描述                                        | 操作符的value支持的数据类型                             |
|-----|-------------------------------------------|----------------------------------------------|
| eq  | 等于。不能为空字符串                                | boolean, numeric, string                     |
| neq | 不等。不能为空字符串                                | boolean, numeric, string                     |
| gt  | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02 15:04:05"） |
| gte | 大于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02 15:04:05"） |
| lt  | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02 15:04:05"） |
| lte | 小于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02 15:04:05"） |
| in  | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string                     |
| nin | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string                     |
| cs  | 模糊查询，区分大小写                                | string                                       |
| cis | 模糊查询，不区分大小写                               | string                                       |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                            |
|---------|--------|-----|---------------------------------|
| start   | int    | 否  | 记录开始位置，start 起始值为0       |
| limit   | int    | 是  | 每页限制条数，最大200              |
| sort    | string | 否  | 排序字段，返回数据将按该字段进行排序  |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "filter":{
    "op": "and",
    "rules":[
      {
        "field":"device_type",
        "op":"eq",
        "value":"B70"
      }
    ]
  },
  "page":{
    "start":0,
    "limit": 100,
    "sort":"device_type"
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data":{
    "count":1,
    "info":[
      {
        "device_type":"B70",
        "cpu":24,
        "mem":128,
        "raid":"RAID1",
        "network":"TENTHOUSAND",
        "remark":"2个12核CPU,128G内存,Raid,2*300GSAS,万兆网卡",
        "label":{},
        "device_group":""
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| result     | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| data	     | object array | 响应数据             |

#### data

| 参数名称 | 参数类型       | 描述                    |
|---------|--------------|-------------------------|
| count   | int          | 当前规则能匹配到的总记录条数 |
| info    | object array | 私有子网信息列表           |

#### data.info

| 参数名称      | 参数类型   | 描述          |
|--------------|----------|---------------|
| device_type  | string   | 设备型号       |
| cpu          | int      | CPU核数，单位个 |
| mem          | int      | 内存大小，单位G  |
| raid         | string   | RAID类型       |
| network      | string   | 网络类型。"ONETHOUSAND": 千兆, "TENTHOUSAND": 万兆  |
| remark       | string   | 备注信息        |
| label        | object   | 实例族信息      |

#### data.info.label

| 参数名称      | 参数类型   | 描述        |
|--------------|----------|-------------|
| device_group | string   | 实例族，当前支持的实例族：GAMESERVER, DBSERVICE, HIGHFREQ |