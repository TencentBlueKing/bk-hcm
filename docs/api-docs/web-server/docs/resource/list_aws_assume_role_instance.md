### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询 AWS 成员账号的 EC2 实例列表。通过 STS AssumeRole 跨账号访问成员账号的 DescribeInstances 接口，透传返回 AWS 原始 Instance 对象。

### URL

POST /api/v1/cloud/vendors/aws/assume_role/instances/list

### 请求参数

| 参数名称           | 参数类型     | 必选 | 描述                                                                                                     |
|----------------|----------|----|--------------------------------------------------------------------------------------------------------|
| root_account_id | string   | 是  | 根账号 ID，用于获取 AWS 根账号凭证                                                                                  |
| main_account_id | string   | 是  | 主账号 ID，用于查询 main_account 表获取成员账号的 AWS Account ID                                                       |
| role_chain     | string[] | 是  | 角色名数组，支持 Role Chaining。中间角色在管理账号中 AssumeRole，最后一个角色在成员账号中 AssumeRole。至少包含 1 个角色名                    |
| region         | string   | 是  | AWS 区域，如 us-east-1                                                                                     |
| external_id    | string   | 否  | STS AssumeRole 的 ExternalId，用于目标角色 Trust Policy 的条件验证。仅应用于 Role Chain 最后一步                             |

### 调用示例

#### 请求参数示例（单步 AssumeRole）

```json
{
  "root_account_id": "00000001",
  "main_account_id": "00000002",
  "role_chain": ["gpu-readonly"],
  "region": "us-east-1"
}
```

#### 请求参数示例（多步 Role Chain）

```json
{
  "root_account_id": "00000001",
  "main_account_id": "00000002",
  "role_chain": ["GPUInventoryCallerRole", "GPUInventoryReadOnlyRole"],
  "region": "us-east-1",
  "external_id": "your-external-id"
}
```

#### 返回参数示例

> **注意**：`data` 为 AWS EC2 DescribeInstances 原始 Instance 对象数组的透传。
> 完整字段列表请参考 [AWS EC2 Instance 文档](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Instance.html)，以下仅展示常用字段。

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "InstanceId": "i-0abcdef1234567890",
      "InstanceType": "p3.2xlarge",
      "State": {
        "Code": 16,
        "Name": "running"
      },
      "PrivateIpAddress": "192.168.1.100",
      "PublicIpAddress": "54.123.45.67",
      "Placement": {
        "AvailabilityZone": "us-east-1a",
        "GroupName": "",
        "Tenancy": "default"
      },
      "Architecture": "x86_64",
      "ImageId": "ami-0abcdef1234567890",
      "LaunchTime": "2025-06-01T12:00:00Z",
      "SubnetId": "subnet-abc123",
      "VpcId": "vpc-abc123",
      "Tags": [
        {
          "Key": "Name",
          "Value": "gpu-worker-01"
        }
      ]
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                                           |
|---------|--------|----------------------------------------------|
| code    | int    | 状态码                                          |
| message | string | 请求信息                                         |
| data    | array  | AWS EC2 Instance 对象数组，完全透传 AWS 原始结构 |

#### data[n] 常用字段

> 以下仅列出常用字段，实际返回包含 AWS EC2 Instance 的全部字段。

| 参数名称             | 参数类型   | 描述                                               |
|------------------|--------|--------------------------------------------------|
| InstanceId       | string | EC2 实例 ID                                        |
| InstanceType     | string | 实例机型，如 p3.2xlarge、g4dn.xlarge                   |
| State            | object | 实例状态，含 Code（int）和 Name（string，如 running/stopped）|
| PrivateIpAddress | string | 内网 IP 地址                                         |
| PublicIpAddress  | string | 公网 IP 地址，无公网 IP 时该字段不存在                         |
| Placement        | object | 放置信息，含 AvailabilityZone、Tenancy 等               |
| Architecture     | string | CPU 架构，如 x86_64、arm64                            |
| ImageId          | string | AMI 镜像 ID                                        |
| LaunchTime       | string | 启动时间，ISO 8601 格式                                |
| SubnetId         | string | 子网 ID                                            |
| VpcId            | string | VPC ID                                           |
| Tags             | array  | 标签列表，每项含 Key 和 Value                            |
