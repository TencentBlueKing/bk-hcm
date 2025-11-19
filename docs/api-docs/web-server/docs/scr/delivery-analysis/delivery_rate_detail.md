### 描述

- 该接口提供版本：v9.9.9。
- 该接口所需权限：自研云管理-交付分析。
- 该接口功能描述：详情页申请单据需求交付率详情统计，按月统计信息。

### URL

POST /api/v1/woa/task/apply/delivery-rate/detail

### 输入参数

| 参数名称       | 参数类型   | 必选 | 描述                                                            |
|------------|--------|----|---------------------------------------------------------------|
| start_time | string | 是  | 开始时间，格式：YYYY-MM-DD，如 "2025-10-01" |
| end_time   | string | 是  | 结束时间，格式：YYYY-MM-DD，如 "2025-10-31" |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "start_time": "2025-10-01",
  "end_time": "2025-10-31"
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
    "details": [
      {
        "bk_biz_id": 104,
        "year_month": "2025-10",
        "total_orders": 5,
        "done_orders": 5,
        "total_num_sum": 7,
        "success_num_sum": 7,
        "host_delivery_rate": 100.00
      },
      {
        "bk_biz_id": 105,
        "year_month": "2025-10",
        "total_orders": 3,
        "done_orders": 3,
        "total_num_sum": 9,
        "success_num_sum": 9,
        "host_delivery_rate": 100.00
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述                         |
|---------|--------------|----------------------------|
| result  | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code    | int          | 错误编码。 0表示success，>0表示失败错误  |
| message | string       | 请求失败返回的错误信息                |
| data    | object array | 响应数据                       |

#### data

| 参数名称 | 参数类型         | 描述                                             |
|------|--------------|------------------------------------------------|
| details | object array | 按业务和月份统计的交付率详情列表，按主机交付率降序排列，交付率相同则按业务ID和月份升序排列 |

#### data.details[n]

| 参数名称                       | 参数类型   | 描述                                                            |
|----------------------------|--------|---------------------------------------------------------------|
| bk_biz_id                  | int    | 业务ID                                                          |
| year_month                 | string | 年月，格式：YYYY-MM，如 "2025-10"                                     |
| total_orders               | int    | 已完成单据总数（只统计stage=DONE且status=DONE的单据）                         |
| done_orders                | int    | 已完成单据数（stage=DONE）                                            |
| total_num_sum              | int    | 需求总数                                                          |
| success_num_sum            | int    | 成功交付数                                                         |
| host_delivery_rate         | number | 主机交付率（百分比），保留2位小数，范围：0.00-100.00。计算公式：成功交付数/需求总数×100%         |