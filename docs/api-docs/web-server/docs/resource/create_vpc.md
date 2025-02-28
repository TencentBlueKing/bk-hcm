### 描述

- 该接口提供版本： v1.1.17+。
- 该接口所需权限：IaaS资源创建。
- 该接口功能描述：创建VPC。

### URL

POST /api/v1/cloud/vpcs/create

### 输入参数

#### tcloud

| 参数名称        | 参数类型   | 必选 | 描述                |
|-------------|--------|----|-------------------|
| account_id  | string | 是  | 账号ID              |
| region      | string | 是  | 地域                |
| name        | string | 是  | 名称                |
| ipv4_cidr   | string | 是  | IPv4 CIDR         |
| subnet      | object | 是  | 子网                |
| memo        | string | 否  | 备注                |

##### subnet

| 参数名称      | 参数类型    | 必选  | 描述        |
|-----------|---------|-----|-----------|
| name      | string  | 是   | 子网名称      |
| ipv4_cidr | string  | 是   | IPv4 CIDR |
| zone      | string  | 是   | 可用区       |

#### aws
| 参数名称              | 参数类型   | 必选 | 描述                        |
|-------------------|--------|----|---------------------------|
| account_id        | string | 是  | 账号ID                      |
| region            | string | 是  | 地域                        |
| name              | string | 是  | 名称                        |
| ipv4_cidr         | string | 是  | IPv4 CIDR                 |
| instance_tenancy  | string | 是  | 租期（枚举值：default、dedicated） |
| memo              | string | 否  | 备注                        |

#### huawei

| 参数名称        | 参数类型   | 必选 | 描述                |
|-------------|--------|----|-------------------|
| account_id  | string | 是  | 账号ID              |
| region      | string | 是  | 地域                |
| name        | string | 是  | 名称                |
| ipv4_cidr   | string | 是  | IPv4 CIDR         |
| subnet      | object | 是  | 子网                |
| memo        | string | 否  | 备注                |

##### subnet
| 参数名称         | 参数类型     | 必选  | 描述        |
|--------------|----------|-----|-----------|
| name         | string   | 是   | 子网名称      |
| ipv4_cidr    | string   | 是   | IPv4 CIDR |
| ipv6_enable  | bool     | 是   | 是否启用IPv6  |
| gateway_ip   | string   | 是   | 网关IP      |

#### azure

| 参数名称                 | 参数类型   | 必选 | 描述                |
|----------------------|--------|----|-------------------|
| account_id           | string | 是  | 账号ID              |
| resource_group_name  | string | 是  | 资源组名称             |
| region               | string | 是  | 地域                |
| name                 | string | 是  | 名称                |
| ipv4_cidr            | string | 是  | IPv4 CIDR         |
| subnet               | object | 是  | 子网                |
| memo                 | string | 否  | 备注                |

##### subnet
| 参数名称      | 参数类型    | 必选  | 描述        |
|-----------|---------|-----|-----------|
| name      | string  | 是   | 子网名称      |
| ipv4_cidr | string  | 是   | IPv4 CIDR |


#### gcp

| 参数名称          | 参数类型   | 必选 | 描述                          |
|---------------|--------|----|-----------------------------|
| account_id    | string | 是  | 账号ID                        |
| region        | string | 是  | 地域                          |
| name          | string | 是  | 名称                          |
| routing_mode  | string | 是  | 动态路由模式（枚举值：REGIONAL、GLOBAL） |
| subnet        | object | 是  | 子网                          |
| memo          | string | 否  | 备注                          |

##### subnet
| 参数名称                     | 参数类型   | 必选  | 描述         |
|--------------------------|--------|-----|------------|
| name                     | string | 是   | 子网名称       |
| ipv4_cidr                | string | 是   | IPv4 CIDR  |
| private_ip_google_access | bool   | 是   | 是否启用专用访问通道 |
| enable_flow_logs         | bool   | 是   | 是否启用流日志    |


### 调用示例

#### tcloud

```json
{
  "account_id": "0000001",
  "region": "ap-hk",
  "name": "xxx",
  "ipv4_cidr": "127.0.0.0/16",
  "subnet": {
    "name": "xxxxx",
    "ipv4_cidr": "127.0.0.0/16",
    "zone": "ap-hk-1"
  },
  "memo": ""
}
```
#### aws

```json
{
  "account_id": "0000001",
  "region": "ap-hk",
  "name": "xxx",
  "ipv4_cidr": "127.0.0.0/16",
  "instance_tenancy": "default",
  "memo": ""
}
```
#### azure

```json
{
  "account_id": "0000001",
  "resource_group_name": "test",
  "region": "ap-hk",
  "name": "xxx",
  "ipv4_cidr": "127.0.0.0/16",
  "subnet": {
    "name": "xxxxx",
    "ipv4_cidr": "127.0.0.0/16"
  },
  "memo": ""
}
```

#### huawei

```json
{
  "account_id": "0000001",
  "region": "ap-hk",
  "name": "xxx",
  "ipv4_cidr": "127.0.0.0/16",
  "subnet": {
    "name": "xxxxx",
    "ipv4_cidr": "127.0.0.0/16",
    "zone": "ap-hk-1",
    "ipv6_enable": false,
    "gateway_ip": "192.168.1.1"
  },
  "memo": ""
}
```

#### gcp

```json
{
  "account_id": "0000001",
  "region": "ap-hk",
  "name": "xxx",
  "routing_mode": "REGIONAL",
  "subnet": {
    "name": "xxxxx",
    "ipv4_cidr": "127.0.0.0/16",
    "private_ip_google_access": false,
    "enable_flow_logs": false
  },
  "memo": ""
}
```





### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000001"
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

| 参数名称 | 参数类型   | 描述   |
|------|--------|------|
| id   | string | 单据ID |