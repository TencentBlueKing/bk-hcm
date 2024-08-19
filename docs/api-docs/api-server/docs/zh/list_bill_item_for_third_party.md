### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：云账单拉取。
- 该接口功能描述：查询云账单接口列表。

### 输入参数

#### url 参数

| 参数名称   | 参数类型   | 必选 | 描述  |
|--------|--------|----|-----|
| vendor | string | 是  | 云厂商 |

##### vendor 列表：

- aws
- azure
- gcp
- huawei

#### Body参数

| 参数名称                   | 参数类型         | 必选 | 描述                |
|------------------------|--------------|----|-------------------|
| bill_year              | uint         | 是  | 账单年份              |
| bill_month             | uint         | 是  | 账单月份              |
| begin_bill_day         | uint         | 否  | 账单开始日，需和账单截止日一起设定 |
| end_bill_day           | uint         | 否  | 账单截止日，需和账单开始日一起设定 |
| root_account_ids       | string array | 否  | 根账号ID列表           |
| root_account_cloud_ids | string array | 否  | 根账号云ID            |
| main_account_ids       | string array | 否  | 主账号ID列表           |
| main_account_cloud_ids | string array | 否  | 主账号云ID            |
| page                   | object       | 是  | 分页设置              |

##### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

### 调用示例

#### 获取详细信息请求参数示例

如查询账号ID为"00000012"的Aws账单列表接口。

```json
{
  "bill_year": 2024,
  "bill_month": 7,
  "begin_bill_day": 30,
  "end_bill_day": 30,
  "filter": {
    "op": "and",
    "rules": []
  },
  "page": {
    "limit": 100,
    "start": 0,
    "count": false
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

| 参数名称   | 参数类型   | 描述             |
|--------|--------|----------------|
| count  | uint64 | 当前规则能匹配到的总记录条数 |
| detail | array  | 查询返回的数据        |

#### detail

| 参数名称                 | 参数类型   | 描述                             |
|----------------------|--------|--------------------------------|
| id                   | string | 账单id                           |
| root_account_id      | string | 一级账号id                         |
| main_account_id      | string | 二级账号id                         |
| vendor               | string | 云厂商                            |
| product_id           | int32  | 产品id                           |
| bk_biz_id            | int32  | 业务id                           |
| bill_year            | int32  | 账单年份, 如: 2024                  |
| bill_month           | int32  | 账单月份, 如: 7                     |
| bill_day             | int32  | 账单日, 如: 1                      |
| version_id           | int32  | 账单版本id                         |
| currency             | string | 货币类型                           |
| cost                 | string | 费用                             |
| hc_product_code      | string | 产品编码                           |
| hc_product_name      | string | 产品名称                           |
| hc_product_type      | string | 产品类型                           |
| hc_product_type_name | string | 产品类型名称                         |
| hc_product_type_code | string | 产品类型编码                         |
| res_amount           | string | 资源用量                           |
| res_amount_unit      | string | 资源用量单位                         |
| creator              | string | 创建者                            |
| reviser              | string | 更新者                            |
| created_at           | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at           | string | 更新时间，标准格式：2006-01-02T15:04:05Z |
| extension            | object | 云上原始账单格式，见下文                   |

### AWS的响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 1,
    "details": [
      {
        "id": "00000001",
        "root_account_id": "00000001",
        "main_account_id": "123456789",
        "vendor": "aws",
        "product_id": -1,
        "bk_biz_id": -1,
        "bill_year": 2024,
        "bill_month": 7,
        "bill_day": 20,
        "version_id": 1,
        "currency": "USD",
        "cost": "0.0000001",
        "hc_product_code": "AmazonS3",
        "hc_product_name": "Amazon Simple Storage Service",
        "res_amount": "1",
        "res_amount_unit": "Requests",
        "creator": "",
        "reviser": "",
        "created_at": "2024-08-05T04:51:03Z",
        "updated_at": "2024-08-05T04:51:03Z",
        "extension": {
          "pricing_term": "OnDemand",
          "pricing_unit": "GB-Mo",
          "product_region": "us-east-1",
          "bill_invoice_id": "123456789",
          "product_location": "US East (N. Virginia)",
          "bill_billing_entity": "AWS",
          "line_item_operation": "StandardStorage",
          "line_item_usage_type": "TimedStorage-ByteHrs",
          "product_product_name": "Amazon Simple Storage Service",
          "bill_payer_account_id": "123456789",
          "identity_line_item_id": "23e6dzvbifkirvrqoad3rer7",
          "line_item_resource_id": "rss_id",
          "product_instance_type": "",
          "line_item_product_code": "AmazonS3",
          "line_item_usage_amount": "0.0000000000",
          "product_product_family": "Storage",
          "line_item_currency_code": "USD",
          "line_item_line_item_type": "Usage",
          "line_item_unblended_cost": "0.0000000000",
          "line_item_unblended_rate": "0.0000000000",
          "line_item_usage_end_date": "",
          "line_item_usage_account_id": "123456789",
          "line_item_usage_start_date": "",
          "reservation_effective_cost": "0.0",
          "line_item_net_unblended_cost": "0.0000000000",
          "line_item_net_unblended_rate": "0.0",
          "pricing_public_on_demand_cost": "0.0000000000",
          "pricing_public_on_demand_rate": "0.0000000000",
          "reservation_net_effective_cost": "0.0",
          "savings_plan_savings_plan_rate": "0.0",
          "line_item_line_item_description": "$0.0000000000 per GB - first 50 TB / month of storage used",
          "savings_plan_savings_plan_a_r_n": "",
          "savings_plan_savings_plan_effective_cost": "0.0",
          "savings_plan_net_savings_plan_effective_cost": "0.0"
        }
      }
    ]
  }
}
```

