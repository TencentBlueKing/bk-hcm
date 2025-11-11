### 描述

- 该接口提供版本：v1.8.7+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询 GCP Cloud Monitoring 时间序列数据。

### URL

POST /api/v1/cloud/vendors/gcp/monitoring/time_series/list

### 输入参数

| 参数名称            | 参数类型   | 必选 | 描述                                            |
|-----------------|--------|----|-----------------------------------------------|
| root_account_id | string | 是  | 根账号 ID                                        |
| main_account_id | string | 是  | 二级账号 ID                                       |
| filter          | string | 是  | 时间序列过滤器，使用 Monitoring Query Language (MQL) 格式 |
| interval        | object | 是  | 查询时间范围                                        |
| aggregation     | object | 否  | 数据聚合配置                                        |
| view            | string | 是  | 输出视图控制，可选值：FULL, HEADERS                      |
| page_size       | uint32 | 否  | 每页限制条数，最大 100000                              |
| page_token      | string | 否  | 分页令牌，用于获取下一页数据                                |

#### interval

| 参数名称       | 参数类型   | 必选 | 描述                                        |
|------------|--------|----|-------------------------------------------|
| start_time | string | 是  | 开始时间，RFC3339 格式，例如：`2024-01-01T00:00:00Z` |
| end_time   | string | 是  | 结束时间，RFC3339 格式，例如：`2024-01-01T01:00:00Z` |

#### aggregation

| 参数名称                 | 参数类型   | 必选 | 描述                                 |
|----------------------|--------|----|------------------------------------|
| alignment_period     | string | 否  | 对齐周期，格式为持续时间字符串，例如：`60s`           |
| per_series_aligner   | string | 否  | 每个序列的对齐方式，例如：ALIGN_MEAN, ALIGN_SUM |
| cross_series_reducer | string | 否  | 跨序列减少器，例如：REDUCE_MEAN, REDUCE_SUM  |
| group_by_fields      | array  | 否  | 分组字段列表                             |

### 调用示例

#### 获取详细信息请求参数示例

查询指定 GCP 账号的 CPU 使用率监控数据。

```json
{
  "root_account_id": "00000001",
  "main_account_id": "00000002",
  "filter": "metric.type=\"compute.googleapis.com/instance/cpu/utilization\"",
  "interval": {
    "start_time": "2024-01-01T00:00:00Z",
    "end_time": "2024-01-01T01:00:00Z"
  },
  "aggregation": {
    "alignment_period": "60s",
    "per_series_aligner": "ALIGN_MEAN"
  },
  "page_size": 100
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "time_series": [
      {
        "metric": {
          "type": "compute.googleapis.com/instance/cpu/utilization",
          "labels": {
            "instance_name": "instance-1"
          }
        },
        "resource": {
          "type": "gce_instance",
          "labels": {
            "project_id": "my-project",
            "instance_id": "1234567890123456789",
            "zone": "us-central1-a"
          }
        },
        "metric_kind": "GAUGE",
        "value_type": "DOUBLE",
        "points": [
          {
            "interval": {
              "start_time": "2024-01-01T00:00:00Z",
              "end_time": "2024-01-01T00:01:00Z"
            },
            "value": {
              "double_value": 0.45
            }
          },
          {
            "interval": {
              "start_time": "2024-01-01T00:01:00Z",
              "end_time": "2024-01-01T00:02:00Z"
            },
            "value": {
              "double_value": 0.52
            }
          }
        ],
        "unit": "1"
      }
    ],
    "next_page_token": "CAAaGAoNCAESCQiAkvK/qK7rARIHCgVfX2tleQ==",
    "execution_errors": []
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

| 参数名称             | 参数类型   | 描述               |
|------------------|--------|------------------|
| time_series      | array  | 时间序列数据列表         |
| next_page_token  | string | 下一页令牌，为空表示没有更多数据 |
| execution_errors | array  | 执行过程中的错误列表       |
| unit             | string | 指标的单位            |

#### data.time_series[n]

| 参数名称        | 参数类型   | 描述                                            |
|-------------|--------|-----------------------------------------------|
| metric      | object | 指标信息                                          |
| resource    | object | 资源信息                                          |
| metric_kind | string | 指标类型，例如：GAUGE（瞬时值）、DELTA（变化量）、CUMULATIVE（累计值） |
| value_type  | string | 值类型，例如：DOUBLE、INT64、BOOL、STRING、DISTRIBUTION  |
| points      | array  | 数据点列表                                         |
| unit        | string | 单位，例如："1"（无单位）、"By"（字节）、"s"（秒）                |

#### metric

| 参数名称   | 参数类型   | 描述                                                        |
|--------|--------|-----------------------------------------------------------|
| type   | string | 指标类型，例如：`compute.googleapis.com/instance/cpu/utilization` |
| labels | object | 指标标签，键值对形式                                                |

#### resource

| 参数名称   | 参数类型   | 描述                              |
|--------|--------|---------------------------------|
| type   | string | 资源类型，例如：gce_instance、gcs_bucket |
| labels | object | 资源标签，键值对形式                      |

#### points[n]

| 参数名称     | 参数类型   | 描述                                                      |
|----------|--------|---------------------------------------------------------|
| interval | object | 时间区间                                                    |
| value    | object | 数据值，根据 value_type 的不同，可能包含 double_value、int64_value 等字段 |

#### interval

| 参数名称       | 参数类型   | 描述              |
|------------|--------|-----------------|
| start_time | string | 开始时间，RFC3339 格式 |
| end_time   | string | 结束时间，RFC3339 格式 |

#### value

| 参数名称               | 参数类型    | 描述                                 |
|--------------------|---------|------------------------------------|
| double_value       | float64 | 双精度浮点值（当 value_type 为 DOUBLE 时）    |
| int64_value        | int64   | 64 位整数值（当 value_type 为 INT64 时）    |
| bool_value         | bool    | 布尔值（当 value_type 为 BOOL 时）         |
| string_value       | string  | 字符串值（当 value_type 为 STRING 时）      |
| distribution_value | object  | 分布值（当 value_type 为 DISTRIBUTION 时） |

#### distribution_value

| 参数名称                     | 参数类型    | 描述             |
|--------------------------|---------|----------------|
| count                    | int64   | 数据点数量          |
| mean                     | float64 | 平均值            |
| sum_of_squared_deviation | float64 | 平方差之和          |
| range                    | object  | 值范围 [min, max] |
| bucket_counts            | array   | 每个直方图桶的计数      |

#### range

| 参数名称 | 参数类型    | 描述  |
|------|---------|-----|
| min  | float64 | 最小值 |
| max  | float64 | 最大值 |

#### data.execution_errors[n]

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 错误码  |
| message | string | 错误信息 |
