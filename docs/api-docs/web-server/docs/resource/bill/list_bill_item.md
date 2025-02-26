### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：账单查看权限。
- 该接口功能描述：获取账单明细。

### URL

POST /api/v1/account/vendors/{vendor}/bills/items/list

### 输入参数

| 参数名称       | 类型     | 必选 | 描述     |
|------------|--------|----|--------|
| bill_year  | int    | 是  | 账单年份   |
| bill_month | int    | 是  | 账单月份   |
| filter     | object | 是  | 查询过滤条件 |
| page       | object | 是  | 分页设置   |


#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

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

| 参数名称            | 参数类型    | 描述     |
|-----------------|---------|--------|
| id              | string  | ID     |
| root_account_id | string  | 一级账号ID |
| main_account_id | string  | 二级账号ID |
| vendor          | string  | 云厂商    |
| product_id      | int64   | 产品ID   |
| bk_biz_id       | int64   | 业务ID   |
| bill_year       | int64   | 账单年份   |
| bill_month      | int32   | 账单月份   |
| bill_day        | int32   | 账单日    |
| version_id      | int64   | 版本号    |
| currency        | string  | 货币类型   |
| cost            | decimal | 费用     |
| hc_product_code | string  | 产品编码   |
| hc_product_name | string  | 产品名称   |
| res_amount      | decimal | 资源金额   |
| res_amount_unit | string  | 资源金额单位 |
| extension       | json    | 扩展字段   |
| creator         | string  | 创建者    |
| reviser         | string  | 修改者    |
| created_at      | string  | 创建时间   |
| updated_at      | string  | 更新时间   |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 响应示例

