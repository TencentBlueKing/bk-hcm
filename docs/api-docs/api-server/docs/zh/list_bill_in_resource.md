### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询云账单接口列表。

### 输入参数

| 参数名称   | 参数类型        | 必选  | 描述                               |
|-----------|---------------|------|------------------------------------|
| vendor     | string       |  是   | 云厂商                             |

### AWS的输入参数

| 参数名称   | 参数类型        | 必选  | 描述                               |
|-----------|---------------|------|------------------------------------|
| account_id | string       |  是   | 账号ID                             |
| begin_date | string       |  是   | 开始日期，格式：Y-m-d                |
| end_date   | string       |  是   | 截止日期，格式：Y-m-d                |
| page       | object       |  否   | 分页设置                            |

### TCloud的输入参数

| 参数名称   | 参数类型        | 必选  | 描述                               |
|-----------|---------------|------|------------------------------------|
| account_id | string       |  是   | 账号ID                              |
| begin_date | string       |  是   | 开始日期，格式：Y-m-d H:i:s           |
| end_date   | string       |  是   | 截止日期，格式：Y-m-d H:i:s           |
| page       | object       |  否   | 分页设置                             |

### HuaWei的输入参数

| 参数名称   | 参数类型        | 必选  | 描述                               |
|-----------|---------------|------|------------------------------------|
| account_id | string       |  是   | 账号ID                             |
| month      | string       |  是   | 开始日期，格式：Y-m，示例:2019-01     |
| page       | object       |  否   | 分页设置                            |

### Azure的输入参数

| 参数名称   | 参数类型        | 必选  | 描述                               |
|-----------|---------------|------|------------------------------------|
| account_id | string       |  是   | 账号ID                              |
| begin_date | string       |  是   | 开始日期，格式：Y-m-d                 |
| end_date   | string       |  是   | 截止日期，格式：Y-m-d                 |
| page       | object       |  否   | 分页设置                             |

### Gcp的输入参数

| 参数名称   | 参数类型        | 必选  | 描述                                      |
|-----------|---------------|------|-------------------------------------------|
| account_id | string       |  是   | 账号ID                                    |
| month      | string       |  是   | 开始日期，格式：Ym，示例:202301              |
| begin_date | string       |  是   | 开始日期，UTC格式：2023-01-01T12:30:00.45Z  |
| end_date   | string       |  是   | 截止日期，UTC格式：2023-01-01T12:30:00.45Z  |
| page       | object       |  否   | 分页设置                                   |

#### page

| 参数名称   | 参数类型    | 必选  | 描述                                   |
|--------|---------|-----|----------------------------------------------|
| offset | uint32  | 否	 | 记录开始位置，起始值为0（Azure不需要传，不会生效）  |
| limit	 | uint32  | 否	 | 每页限制条数                                   |
| next_link	| string | 否 | 获取下一页数据的链接 (url) ，为空表示没有更多数据，不为空则需要获取下一页（仅Azure会返回） |

### 调用示例

#### 获取详细信息请求参数示例

如查询账号ID为"00000012"的Aws账单列表接口。

```json
{
  "account_id": "00000012",
  "begin_date": "2023-04-01",
  "end_date": "2023-04-01",
  "page": {
    "offset": 0,
    "limit": 10
  }
}
```

