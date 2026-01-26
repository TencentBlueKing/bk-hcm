### 描述

- 该接口提供版本：v1.8.9+。
- 该接口所需权限：资源查看。
- 该接口功能描述：获取云主机监控数据。

### URL

POST /api/v1/cloud/vendors/{vendor}/cvms/monitor/data

### 请求参数

| 参数名称        | 参数类型         | 必选 | 描述                             |
|-------------|--------------|----|--------------------------------|
| vendor      | string       | 是  | 云厂商（枚举值：tcloud，当前版本暂只支持tcloud） |
| metric_name | string       | 是  | 指标名称，例如：CPUUsage、MemUsage等     |
| period      | int64        | 是  | 监控统计周期，单位：秒，最小值：60             |
| start_time  | string       | 是  | 起始时间，格式：2006-01-02 15:04:05    |
| end_time    | string       | 是  | 结束时间，格式：2006-01-02 15:04:05    |
| ids         | string array | 是  | CVM ID列表，最多20个                 |

### 调用示例

#### 请求参数示例

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
        ]
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

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | Data   | 响应数据 |

#### Data

| 参数名称        | 参数类型            | 描述      |
|-------------|-----------------|---------|
| data_points | DataPoint Array | 监控数据点列表 |

#### DataPoint[n]

| 参数名称        | 参数类型          | 描述                  |
|-------------|---------------|---------------------|
| id          | string        | CVM ID（内部ID）        |
| ip          | string array  | 内网IP地址列表            |
| region      | string        | 地域                  |
| instance_id | string        | 云实例ID（云厂商侧的实例ID）    |
| timestamps  | int64 array   | 时间戳列表（Unix时间戳，单位：秒） |
| values      | float64 array | 监控值列表               |

### 说明

1. 权限要求：需要对所有查询的实例具有"资源查看"权限
2. 时间格式：入参时间格式为 `2006-01-02 15:04:05`
3. 实例数量限制：单次请求最多支持20个实例
4. 统计周期：period 参数最小值为60秒
5. 智能分组查询：系统会根据实例ID自动识别所属的账号（account_id）和地域（region），并按照 `(account_id, region)` 组合进行分组，对每个分组单独调用云厂商API查询监控数据
6. 跨账号跨地域支持：支持查询不同账号、不同地域的实例监控数据，系统会自动处理分组和聚合
7. 不指定SpecifyStatistics参数：本接口不指定统计方式（avg/min/max），按照云上该资源的默认方式统计
8. 返回数据说明：
   - `id`：HCM系统内部的实例ID
   - `instance_id`：云厂商侧的实例ID（如腾讯云的 ins-xxx）
   - `ip`：实例的内网IP地址列表
   - `region`：实例所在地域

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

更多指标请参考腾讯云官方文档：https://cloud.tencent.com/document/product/248/6843
