### 描述

- 该接口提供版本：9.9.9。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询 AWS 成员账号的 SageMaker Training Plan Offering 搜索结果。通过 STS AssumeRole 跨账号访问成员账号的 SageMaker 原生接口，透传返回 AWS 原始响应。

### URL

POST /api/v1/cloud/vendors/aws/assume_role/sagemaker/training_plan_offerings/search

### 请求参数

| 参数名称           | 参数类型     | 必选 | 描述 |
|----------------|----------|----|------|
| root_account_id | string   | 是  | 根账号 ID，用于获取 AWS 根账号凭证 |
| main_account_id | string   | 是  | 主账号 ID，用于查询 main_account 表获取成员账号的 AWS Account ID |
| role_chain     | string[] | 是  | 角色名数组，支持 Role Chaining。中间角色在管理账号中 AssumeRole，最后一个角色在成员账号中 AssumeRole。至少包含 1 个角色名 |
| region         | string   | 是  | AWS 区域，如 us-east-1 |
| external_id    | string   | 否  | STS AssumeRole 的 ExternalId，用于目标角色 Trust Policy 的条件验证。仅应用于 Role Chain 最后一步 |
| duration_hours | int64 | 否 | 期望时长（小时） |
| end_time_before | string | 否 | 结束时间上界，RFC3339 时间格式 |
| instance_count | int32 | 否 | 期望实例数量，最小值 1 |
| instance_type | string | 否 | 期望实例类型，如 ml.p5.48xlarge |
| start_time_after | string | 否 | 开始时间下界，RFC3339 时间格式 |
| target_resources | string[] | 否 | 目标资源类型，如 SageMakerTrainingJob、SageMakerHyperPodCluster、SageMakerEndpoint |
| training_plan_arn | string | 否 | 已有 Training Plan ARN，用于搜索扩展 offering |
| ultra_server_count | int32 | 否 | 期望 UltraServer 数量，最小值 1 |
| ultra_server_type | string | 否 | 期望 UltraServer 类型，如 ml.u-p6e-gb200x72 |


### 调用示例

#### 请求参数示例

```json
{
  "root_account_id": "00000001",
  "main_account_id": "00000002",
  "role_chain": ["SageMakerReadOnlyRole"],
  "region": "us-east-1",
  "instance_type": "ml.p5.48xlarge",
  "instance_count": 8,
  "duration_hours": 24,
  "target_resources": ["SageMakerTrainingJob"]
}
```

#### 返回参数示例

> **注意**：`data` 为 AWS SageMaker 原始响应结构透传，完整字段请参考 AWS SageMaker 对应 API 文档，以下仅展示常用字段。HCM 不在该接口中做 GPU 或实例规格业务推导。

```json
{
  "code": 0,
  "message": "",
  "data": {
    "TrainingPlanOfferings": [
      {
        "TrainingPlanOfferingId": "tpo-xxxxxxxx",
        "TargetResources": ["SageMakerTrainingJob"],
        "ReservedCapacityOfferings": [
          {
            "InstanceType": "ml.p5.48xlarge",
            "InstanceCount": 8
          }
        ]
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|--------|--------|------|
| code | int | 状态码 |
| message | string | 请求信息 |
| data | object | AWS SageMaker `SearchTrainingPlanOfferings` 原始响应透传 |