### AWS的响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 4102,
    "details": [
    {
      "bill_bill_type": "Anniversary",
      "bill_billing_entity": "AWS",
      "bill_billing_period_end_date": "2023-05-01 00:00:00.000",
      "bill_billing_period_start_date": "2023-04-01 00:00:00.000",
      "bill_invoice_id": "",
      "bill_invoicing_entity": "Amazon Web Services, Inc.",
      "bill_payer_account_id": "706937564898",
      "discount_edp_discount": "0.0",
      "discount_total_discount": "0.0",
      "identity_line_item_id": "ggj6pdvtzu5egl3vnwl6hjkrvuiit2jkeqi4istfft7cqhedc6xa",
      "identity_time_interval": "2023-04-01T09:00:00Z/2023-04-01T10:00:00Z",
      "line_item_availability_zone": "",
      "line_item_blended_cost": "0.0",
      "line_item_blended_rate": "",
      "line_item_currency_code": "USD",
      "line_item_legal_entity": "",
      "line_item_line_item_description": "Enterprise Discount Program Discount for AmazonEC2",
      "line_item_line_item_type": "EdpDiscount",
      "line_item_net_unblended_cost": "0.0",
      "line_item_net_unblended_rate": "",
      "line_item_normalization_factor": "0.0",
      "line_item_normalized_usage_amount": "0.0",
      "line_item_operation": "AssociateAddressVPC",
      "line_item_product_code": "AmazonEC2",
      "line_item_resource_id": "",
      "line_item_tax_type": "",
      "line_item_unblended_cost": "-0.00133",
      "line_item_unblended_rate": "",
      "line_item_usage_account_id": "239069856473",
      "line_item_usage_amount": "0.0",
      "line_item_usage_end_date": "2023-04-01 10:00:00.000",
      "line_item_usage_start_date": "2023-04-01 09:00:00.000",
      "line_item_usage_type": "ElasticIP:IdleAddress",
      "pricing_currency": "",
      "pricing_lease_contract_length": "",
      "pricing_offering_class": "",
      "pricing_public_on_demand_cost": "0.0",
      "pricing_public_on_demand_rate": "",
      "pricing_purchase_option": "",
      "pricing_term": "",
      "pricing_unit": "",
      "product_database_engine": "",
      "product_from_location": "",
      "product_from_location_type": "",
      "product_from_region_code": "",
      "product_instance_type": "",
      "product_instance_type_family": "",
      "product_location": "",
      "product_location_type": "",
      "product_marketoption": "",
      "product_normalization_size_factor": "",
      "product_operation": "",
      "product_product_family": "",
      "product_product_name": "",
      "product_purchase_option": "",
      "product_purchaseterm": "",
      "product_region": "",
      "product_region_code": "",
      "product_servicecode": "",
      "product_servicename": "",
      "product_tenancy": "",
      "product_to_location": "",
      "product_to_location_type": "",
      "product_to_region_code": "",
      "product_transfer_type": "",
      "reservation_amortized_upfront_cost_for_usage": "0.0",
      "reservation_amortized_upfront_fee_for_billing_period": "0.0",
      "reservation_effective_cost": "0.0",
      "reservation_end_time": "",
      "reservation_modification_status": "",
      "reservation_net_amortized_upfront_cost_for_usage": "0.0",
      "reservation_net_amortized_upfront_fee_for_billing_period": "0.0",
      "reservation_net_effective_cost": "0.0",
      "reservation_net_recurring_fee_for_usage": "0.0",
      "reservation_net_unused_amortized_upfront_fee_for_billing_period": "0.0",
      "reservation_net_unused_recurring_fee": "0.0",
      "reservation_net_upfront_value": "0.0",
      "reservation_normalized_units_per_reservation": "",
      "reservation_number_of_reservations": "",
      "reservation_recurring_fee_for_usage": "0.0",
      "reservation_reservation_a_r_n": "",
      "reservation_start_time": "",
      "reservation_subscription_id": "9300435416",
      "reservation_total_reserved_normalized_units": "",
      "reservation_total_reserved_units": "",
      "reservation_units_per_reservation": "",
      "reservation_unused_amortized_upfront_fee_for_billing_period": "0.0",
      "reservation_unused_normalized_unit_quantity": "0.0",
      "reservation_unused_quantity": "0.0",
      "reservation_unused_recurring_fee": "0.0",
      "reservation_upfront_value": "0.0",
      "savings_plan_amortized_upfront_commitment_for_billing_period": "0.0",
      "savings_plan_end_time": "",
      "savings_plan_net_amortized_upfront_commitment_for_billing_period": "0.0",
      "savings_plan_net_recurring_commitment_for_billing_period": "0.0",
      "savings_plan_net_savings_plan_effective_cost": "0.0",
      "savings_plan_offering_type": "",
      "savings_plan_payment_option": "",
      "savings_plan_purchase_term": "",
      "savings_plan_recurring_commitment_for_billing_period": "0.0",
      "savings_plan_region": "",
      "savings_plan_savings_plan_a_r_n": "",
      "savings_plan_savings_plan_effective_cost": "0.0",
      "savings_plan_savings_plan_rate": "0.0",
      "savings_plan_start_time": "",
      "savings_plan_total_commitment_to_date": "0.0",
      "savings_plan_used_commitment": "0.0"
    }]
  }
}
```

### AWS的响应参数说明

| 参数名称  | 参数类型 | 描述 |
|---------|--------|---------|
| code    | int32  | 状态码   |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称   | 参数类型   | 描述                                |
|--------|--------|------------------------------------------|
| count  | uint64 | 当前规则能匹配到的总记录条数                  |
| detail | array  | 查询返回的数据                              |

#### data.detail[n]

| 参数名称        | 参数类型   | 描述                            |
|-------------|--------|--------------------------------------|
| bill_bill_type           | string | 计费类别                  |
| bill_billing_entity      | string | 账单实体                  |
| bill_billing_period_end_date | string | 账单周期截止日期        |
| bill_billing_period_start_date | string | 账单周期开始日期      |
| bill_invoice_id          | string | 账单清单ID                 |
| bill_invoicing_entity    | string | 账单清单实体                |
| bill_payer_account_id    | string | 账单支付账号ID              |
| discount_edp_discount    | string | edp优惠金额                |
| discount_total_discount  | string | 总优惠金额                 |
| identity_line_item_id    | string | 项目ID                    |
| identity_time_interval   | string | 标识时间间隔                |
| line_item_availability_zone | string | 可用区                  |
| line_item_blended_cost   | string | 混合成本                   |
| line_item_blended_rate   | string | 混合费率                   |
| line_item_currency_code  | int     | 项目当前代码               |
| line_item_legal_entity   | string  | 项目合法实体               |
| line_item_line_item_description | string | 计费描述             |
| line_item_line_item_type     | string | 项目类型                |
| line_item_net_unblended_cost  | string | 未混合折后成本          |
| line_item_net_unblended_rate  | string | 未混合折后费率          |
| line_item_normalization_factor  | string | 项目规范因子          |
| line_item_normalized_usage_amount  | string | 规范化使用费用     |
| line_item_operation  | string | 项目操作                        |
| line_item_product_code  | string | 项目产品代码                  |
| line_item_resource_id  | string | 资源ID                        |
| line_item_tax_type  | string | 项目税费类型                      |
| line_item_unblended_cost   | string | 未混合成本                 |
| line_item_unblended_rate   | string | 未混合费率                 |
| line_item_usage_account_id | string | 使用的账号ID               |
| line_item_usage_amount     | string | 使用金额                   |
| line_item_usage_end_date   | string | 使用截止日期                |
| line_item_usage_start_date | string | 使用开始日期                |
| line_item_usage_type | string | 使用类型                         |
| pricing_currency | string | 定价货币                             |
| pricing_lease_contract_length | string | 定价合同长度             |
| pricing_offering_class | string | 报价类别                       |
| pricing_public_on_demand_cost  | string | 定价公开需求成本        |
| pricing_public_on_demand_rate  | string | 定价公开需求费率        |
| pricing_purchase_option  | string | 付款方式：全量预付、部分预付、无预付 |
| pricing_term  | string | 使用量是预留还是按需                      |
| pricing_unit  | string | 费用计价单位                             |
| product_database_engine  | string | 产品数据库引擎                 |
| product_from_location  | string | 产品来源                        |
| product_from_location_type  | string | 产品来源类型               |
| product_from_region_code  | string | 产品区域编码                 |
| product_instance_type  | string | 产品实例类型                    |
| product_instance_type_family  | string | 产品实例类型系列          |
| product_location  | string | 产品定位                            |
| product_location_type  | string | 产品定位类型                    |
| product_marketoption  | string | 市场选项                        |
| product_normalization_size_factor  | string | 产品规格因子        |
| product_operation  | string | 产品操作                           |
| product_product_family  | string | 产品系列                      |
| product_product_name  | string | 产品名称                        |
| product_purchase_option  | string | 产品采购选项                  |
| product_purchaseterm  | string | 产品采购条款                     |
| product_region  | string | 产品区域                               |
| product_region_code  | string | 产品区域编码                       |
| product_servicecode  | string | 产品服务编码                       |
| product_servicename     | string | 产品服务名称                    |
| product_tenancy     | string | 产品库存                           |
| product_to_location     | string | 产品指向的位置                  |
| product_to_location_type     | string | 产品指向的位置类型          |
| product_to_region_code     | string | 产品指向的区域的编码          |
| product_transfer_type     | string | 产品传输类型                  |
| reservation_amortized_upfront_cost_for_usage | string | 预留摊销前期使用成本 |
| reservation_amortized_upfront_fee_for_billing_period | string | 预留摊销预付费账单周期 |
| reservation_effective_cost | string | 预留有效成本                 |
| reservation_end_time     | string | 预留截止时间                   |
| reservation_modification_status | string | 预留修改状态            |
| reservation_net_amortized_upfront_cost_for_usage | string | 预留网络摊销可用成本 |
| reservation_net_amortized_upfront_fee_for_billing_period | string | 预留网络摊销预付费账单周期 |
| reservation_net_effective_cost | string | 预留网络有效成本          |
| reservation_net_recurring_fee_for_usage | string | 预留可用的常用费用 |
| reservation_net_unused_amortized_upfront_fee_for_billing_period | string | 预留网络未使用预付费账单周期 |
| reservation_net_unused_recurring_fee | string | 预留网络未使用常用费用 |
| reservation_net_upfront_value | string | 预留前期净值              |
| reservation_normalized_units_per_reservation | string | 预留规范化单位每次保留量 |
| reservation_number_of_reservations | string | 预留数量             |
| reservation_recurring_fee_for_usage | string | 预留可用的常用费用    |
| reservation_reservation_a_r_n     | string | 预留的ARN             |
| reservation_start_time | string | 预留开始时间                      |
| reservation_subscription_id     | string | 预留的订阅ID             |
| reservation_total_reserved_normalized_units | string | 预留总服务标准化单位 |
| reservation_total_reserved_units  | string | 预留总服务单位          |
| reservation_units_per_reservation | string | 预留每次保留的单位       |
| reservation_unused_amortized_upfront_fee_for_billing_period | string | 预留未使用冻结的预付费计费周期 |
| reservation_unused_normalized_unit_quantity | string | 预留未使用规范化单位数量 |
| reservation_unused_quantity | string | 预留未使用的数量               |
| reservation_unused_recurring_fee | string | 预留未使用的现金          |
| reservation_upfront_value | string | 预留上行数值                    |
| savings_plan_amortized_upfront_commitment_for_billing_period | string | 账单期的计划摊销前期承诺 |
| savings_plan_end_time     | string | SavingsPlan截止时间             |
| savings_plan_net_amortized_upfront_commitment_for_billing_period | string | SavingsPlan承诺账单周期 |
| savings_plan_net_recurring_commitment_for_billing_period | string | SavingsPlan计划净现金周期 |
| savings_plan_net_savings_plan_effective_cost | string | SavingsPlan网络有效成本 |
| savings_plan_offering_type     | string | SavingsPlan报价类型        |
| savings_plan_payment_option    | string | SavingsPlan支付类型        |
| savings_plan_purchase_term     | string | SavingsPlan采购期限        |
| savings_plan_recurring_commitment_for_billing_period | string | SavingsPlan现金承诺账单周期 |
| savings_plan_region     | string | SavingsPlan区域                  |
| savings_plan_savings_plan_a_r_n  | string | SavingsPlanARN         |
| savings_plan_savings_plan_effective_cost | string | SavingsPlan有效成本 |
| savings_plan_savings_plan_rate     | string | SavingsPlan计划费率    |
| savings_plan_start_time     | string | SavingsPlan计划开始时间        |
| savings_plan_total_commitment_to_date | string | SavingsPlan总承诺日期 |
| savings_plan_used_commitment | string | SavingsPlan已使用承诺         |


### TCloud的响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 1155,
    "details": [
      {
          "BusinessCodeName": "公网 IP",
          "ProductCodeName": "公网 IP",
          "PayModeName": "按量计费",
          "ProjectName": "默认项目",
          "RegionName": "华南地区（广州）",
          "ZoneName": "其他",
          "ResourceId": "eip-noaqswaw",
          "ResourceName": "test_20230324",
          "ActionTypeName": "按量计费小时结",
          "OrderId": "eip-noaqswaw",
          "BillId": "202304018********1",
          "PayTime": "2023-04-01 01:12:22",
          "FeeBeginTime": "2023-04-01 00:00:00",
          "FeeEndTime": "2023-04-01 00:59:59",
          "ComponentSet": [
            {
              "ComponentCodeName": "公网IP资源费",
              "ItemCodeName": "公网IP资源费",
              "SinglePrice": "0.00005556",
              "SpecifiedPrice": "0.00005556",
              "PriceUnit": "元/个/秒",
              "UsedAmount": "1",
              "UsedAmountUnit": "个",
              "TimeSpan": "3600",
              "TimeUnitName": "秒",
              "Cost": "0.20001600",
              "Discount": "0.036207",
              "ReduceType": "折扣",
              "RealCost": "0.00724198",
              "VoucherPayAmount": "0",
              "CashPayAmount": "0",
              "IncentivePayAmount": "0.00724198",
              "ItemCode": "sv_eip_hour",
              "ComponentCode": "v_eip_hour",
              "ContractPrice": "0.00000201",
              "InstanceType": "0.00724198",
              "RiTimeSpan": "0.00000000",
              "OriginalCostWithRI": "0.00000000",
              "SPDeductionRate": "0.00000000",
              "SPDeduction": "0.00000000",
              "OriginalCostWithSP": "0.00000000",
              "BlendedDiscount": "0.03620700"
            }],
          "PayerUin": "10000000000001",
          "OwnerUin": "10000000000002",
          "OperateUin": "10000000000003",
          "BusinessCode": "p_eip",
          "ProductCode": "sp_eip",
          "ActionType": "postpay_deduct_h",
          "RegionId": "1",
          "ProjectId": 0
      }]
  }
}
```

