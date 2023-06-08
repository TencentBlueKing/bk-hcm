### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务-IaaS资源创建。
- 该接口功能描述：创建子网。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/subnets/create

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/subnets/create

### 输入参数

输入参数由接口通用参数和vendor对应的云厂商差异参数组成。

#### 接口通用参数

| 参数名称         | 参数类型   | 必选  | 描述                                   |
|--------------|--------|-----|--------------------------------------|
| bk_biz_id    | int64  | 是   | 业务ID                                 |
| vendor       | string | 是   | 云厂商（枚举值：tcloud、aws、gcp、azure、huawei） |
| account_id   | string | 是   | 账号ID                                 |
| cloud_vpc_id | string | 是   | VPC的云ID                              |
| name         | string | 是   | 子网名称                                 |
| memo         | string | 否   | 备注                                   |

#### 云厂商差异参数[tcloud]

| 参数名称                 | 参数类型   | 必选  | 描述         |
|----------------------|--------|-----|------------|
| region               | string | 是   | 地域         |
| zone                 | string | 是   | 可用区        |
| ipv4_cidr            | string | 是   | IPv4 CIDR  |
| cloud_route_table_id | string | 否   | 关联的路由表的云ID |

#### 云厂商差异参数[aws]

| 参数名称      | 参数类型   | 必选                           | 描述        |
|-----------|--------|------------------------------|-----------|
| region    | string | 是                            | 地域        |
| zone      | string | 是                            | 可用区       |
| ipv4_cidr | string | ipv4_cidr和ipv6_cidr中至少需要填写一个 | IPv4 CIDR |
| ipv6_cidr | string | ipv4_cidr和ipv6_cidr中至少需要填写一个 | IPv6 CIDR |

#### 云厂商差异参数[gcp]

| 参数名称                     | 参数类型         | 必选  | 描述               |
|--------------------------|--------------|-----|------------------|
| region                   | string       | 是   | 地域               |
| ipv4_cidr                | string array | 是   | IPv4 CIDR        |
| private_ip_google_access | boolean      | 否   | 是否启用专用Google访问通道 |
| enable_flow_logs         | boolean      | 否   | 是否启用流日志          |

#### 云厂商差异参数[azure]

| 参数名称                    | 参数类型         | 必选  | 描述         |
|-------------------------|--------------|-----|------------|
| resource_group          | string       | 是   | 资源组        |
| ipv4_cidr               | string array | 是   | IPv4 CIDR  |
| ipv6_cidr               | string array | 否   | IPv6 CIDR  |
| nat_gateway             | string       | 否   | NAT网关      |
| cloud_security_group_id | string       | 否   | 云安全组ID     |
| cloud_route_table_id    | string       | 否   | 关联的路由表的云ID |

#### 云厂商差异参数[huawei]

| 参数名称        | 参数类型    | 必选  | 描述        |
|-------------|---------|-----|-----------|
| region      | string  | 是   | 地域        |
| zone        | string  | 否   | 可用区       |
| ipv4_cidr   | string  | 是   | IPv4 CIDR |
| ipv6_enable | boolean | 否   | 是否支持IPv6  |
| gateway_ip  | string  | 是   | 网关地址      |

### 腾讯云调用示例

```json
{
  "vendor": "tcloud",
  "account_id": "00000001",
  "cloud_vpc_id": "vpc-xxxxxxxx",
  "name": "test-subnet",
  "memo": "test subnet",
  "region": "ap-guangzhou",
  "zone": "ap-guangzhou-6",
  "ipv4_cidr": "127.0.0.0/16",
  "cloud_route_table_id": "rtb-xxxxxxxx"
}
```

### AWS调用示例

```json
{
  "vendor": "aws",
  "account_id": "00000001",
  "cloud_vpc_id": "vpc-xxxxxxxx",
  "name": "test-subnet",
  "memo": "test subnet",
  "region": "us-east-1",
  "zone": "us-east-1a",
  "ipv4_cidr": "127.0.0.0/16",
  "ipv6_cidr": "::/24"
}
```

### GCP调用示例

```json
{
  "vendor": "gcp",
  "account_id": "00000001",
  "cloud_vpc_id": "https://www.googleapis.com/compute/v1/projects/xxx/global/networks/test",
  "name": "test-subnet",
  "memo": "test subnet",
  "region": "https://www.googleapis.com/compute/v1/projects/xxx/regions/us-west1",
  "ipv4_cidr": "127.0.0.0/16",
  "private_ip_google_access": false,
  "enable_flow_logs": false
}
```

### Azure调用示例

```json
{
  "vendor": "azure",
  "account_id": "00000001",
  "cloud_vpc_id": "xxx",
  "name": "test-subnet",
  "memo": "test subnet",
  "resource_group": "test",
  "ipv4_cidr": [
    "127.0.0.0/16"
  ],
  "ipv6_cidr": [
    "::/24"
  ],
  "cloud_route_table_id": "xxx",
  "nat_gateway": "xxx",
  "cloud_security_group_id": "xxx"
}
```

### 华为云调用示例

```json
{
  "vendor": "huawei",
  "account_id": "00000001",
  "cloud_vpc_id": "xxx",
  "name": "test-subnet",
  "memo": "test subnet",
  "region": "ap-southeast-1",
  "zone": "ap-southeast-1f",
  "ipv4_cidr": "127.0.0.0/16",
  "ipv6_enable": false,
  "gateway_ip": "127.0.0.2"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000003"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 调用数据 |

#### data

| 参数名称 | 参数类型   | 描述      |
|------|--------|---------|
| id   | string | 创建的子网ID |
