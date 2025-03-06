### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：账单查看权限。
- 该接口功能描述：账单同步记录查询。

### URL

POST /api/v1/account/bills/sync_records/list


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

| 参数名称  | 参数类型   | 必选  | 描述                                                                                                                                                  |
|-------|--------|-----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是   | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否   | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否   | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否   | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否   | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

#### 查询参数介绍：

| 参数名称       | 参数类型    | 描述         |
|------------|---------|------------|
| id         | string  | 资源ID       |
| vendor     | string  | 云服务商名称     |
| bill_year  | int     | 账单年份       |
| bill_month | int     | 账单月份       |
| state      | string  | 账单状态       |
| currency   | string  | 货币         |
| cost       | decimal | 账单总金额      |
| rmb_cost   | decimal | 账单总金额（人民币） |
| detail     | string  | 账单详情       |
| operator   | string  | 操作人        |
| creator    | string  | 创建人        |
| reviser    | string  | 最后修改人      |
| created_at | string  | 创建时间       |
| updated_at | string  | 更新时间       |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。


### 响应示例

#### 导出成功结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 0,
    "details": [
      {
        "id": "0000000c",
        "vendor": "huawei",
        "bill_year": 2024,
        "bill_month": 5,
        "state": "asyncing",
        "currency": "USD",
        "cost": "669064.3429381",
        "rmb_cost": "4755408.2706496926",
        "detail": "",
        "operator": "xxx",
        "creator": "xxxx",
        "reviser": "xxx",
        "created_at": "2024-06-26T00:07:50Z",
        "updated_at": "2024-06-26T00:07:50Z"
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

#### data参数说明

| 参数名称    | 参数类型  | 描述    |
|---------|-------|-------|
| count   | int   | 总记录条数 |
| details | array | 账单列表  |

#### details参数说明

| 参数名称       | 参数类型    | 描述         |
|------------|---------|------------|
| id         | string  | 资源ID       |
| vendor     | string  | 云服务商名称     |
| bill_year  | int     | 账单年份       |
| bill_month | int     | 账单月份       |
| state      | string  | 账单状态       |
| currency   | string  | 货币         |
| cost       | decimal | 账单总金额      |
| rmb_cost   | decimal | 账单总金额（人民币） |
| detail     | string  | 账单详情       |
| operator   | string  | 操作人        |
| creator    | string  | 创建人        |
| reviser    | string  | 最后修改人      |
| created_at | string  | 创建时间       |
| updated_at | string  | 更新时间       |