### TCloud的响应参数说明

| 参数名称  | 参数类型 | 描述 |
|---------|--------|---------|
| code    | int32  | 状态码   |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称   | 参数类型   | 描述                                |
|--------|--------|------------------------------------------|
| count  | uint64 | 当前规则能匹配到的总记录条数                  |
| detail | array  | 查询返回的数据                              |

#### data.detail[n]

| 参数名称        | 参数类型   | 描述                            |
|-------------|--------|--------------------------------------|
| BusinessCodeName | string | 产品名称：云产品大类，如云服务器CVM、云数据库MySQL |
| ProductCodeName | string | 子产品名称：云产品子类，如云服务器CVM-标准型S1 |
| PayModeName | string | 计费模式：包年包月和按量计费 |
| ProjectName | string | 项目:资源所属项目 |
| RegionName | string | 区域：资源所属地域，如华南地区（广州） |
| ZoneName | string | 可用区：资源所属可用区，如广州三区 |
| ResourceId | string | 资源实例ID |
| ResourceName | string | 实例名称 |
| ActionTypeName | string | 交易类型 |
| OrderId | string | 订单ID |
| BillId | string | 交易ID |
| PayTime | timestamp | 扣费时间 |
| FeeBeginTime | timestamp | 开始使用时间 |
| FeeEndTime | timestamp | 结束使用时间 |
| ComponentSet | array of ComponentSet | 组件列表 |
| PayerUin | string | 支付者UIN |
| OwnerUin | string | 使用者UIN |
| OperateUin| string | 操作者UIN |
| Tags | array of BillTagInfo | Tag 信息 注意：此字段可能返回 null，表示取不到有效值。 |
| BusinessCode | string | 产品名称代码 注意：此字段可能返回 null，表示取不到有效值。 |
| ProductCode | string | 子产品名称代码 注意：此字段可能返回 null，表示取不到有效值。 |
| ActionType | string | 交易类型代码 注意：此字段可能返回 null，表示取不到有效值。 |
| RegionId | string | 区域ID 注意：此字段可能返回 null，表示取不到有效值。 |
| ProjectId | int64 | 项目ID:资源所属项目ID |
| PriceInfo | string array | 价格属性 注意：此字段可能返回 null，表示取不到有效值。 |

