### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询 AWS 成员账号的 GPU 实例类型列表。通过 STS AssumeRole 跨账号访问成员账号的 DescribeInstanceTypes 接口，返回含 GPU 字段（显存、型号、制造商）的实例类型列表。

### URL

POST /api/v1/cloud/vendors/aws/gpu/instance_types/list

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
      "instance_family": "p3",
      "instance_type": "p3.2xlarge",
      "gpu": 1,
      "gpu_memory": 16384,
      "gpu_name": "V100",
      "gpu_manufacturer": "NVIDIA",
      "cpu": 8,
      "memory": 62464,
      "fpga": 0,
      "network_performance": "Up to 10 Gigabit",
      "disk_size_in_gb": 0,
      "architecture": "x86_64",
      "disk_type": ""
    },
    {
      "instance_family": "t3",
      "instance_type": "t3.micro",
      "gpu": 0,
      "gpu_memory": 0,
      "gpu_name": "",
      "gpu_manufacturer": "",
      "cpu": 2,
      "memory": 1024,
      "fpga": 0,
      "network_performance": "Up to 5 Gigabit",
      "disk_size_in_gb": 0,
      "architecture": "x86_64",
      "disk_type": ""
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

| 参数名称                | 参数类型   | 描述                       |
|---------------------|--------|--------------------------|
| instance_family     | string | 实例机型族，如 p3、g4dn          |
| instance_type       | string | 实例机型，如 p3.2xlarge        |
| gpu                 | int64  | GPU 数量                   |
| gpu_memory          | int64  | 单卡 GPU 显存，单位 MiB         |
| gpu_name            | string | GPU 型号，如 V100、T4、A100    |
| gpu_manufacturer    | string | GPU 制造商，如 NVIDIA         |
| cpu                 | int64  | CPU 核数                   |
| memory              | int64  | 内存容量，单位 MiB              |
| fpga                | int64  | FPGA 数量                  |
| network_performance | string | 网络性能                     |
| disk_size_in_gb     | int64  | 本地磁盘大小，单位 GB             |
| architecture        | string | CPU 架构，如 x86_64、arm64    |
| disk_type           | string | 本地磁盘类型                   |
