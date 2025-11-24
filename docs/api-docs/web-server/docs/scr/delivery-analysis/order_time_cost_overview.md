### 描述

- 该接口提供版本：v9.9.9。
- 该接口所需权限：自研云管理-交付分析。
- 该接口功能描述：查询单据从创建到完成的耗时统计（剔除审批阶段耗时的口径由后端实现约束），按月聚合 主机申请子单表 数据。

### URL

POST /api/v1/woa/task/apply/analysis/order_time_cost/overview

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
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "details": [
      {
        "year_month": "2025-01",
        "avg_duration_hours": 14.33
      },
      {
        "year_month": "2025-02",
        "avg_duration_hours": 14.58
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
