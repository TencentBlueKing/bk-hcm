### 描述

- 该接口提供版本：v1.8.9+。
- 该接口所需权限：资源查看。
- 该接口功能描述：获取云主机监控数据。

### URL

POST /api/v1/cloud/vendors/{vendor}/cvms/monitor/data

### 请求参数

#### 通用参数

| 参数名称      | 参数类型      | 必选 | 描述                                            |
|-------------|--------------|------|------------------------------------------------|
| vendor      | string       | 是   | 云厂商（枚举值：`tcloud`、`huawei`、`aws`）         |
| metric_name | string       | 是   | 指标名称，例如：`CPUUsage`、`MemUsage`、`cpu_util` |
| period      | int64        | 是   | 监控统计周期，单位：秒                              |
| ids         | string array | 是   | CVM ID 列表，最多 20 个                           |

#### 腾讯云（tcloud）参数

| 参数名称     | 参数类型 | 必选 | 描述                                |
|-------------|--------|------|------------------------------------|
| start_time  | string | 是   | 起始时间，格式：`2006-01-02 15:04:05` |
| end_time    | string | 是   | 结束时间，格式：`2006-01-02 15:04:05` |

#### 华为云（huawei）参数

| 参数名称     | 参数类型 | 必选 | 描述                                                    |
|------------|---------|------|--------------------------------------------------------|
| start_time | int64   | 是   | 起始时间，Unix 毫秒时间戳                                  |
| end_time   | int64   | 是   | 结束时间，Unix 毫秒时间戳                                  |
| namespace  | string  | 否   | 监控命名空间，例如：`SYS.ECS`、`SYS.VPC`，不传默认 `SYS.ECS` |
| filter     | string  | 否   | 聚合方式，period为1（原始值）时，filter字段不生效，枚举值：`average`、`variance`、`max`、`min`、`sum`，不传默认 `average`  |

#### AWS（aws）参数

| 参数名称     | 参数类型 | 必选 | 描述                                                                 |
|------------|---------|------|----------------------------------------------------------------------|
| start_time | string  | 是   | 起始时间，RFC3339 UTC 格式，例如：`2026-04-09T00:00:00Z`              |
| end_time   | string  | 是   | 结束时间，RFC3339 UTC 格式，例如：`2026-04-09T01:00:00Z`              |

### 调用示例

#### 请求参数示例（tcloud）

```json
{
  "metric_name": "CPUUsage",
  "period": 60,
  "start_time": "2024-01-20 10:00:00",
  "end_time": "2024-01-20 11:00:00",
  "ids": [
    "00000001",
    "00000002"
  ]
}
```

#### 请求参数示例（huawei）

```json
{
  "metric_name": "cpu_util",
  "period": 1,
  "start_time": 1705718400000,
  "end_time": 1705718460000,
  "namespace": "SYS.ECS",
  "filter": "average",
  "ids": [
    "00000001",
    "00000002"
  ]
}
```

#### 请求参数示例（aws）

```json
{
  "metric_name": "LanOuttraffic",
  "period": 300,
  "start_time": "2026-04-09T00:00:00Z",
  "end_time": "2026-04-09T01:00:00Z",
  "ids": [
    "00000001",
    "00000002"
  ]
}
```

