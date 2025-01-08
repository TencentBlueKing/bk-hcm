### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：账单查看权限。
- 该接口功能描述：获取当月筛选出来的二级账号总金额。

### URL

POST /api/v1/account/bills/main_account_summarys/sum

### 输入参数

| 参数名称            | 类型     | 必选 | 描述     |
|-----------------|--------|----|--------|
| bill_year       | int    | 是  | 账单年份   |
| bill_month      | int    | 是  | 账单月份   |
| filter          | object | 是  | 查询过滤条件 |

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

| 参数名称                          | 参数类型    | 描述        |
|-------------------------------|---------|-----------|
| id                            | string  | ID        |
| root_account_id               | string  | 一级账号ID    |
| root_account_cloud_id         | string  | 一级账号云ID   |
| main_account_id               | string  | 二级账号ID    |
| main_account_cloud_id         | string  | 二级账号云ID   |
| vendor                        | string  | 云厂商       |
| product_id                    | int64   | 产品ID      |
| product_name                  | string  | 产品名称      |
| bk_biz_id                     | int64   | 业务ID      |
| bk_biz_name                   | string  | 业务名称      |
| bill_year                     | int64   | 账单年份      |
| bill_month                    | int32   | 账单月份      |
| last_synced_version           | int64   | 上一次同步的版本号 |
| current_version               | int64   | 当前版本号     |
| currency                      | string  | 货币类型      |
| last_month_cost_synced        | decimal | 上月费用      |
| last_month_rmb_cost_synced    | decimal | 上月人民币费用   |
| current_month_cost_synced     | decimal | 本月费用      |
| current_month_rmb_cost_synced | decimal | 本月人民币费用   |
| month_on_month_value          | float   | 月同比增长值    |
| current_month_cost            | decimal | 本月费用      |
| current_month_rmb_cost        | decimal | 本月人民币费用   |
| rate                          | float   | 费用比率      |
| adjustment_cost               | decimal | 调整金额      |
| adjustment_rmb_cost           | decimal | 调整金额（人民币） |
| state                         | string  | 账单状态      |
| created_at                    | string  | 创建时间      |
| updated_at                    | string  | 更新时间      |


接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 响应示例

#### 导出成功结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 123412,
    "cost_map": {
      "USD": {
        "Cost": "781223.32650609",
        "RMBCost": "5552583.8543083598",
        "Currency": "USD"
      }
    }
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 返回数据 |

#### data

| 参数名称     | 参数类型              | 描述  |
|----------|-------------------|-----|
| count    | int64             | 总数  |
| cost_map | map[string]object | 费用表 |


#### cost_map

key: 币种
value: 费用

| 参数名称     | 参数类型   | 描述    |
|----------|--------|-------|
| Cost     | string | 费用    |
| RMBCost  | string | 人民币费用 |
| Currency | string | 币种    |
