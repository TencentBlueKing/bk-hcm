### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询 AWS 成员账号的 EC2 实例列表。通过 STS AssumeRole 跨账号访问成员账号的 DescribeInstances 接口，返回指定 region 下的实例列表。该接口为数据透传，不持久化到 HCM 本地数据库。

### URL

POST /api/v1/cloud/vendors/aws/gpu/instances/list

### 请求参数

| 参数名称      | 参数类型   | 必选 | 描述                                           |
|-----------|--------|----|----------------------------------------------|
| cloud_id  | string | 是  | 成员账号 AWS Account ID（全球唯一），HCM 自动反查对应根账号      |
| role_name | string | 是  | 成员账号中的 IAM Role 名称，用于拼接 Role ARN 执行 AssumeRole |
| region    | string | 是  | AWS 区域，如 us-east-1                           |

### 调用示例

#### 请求参数示例

```json
{
  "cloud_id": "123456789012",
  "role_name": "gpu-readonly",
  "region": "us-east-1"
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
      "private_ip": "10.0.1.100",
      "public_ip": "54.123.45.67",
      "region": "us-east-1",
      "zone": "us-east-1a"
    },
    {
      "instance_id": "i-0fedcba0987654321",
      "instance_type": "g4dn.xlarge",
      "state": "stopped",
      "private_ip": "10.0.2.200",
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