#### details[n].ComponentSet

| 参数名称   | 参数类型   | 描述                                |
|--------|--------|------------------------------------------|
| ComponentCodeName | string | 组件类型:资源组件类型的名称，如内存、硬盘等 |
| ItemCodeName | string | 组件名称:资源组件的名称，如云数据库MySQL-内存等 |
| SinglePrice | string | 组件刊例价:资源组件的原始价格，保持原始粒度 |
| SpecifiedPrice | string | 组件指定价 |
| PriceUnit | string | 价格单位 |
| UsedAmount | string | 组件用量 |
| UsedAmountUnit | string | 组件用量单位 |
| TimeSpan | string | 使用时长 |
| TimeUnitName | string | 时长单位 |
| Cost | string	 | 组件原价 |
| Discount | string | 折扣率 |
| ReduceType | string | 优惠类型 |
| RealCost | string | 优惠后总价 |
| VoucherPayAmount | string | 代金券支付金额 |
| CashPayAmount | string | 现金支付金额 |
| IncentivePayAmount | string | 赠送账户支付金额 |
| ItemCode | string | 组件类型代码 注意：此字段可能返回 null，表示取不到有效值。 |
| ComponentCode | string | 组件名称代码 注意：此字段可能返回 null，表示取不到有效值。 |
| ContractPrice | string | 合同价 注意：此字段可能返回 null，表示取不到有效值。 |
| InstanceType | string | 资源包、预留实例、节省计划、竞价实例这四类特殊实例本身的扣费行为，此字段体现对应的实例类型。枚举值如下： 注意：此字段可能返回 null，表示取不到有效值。 |
| RiTimeSpan | string | 预留实例抵扣的使用时长，时长单位与被抵扣的时长单位保持一致 注意：此字段可能返回 null，表示取不到有效值。 |
| OriginalCostWithRI | string | 按组件原价的口径换算的预留实例抵扣金额 注意：此字段可能返回 null，表示取不到有效值。 |
| SPDeductionRate | string | 节省计划可用余额额度范围内，节省计划对于此组件打的折扣率 注意：此字段可能返回 null，表示取不到有效值。 |
| SPDeduction | string | 节省计划抵扣的SP包面值 注意：此字段可能返回 null，表示取不到有效值。 |
| OriginalCostWithSP | string | 按组件原价的口径换算的节省计划抵扣金额 注意：此字段可能返回 null，表示取不到有效值。 |
| BlendedDiscount | string | 综合了官网折扣、预留实例抵扣、节省计划抵扣的混合折扣率。若没有预留实例抵扣、节省计划抵扣,混合折扣率等于折扣率 注意：此字段可能返回 null，表示取不到有效值 |


