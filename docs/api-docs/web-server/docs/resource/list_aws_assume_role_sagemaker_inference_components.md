### 描述

- 该接口提供版本：9.9.9。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询 AWS 成员账号 SageMaker Inference Component 列表。通过 STS AssumeRole 跨账号访问成员账号的 SageMaker 原生接口，透传返回 AWS 原始响应。

### URL

POST /api/v1/cloud/vendors/aws/assume_role/sagemaker/inference_components/list

### 请求参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|--------|--------|----|------|
| root_account_id | string | 是 | 根账号 ID，用于获取 AWS 根账号凭证 |
| main_account_id | string | 是 | 主账号 ID，用于查询 main_account 表获取成员账号的 AWS Account ID |
| role_chain | string[] | 是 | 角色名数组，支持 Role Chaining。中间角色在管理账号中 AssumeRole，最后一个角色在成员账号中 AssumeRole。至少包含 1 个角色名 |
| region | string | 是 | AWS 区域，如 us-east-1 |
| external_id | string | 否 | STS AssumeRole 的 ExternalId，用于目标角色 Trust Policy 的条件验证。仅应用于 Role Chain 最后一步 |
| creation_time_after | string | 否 | 创建时间下界，RFC3339 时间格式 |
| creation_time_before | string | 否 | 创建时间上界，RFC3339 时间格式 |
| endpoint_name_equals | string | 否 | 按 endpoint 名称过滤 |
| last_modified_time_after | string | 否 | 最近修改时间下界，RFC3339 时间格式 |
| last_modified_time_before | string | 否 | 最近修改时间上界，RFC3339 时间格式 |
| max_results | int32 | 否 | 分页大小，最小值 1 |
| name_contains | string | 否 | 名称包含过滤 |
| next_token | string | 否 | AWS 分页 token |
| sort_by | string | 否 | AWS 原生排序字段 |
| sort_order | string | 否 | AWS 原生排序方向 |
| status_equals | string | 否 | 状态过滤 |
| variant_name_equals | string | 否 | 按 production variant 名称过滤 |


### 调用示例

#### 请求参数示例

```json
{
  "root_account_id": "00000001",
  "main_account_id": "00000002",
  "role_chain": ["SageMakerReadOnlyRole"],
  "region": "us-east-1",
  "endpoint_name_equals": "demo-endpoint"
}
```

#### 返回参数示例

> **注意**：`data` 为 AWS SageMaker 原始响应结构透传，完整字段请参考 AWS SageMaker 对应 API 文档，以下仅展示常用字段。HCM 不在该接口中做 GPU 或实例规格业务推导。

```json
{
  "code": 0,
  "message": "",
  "data": {"InferenceComponentSummaries":[{"InferenceComponentName":"demo-component","EndpointName":"demo-endpoint"}]}
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|--------|--------|------|
| code | int | 状态码 |
| message | string | 请求信息 |
| data | object | AWS SageMaker `ListInferenceComponents` 原始响应透传 |
