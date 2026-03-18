### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询 AWS 成员账号的 CloudWatch 指标时序数据。通过 STS AssumeRole 跨账号访问成员账号的 CloudWatch GetMetricData 接口，返回指定指标的时序数据点。支持单次查询多个指标。

### URL

POST /api/v1/cloud/vendors/aws/assume_role/cloudwatch/metric_data/list

### 请求参数

| 参数名称                | 参数类型     | 必选 | 描述                                                                                                     |
|---------------------|----------|----|--------------------------------------------------------------------------------------------------------|
| root_account_id     | string   | 是  | 根账号 ID，用于获取 AWS 根账号凭证                                                                                  |
| main_account_id     | string   | 是  | 主账号 ID，用于查询 main_account 表获取成员账号的 AWS Account ID                                                       |
| role_chain          | string[] | 是  | 角色名数组，支持 Role Chaining。中间角色在管理账号中 AssumeRole，最后一个角色在成员账号中 AssumeRole。至少包含 1 个角色名                    |
| region              | string   | 是  | AWS 区域，如 us-east-1                                                                                     |
| external_id         | string   | 否  | STS AssumeRole 的 ExternalId，用于目标角色 Trust Policy 的条件验证。仅应用于 Role Chain 最后一步                             |
| metric_data_queries | array    | 是  | 指标查询数组，每个元素定义一个指标查询，最多 500 个                                                                          |
| start_time          | string   | 是  | 查询起始时间，ISO 8601 格式，如 2024-01-01T00:00:00Z                                                              |
| end_time            | string   | 是  | 查询截止时间，ISO 8601 格式，如 2024-01-02T00:00:00Z                                                              |

#### metric_data_queries[n]

| 参数名称       | 参数类型   | 必选 | 描述                                                                                 |
|------------|--------|----|------------------------------------------------------------------------------------|
| id         | string | 是  | 查询 ID，用于在响应中关联结果，需唯一，仅支持字母数字和下划线                                                   |
| namespace  | string | 是  | CloudWatch 命名空间，如 AWS/EC2（内置指标）、CWAgent（Agent 采集指标）                                 |
| metric_name | string | 是  | 指标名称，如 CPUUtilization、nvidia_smi_utilization_gpu                                   |
| dimensions | array  | 否  | 维度过滤条件，每个元素包含 name 和 value                                                         |
| stat       | string | 是  | 统计方式，如 Average、Sum、Maximum、Minimum、SampleCount                                      |
| period     | int    | 是  | 数据粒度（秒），如 300（5 分钟）、3600（1 小时）                                                     |

#### dimensions[n]

| 参数名称  | 参数类型   | 必选 | 描述                             |
|-------|--------|----|--------------------------------|
| name  | string | 是  | 维度名称，如 InstanceId              |
| value | string | 是  | 维度值，如 i-0abcdef1234567890      |

### 调用示例

#### 请求参数示例（查询单个 CPU 利用率）

```json
{
  "root_account_id": "00000001",
  "main_account_id": "00000002",
  "role_chain": ["gpu-readonly"],
  "region": "us-east-1",
  "metric_data_queries": [
    {
      "id": "cpu_util",
      "namespace": "AWS/EC2",
      "metric_name": "CPUUtilization",
      "dimensions": [
        {"name": "InstanceId", "value": "i-0abcdef1234567890"}
      ],
      "stat": "Average",
      "period": 300
    }
  ],
  "start_time": "2024-01-01T00:00:00Z",
  "end_time": "2024-01-02T00:00:00Z"
}
```

#### 请求参数示例（同时查询 CPU + GPU 利用率）

```json
{
  "root_account_id": "00000001",
  "main_account_id": "00000002",
  "role_chain": ["GPUInventoryCallerRole", "GPUInventoryReadOnlyRole"],
  "region": "us-east-1",
  "external_id": "your-external-id",
  "metric_data_queries": [
    {
      "id": "cpu_util",
      "namespace": "AWS/EC2",
      "metric_name": "CPUUtilization",
      "dimensions": [
        {"name": "InstanceId", "value": "i-0abcdef1234567890"}
      ],
      "stat": "Average",
      "period": 3600
    },
    {
      "id": "gpu_util",
      "namespace": "CWAgent",
      "metric_name": "nvidia_smi_utilization_gpu",
      "dimensions": [
        {"name": "InstanceId", "value": "i-0abcdef1234567890"}
      ],
      "stat": "Average",
      "period": 3600
    }
  ],
  "start_time": "2024-01-01T00:00:00Z",
  "end_time": "2024-01-02T00:00:00Z"
}
```

#### 返回参数示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "id": "cpu_util",
      "label": "CPUUtilization",
      "status_code": "Complete",
      "timestamps": [1704067200, 1704070800, 1704074400],
      "values": [23.5, 45.2, 12.8]
    },
    {
      "id": "gpu_util",
      "label": "nvidia_smi_utilization_gpu",
      "status_code": "Complete",
      "timestamps": [1704067200, 1704070800, 1704074400],
      "values": [78.3, 92.1, 65.7]
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | array  | 响应数据 |

#### data[n]

| 参数名称        | 参数类型     | 描述                                                                                     |
|-------------|----------|----------------------------------------------------------------------------------------|
| id          | string   | 对应请求中 metric_data_queries 的 id                                                         |
| label       | string   | 指标标签（可选），通常为指标名称                                                                       |
| status_code | string   | 查询状态码（可选），如 Complete（数据完整）、InternalError、PartialData、Forbidden                          |
| messages    | array    | 查询相关的警告或错误消息列表（可选），每个元素包含 code 和 value                                                 |
| timestamps  | int[]    | 时间戳数组（Unix 秒），与 values 一一对应                                                            |
| values      | float[]  | 指标值数组，与 timestamps 一一对应                                                                 |

#### messages[n]

| 参数名称  | 参数类型   | 描述           |
|-------|--------|--------------|
| code  | string | 消息错误码        |
| value | string | 消息详细内容       |
