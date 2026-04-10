### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询 AWS 成员账号在 CloudWatch 中实际存在的指标列表。通过 STS AssumeRole 跨账号访问成员账号的 CloudWatch ListMetrics 接口，透传返回 AWS 原始 Metric 对象。可用于发现实例上 CloudWatch Agent 实际采集了哪些指标。

### URL

POST /api/v1/cloud/vendors/aws/assume_role/cloudwatch/metrics/list

### 请求参数

| 参数名称           | 参数类型     | 必选 | 描述                                                                                                     |
|----------------|----------|----|--------------------------------------------------------------------------------------------------------|
| root_account_id | string   | 是  | 根账号 ID，用于获取 AWS 根账号凭证                                                                                  |
| main_account_id | string   | 是  | 主账号 ID，用于查询 main_account 表获取成员账号的 AWS Account ID                                                       |
| role_chain     | string[] | 是  | 角色名数组，支持 Role Chaining。中间角色在管理账号中 AssumeRole，最后一个角色在成员账号中 AssumeRole。至少包含 1 个角色名                    |
| region         | string   | 是  | AWS 区域，如 us-east-1                                                                                     |
| external_id    | string   | 否  | STS AssumeRole 的 ExternalId，用于目标角色 Trust Policy 的条件验证。仅应用于 Role Chain 最后一步                             |
| namespace      | string   | 否  | CloudWatch 命名空间过滤，如 CWAgent、AWS/EC2。不指定则返回所有命名空间                                                       |
| metric_name    | string   | 否  | 指标名称过滤，如 nvidia_smi_utilization_gpu。不指定则返回所有指标                                                         |
| dimensions     | array    | 否  | 维度过滤条件，每个元素包含 name 和 value                                                                             |

#### dimensions[n]

| 参数名称  | 参数类型   | 必选 | 描述                             |
|-------|--------|----|--------------------------------|
| name  | string | 是  | 维度名称，如 InstanceId              |
| value | string | 是  | 维度值，如 i-0abcdef1234567890      |

### 调用示例

#### 请求参数示例（查看实例的所有 CWAgent 指标）

```json
{
  "root_account_id": "00000001",
  "main_account_id": "00000002",
  "role_chain": ["gpu-readonly"],
  "region": "us-east-1",
  "namespace": "CWAgent",
  "dimensions": [
    {"name": "InstanceId", "value": "i-0abcdef1234567890"}
  ]
}
```

#### 请求参数示例（查看实例在所有命名空间的指标）

```json
{
  "root_account_id": "00000001",
  "main_account_id": "00000002",
  "role_chain": ["GPUInventoryCallerRole", "GPUInventoryReadOnlyRole"],
  "region": "us-east-1",
  "external_id": "your-external-id",
  "dimensions": [
    {"name": "InstanceId", "value": "i-0abcdef1234567890"}
  ]
}
```

#### 返回参数示例

> **注意**：`data` 为 AWS CloudWatch ListMetrics 原始 Metric 对象数组的透传。
> 完整字段列表请参考 [AWS CloudWatch Metric 文档](https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_Metric.html)。

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "Namespace": "CWAgent",
      "MetricName": "nvidia_smi_utilization_gpu",
      "Dimensions": [
        {"Name": "InstanceId", "Value": "i-0abcdef1234567890"},
        {"Name": "InstanceType", "Value": "p3.2xlarge"}
      ]
    },
    {
      "Namespace": "CWAgent",
      "MetricName": "nvidia_smi_memory_used",
      "Dimensions": [
        {"Name": "InstanceId", "Value": "i-0abcdef1234567890"},
        {"Name": "InstanceType", "Value": "p3.2xlarge"}
      ]
    },
    {
      "Namespace": "AWS/EC2",
      "MetricName": "CPUUtilization",
      "Dimensions": [
        {"Name": "InstanceId", "Value": "i-0abcdef1234567890"}
      ]
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                                                  |
|---------|--------|-----------------------------------------------------|
| code    | int    | 状态码                                                 |
| message | string | 请求信息                                                |
| data    | array  | AWS CloudWatch Metric 对象数组，完全透传 AWS 原始结构 |

#### data[n] 字段

| 参数名称       | 参数类型   | 描述                                                  |
|------------|--------|-----------------------------------------------------|
| Namespace  | string | CloudWatch 命名空间，如 AWS/EC2、CWAgent                  |
| MetricName | string | 指标名称                                                |
| Dimensions | array  | 该指标关联的维度列表，每个元素包含 Name（string）和 Value（string） |
