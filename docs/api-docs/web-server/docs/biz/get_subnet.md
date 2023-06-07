### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询子网详情。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/subnets/{id}

### 输入参数

| 参数名称       | 参数类型   | 必选 | 描述    |
|------------|--------|----|-------|
| bk_biz_id  | int64  | 是  | 业务ID  |
| id         | string | 是  | 子网的ID |

### 调用示例

```json
```

### 腾讯云响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001",
    "vendor": "tcloud",
    "account_id": "00000001",
    "cloud_vpc_id": "vpc-xxxxxxxx",
    "cloud_id": "subnet-xxxxxxxx",
    "name": "subnet-default",
    "region": "ap-guangzhou",
    "zone": "ap-guangzhou-6",
    "ipv4_cidr": [
      "127.0.0.0/16"
    ],
    "ipv6_cidr": [
      "::/24"
    ],
    "memo": "default subnet",
    "vpc_id": "00000006",
    "bk_biz_id": 123,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "is_default": false,
      "network_acl_id": ""
    }
  }
}
```

### AWS响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001",
    "vendor": "aws",
    "account_id": "00000001",
    "cloud_vpc_id": "vpc-xxxxxxxx",
    "cloud_id": "subnet-xxxxxxxx",
    "name": "subnet-default",
    "region": "us-east-1",
    "zone": "us-east-1a",
    "ipv4_cidr": [
      "127.0.0.0/16"
    ],
    "ipv6_cidr": [
      "::/24"
    ],
    "memo": "default subnet",
    "vpc_id": "00000006",
    "bk_biz_id": 123,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "state": "available",
      "is_default": false,
      "map_public_ip_on_launch": false,
      "assign_ipv6_address_on_creation": false,
      "hostname_type": "ip-name"
    }
  }
}
```

### GCP响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001",
    "vendor": "gcp",
    "account_id": "00000001",
    "cloud_vpc_id": "https://www.googleapis.com/compute/v1/projects/xxx/global/networks/test",
    "cloud_id": "456",
    "name": "test",
    "region": "https://www.googleapis.com/compute/v1/projects/xxx/regions/us-west1",
    "ipv4_cidr": [
      "127.0.0.0/16"
    ],
    "ipv6_cidr": [
      "::/24"
    ],
    "memo": "default subnet",
    "vpc_id": "00000006",
    "bk_biz_id": 123,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "self_link": "https://www.googleapis.com/compute/v1/projects/xxx/regions/us-west1/subnetworks/test",
      "stack_type": "IPV4_IPV6",
      "ipv6_access_type": "INTERNAL",
      "gateway_address": "127.0.0.1",
      "enable_flow_logs": false,
      "private_ip_google_access": false
    }
  }
}
```

### Azure响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001",
    "vendor": "azure",
    "account_id": "00000001",
    "cloud_vpc_id": "xxx",
    "cloud_id": "subnet-xxxxxxxx",
    "name": "subnet-default",
    "ipv4_cidr": [
      "127.0.0.0/16"
    ],
    "ipv6_cidr": [
      "::/24"
    ],
    "memo": "default subnet",
    "vpc_id": "00000006",
    "bk_biz_id": 123,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "resource_group": "test",
      "nat_gateway": "xxx",
      "cloud_security_group_id": "xxx"
    }
  }
}
```

### 华为云响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001",
    "vendor": "huawei",
    "account_id": "00000001",
    "cloud_vpc_id": "xxx",
    "cloud_id": "subnet-xxxxxxxx",
    "name": "subnet-default",
    "region": "ap-southeast-1",
    "ipv4_cidr": [
      "127.0.0.0/16"
    ],
    "ipv6_cidr": [
      "::/24"
    ],
    "memo": "default subnet",
    "vpc_id": "00000006",
    "bk_biz_id": 123,
    "creator": "tom",
    "reviser": "tom",
    "created_at": "2019-07-29 11:57:20",
    "updated_at": "2019-07-29 11:57:20",
    "extension": {
      "status": "ACTIVE",
      "dhcp_enable": true,
      "dns_list": [
        "127.0.0.1"
      ],
      "gateway_ip": "127.0.0.2",
      "ntp_addresses": [
        "127.0.0.3"
      ]
    }
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称         | 参数类型         | 描述                            |
|--------------|--------------|-------------------------------|
| id           | string       | 子网的ID                         |
| vendor       | string       | 云厂商                           |
| account_id   | string       | 账号ID                          |
| cloud_vpc_id | string       | VPC的云ID                       |
| cloud_id     | string       | 子网的云ID                        |
| name         | string       | 子网名称                          |
| region       | string       | 地域                            |
| zone         | string       | 可用区                           |
| ipv4_cidr    | string array | IPv4 CIDR                     |
| ipv6_cidr    | string array | IPv6 CIDR                     |
| memo         | string       | 备注                            |
| vpc_id       | string       | VPC的云ID                       |
| bk_biz_id    | int64        | 业务ID，-1表示没有分配到业务              |
| creator      | string       | 创建者                           |
| reviser      | string       | 更新者                           |
| created_at   | string       | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at   | string       | 更新时间，标准格式：2006-01-02T15:04:05Z |
| extension    | object       | 云厂商私有结构                       |

#### data.extension(tcloud)

| 参数名称                 | 参数类型    | 描述        |
|----------------------|---------|-----------|
| is_default           | boolean | 是否是默认子网   |
| cloud_network_acl_id | string  | 关联的ACL的ID |

#### data.extension(aws)

| 参数名称                            | 参数类型    | 描述                                |
|---------------------------------|---------|-----------------------------------|
| state                           | string  | 状态（枚举值：pending、available）         |
| is_default                      | boolean | 是否是默认子网                           |
| map_public_ip_on_launch         | boolean | 是否自动分配公有 IPv4 地址                  |
| assign_ipv6_address_on_creation | boolean | 是否自动分配 IPv6 地址                    |
| hostname_type                   | string  | 主机名称类型（枚举值：ip-name、resource-name） |

#### data.extension(gcp)

| 参数名称                     | 参数类型    | 描述                              |
|--------------------------|---------|---------------------------------|
| vpc_self_link            | string  | Vpc资源URL                        |
| self_link                | string  | 资源URL                           |
| stack_type               | string  | IP栈类型（枚举值：IPV4_IPV6、IPV4_ONLY）  |
| ipv6_access_type         | string  | IPv6权限类型（枚举值：EXTERNAL、INTERNAL） |
| gateway_address          | string  | 网关地址                            |
| private_ip_google_access | boolean | 是否启用专用Google访问通道                |
| enable_flow_logs         | boolean | 是否启用流日志                         |

#### data.extension(azure)

| 参数名称                    | 参数类型   | 描述     |
|-------------------------|--------|--------|
| resource_group          | string | 资源组    |
| nat_gateway             | string | NAT网关  |
| cloud_security_group_id | string | 云安全组ID |

#### data.extension(huawei)

| 参数名称          | 参数类型         | 描述                           |
|---------------|--------------|------------------------------|
| status        | string       | 状态（枚举值：ACTIVE、UNKNOWN、ERROR） |
| dhcp_enable   | boolean      | 是否开启dhcp功能                   |
| gateway_ip    | string       | 网关地址                         |
| dns_list      | string array | DNS服务器地址                     |
| ntp_addresses | string array | NTP服务器地址                     |