### HuaWei的响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 69,
    "details": [
      {
        "cycle": "2023-04",
        "bill_type": 1,
        "customer_id": "********************",
        "region": "ap-southeast-1",
        "region_name": "CN-Hong Kong",
        "cloud_service_type": "hws.service.type.ebs",
        "resource_Type_code": "hws.resource.type.volume",
        "cloud_service_type_name": "Elastic Volume Service",
        "resource_type_name": "Volume",
        "res_instance_id": "******-****-****-****-*******",
        "resource_name": "test-20230419",
        "sku_code": "SAS",
        "enterprise_project_id": "0",
        "enterprise_project_name": "default",
        "charge_mode": 1,
        "consume_amount": 4.6,
        "cash_amount": 0,
        "credit_amount": 0,
        "coupon_amount": 0,
        "flexipurchase_coupon_amount": 0,
        "stored_card_amount": 0,
        "bonus_amount": 0,
        "debt_amount": 4.6,
        "official_amount": 4.6,
        "discount_amount": 0,
        "measure_id": 1,
        "period_type": 20,
        "trade_id": "*************",
        "product_spec_desc": "High IO|50GB"
        }],
    "currency": "USD"
  }
}
```

### HuaWei的响应参数说明

| 参数名称  | 参数类型 | 描述 |
|---------|--------|---------|
| code    | int32  | 状态码   |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称   | 参数类型   | 描述                                |
|--------|--------|------------------------------------------|
| count  | uint64 | 当前规则能匹配到的总记录条数                  |
| detail | array  | 查询返回的数据                              |

#### data.detail[n]

| 参数名称        | 参数类型   | 描述                            |
|-------------|--------|--------------------------------------|
| cycle | string | 资源详单数据所在账期,东八区时间,格式为YYYY-MM。 例如2020-01 |
| bill_date | string | 消费日期,东八区时间,格式为YYYY-MM-DD。  说明: 当statistic_type=2时该字段才有值,否则返回null。 |
| bill_type | int64 | 账单类型。 1:消费-新购2:消费-续订3:消费-变更4:退款-退订5:消费-使用8:消费-自动续订9:调账-补偿14:消费-服务支持计划月末扣费15:消费-税金16:调账-扣费17:消费-保底差额 说明: 保底差额=客户签约保底合同后,如果没有达到保底消费,客户需要补交的费用,仅限于直销或者伙伴顾问销售类子客户,且为后付费用户。 20:退款-变更100:退款-退订税金101:调账-补偿税金102:调账-扣费税金 |
| customer_id | string | 消费的客户账号ID |
| region | string | 云服务区编码,例如:“ap-southeast-1”|
| region_name | string | 云服务区名称,例如:“中国-香港” |
| cloud_service_type | string | 云服务类型编码 |
| resource_Type_code | string | 资源类型编码 |
| cloud_service_type_name | string | 云服务类型名称 |
| resource_type_name | string | 资源类型名称 |
| res_instance_id | string | 资源实例ID |
| resource_name | string | 资源名称 |
| resource_tag | string | 资源标签 |
| sku_code | string | SKU编码,在账单中唯一标识一个资源的规格 |
| enterprise_project_id | string | 企业项目标识(企业项目ID) |
| enterprise_project_name | string | 企业项目名称 |
| charge_mode | int64 | 计费模式。 1 : 包年/包月3:按需10:预留实例 |
| consume_amount | float64 | 客户购买云服务类型的消费金额,包含代金券、现金券,精确到小数点后8位。说明: consume_amount的值等于cash_amount,credit_amount,coupon_amount,flexipurchase_coupon_amount,stored_card_amount,bonus_amount,debt_amount,adjustment_amount的总和 |
| cash_amount | float64 | 现金支付金额 |
| credit_amount | float64 | 信用额度支付金额 |
| coupon_amount | float64 | 代金券支付金额 |
| flexipurchase_coupon_amount | float64 | 现金券支付金额 |
| stored_card_amount | float64 | 储值卡支付金额 |
| bonus_amount | float64 | 奖励金支付金额 |
| debt_amount | float64 | 欠费金额 |
| adjustment_amount | float64 | 欠费核销金额 |
| official_amount | float64 | 官网价 |
| discount_amount | float64 | 对应官网价折扣金额 |
| measure_id | int64 | 金额单位。 1:元 |
| period_type | int64 | 周期类型: 19:年20:月24:天25:小时5:一次性 |
| root_resource_id | string | 根资源标识 |
| parent_resource_id | string | 父资源标识 |
| trade_id | string | 订单ID 或 交易ID。 账单类型为1,2,3,4,8时为订单ID;其它场景下为: 交易ID(非月末扣费:应收ID;月末扣费:账单ID) |
| product_spec_desc | string | 产品的规格描述 |
| sub_service_type_code | string | 该字段为预留字段 |
| sub_service_type_name | string | 该字段为预留字段 |
| sub_resource_type_code | string | 该字段为预留字段 |
| sub_resource_type_name | string | 该字段为预留字段 |
| sub_resource_id | string | 该字段为预留字段 |
| sub_resource_name | string | 该字段为预留字段 |
| pre_order_id | string | 该字段为预留字段 |
| az_code_infos | array of AzCodeInfo | 该字段为预留字段 |

#### details[n].AzCodeInfo

| 参数名称   | 参数类型   | 描述                                |
|--------|--------|------------------------------------------|
| az_code | string | 参数名称:可用区编码，参数的约束及描述:该参数非必填,且只允许字符串 |

### Azure的响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "id": "/subscriptions/7c99b444-456a-4eef-a083-40f6ab39ceaa/providers/Microsoft.Billing/billingPeriods/20230301/providers/Microsoft.Consumption/usageDetails/ca27d77d-5730-35bd-6ab5-f916ff1f7b01",
        "kind": "legacy",
        "name": "ca27d77d-5730-35bd-6ab5-f916ff1f7b01",
        "properties": {
        "accountName": "tencent_azure_IEG10",
        "accountOwnerId": "tencent_azure_ieg10@tencent.com",
        "billingAccountId": "81911009",
        "billingAccountName": "Aceville Pte. Ltd.",
        "billingCurrency": "USD",
        "billingPeriodEndDate": "2023-03-31T00:00:00Z",
        "billingPeriodStartDate": "2023-03-01T00:00:00Z",
        "billingProfileId": "81911009",
        "billingProfileName": "Aceville Pte. Ltd.",
        "chargeType": "Usage",
        "consumedService": "Microsoft.Compute",
        "cost": 0.144477228992944,
        "date": "2023-03-01T00:00:00Z",
        "effectivePrice": 4.47908075994991,
        "frequency": "UsageBased",
        "invoiceSection": "IEG",
        "isAzureCreditEligible": true,
        "meterId": "ed9e91d2-0f0c-4d55-b3dd-7f69d4708b22",
        "offerId": "MS-AZR-0017P",
        "partNumber": "AAD-18156",
        "payGPrice": 5.27,
        "pricingModel": "OnDemand",
        "product": "Premium SSD Managed Disks - P4 LRS - IN Central",
        "publisherType": "Azure",
        "quantity": 0.032256,
        "resourceGroup": "DOMMYTEST1_GROUP",
        "resourceId": "/subscriptions/7c99b444-456a-4eef-a083-40f6ab39ceaa/resourceGroups/DOMMYTEST1_GROUP/providers/Microsoft.Compute/disks/dommytest1_disk1_9c2b465e56b7452291d3261622f8c6ed",
        "resourceLocation": "centralindia",
        "resourceName": "dommytest1_disk1_9c2b465e56b7452291d3261622f8c6ed",
        "subscriptionId": "7c99b444-456a-4eef-a083-40f6ab39ceaa",
        "subscriptionName": "tencent_azure_IEG10",
        "unitPrice": 4.48
      },
      "type": "Microsoft.Consumption/usageDetails"
    }]
  }
}
```