#### 导出成功结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 123412,
    "details": [
      {
        "id": "00000001",
        "root_account_id": "00000001",
        "main_account_id": "00000001",
        "vendor": "aws",
        "product_id": 123456,
        "bk_biz_id": 123456,
        "bill_year": 2021,
        "bill_month": 1,
        "bill_day": 1,
        "version_id": 123456,
        "currency": "USD",
        "cost": "100.00",
        "hc_product_code": "hcbm001",
        "hc_product_name": "云主机",
        "res_amount": "100.00",
        "res_amount_unit": "USD",
        "creator": "admin",
        "reviser": "admin",
        "created_at": "2021-01-01T00:00:00Z",
        "updated_at": "2021-01-01T00:00:00Z",
        "extension": null
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
| data    | object | 返回数据 |

#### data

| 参数名称    | 参数类型  | 描述   |
|---------|-------|------|
| count   | int64 | 总数   |
| details | array | 详细数据 |

####  details

| 参数名称              | 参数类型    | 描述         |
|-------------------|---------|------------|
| id                | string  | ID         |
| root_account_id   | string  | 一级账号ID     |
| main_account_id   | string  | 二级账号ID     |
| vendor            | string  | 云厂商        |
| product_id        | int64   | 产品ID       |
| bk_biz_id         | int64   | 业务ID       |
| bill_year         | int64   | 账单年份       |
| bill_month        | int32   | 账单月份       |
| bill_day          | int32   | 账单日        |
| version_id        | int64   | 版本号        |
| currency          | string  | 货币类型       |
| cost              | decimal | 费用         |
| hc_product_code   | string  | 产品编码       |
| hc_product_name   | string  | 产品名称       |
| res_amount        | decimal | 资源金额       |
| res_amount_unit   | string  | 资源金额单位     |
| created_at        | string  | 创建时间       |
| updated_at        | string  | 更新时间       |
| main_account_name | string  | 二级账号名称     |
| root_account_name | string  | 一级账号名称     |
| extension         | object  | 云厂商的原始账单结构 |

#### extension

##### aws

| 参数名称                                                             | 参数类型   | 描述                  |
|------------------------------------------------------------------|--------|---------------------|
| bill_bill_type                                                   | string | 计费类别                |
| bill_billing_entity                                              | string | 账单实体                |
| bill_billing_period_end_date                                     | string | 账单周期截止日期            |
| bill_billing_period_start_date                                   | string | 账单周期开始日期            |
| bill_invoice_id                                                  | string | 账单清单ID              |
| bill_invoicing_entity                                            | string | 账单清单实体              |
| bill_payer_account_id                                            | string | 账单支付账号ID            |
| discount_edp_discount                                            | string | edp优惠金额             |
| discount_total_discount                                          | string | 总优惠金额               |
| identity_line_item_id                                            | string | 项目ID                |
| identity_time_interval                                           | string | 标识时间间隔              |
| line_item_availability_zone                                      | string | 可用区                 |
| line_item_blended_cost                                           | string | 混合成本                |
| line_item_blended_rate                                           | string | 混合费率                |
| line_item_currency_code                                          | int    | 项目当前代码              |
| line_item_legal_entity                                           | string | 项目合法实体              |
| line_item_line_item_description                                  | string | 计费描述                |
| line_item_line_item_type                                         | string | 项目类型                |
| line_item_net_unblended_cost                                     | string | 未混合折后成本             |
| line_item_net_unblended_rate                                     | string | 未混合折后费率             |
| line_item_normalization_factor                                   | string | 项目规范因子              |
| line_item_normalized_usage_amount                                | string | 规范化使用费用             |
| line_item_operation                                              | string | 项目操作                |
| line_item_product_code                                           | string | 项目产品代码              |
| line_item_resource_id                                            | string | 资源ID                |
| line_item_tax_type                                               | string | 项目税费类型              |
| line_item_unblended_cost                                         | string | 未混合成本               |
| line_item_unblended_rate                                         | string | 未混合费率               |
| line_item_usage_account_id                                       | string | 使用的账号ID             |
| line_item_usage_amount                                           | string | 使用金额                |
| line_item_usage_end_date                                         | string | 使用截止日期              |
| line_item_usage_start_date                                       | string | 使用开始日期              |
| line_item_usage_type                                             | string | 使用类型                |
| pricing_currency                                                 | string | 定价货币                |
| pricing_lease_contract_length                                    | string | 定价合同长度              |
| pricing_offering_class                                           | string | 报价类别                |
| pricing_public_on_demand_cost                                    | string | 定价公开需求成本            |
| pricing_public_on_demand_rate                                    | string | 定价公开需求费率            |
| pricing_purchase_option                                          | string | 付款方式：全量预付、部分预付、无预付  |
| pricing_term                                                     | string | 使用量是预留还是按需          |
| pricing_unit                                                     | string | 费用计价单位              |
| product_database_engine                                          | string | 产品数据库引擎             |
| product_from_location                                            | string | 产品来源                |
| product_from_location_type                                       | string | 产品来源类型              |
| product_from_region_code                                         | string | 产品区域编码              |
| product_instance_type                                            | string | 产品实例类型              |
| product_instance_type_family                                     | string | 产品实例类型系列            |
| product_location                                                 | string | 产品定位                |
| product_location_type                                            | string | 产品定位类型              |
| product_marketoption                                             | string | 市场选项                |
| product_normalization_size_factor                                | string | 产品规格因子              |
| product_operation                                                | string | 产品操作                |
| product_product_family                                           | string | 产品系列                |
| product_product_name                                             | string | 产品名称                |
| product_purchase_option                                          | string | 产品采购选项              |
| product_purchaseterm                                             | string | 产品采购条款              |
| product_region                                                   | string | 产品区域                |
| product_region_code                                              | string | 产品区域编码              |
| product_servicecode                                              | string | 产品服务编码              |
| product_servicename                                              | string | 产品服务名称              |
| product_tenancy                                                  | string | 产品库存                |
| product_to_location                                              | string | 产品指向的位置             |
| product_to_location_type                                         | string | 产品指向的位置类型           |
| product_to_region_code                                           | string | 产品指向的区域的编码          |
| product_transfer_type                                            | string | 产品传输类型              |
| reservation_amortized_upfront_cost_for_usage                     | string | 预留摊销前期使用成本          |
| reservation_amortized_upfront_fee_for_billing_period             | string | 预留摊销预付费账单周期         |
| reservation_effective_cost                                       | string | 预留有效成本              |
| reservation_end_time                                             | string | 预留截止时间              |
| reservation_modification_status                                  | string | 预留修改状态              |
| reservation_net_amortized_upfront_cost_for_usage                 | string | 预留网络摊销可用成本          |
| reservation_net_amortized_upfront_fee_for_billing_period         | string | 预留网络摊销预付费账单周期       |
| reservation_net_effective_cost                                   | string | 预留网络有效成本            |
| reservation_net_recurring_fee_for_usage                          | string | 预留可用的常用费用           |
| reservation_net_unused_amortized_upfront_fee_for_billing_period  | string | 预留网络未使用预付费账单周期      |
| reservation_net_unused_recurring_fee                             | string | 预留网络未使用常用费用         |
| reservation_net_upfront_value                                    | string | 预留前期净值              |
| reservation_normalized_units_per_reservation                     | string | 预留规范化单位每次保留量        |
| reservation_number_of_reservations                               | string | 预留数量                |
| reservation_recurring_fee_for_usage                              | string | 预留可用的常用费用           |
| reservation_reservation_a_r_n                                    | string | 预留的ARN              |
| reservation_start_time                                           | string | 预留开始时间              |
| reservation_subscription_id                                      | string | 预留的订阅ID             |
| reservation_total_reserved_normalized_units                      | string | 预留总服务标准化单位          |
| reservation_total_reserved_units                                 | string | 预留总服务单位             |
| reservation_units_per_reservation                                | string | 预留每次保留的单位           |
| reservation_unused_amortized_upfront_fee_for_billing_period      | string | 预留未使用冻结的预付费计费周期     |
| reservation_unused_normalized_unit_quantity                      | string | 预留未使用规范化单位数量        |
| reservation_unused_quantity                                      | string | 预留未使用的数量            |
| reservation_unused_recurring_fee                                 | string | 预留未使用的现金            |
| reservation_upfront_value                                        | string | 预留上行数值              |
| savings_plan_amortized_upfront_commitment_for_billing_period     | string | 账单期的计划摊销前期承诺        |
| savings_plan_end_time                                            | string | SavingsPlan截止时间     |
| savings_plan_net_amortized_upfront_commitment_for_billing_period | string | SavingsPlan承诺账单周期   |
| savings_plan_net_recurring_commitment_for_billing_period         | string | SavingsPlan计划净现金周期  |
| savings_plan_net_savings_plan_effective_cost                     | string | SavingsPlan网络有效成本   |
| savings_plan_offering_type                                       | string | SavingsPlan报价类型     |
| savings_plan_payment_option                                      | string | SavingsPlan支付类型     |
| savings_plan_purchase_term                                       | string | SavingsPlan采购期限     |
| savings_plan_recurring_commitment_for_billing_period             | string | SavingsPlan现金承诺账单周期 |
| savings_plan_region                                              | string | SavingsPlan区域       |
| savings_plan_savings_plan_a_r_n                                  | string | SavingsPlanARN      |
| savings_plan_savings_plan_effective_cost                         | string | SavingsPlan有效成本     |
| savings_plan_savings_plan_rate                                   | string | SavingsPlan计划费率     |
| savings_plan_start_time                                          | string | SavingsPlan计划开始时间   |
| savings_plan_total_commitment_to_date                            | string | SavingsPlan总承诺日期    |
| savings_plan_used_commitment                                     | string | SavingsPlan已使用承诺    |

##### huawei


| 参数名称                        | 参数类型                | 描述                                                                                                                                                                                                               |
|-----------------------------|---------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| cycle                       | string              | 资源详单数据所在账期,东八区时间,格式为YYYY-MM。 例如2020-01                                                                                                                                                                           |
| bill_date                   | string              | 消费日期,东八区时间,格式为YYYY-MM-DD。  说明: 当statistic_type=2时该字段才有值,否则返回null。                                                                                                                                                |
| bill_type                   | int64               | 账单类型。 1:消费-新购2:消费-续订3:消费-变更4:退款-退订5:消费-使用8:消费-自动续订9:调账-补偿14:消费-服务支持计划月末扣费15:消费-税金16:调账-扣费17:消费-保底差额 说明: 保底差额=客户签约保底合同后,如果没有达到保底消费,客户需要补交的费用,仅限于直销或者伙伴顾问销售类子客户,且为后付费用户。 20:退款-变更100:退款-退订税金101:调账-补偿税金102:调账-扣费税金 |
| customer_id                 | string              | 消费的客户账号ID                                                                                                                                                                                                        |
| region                      | string              | 云服务区编码,例如:“ap-southeast-1”                                                                                                                                                                                       |
| region_name                 | string              | 云服务区名称,例如:“中国-香港”                                                                                                                                                                                                |
| cloud_service_type          | string              | 云服务类型编码                                                                                                                                                                                                          |
| resource_Type_code          | string              | 资源类型编码                                                                                                                                                                                                           |
| cloud_service_type_name     | string              | 云服务类型名称                                                                                                                                                                                                          |
| resource_type_name          | string              | 资源类型名称                                                                                                                                                                                                           |
| res_instance_id             | string              | 资源实例ID                                                                                                                                                                                                           |
| resource_name               | string              | 资源名称                                                                                                                                                                                                             |
| resource_tag                | string              | 资源标签                                                                                                                                                                                                             |
| sku_code                    | string              | SKU编码,在账单中唯一标识一个资源的规格                                                                                                                                                                                            |
| enterprise_project_id       | string              | 企业项目标识(企业项目ID)                                                                                                                                                                                                   |
| enterprise_project_name     | string              | 企业项目名称                                                                                                                                                                                                           |
| charge_mode                 | int64               | 计费模式。 1 : 包年/包月3:按需10:预留实例                                                                                                                                                                                       |
| consume_amount              | float64             | 客户购买云服务类型的消费金额,包含代金券、现金券,精确到小数点后8位。说明: consume_amount的值等于cash_amount,credit_amount,coupon_amount,flexipurchase_coupon_amount,stored_card_amount,bonus_amount,debt_amount,adjustment_amount的总和                    |
| cash_amount                 | float64             | 现金支付金额                                                                                                                                                                                                           |
| credit_amount               | float64             | 信用额度支付金额                                                                                                                                                                                                         |
| coupon_amount               | float64             | 代金券支付金额                                                                                                                                                                                                          |
| flexipurchase_coupon_amount | float64             | 现金券支付金额                                                                                                                                                                                                          |
| stored_card_amount          | float64             | 储值卡支付金额                                                                                                                                                                                                          |
| bonus_amount                | float64             | 奖励金支付金额                                                                                                                                                                                                          |
| debt_amount                 | float64             | 欠费金额                                                                                                                                                                                                             |
| adjustment_amount           | float64             | 欠费核销金额                                                                                                                                                                                                           |
| official_amount             | float64             | 官网价                                                                                                                                                                                                              |
| discount_amount             | float64             | 对应官网价折扣金额                                                                                                                                                                                                        |
| measure_id                  | int64               | 金额单位。 1:元                                                                                                                                                                                                        |
| period_type                 | int64               | 周期类型: 19:年20:月24:天25:小时5:一次性                                                                                                                                                                                     |
| root_resource_id            | string              | 根资源标识                                                                                                                                                                                                            |
| parent_resource_id          | string              | 父资源标识                                                                                                                                                                                                            |
| trade_id                    | string              | 订单ID 或 交易ID。 账单类型为1,2,3,4,8时为订单ID;其它场景下为: 交易ID(非月末扣费:应收ID;月末扣费:账单ID)                                                                                                                                             |
| product_spec_desc           | string              | 产品的规格描述                                                                                                                                                                                                          |
| sub_service_type_code       | string              | 该字段为预留字段                                                                                                                                                                                                         |
| sub_service_type_name       | string              | 该字段为预留字段                                                                                                                                                                                                         |
| sub_resource_type_code      | string              | 该字段为预留字段                                                                                                                                                                                                         |
| sub_resource_type_name      | string              | 该字段为预留字段                                                                                                                                                                                                         |
| sub_resource_id             | string              | 该字段为预留字段                                                                                                                                                                                                         |
| sub_resource_name           | string              | 该字段为预留字段                                                                                                                                                                                                         |
| pre_order_id                | string              | 该字段为预留字段                                                                                                                                                                                                         |
| az_code_infos               | array of AzCodeInfo | 该字段为预留字段                                                                                                                                                                                                         |

###### details[n].AzCodeInfo

| 参数名称   | 参数类型   | 描述                                |
|--------|--------|------------------------------------------|
| az_code | string | 参数名称:可用区编码，参数的约束及描述:该参数非必填,且只允许字符串 |


##### azure/kaopu

null


##### gcp

| 参数名称                          | 参数类型    | 描述                             |
|-------------------------------|---------|--------------------------------|
| billing_account_id            | string  | 与使用量相关的 Cloud Billing 账号ID     |
| cost                          | float64 | 成本                             |
| cost_type                     | string  | 费用类型                           |
| country                       | string  | 国家                             |
| credits_amount                | string  | 赠送金信息                          |
| currency                      | string  | 币种                             |
| currency_conversion_rate      | float64 | 货币转换率                          |
| location                      | string  | 区域信息                           |
| month                         | string  | 账单年月                           |
| project_id                    | string  | 项目ID                           |
| project_name                  | string  | 项目名称                           |
| project_number                | string  | 项目编号                           |
| region                        | string  | 区域                             |
| resource_global_name          | string  | 资源全局唯一标识符                      |
| resource_name                 | string  | 资源名称                           |
| service_description           | string  | 服务描述                           |
| service_id                    | string  | 服务ID                           |
| sku_description               | string  | 资源类型描述                         |
| sku_id                        | string  | 资源类型ID                         |
| total_cost                    | float64 | 总成本                            |
| return_cost                   | float64 | 退款成本                           |
| usage_amount                  | string  | 可用金额                           |
| usage_amount_in_pricing_units | string  | 可用金额单价                         |
| usage_end_time                | string  | 可用结束时间，示例：2023-04-16T15:00:00Z |
| usage_pricing_unit            | float64 | 可用金额单价的单位                      |
| usage_start_time              | string  | 可用开始时间，示例：2023-04-16T15:00:00Z |
| usage_unit                    | string  | 可用金额单位                         |
| zone                          | string  | 可用区                            |
| credit_infos                  | array   | credit_info                    |

###### credit_infos[n]

| 参数名称      | 参数类型    | 描述 |
|-----------|---------|----|
| id        | string  | ID |
| amount    | float64 | 金额 |
| type      | string  | 类型 |
| name      | string  | 名称 |
| full_name | string  | 全称 |

##### zenlayer

| 参数名称            | 参数类型    | 描述           |
|-----------------|---------|--------------|
| bill_id         | string  | 账单ID         |
| zenlayer_order  | string  | Zenlayer订单编号 |
| cid             | string  | CID          |
| group_id        | string  | GROUP ID     |
| currency        | string  | 币种           |
| city            | string  | 城市           |
| pay_content     | string  | 付费内容         |
| type            | string  | 类型           |
| acceptance_num  | decimal | 验收数量         |
| pay_num         | decimal | 付费数量         |
| unit_price_usd  | decimal | 单价USD        |
| total_payable   | decimal | 应付USD        |
| billing_period  | string  | 账期           |
| contract_period | string  | 合约周期         |
| remarks         | string  | 备注           |
| business_group  | string  | 业务组          |
| cpu             | string  | CPU          |
| disk            | string  | 硬盘           |
| memory          | string  | 内存           |
