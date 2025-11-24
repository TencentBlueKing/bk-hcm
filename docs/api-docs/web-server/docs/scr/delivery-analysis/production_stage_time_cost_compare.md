### 描述

- 该接口提供版本：v9.9.9。
- 该接口所需权限：自研云管理-交付分析。
- 该接口功能描述：查询生产阶段耗时的对比统计，按月、业务聚合生成记录数据

### URL

POST /api/v1/woa/task/apply/analysis/production_stage_time_cost/compare

### 请求参数

| 参数名称         | 参数类型 | 必填 | 描述                         |
|--------------|---------|------|----------------------------|
| current_date | string  | 是   | 开始年月，YYYY-MM格式，例如：2025-11（表示2025年11月） |
| compare_date | string  | 是   | 结束年月，YYYY-MM格式，例如：2025-12（表示2025年12月） |

#### 请求示例

```json
{
  "current_date": "2025-11",
  "compare_date": "2025-12"
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "current": [
      {
        "bk_biz_id": 1,
        "year_month": "2025-01",
        "done_orders": 1519,
        "avg_duration_hours": 0.54
      },
      {
        "bk_biz_id": 1,
        "year_month": "2025-02",
        "done_orders": 1061,
        "avg_duration_hours": 0.54
      }
    ],
    "compare": [
      {
        "bk_biz_id": 1,
        "year_month": "2025-01",
        "done_orders": 1519,
        "avg_duration_hours": 0.54
      },
      {
        "bk_biz_id": 1,
        "year_month": "2025-02",
        "done_orders": 1061,
        "avg_duration_hours": 0.54
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data	   | object | 响应数据 |

#### data

| 参数名称    | 参数类型         | 描述              |
|---------|--------------|-----------------|
| current | object array | 指定的当前月份数据，按业务分组 |
| compare | object array | 指定的对比月份数据，按业务分组 |

#### current[n]/compare[n]

| 参数名称        | 参数类型 | 描述                      |
|----------------|---------|-------------------------|
| bk_biz_id      | int64   | 业务ID                    |
| year_month     | string  | 年月，格式：YYYY-MM           |
| done_orders    | int64   | 完成记录数                  |
| avg_duration_hours | float64 | 平均耗时（小时），保留2位小数 |

