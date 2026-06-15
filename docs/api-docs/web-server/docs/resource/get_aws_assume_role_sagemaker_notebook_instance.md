### 描述

- 该接口提供版本：9.9.9。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询 AWS 成员账号指定 SageMaker Notebook Instance 的详情。通过 STS AssumeRole 跨账号访问成员账号的 SageMaker 原生详情接口，透传返回 AWS 原始响应。

### URL

POST /api/v1/cloud/vendors/aws/assume_role/sagemaker/notebook_instances/get

### 请求参数

| 参数名称           | 参数类型     | 必选 | 描述 |
|----------------|----------|----|------|
| root_account_id | string   | 是  | 根账号 ID，用于获取 AWS 根账号凭证 |
| main_account_id | string   | 是  | 主账号 ID，用于查询 main_account 表获取成员账号的 AWS Account ID |
| role_chain     | string[] | 是  | 角色名数组，支持 Role Chaining。中间角色在管理账号中 AssumeRole，最后一个角色在成员账号中 AssumeRole。至少包含 1 个角色名 |
| region         | string   | 是  | AWS 区域，如 us-east-1 |
| external_id    | string   | 否  | STS AssumeRole 的 ExternalId，用于目标角色 Trust Policy 的条件验证。仅应用于 Role Chain 最后一步 |
| notebook_instance_name | string | 是 | Notebook Instance 名称 |


### 调用示例

#### 请求参数示例

```json
{
  "root_account_id": "00000001",
  "main_account_id": "00000002",
  "role_chain": ["SageMakerReadOnlyRole"],
  "region": "us-east-1",
  "notebook_instance_name": "demo-notebook"
}
```

#### 返回参数示例

> **注意**：`data` 为 AWS SageMaker 原始响应结构透传，完整字段请参考 AWS SageMaker 对应 API 文档，以下仅展示常用字段。

```json
{
  "code": 0,
  "message": "",
  "data": {
    "NotebookInstanceName": "demo-notebook",
    "NotebookInstanceStatus": "InService",
    "InstanceType": "ml.g4dn.xlarge",
    "RoleArn": "arn:aws:iam::123456789012:role/SageMakerExecutionRole",
    "SubnetId": "subnet-abc123",
    "SecurityGroups": ["sg-abc123"],
    "Url": "https://notebook.example.aws"
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|--------|--------|------|
| code | int | 状态码 |
| message | string | 请求信息 |
| data | object | AWS SageMaker `DescribeNotebookInstance` 原始响应透传 |
| data.NotebookInstanceName | string | Notebook Instance 名称 |
| data.NotebookInstanceStatus | string | Notebook Instance 状态 |
| data.InstanceType | string | Notebook 使用的实例规格，如 `ml.g4dn.xlarge` |
| data.RoleArn | string | SageMaker 执行角色 ARN |
| data.SubnetId | string | 关联子网 ID |
| data.SecurityGroups | array | 安全组 ID 列表 |
| data.Url | string | Notebook 访问地址 |
