### 描述

- 该接口提供版本：v9.9.9。
- 该接口所需权限：自研云管理-交付分析。
- 该接口功能描述：查询生产阶段耗时统计，按月聚合生成记录数据

### URL

POST /api/v1/woa/task/apply/analysis/production_stage_time_cost/overview

### 请求参数

| 参数名称 | 参数类型 | 必填 | 描述                         |
|---------|---------|------|----------------------------|
| start_time | string  | 是   | 时间范围开始日期，格式如"2025-01-01" |
| end_time | string  | 是   | 时间范围结束日期，格式如"2025-01-31" |

#### 请求示例

```json
{
  "start_time": "2025-01-01",
  "end_time": "2025-01-31"
}
```

### 响应示例

#### 获取详细信息返回结果示例
```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "year_month": "2025-01",
        "avg_duration_hours": 0.54
      },
      {
        "year_month": "2025-02",
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
| details | object array | 按月统计的数组，按年月升序排列 |

#### details[n]

| 参数名称        | 参数类型 | 描述                      |
|----------------|---------|-------------------------|
| year_month     | string  | 年月，格式：YYYY-MM           |
| avg_duration_hours | float64 | 平均耗时（小时），保留2位小数 |