#### aws extension 说明

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

### HuaWei的响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 1,
    "details": [
      {
        "id": "00000001",
        "root_account_id": "00000001",
        "main_account_id": "123456789",
        "vendor": "aws",
        "product_id": -1,
        "bk_biz_id": -1,
        "bill_year": 2024,
        "bill_month": 7,
        "bill_day": 20,
        "version_id": 1,
        "currency": "USD",
        "cost": "0.0000001",
        "hc_product_code": "hws.service.type.vpc",
        "hc_product_name": "弹性云服务器",
        "res_amount": "1",
        "res_amount_unit": "1",
        "creator": "",
        "reviser": "",
        "created_at": "2024-08-05T04:51:03Z",
        "updated_at": "2024-08-05T04:51:03Z",
        "extension": {
          "bill_date": "2020-12-05",
          "bill_type": 1,
          "customer_id": "52190d93cb844a249c70fd1e1d416f8b",
          "region": "cn-north-1",
          "region_name": "CN North-Beijing1",
          "cloud_service_type": "hws.service.type.ec2",
          "resource_type": "hws.resource.type.ip",
          "effective_time": "2020-12-05T11:06:55Z",
          "expire_time": "2020-12-06T11:06:55Z",
          "resource_id": "4251f987c09c4d97a6b4784e4661f8ce",
          "resource_name": "hws.service.type.vpcname",
          "resource_tag": "resourceTag",
          "product_id": "00301-110660-0--0",
          "product_name": "调试15_4核8G_linux 包年",
          "product_spec_desc": "调试15_4核8G_linux",
          "sku_code": "comtest15.linux",
          "spec_size": 40,
          "spec_size_measure_id": 0,
          "trade_id": "BC0883684711",
          "id": "037e8a2b-bde******9eb5153cba_1",
          "trade_time": "2020-12-05T11:07:00Z",
          "enterprise_project_id": "0",
          "enterprise_project_name": "default",
          "charge_mode": "1",
          "order_id": "BC0883684711",
          "period_type": "20",
          "usage_type": "dsfhjgbk",
          "usage": 147,
          "usage_measure_id": 1,
          "free_resource_usage": 258,
          "free_resource_measure_id": 1,
          "ri_usage": 30,
          "ri_usage_measure_id": 0,
          "unit_price": 0,
          "unit": "元/1个(次)",
          "official_amount": 0.81,
          "discount_amount": 0.01,
          "amount": 0.81,
          "cash_amount": 2.25,
          "credit_amount": 1.23,
          "coupon_amount": 0.07,
          "flexipurchase_coupon_amount": 0.4,
          "stored_card_amount": 0.34,
          "bonus_amount": 4.63,
          "debt_amount": -8.11,
          "adjustment_amount": 3.69,
          "measure_id": 1,
          "sub_service_type_code": null,
          "sub_service_type_name": null,
          "sub_resource_type_code": null,
          "sub_resource_type_name": null,
          "sub_resource_id": null,
          "sub_resource_name": null,
          "formula": "（2月）【周期数】/（1）【周期转换】*（5997.5641元/月）【单价】-0.00【优惠金额】-0.00【代金券抵扣】"
        }
      }
    ]
  }
}
```

### HuaWei Extension说明

| 字段名称                        | 数据类型    | 说明                                           |
|-----------------------------|---------|----------------------------------------------|
| bill_date                   | string  | 账单日期，格式为YYYY-MM-DD                           |
| bill_type                   | int     | 账单类型                                         |
| customer_id                 | string  | 消费的客户账号ID。                                   |
| region                      | string  | 云服务区编码                                       |
| region_name                 | string  | 云服务区名称，例如："华北-北京"                            |
| cloud_service_type          | string  | 云服务类型, 例如OBS的云服务类型编码为"hws.service.type.obs"。 |
| resource_type_name          | string  | 资源类型名称。例如ECS的资源类型名称为“云主机”。                   |
| effective_time              | string  | 费用对应的资源使用的开始时间，按需有效，包年/包月该字段保留。              |
| expire_time                 | string  | 费用对应的资源使用的结束时间，按需有效，包年/包月该字段保留。              |
| resource_id                 | string  | 资源ID                                         |
| resource_name               | string  | 资源名称                                         |
| resource_tag                | string  | 资源标签                                         |
| product_id                  | string  | 产品ID。                                        |
| product_name                | string  | 产品名称。                                        |
| product_spec_desc           | string  | 产品的规格描述。                                     |
| sku_code                    | string  | SKU编码，在账单中唯一标识一个资源的规格。                       |
| spec_size                   | string  | 产品的实例大小，仅线性产品有效。                             |
| spec_size_measure_id        | string  | 产品实例大小的单位，仅线性产品有该字段。                         |
| trade_id                    | string  | 订单ID或交易ID，扣费维度的唯一标识。                         |
| id                          | string  | 唯一标识。按账期类型统计时不返回唯一标识                         |
| trade_time                  | string  | 交易时间                                         |
| enterprise_project_id       | string  | 企业项目标识（企业项目ID）                               |
| enterprise_project_name     | string  | 企业项目的名称                                      |
| charge_mode                 | string  | 计费模式。1：包年/包月 3：按需 10：预留实例 11：节省计划            |
| order_id                    | string  | 订单ID。                                        |
| period_type                 | string  | 周期类型：19：年 20：月 24：天 25：小时 5：一次性              |
| usage_type                  | string  | 使用量类型编码                                      |
| usage                       | int     | 资源的使用量                                       |
| usage_measure_id            | int     | 资源使用量的度量单位                                   |
| free_resource_usage         | float   | 套餐内使用量                                       |
| free_resource_measure_id    | int     | 套餐内使用量的度量单位                                  |
| ri_resource_usage           | float   | 预留实例使用量                                      |
| ri_resource_measure_id      | int     | 预留实例使用量单位                                    |
| unit_price                  | float   | 产品的单价                                        |
| unit                        | string  | 产品的单价单位                                      |
| discount_amount             | float64 | 优惠金额                                         |
| official_amount             | float64 | 官方金额                                         |
| trade_amount                | float64 | 交易金额                                         |
| amount                      | float64 | 应付金额                                         |
| cash_amount                 | float64 | 现金支付金额                                       |
| credit_amount               | float64 | 信用额度支付金额                                     |
| coupon_amount               | float64 | 代金券支付金额                                      |
| flexipurchase_coupon_amount | float64 | 现金券支付金额                                      |
| stored_card_amount          | float64 | 储值卡金额                                        |
| debt_amount                 | float64 | 欠费金额                                         |
| adjustment_amount           | float64 | 欠费核销金额                                       |
| measure_id                  | int     | 金额单位。 1: 元                                   |
| formula                     | string  | 实付金额计算公式                                     |
| sub_service_type_code       | string  | 整机的子云服务的自身的云服务类型编码                           |
| sub_service_type_name       | string  | 整机的子云服务的自身的云服务类型名称                           |
| sub_resource_type_code      | string  | 整机的子云服务的自身的资源类型编码                            |
| sub_resource_type_name      | string  | 整机的子云服务的自身的资源类型名称                            |
| sub_resource_name           | string  | 整机的子云服务的自身的资源ID，资源标识。（如果为预留实例，则为预留实例标识）      |
| sub_resource_id             | string  | 整机的子云服务的自身的资源名称，资源标识。（如果为预留实例，则为预留实例标识）      |

#### bill_type 账单类型说明

- 1：消费-新购
- 2：消费-续订
- 3：消费-变更
- 4：退款-退订
- 5：消费-使用
- 8：消费-自动续订
- 9：调账-补偿
- 12：消费-按时计费
- 13：消费-退订手续费
- 14：消费-服务支持计划月末扣费
- 16：调账-扣费
- 18：消费-按月付费
- 20：退款-变更
- 23：消费-节省计划抵扣
- 24：退款-包年/包月转按需

#### details[n].AzCodeInfo

| 参数名称    | 参数类型   | 描述                                 |
|---------|--------|------------------------------------|
| az_code | string | 参数名称:可用区编码，参数的约束及描述:该参数非必填,且只允许字符串 |

### Azure的响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 0,
    "details": [
      {
        "id": "00000001",
        "root_account_id": "00000002",
        "main_account_id": "10000001",
        "vendor": "azure",
        "product_id": -1,
        "bk_biz_id": -1,
        "bill_year": 2024,
        "bill_month": 7,
        "bill_day": 30,
        "version_id": 1,
        "currency": "USD",
        "cost": "0.00000000",
        "hc_product_code": "Microsoft.EventHub",
        "hc_product_name": "Service Bus - Basic Messaging Operations",
        "res_amount": "0.00000000",
        "res_amount_unit": "1M",
        "creator": "",
        "reviser": "",
        "created_at": "2024-01-08T13:57:52Z",
        "updated_at": "2024-01-08T13:57:52Z",
        "extension": {
          "id": "/subscriptions/abcdefg-abcdefg-abcdefg-abcdefg-abcdefg/providers/Microsoft.Billing/billingPeriods/20240701/providers/Microsoft.Consumption/usageDetails/abcdefg-abcdefg-abcdefg-abcdefg-abcdefg",
          "kind": "legacy",
          "name": "abcdefg-3369-a2e2-abcdefg-abcdefg",
          "type": "Microsoft.Consumption/usageDetails",
          "properties": {
            "cost": 0.000,
            "date": "2022-07-10T00:00:00Z",
            "meterId": "abcdefg-28c7-49f0-9456-abcdefg",
            "offerId": "MS-AZR-0017P",
            "product": "Service Bus - Basic Messaging Operations",
            "planName": "Basic",
            "quantity": 0.000,
            "frequency": "UsageBased",
            "payGPrice": 0.000,
            "unitPrice": 0.000,
            "chargeType": "Usage",
            "partNumber": "P5H-00052",
            "resourceId": "/subscriptions/adfsaf-456a-adfsaf-a083-adfsaf/resourceGroups/test_group/",
            "accountName": "tencent_azure_IEG10",
            "meterDetails": {
              "meterName": "Basic Messaging Operations",
              "meterCategory": "Service Bus",
              "unitOfMeasure": "1M",
              "meterSubCategory": ""
            },
            "pricingModel": "OnDemand",
            "resourceName": "hcmEvent",
            "publisherName": "Microsoft",
            "publisherType": "Azure",
            "resourceGroup": "test_group",
            "accountOwnerId": "abc@abc.com",
            "effectivePrice": 0.0425,
            "invoiceSection": "ABC",
            "subscriptionId": "abcdefg-abcdefg-abcdefg-abcdefg-abcdefg",
            "billingCurrency": "USD",
            "consumedService": "Microsoft.EventHub",
            "billingAccountId": "123456789",
            "billingProfileId": "123456789",
            "resourceLocation": "eastasia",
            "subscriptionName": "subname",
            "billingAccountName": "xxx xx. Ltd.",
            "billingProfileName": "xxx xx. Ltd.",
            "billingPeriodEndDate": "2021-07-31T00:00:00Z",
            "isAzureCreditEligible": true,
            "billingPeriodStartDate": "2021-07-01T00:00:00Z"
          }
        }
      }
    ]
  }
}
```

