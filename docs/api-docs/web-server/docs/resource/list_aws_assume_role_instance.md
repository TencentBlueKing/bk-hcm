### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询 AWS 成员账号的 EC2 实例列表。通过 STS AssumeRole 跨账号访问成员账号的 DescribeInstances 接口，返回指定 region 下的实例列表。

### URL

POST /api/v1/cloud/vendors/aws/assume_role/instances/list

### 请求参数

| 参数名称        | 参数类型     | 必选 | 描述                                                                                                     |
|-------------|----------|----|--------------------------------------------------------------------------------------------------------|
| cloud_id    | string   | 是  | 成员账号 AWS Account ID（全球唯一），HCM 自动反查对应根账号                                                                |
| role_chain  | string[] | 是  | 角色名数组，支持 Role Chaining。中间角色在管理账号中 AssumeRole，最后一个角色在成员账号（cloud_id）中 AssumeRole。至少包含 1 个角色名 |
| region      | string   | 是  | AWS 区域，如 us-east-1                                                                                     |
| external_id | string   | 否  | STS AssumeRole 的 ExternalId，用于目标角色 Trust Policy 的条件验证。仅应用于 Role Chain 最后一步                             |

### 调用示例

#### 请求参数示例（单步 AssumeRole）

```json
{
  "cloud_id": "123456789012",
  "role_chain": ["gpu-readonly"],
  "region": "us-east-1"
}
```

#### 请求参数示例（多步 Role Chain）

```json
{
  "cloud_id": "123456789012",
  "role_chain": ["GPUInventoryCallerRole", "GPUInventoryReadOnlyRole"],
  "region": "us-east-1",
  "external_id": "your-external-id"
}
```

#### 返回参数示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "instance_id": "i-0abcdef1234567890",
      "instance_type": "p3.2xlarge",
      "state": "running",
      "private_ip": "192.168.1.100",
      "public_ip": "54.123.45.67",
      "region": "us-east-1",
      "zone": "us-east-1a"
    },
    {
      "instance_id": "i-0fedcba0987654321",
      "instance_type": "g4dn.xlarge",
      "state": "stopped",
      "private_ip": "192.168.2.200",
      "public_ip": "",
      "region": "us-east-1",
      "zone": "us-east-1b"
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

| 参数名称          | 参数类型   | 描述                                     |
|---------------|--------|----------------------------------------|
| instance_id   | string | EC2 实例 ID                              |
| instance_type | string | 实例机型，如 p3.2xlarge                      |
| state         | string | 实例状态，如 running、stopped、terminated 等    |
| private_ip    | string | 内网 IP 地址                               |
| public_ip     | string | 公网 IP 地址，无公网 IP 时为空字符串                 |
| region        | string | 所在区域                                   |
| zone          | string | 所在可用区                                  |
