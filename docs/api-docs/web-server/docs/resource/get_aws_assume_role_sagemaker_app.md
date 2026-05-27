### 描述

- 该接口提供版本：9.9.9。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询 AWS 成员账号的 SageMaker Studio App 详情。通过 STS AssumeRole 跨账号访问成员账号的 SageMaker 原生接口，透传返回 AWS 原始响应。

### URL

POST /api/v1/cloud/vendors/aws/assume_role/sagemaker/apps/get

### 请求参数

| 参数名称           | 参数类型     | 必选 | 描述 |
|----------------|----------|----|------|
| root_account_id | string   | 是  | 根账号 ID，用于获取 AWS 根账号凭证 |
| main_account_id | string   | 是  | 主账号 ID，用于查询 main_account 表获取成员账号的 AWS Account ID |
| role_chain     | string[] | 是  | 角色名数组，支持 Role Chaining。中间角色在管理账号中 AssumeRole，最后一个角色在成员账号中 AssumeRole。至少包含 1 个角色名 |
| region         | string   | 是  | AWS 区域，如 us-east-1 |
| external_id    | string   | 否  | STS AssumeRole 的 ExternalId，用于目标角色 Trust Policy 的条件验证。仅应用于 Role Chain 最后一步 |
| domain_id | string | 是 | Studio Domain ID |
| app_type | string | 是 | App 类型，如 JupyterServer、KernelGateway、JupyterLab、CodeEditor 等 AWS 原生枚举值 |
| app_name | string | 是 | App 名称 |
| user_profile_name | string | 否 | User Profile 名称。未设置时需要设置 space_name |
| space_name | string | 否 | Space 名称。未设置时需要设置 user_profile_name |


### 调用示例

#### 请求参数示例

```json
{
  "root_account_id": "00000001",
  "main_account_id": "00000002",
  "role_chain": ["SageMakerReadOnlyRole"],
  "region": "us-east-1",
  "domain_id": "d-xxxxxxxxxxxx",
  "user_profile_name": "demo-user",
  "app_type": "JupyterLab",
  "app_name": "default"
}
```

#### 返回参数示例

> **注意**：`data` 为 AWS SageMaker 原始响应结构透传，完整字段请参考 AWS SageMaker 对应 API 文档，以下仅展示常用字段。

```json
{
  "code": 0,
  "message": "",
  "data": {
    "DomainId": "d-xxxxxxxxxxxx",
    "AppName": "default",
    "AppType": "JupyterLab",
    "Status": "InService",
    "ResourceSpec": {
      "InstanceType": "ml.g5.xlarge"
    }
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|--------|--------|------|
| code | int | 状态码 |
| message | string | 请求信息 |
| data | object | AWS SageMaker `DescribeApp` 原始响应透传 |