### Azure Extension 参数说明


| 参数名称       | 参数类型       | 描述             |
|------------|------------|----------------|
| id         | string     | 事件的完整限定 ARM ID |
| kind       | string     | 指定使用详细信息的类型    |
| name       | int64      | 唯一标识事件的 ID     |
| properties | Properties | 帐户名称           |
| type       | string     | 资源类型           |

#### Properties

| 参数名称                   | 参数类型                   | 描述                                                                                                        |
|------------------------|------------------------|-----------------------------------------------------------------------------------------------------------|
| accountName            | string                 | 帐户名称                                                                                                      |
| accountOwnerId         | string                 | 帐户所有者 ID                                                                                                  |
| additionalInfo         | string                 | 此使用情况项的其他详细信息。 默认情况下，除非在 $expand 中指定，否则不会填充它。 使用此字段可获取特定于使用情况行项的详细信息，例如实际 VM 大小 (ServiceType) 或应用预留折扣的比率。 |
| benefitId              | string                 | 适用权益的唯一标识符                                                                                                |
| benefitName            | string                 | 适用权益的名称                                                                                                   |
| billingAccountId       | string                 | 计费帐户标识符                                                                                                   |
| billingAccountName     | string                 | 计费帐户名称                                                                                                    |
| billingCurrency        | string                 | 计费货币                                                                                                      |
| billingPeriodEndDate   | string                 | 计费周期结束日期                                                                                                  |
| billingPeriodStartDate | string                 | 计费周期开始日期                                                                                                  |
| billingProfileId       | string                 | 计费配置文件标识符                                                                                                 |
| billingProfileName     | string                 | 计费对象信息名称                                                                                                  |
| chargeType             | string                 | 指示费用表示额度、使用情况、市场购买、预留费用或退款                                                                                |
| consumedService        | string                 | 使用的服务名称。 发出使用情况或已购买的 Azure 资源提供程序的名称。 未为市场使用提供此值                                                          |
| cost                   | int                    | 税前成本金额                                                                                                    |
| costCenter             | string                 | 此部门的成本中心（如果它是一个部门，并且提供了一个成本中心）                                                                            |
| date                   | string                 | 使用情况记录的日期，格式示例：2023-03-02T00:00:00Z                                                                       |
| effectivePrice         | int                    | 按使用量收费的有效价格                                                                                               |
| frequency              | string                 | 指示此费用的发生频率。 OneTime 用于仅发生一次的购买，每月针对每月重复的费用，使用基于服务使用量的费用                                                   |
| invoiceSection         | string                 | 发票科目名称                                                                                                    |
| isAzureCreditEligible  | boolean                | 是否符合 Azure 额度条件                                                                                           |
| meterDetails           | object of MeterDetails | 有关计量的详细信息。 默认情况下，除非在 $expand 中指定，否则不会填充它                                                                  |
| meterId                | string                 | 计量 ID (GUID) 。 不适用于市场。 对于预留实例，它表示为其购买预留的主计量                                                               |
| offerId                | string                 | 产品/服务 ID。例如：MS-AZR-0017P、MS-AZR-0148P                                                                     |
| partNumber             | string                 | 所用服务的部件号。 可用于联接价目表。 不适用于市场                                                                                |
| payGPrice              | int                    | 资源的零售价格                                                                                                   |
| planName               | string                 | 计划名称                                                                                                      |
| pricingModel           | string                 | 指示计量器定价方式的标识符(OnDemand\)                                                                                  |
| product                | string                 | 使用的服务或购买的产品名称。 不适用于市场                                                                                     |
| productOrderId         | string                 | 产品订单 ID。对于预留，这是预留订单 ID                                                                                    |
| productOrderName       | string                 | 产品订单名称。 对于预留，这是购买的 SKU                                                                                    |
| publisherName          | string                 | 发布者名称                                                                                                     |
| publisherType          | string                 | 发布服务器类型                                                                                                   |
| quantity               | int                    | 使用数量                                                                                                      |
| reservationId          | string                 | 预留的 ARM 资源 ID。 仅适用于与预留相关的记录                                                                               |
| reservationName        | string                 | 用户提供的预留的显示名称。 特定日期的姓氏将填充在每日数据中。 仅适用于与预留相关的记录                                                              |
| resourceGroup          | string                 | 资源组名称                                                                                                     |
| resourceId             | string                 | Azure 资源管理器使用情况详细信息资源的唯一标识符                                                                               |
| resourceLocation       | string                 | 资源位置                                                                                                      |
| resourceName           | string                 | 资源名称                                                                                                      |
| serviceInfo1           | string                 | 服务特定的元数据                                                                                                  |
| serviceInfo2           | string                 | 旧字段，具有可选的特定于服务的元数据                                                                                        |
| subscriptionId         | string                 | 订阅 guid                                                                                                   |
| subscriptionName       | string                 | 订阅名称                                                                                                      |
| term                   | string                 | 学期 (（以月) 为单位）。 每月定期购买 1 个月。 1 年预留 12 个月。 3 年预留 36 个月                                                      |
| unitPrice              | int                    | 单价是适用于你的价格。 (EA 或其他合同价格)                                                                                  |

