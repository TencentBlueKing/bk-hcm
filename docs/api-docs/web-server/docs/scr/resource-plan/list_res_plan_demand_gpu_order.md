### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-GPU需求。
- 该接口功能描述：查询资源下GPU需求提报主单列表。

### URL

POST /api/v1/woa/plans/resources/gpu/demands/orders/list

### 输入参数

| 参数名称   | 参数类型   | 必选 | 描述     |
|--------|--------|----|--------|
| filter | object | 是  | 查询过滤条件 |
| page   | object | 是  | 分页设置   |

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

##### 2. 查询参数介绍

| 参数名称            | 参数类型   | 描述                                                                                       |
|-----------------|--------|------------------------------------------------------------------------------------------|
| id              | string | 主单ID                                                                                     |
| bk_biz_id       | int64  | 业务ID                                                                                     |
| op_product_id   | int64  | 运营产品ID                                                                                   |
| op_product_name | string | 运营产品名称                                                                                   |
| template_id     | string | 模版ID                                                                                     |
| status          | string | 主单状态，枚举值：INIT（待评审）、PENDING（评审中）、DONE（已评审）、REJECT（部分已驳回）、REJECT_ALL（全部已驳回）、TERMINATE（已终止） |
| creator         | string | 创建者                                                                                      |
| reviser         | string | 最后修改者                                                                                    |
| created_at      | string | 创建时间，格式为 RFC3339，例如 2006-01-02T15:04:05Z                                                 |
| updated_at      | string | 更新时间，格式为 RFC3339，例如 2006-01-02T15:04:05Z                                                 |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                              |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                              |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                               |
| sort  | string | 否  | 排序字段，默认按 created_at 倒序排序，枚举值：created_at（创建时间）、updated_at（更新时间）                                                                |
| order | string | 否  | 排序顺序，枚举值：ASC（升序）、DESC（降序）                                                                                                       |

### 调用示例

#### 请求参数示例

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "bk_biz_id",
        "op": "eq",
        "value": 100
      },
      {
        "field": "status",
        "op": "in",
        "value": ["INIT", "PENDING"]
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
        "id": "order-001",
        "bk_biz_id": 100,
        "op_product_id": 123,
        "op_product_name": "运营产品A",
        "template_id": "tpl-001",
        "status": "INIT",
        "remark": "备注信息",
        "total_gpu_num": 16,
        "total_qpm_max": 10000,
        "creator": "admin",
        "reviser": "admin",
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
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
| data    | object | 响应数据                      |

#### data

| 参数名称    | 参数类型         | 描述                                        |
|---------|--------------|-------------------------------------------|
| count   | int          | 当前规则能匹配到的总记录条数，仅在 count 查询参数设置为 true 时返回  |
| details | object array | 查询返回的数据列表，仅在 count 查询参数设置为 false 时返回      |

#### data.details[n]

| 参数名称            | 参数类型   | 描述                                                                                       |
|-----------------|--------|------------------------------------------------------------------------------------------|
| id              | string | 主单ID                                                                                     |
| bk_biz_id       | int64  | 业务ID                                                                                     |
| op_product_id   | int64  | 运营产品ID                                                                                   |
| op_product_name | string | 运营产品名称                                                                                   |
| template_id     | string | 模版ID                                                                                     |
| status          | string | 主单状态，枚举值：INIT（待评审）、PENDING（评审中）、DONE（已评审）、REJECT（部分已驳回）、REJECT_ALL（全部已驳回）、TERMINATE（已终止） |
| remark          | string | 备注                                                                                       |
| total_gpu_num   | int64  | 需求卡数，由关联子单的 gpu_num 字段汇总求和得出                                                            |
| total_qpm_max   | int64  | QPM（月调用峰值），由关联子单的 qpm_max 字段汇总求和得出                                                       |
| creator         | string | 创建者                                                                                      |
| reviser         | string | 最后修改者                                                                                    |
| created_at      | string | 创建时间，格式为 RFC3339，例如 2006-01-02T15:04:05Z                                                 |
| updated_at      | string | 更新时间，格式为 RFC3339，例如 2006-01-02T15:04:05Z                                                 |
