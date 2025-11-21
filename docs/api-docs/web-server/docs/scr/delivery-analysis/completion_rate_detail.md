### 描述

- 该接口提供版本：v9.9.9。
- 该接口所需权限：自研云管理-交付分析。
- 该接口功能描述：详情页申请单据结单率详情统计，按月统计信息。

### URL

POST /api/v1/woa/task/apply/completion-rate/detail

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
        "completion_rate": 100.00
      },
      {
        "bk_biz_id": 105,
        "year_month": "2025-10",
        "total_orders": 3,
        "done_orders": 3,
        "completion_rate": 100.00
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

| 参数名称 | 参数类型         | 描述                                           |
|------|--------------|----------------------------------------------|
| details | object array | 按业务和月份统计的结单率详情列表，按结单率降序排列，结单率相同则按业务ID和月份升序排列 |

#### data.details[n]

| 参数名称                    | 参数类型   | 描述                                                   |
|-------------------------|--------|------------------------------------------------------|
| bk_biz_id               | int    | 业务ID                                                 |
| year_month              | string | 年月，格式：YYYY-MM，如 "2025-10"                            |
| total_orders            | int    | 总单据数                                                 |
| done_orders             | int    | 已完成单据数（stage=DONE）                                   |
| completion_rate         | number | 结单率（百分比），保留2位小数，范围：0.00-100.00。计算公式：已完成单据数/总单据数×100% |