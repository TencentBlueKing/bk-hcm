### 描述

- 该接口提供版本：v1.6.4+。
- 该接口所需权限：账单查看。
- 该接口功能描述：导出二级账号账单汇总数据。

### URL

POST /api/v1/account/bills/main_account_summarys/export

### 输入参数

| 参数名称         | 参数类型   | 必选 | 描述               |
|--------------|--------|----|------------------|
| bill_year    | int    | 是  | 账单年份             |
| bill_month   | int    | 是  | 账单月份             |
| export_limit | int    | 是  | 导出限制条数, 0-200000 |
| filter       | object | 是  | 查询过滤条件           |

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

#### 查询参数介绍：



| 参数名称                          | 参数类型    | 描述                                                  |
|-------------------------------|---------|-----------------------------------------------------|
| id                            | string  | ID                                                  |
| root_account_id               | string  | 一级账号ID                                              |
| root_account_name             | string  | 一级账号名称                                              |
| main_account_id               | string  | 二级账号ID                                              |
| main_account_name             | string  | 二级账号名称                                              |
| vendor                        | string  | 供应商（枚举值：tcloud、aws、azure、gcp、huawei）                |
| product_id                    | int64   | 产品ID                                                |
| product_name                  | string  | 产品名称                                                |
| bk_biz_id                     | int64   | 业务ID                                                |
| bk_biz_name                   | string  | 业务名称                                                |
| bill_year                     | int     | 账单年份                                                |
| bill_month                    | int     | 账单月份                                                |
| last_synced_version           | int64   | 最后同步的账单版本                                           |
| current_version               | int64   | 当前账单版本                                              |
| currency                      | string  | 币种                                                  |
| last_month_cost_synced        | decimal | 上月已同步账单                                             |
| last_month_rmb_cost_synced    | decimal | 上月已同步人民币账单                                          |
| current_month_cost_synced     | decimal | 本月已同步账单                                             |
| current_month_rmb_cost_synced | decimal | 本月已同步人民币账单                                          |
| month_on_month_value          | float   | 本月已同步账单环比                                           |
| current_month_cost            | decimal | 实时账单                                                |
| current_month_rmb_cost        | decimal | 实时人民币账单                                             |
| rate                          | float   | 汇率                                                  |
| adjustment_cost               | decimal | 实时调账账单                                              |
| adjustment_rmb_cost           | decimal | 实时人民币调账账单                                           |
| state                         | string  | 状态(accounting, accounted, syncing, synced, stopped) |


接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 获取详细信息请求参数示例

导出2024年1月份的账单数据，限制导出条数为100条。

```json
{
  "bill_year": 2024,
  "bill_month": 1,
  "export_limit": 100,
  "filter":{
    "op": "and",
    "rules": [

    ]
  }
}
```



### 响应示例

#### 导出成功结果示例

Content-Type: application/octet-stream
Content-Disposition: attachment; filename="bill_summary_main.csv.zip"
[二进制文件流]