#### 返回参数示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data_points": [
      {
        "id": "00000001",
        "ip": [
          "10.0.0.1",
          "10.0.0.2"
        ],
        "region": "ap-guangzhou",
        "instance_id": "ins-xxxxxxxx",
        "timestamps": [
          1705718400,
          1705718460,
          1705718520
        ],
        "values": [
          10.5,
          12.3,
          11.8
        ],
        "extensions": {
          "namespace": "SYS.ECS",
          "metric_name": "cpu_util",
          "unit": "%",
          "filter": "average"
        }
      },
      {
        "id": "00000002",
        "ip": [
          "10.0.1.1"
        ],
        "region": "ap-shanghai",
        "instance_id": "ins-yyyyyyyy",
        "timestamps": [
          1705718400,
          1705718460,
          1705718520
        ],
        "values": [
          8.2,
          9.1,
          8.7
        ]
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称  | 参数类型 | 描述    |
|---------|---------|---------|
| code    | int     | 状态码   |
| message | string  | 请求信息 |
| data    | Data    | 响应数据 |

#### Data

| 参数名称     | 参数类型          | 描述         |
|-------------|-----------------|--------------|
| data_points | DataPoint Array | 监控数据点列表 |

#### DataPoint[n]

| 参数名称     | 参数类型        | 描述                                   |
|-------------|---------------|----------------------------------------|
| id          | string        | CVM ID（内部ID）                        |
| ip          | string array  | 内网IP地址列表                           |
| region      | string        | 地域                                    |
| instance_id | string        | 云实例ID（云厂商侧的实例ID）               |
| timestamps  | int64 array   | 时间戳列表（Unix时间戳，单位：秒）          |
| values      | float64 array | 监控值列表                               |
| extensions  | object        | 厂商扩展字段（可选），用于返回厂商特有监控信息 |

### 说明

1. 权限要求：需要对所有查询的实例具有"资源查看"权限
2. 时间格式：
   - `tcloud`：使用 `start_time/end_time`，格式 `2006-01-02 15:04:05`
   - `huawei`：使用 `start_time/end_time`，Unix 毫秒时间戳
   - `aws`：使用 `start_time/end_time`，RFC3339 UTC 格式（必须为 UTC 时区）
3. 实例数量限制：单次请求最多支持20个实例
4. 统计周期：
   - `tcloud`：period 最小值为 60
   - `huawei`：支持 `period=1` 实时数据，其他取值遵循华为云 CES 约束
5. 华为云扩展参数：
   - `namespace`：可选，不传默认 `SYS.ECS`
   - `filter`：可选，不传默认 `average`
6. AWS 流量语义（Phase 1）：
   - `LanOuttraffic` 与 `WanOuttraffic` 均映射 AWS `NetworkOut`（实例总流量）
   - `LanIntraffic` 与 `WanIntraffic` 均映射 AWS `NetworkIn`（实例总流量）
   - 返回值保持 AWS 云厂商原始语义，不做 Mbps 转换
   - 可通过 `extensions` 中的 `source_metric_name`、`semantic_phase`、`traffic_scope` 等字段识别语义
7. 智能分组查询：系统会根据实例ID自动识别所属的账号（account_id）和地域（region），并按照 `(account_id, region)` 组合进行分组，对每个分组单独调用云厂商API查询监控数据
8. 跨账号跨地域支持：支持查询不同账号、不同地域的实例监控数据，系统会自动处理分组和聚合
9. 返回数据说明：
   - `id`：HCM系统内部的实例ID
   - `instance_id`：云厂商侧的实例ID（如腾讯云的 ins-xxx）
   - `ip`：实例的内网IP地址列表
   - `region`：实例所在地域
   - `extensions`：厂商扩展字段，可能包含 `namespace`、`metric_name`、`unit`、`filter`、`dimensions` 等

### 常用指标

| 指标名称          | 说明      | 单位   |
|---------------|---------|------|
| CPUUsage      | CPU使用率  | %    |
| CPULoadAvg    | CPU平均负载 | -    |
| MemUsage      | 内存使用率   | %    |
| MemUsed       | 内存使用量   | MB   |
| TcpCurrEstab  | TCP连接数  | 个    |
| LanOuttraffic | 内网出带宽   | Mbps |
| LanIntraffic  | 内网入带宽   | Mbps |
| WanOuttraffic | 外网出带宽   | Mbps |
| WanIntraffic  | 外网入带宽   | Mbps |

### 更多指标请参考：
- 腾讯云官方文档：https://cloud.tencent.com/document/product/248/6843
- 华为云官方文档：
  - https://support.huaweicloud.com/api-ces/ces_03_0059.html
  - https://support.huaweicloud.com/usermanual-ecs/ecs_03_1002.html
- AWS官方文档：https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html