### Azure的响应参数说明

| 参数名称  | 参数类型 | 描述 |
|---------|--------|---------|
| code    | int32  | 状态码   |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称   | 参数类型   | 描述                                |
|--------|--------|------------------------------------------|
| next_link | string | 获取下一页数据的链接 (url) ，为空表示没有更多数据，不为空则需要获取下一页 |
| detail | array  | 查询返回的数据                              |

#### data.detail[n]

| 参数名称        | 参数类型   | 描述                            |
|-------------|--------|--------------------------------------|
| id | string | 事件的完整限定 ARM ID      |
| kind | string | 指定使用详细信息的类型    |
| name | int64 | 唯一标识事件的 ID         |
| properties | Properties | 帐户名称      |
| type | string | 资源类型                |

#### Properties
| 参数名称        | 参数类型   | 描述                            |
|-------------|--------|--------------------------------------|
| accountName | string | 帐户名称 |
| accountOwnerId | string | 帐户所有者 ID |
| additionalInfo | string | 此使用情况项的其他详细信息。 默认情况下，除非在 $expand 中指定，否则不会填充它。 使用此字段可获取特定于使用情况行项的详细信息，例如实际 VM 大小 (ServiceType) 或应用预留折扣的比率。 |
| benefitId | string | 适用权益的唯一标识符 |
| benefitName | string | 适用权益的名称 |
| billingAccountId | string | 计费帐户标识符 |
| billingAccountName | string | 计费帐户名称 |
| billingCurrency | string | 计费货币 |
| billingPeriodEndDate | string | 计费周期结束日期 |
| billingPeriodStartDate | string | 计费周期开始日期 |
| billingProfileId | string | 计费配置文件标识符 |
| billingProfileName | string | 计费对象信息名称 |
| chargeType | string | 指示费用表示额度、使用情况、市场购买、预留费用或退款 |
| consumedService | string | 使用的服务名称。 发出使用情况或已购买的 Azure 资源提供程序的名称。 未为市场使用提供此值 |
| cost | int | 税前成本金额 |
| costCenter | string | 此部门的成本中心（如果它是一个部门，并且提供了一个成本中心） |
| date | string | 使用情况记录的日期，格式示例：2023-03-02T00:00:00Z |
| effectivePrice | int | 按使用量收费的有效价格 |
| frequency | string | 指示此费用的发生频率。 OneTime 用于仅发生一次的购买，每月针对每月重复的费用，使用基于服务使用量的费用 |
| invoiceSection | string | 发票科目名称 |
| isAzureCreditEligible | boolean | 是否符合 Azure 额度条件 |
| meterDetails | object of MeterDetails | 有关计量的详细信息。 默认情况下，除非在 $expand 中指定，否则不会填充它 |
| meterId | string | 计量 ID (GUID) 。 不适用于市场。 对于预留实例，它表示为其购买预留的主计量 |
| offerId | string | 产品/服务 ID。例如：MS-AZR-0017P、MS-AZR-0148P |
| partNumber | string | 所用服务的部件号。 可用于联接价目表。 不适用于市场 |
| payGPrice | int | 资源的零售价格 |
| planName | string | 计划名称 |
| pricingModel | string | 指示计量器定价方式的标识符(OnDemand\) |
| product | string | 使用的服务或购买的产品名称。 不适用于市场 |
| productOrderId | string | 产品订单 ID。对于预留，这是预留订单 ID |
| productOrderName | string | 产品订单名称。 对于预留，这是购买的 SKU |
| publisherName | string | 发布者名称 |
| publisherType | string | 发布服务器类型 |
| quantity | int | 使用数量 |
| reservationId | string | 预留的 ARM 资源 ID。 仅适用于与预留相关的记录 |
| reservationName | string | 用户提供的预留的显示名称。 特定日期的姓氏将填充在每日数据中。 仅适用于与预留相关的记录 |
| resourceGroup | string | 资源组名称 |
| resourceId | string | Azure 资源管理器使用情况详细信息资源的唯一标识符 |
| resourceLocation | string | 资源位置 |
| resourceName | string | 资源名称 |
| serviceInfo1 | string | 服务特定的元数据 |
| serviceInfo2 | string | 旧字段，具有可选的特定于服务的元数据 |
| subscriptionId | string | 订阅 guid |
| subscriptionName | string | 订阅名称 |
| term | string | 学期 (（以月) 为单位）。 每月定期购买 1 个月。 1 年预留 12 个月。 3 年预留 36 个月 |
| unitPrice | int | 单价是适用于你的价格。 (EA 或其他合同价格)  |


