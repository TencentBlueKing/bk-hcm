### 描述

- 该接口提供版本：v1.1.18+。
- 该接口所需权限：资源-IaaS资源创建。
- 该接口功能描述：创建 eip。

### URL

POST /api/v1/cloud/eips/create

### 输入参数

#### TCloud

| 参数名称              | 参数类型   | 必选  | 描述             |
|-------------------|--------|-----|----------------|
| account_id        | string | 是   | 账号ID           |
| region            | string | 是   | 地域             |
| eip_name          | string | 是   | Eip 名称         |
| eip_count         | string | 是   | Eip 数量         |
| service_provider  | string | 是   | 服务提供者（枚举值：BGP） |
| address_type      | string | 是   | 地址类型（枚举值：EIP）  |

#### Aws

| 参数名称                 | 参数类型   | 必选 | 描述      |
|----------------------|--------|----|---------|
| account_id           | string | 是  | 账号ID    |
| region               | string | 是  | 地域      |
| public_ipv4_pool     | string | 否  | 公共ipv4池 |
| network_border_group | string | 是  | 网络边界组   |

#### HuaWei

| 参数名称                    | 参数类型   | 必选 | 描述                           |
|-------------------------|--------|----|------------------------------|
| account_id              | string | 是  | 账号ID                         |
| region                  | string | 是  | 地域                           |
| eip_name                | string | 是  | Eip 名称                       |
| eip_type                | string | 是  | Eip 类别（枚举值：5_bgp、5_sbgp）     |
| eip_count               | string | 是  | Eip 数量                       |
| internet_charge_type    | string | 是  | Eip 类别（枚举值：prePaid、postPaid） |
| internet_charge_prepaid | object | 否  | 网费预付                         |
| bandwidth_option        | object | 是  | 带宽选项                         |

##### internet_charge_prepaid
| 参数名称             | 参数类型   | 必选 | 描述      |
|------------------|--------|----|---------|
| period_num       | int32  | 否  | 期间编号    |
| period_type      | string | 否  | 期间类型    |
| is_auto_renew    | bool   | 否  | 是否自动刷新  |

##### bandwidth_option
| 参数名称        | 参数类型   | 必选  | 描述   |
|-------------|--------|-----|------|
| share_type  | string | 是   | 共享类型 |
| charge_mode | string | 是   | 充电模式 |
| name        | string | 否   | 名称   |
| id          | string | 否   | ID   |
| size        | int32  | 否   | 大小   |

#### Gcp

| 参数名称         | 参数类型   | 必选  | 描述                         |
|--------------|--------|-----|----------------------------|
| account_id   | string | 是   | 账号ID                       |
| eip_name     | string | 是   | Eip 名称                     |
| region       | string | 是   | 地域                         |
| network_tier | string | 是   | 网络等级（枚举值：PREMIUM、STANDARD） |
| ip_version   | string | 是   | IP版本                       |

#### Gcp

| 参数名称         | 参数类型   | 必选  | 描述                         |
|--------------|--------|-----|----------------------------|
| account_id   | string | 是   | 账号ID                       |
| eip_name     | string | 是   | Eip 名称                     |
| region       | string | 是   | 地域                         |
| network_tier | string | 是   | 网络等级（枚举值：PREMIUM、STANDARD） |
| ip_version   | string | 是   | IP版本                       |

#### Azure

| 参数名称                     | 参数类型   | 必选  | 描述                          |
|--------------------------|--------|-----|-----------------------------|
| account_id               | string | 是   | 账号ID                        |
| resource_group_name      | string | 是   | 资源组名称                       |
| eip_name                 | string | 是   | Eip 名称                      |
| region                   | string | 是   | 地域                          |
| zone                     | string | 否   | 区域                          |
| sku_name                 | string | 是   | Sku 名称（枚举值：Standard、Basic）  |
| sku_tier                 | string | 是   | Sku 等级（枚举值：Regional、Global） |
| allocation_method        | string | 是   | 分配方法（枚举值：Dynamic、Static）    |
| ip_version               | string | 是   | IP版本（枚举值：ipv6、ipv4）         |
| idle_timeout_in_minutes  | int32  | 是   | 以分钟为单位的闲置超时                 |

### TCloud调用示例

```json
{
  "account_id": "00000001",
  "memo": "test subnet",
  "region": "ap-guangzhou",
  "eip_name": "test-eip",
  "eip_count": "0",
  "zone": "ap-guangzhou-6",
  "service_provider": "BGP",
  "address_type": "EIP"
}
```

### Aws调用示例

```json
{
  "account_id": "00000001",
  "region": "us-east-1",
  "zone": "us-east-1a",
  "network_border_group": "127.0.0.0/16"
}
```

### HuaWei调用示例

```json
{
  "account_id": "00000001",
  "region": "ap-southeast-1",
  "eip_name": "test-eip",
  "eip_type": "5_bgp",
  "eip_count": "0",
  "internet_charge_type": "prePaid",
  "bandwidth_option": {
    "share_type": "xxx",
    "charge_mode": "xxx"
  }
}
```

### Gcp调用示例

```json
{
  "account_id": "00000001",
  "eip_name": "test-eip",
  "region": "https://www.googleapis.com/compute/v1/projects/xxx/regions/us-west1",
  "network_tier": "PREMIUM",
  "ip_version": "ipv4"
}
```

### Azure调用示例

```json
{
  "account_id": "00000001",
  "resource_group_name": "test",
  "eip_name": "test-eip",
  "region": "us-east-1",
  "zone": "us-east-1a",
  "sku_name": "Standard",
  "sku_tier": "Regional",
  "allocation_method": "Dynamic",
  "ip_version": "ipv4",
  " idle_timeout_in_minutes": 1
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "ids": ["00000003"]
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

| 参数名称  | 参数类型          | 描述           |
|-------|---------------|--------------|
| ids   | string array  | 创建的eip ID列表  |
