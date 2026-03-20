### 描述

- 该接口提供版本：v1.8.10.3+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询业务下GPU需求提报子单列表。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/gpu/demands/suborders/list

### 输入参数

| 参数名称   | 参数类型  | 必选 | 描述             |
|-----------|---------|------|-----------------|
| bk_biz_id | int64   | 是   | 业务ID（路径参数） |
| filter    | object  | 是   | 查询过滤条件      |
| page      | object  | 是   | 分页设置          |

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

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

#### 查询参数介绍：

| 参数名称            | 参数类型   | 描述                                                                  |
|-----------------|--------|----------------------------------------------------------------------------|
| id              | string | 申请ID                                                                      |
| order_id        | string | 需求主单ID                                                                   |
| demand_type     | string | 需求分类                                                                     |
| demand_year     | int    | 需求年份                                                                     |
| demand_month    | int    | 需求月份                                                                     |
| gpu_num         | int    | GPU预算卡数                                                                  |
| qpm_max         | int    | 峰值QPM                                                                     |
| status          | string | 状态(INIT:待评审 PENDING:评审中 DONE:已评审 REJECT:已驳回 TERMINATE:已终止)      |
| creator         | string | 创建者                                                                      |
| reviser         | string | 更新者                                                                      |
| created_at      | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                      |
| updated_at      | string | 更新时间，标准格式：2006-01-02T15:04:05Z                                      |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 请求参数示例

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "status",
        "op": "in",
        "value": ["INIT", "PENDING"]
      },
      {
        "field": "creator",
        "op": "eq",
        "value": "admin"
      }
    ]
  },
  "page": {
    "count": false,
    "start": 0,
    "limit": 20
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "count": 0,
    "details": [
      {
        "id": "00000001",
        "order_id": "00000002",
        "bk_biz_id": 111,
        "op_product_id": 222,
        "op_product_name": "运营产品",
        "demand_type": "xxxx",
        "demand_year": 2026,
        "demand_month": 3,
        "gpu_num": 10,
        "qpm_max": 100,
        "status": "INIT",
        "comment": ["xxxx"],
        "extension": [],
        "remark": "xxxx",
        "creator": "xxxxxxx",
        "reviser": "xxxxxxx",
        "created_at": "2023-06-09T11:00:08Z",
        "updated_at": "2023-06-09T11:01:10Z"
      }
    ],
    "tpl_config": [
        {
            "id": "00000001",
            "sheets": [
                {
                  "name": "大语言模型训练-文生文",
                  "head_row": 2,
                  "row_start": 5,
                  "fixed_headers": [
                  {
                      "name": "年份",
                      "field": "A",
                      "db_field":"demand_year",
                      "required": true,
                      "type": "enum",
                      "value": [
                        2026,
                        2027,
                        2028
                      ]
                  },
                  {
                      "name": "月份",
                      "field": "B",
                      "db_field":"demand_month",
                      "required": true,
                      "type": "enum",
                      "value": [1,2,3,4,5,6,7,8,9,10,11,12]
                  },
                  {
                      "name": "预算卡数(H20)",
                      "field": "C",
                      "db_field":"gpu_num",
                      "required": true,
                      "formula": "ROUNDUP(M4*1000000000/Q4/3600/P4,0)",
                      "readonly": true,
                      "type": "int"
                  },
                  {
                    "name": "峰值QPM",
                    "field": "-",
                    "db_field":"qpm_max",
                    "formula": "MAX(H4:K4)",
                    "hidden": true
                  }
                ],
                "headers": [
                  {
                    "name": "使用场景",
                    "field": "D",
                    "required": true,
                    "type": "string"
                  },
                  {
                    "name": "地域",
                    "field": "E",
                    "required": true,
                    "type": "enum",
                    "value": [
                      "国内",
                      "海外"
                    ]
                  },
                  {
                    "name": "训练类型",
                    "field": "F",
                    "type": "enum",
                    "value": [
                      "预训练",
                      "全参精调",
                      "混元精调",
                      "LoRA",
                      "PPO",
                      "DPO",
                      "RLHF_RM",
                      "GRPO"
                    ]
                  },
                  {
                    "name": "当前额度",
                    "field": "G",
                    "type": "int"
                  },
                  {
                    "name": "额度使用率(周均)",
                    "field": "H",
                    "type": "int"
                  },
                  {
                    "name": "当前利用率(周均)",
                    "field": "I",
                    "type": "int"
                  },
                  {
                    "name": "模型名称(模型结构)",
                    "field": "J",
                    "type": "string"
                  },
                  {
                    "name": "训练精度",
                    "field": "K",
                    "type": "enum",
                    "value": [
                      "FP32",
                      "BF16",
                      "FP16",
                      "INT8",
                      "混合精度",
                      "INT4"
                    ]
                  },
                  {
                    "name": "参数量(B)",
                    "field": "L",
                    "type": "float(1)"
                  },
                  {
                    "name": "训练token量(B)",
                    "field": "M",
                    "type": "int"
                  },
                  {
                    "name": "序列长度",
                    "field": "N",
                    "type": "int"
                  },
                  {
                    "name": "训练框架",
                    "field": "O",
                    "type": "enum",
                    "value": [
                      "PyTorch",
                      "Tensorflow",
                      "AnglePTM",
                      "LlamaFactory",
                      "Verl",
                      "Openr1",
                      "Gcore",
                      "MegatronLM"
                    ]
                  },
                  {
                    "name": "模型训练时长(小时)",
                    "field": "P",
                    "type": "float(1)"
                  },
                  {
                    "name": "训练速度(tokens/s/卡)",
                    "field": "Q",
                    "type": "float(1)"
                  },
                  {
                    "name": "单模型卡数",
                    "field": "R",
                    "type": "int"
                  },
                  {
                    "name": "资源推导说明",
                    "field": "S",
                    "type": "string"
                  }
                ]
              }
            ]
        }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                        |
|---------|--------|---------------------------|
| code    | int    | 错误编码。 0表示success，>0表示失败错误 |
| message | string | 请求失败返回的错误信息               |
| data	   | object | 响应数据                      |

#### data

| 参数名称    | 参数类型       | 描述                                                       |
|------------|--------------|------------------------------------------------------------|
| count      | int          | 当前规则能匹配到的总记录条数，仅在 count 查询参数设置为 true 时返回 |
| details    | object array | 查询返回的数据，仅在 count 查询参数设置为 false 时返回            |
| tpl_config | object array | Excel模版配置                                               |

#### data.details[n]

| 参数名称         | 参数类型      | 描述                                                                              |
|-----------------|-------------|-----------------------------------------------------------------------------------|
| id              | string      | 需求子单ID                                                                          |
| order_id        | string      | 需求主单ID                                                                          |
| bk_biz_id       | int         | 业务ID                                                                             |
| op_product_id   | int         | 运营产品ID                                                                          |
| op_product_name | string      | 运营产品名称                                                                         |
| demand_type     | string      | 需求分类                                                                            |
| demand_year     | int         | 需求年份                                                                            |
| demand_month    | int         | 需求月份                                                                            |
| gpu_num         | int         | GPU预算卡数                                                                          |
| qpm_max         | int         | 峰值QPM                                                                             |
| status          | string      | 状态(INIT:待评审 PENDING:评审中 DONE:已评审 REJECT:已驳回 TERMINATE:已终止                |
| comment         | json array  | 评审意见                                                                              |
| extension       | json array  | 需求明细                                                                              |
| remark          | string      | 备注                                                                                 |
| creator         | string      | 创建者                                                                               |
| reviser         | string      | 更新者                                                                               |
| created_at      | string      | 创建时间，标准格式：2006-01-02T15:04:05Z                                               |
| updated_at      | string      | 更新时间，标准格式：2006-01-02T15:04:05Z                                               |

#### data.tpl_config[n]

| 参数名称         | 参数类型       | 描述                                                                              |
|-----------------|--------------|-----------------------------------------------------------------------------------|
| id              | string       | 模版ID                                                                             |
| sheets          | object array | Excel模版的Sheet配置                                                                |

#### data.tpl_config[n].sheets[n]

| 参数名称         | 参数类型             | 描述                                                                        |
|-----------------|--------------------|-----------------------------------------------------------------------------|
| name            | string             | Sheet名称                                                                    |
| head_row        | int                | 列头所在行                                                                    |
| row_start       | int                | 业务数据起始行                                                                 |
| fixed_headers   | HeaderObject array | 业务数据固定列                                                                 |
| headers         | HeaderObject array | 业务数据动态列                                                                 |

#### HeaderObject

| 参数名称         | 参数类型      | 描述                                                                              |
|-----------------|-------------|-----------------------------------------------------------------------------------|
| name            | string      | 列名                                                                               |
| field           | string      | 列所在的列号                                                                        |
| db_field        | string      | 列名对应的DB字段名称，默认：空                                                         |
| type            | string      | 字段类型(int、string、enum、bool等)                                                  |
| formula         | string      | 列公式，默认：空                                                                     |
| required        | bool        | 数据是否必填项，默认：false                                                           |
| readonly        | bool        | 数据是否只读，默认：false                                                             |
| hidden          | bool        | 该列是否隐藏，默认：false                                                             |