#### MeterDetails

| 参数名称             | 参数类型   | 描述                                       |
|------------------|--------|------------------------------------------|
| meterCategory    | string | 计量的类别，例如“云服务”、“网络”等                      |
| meterName        | string | 给定计量类别中的计量名称                             |
| meterSubCategory | string | 计量的子类别，例如“A6 云服务”、“ExpressRoute (IXP) ”等 |
| serviceFamily    | string | 服务系列                                     |
| unitOfMeasure    | string | 计量消耗量计费的单位，例如“小时”、“GB”等                  |

### Gcp的响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 0,
    "details": [
      {
        "billing_account_id": "ABCDEFG-ABCD-ABCD",
        "cost": 0,
        "cost_type": "regular",
        "country": "US",
        "credits": "[]",
        "currency": "USD",
        "location": "us",
        "month": "202304",
        "project_id": "project_id",
        "project_name": "ProjectID",
        "project_number": "123456789",
        "region": null,
        "resource_global_name": null,
        "resource_name": null,
        "service_description": "Cloud Logging",
        "service_id": "5490-F7B7-8DF6",
        "sku_description": "Log Volume",
        "sku_id": "143F-A1B0-E0BE",
        "usage_amount": 837,
        "usage_amount_in_pricing_units": 7.8e-7,
        "usage_end_time": "2023-04-16T11:00:00Z",
        "usage_pricing_unit": "gibibyte",
        "usage_start_time": "2023-04-16T10:00:00Z",
        "usage_unit": "bytes",
        "zone": null
      }
    ]
  }
}
```

### Gcp Extension 说明

| 参数名称                          | 参数类型    | 描述                             |
|-------------------------------|---------|--------------------------------|
| billing_account_id            | string  | 与使用量相关的 Cloud Billing 帐号ID     |
| cost                          | float64 | 成本                             |
| cost_type                     | string  | 费用类型                           |
| country                       | string  | 国家                             |
| credits                       | json    | 赠送金信息                          |
| currency                      | string  | 币种                             |
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
| usage_amount                  | string  | 可用金额                           |
| usage_amount_in_pricing_units | string  | 可用金额单价                         |
| usage_end_time                | string  | 可用结束时间，示例：2023-04-16T15:00:00Z |
| usage_pricing_unit            | float64 | 可用金额单价的单位                      |
| usage_start_time              | string  | 可用开始时间，示例：2023-04-16T15:00:00Z |
| usage_unit                    | string  | 可用金额单位                         |
| zone                          | string  | 可用区                            |