#### MeterDetails
| 参数名称        | 参数类型   | 描述                            |
|-------------|--------|--------------------------------------|
| meterCategory | string | 计量的类别，例如“云服务”、“网络”等 |
| meterName | string | 给定计量类别中的计量名称 |
| meterSubCategory | string | 计量的子类别，例如“A6 云服务”、“ExpressRoute (IXP) ”等 |
| serviceFamily | string | 服务系列 |
| unitOfMeasure | string | 计量消耗量计费的单位，例如“小时”、“GB”等 |


### Gcp的响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 55760,
    "details": [
    {
      "billing_account_id": "01754B-D65E19-D9EA79",
      "cost": 0,
      "cost_type": "regular",
      "country": "US",
      "credits": "[]",
      "currency": "USD",
      "location": "us",
      "month": "202304",
      "project_id": "tencentgcpieg6",
      "project_name": "TencentGcpIEG6",
      "project_number": "904277337334",
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
    }]
  }
}
```

### Gcp的响应参数说明

| 参数名称  | 参数类型 | 描述 |
|---------|--------|---------|
| code    | int32  | 状态码   |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称   | 参数类型   | 描述                                |
|--------|--------|------------------------------------------|
| count  | uint64 | 当前规则能匹配到的总记录条数                  |
| detail | array  | 查询返回的数据                              |

#### data.detail[n]

| 参数名称        | 参数类型   | 描述                            |
|-------------|--------|--------------------------------------|
| billing_account_id | string | 与使用量相关的 Cloud Billing 帐号ID |
| cost | float64 | 成本 |
| cost_type | string | 费用类型 |
| country | string | 国家 |
| credits | json | 赠送金信息 |
| currency | string | 币种 |
| location | string | 区域信息 |
| month | string | 账单年月 |
| project_id | string | 项目ID |
| project_name | string | 项目名称 |
| project_number | string | 项目编号 |
| region | string | 区域 |
| resource_global_name | string | 资源全局唯一标识符 |
| resource_name | string | 资源名称 |
| service_description | string | 服务描述 |
| service_id | string | 服务ID |
| sku_description | string | 资源类型描述 |
| sku_id | string | 资源类型ID |
| usage_amount | string | 可用金额 |
| usage_amount_in_pricing_units | string | 可用金额单价 |
| usage_end_time | string | 可用结束时间，示例：2023-04-16T15:00:00Z |
| usage_pricing_unit | float64 | 可用金额单价的单位 |
| usage_start_time | string | 可用开始时间，示例：2023-04-16T15:00:00Z |
| usage_unit | string | 可用金额单位 |
| zone | string | 可用区 |