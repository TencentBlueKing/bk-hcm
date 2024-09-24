### 描述

- 该接口提供版本：v1.1.17+。
- 该接口所需权限：IaaS资源创建。
- 该接口功能描述：创建虚拟机。

### URL

POST /api/v1/cloud/cvms/create

### 输入参数

#### tcloud

| 参数名称                        | 参数类型          | 必选 | 描述                                                                                                                   |
|-----------------------------|---------------|----|----------------------------------------------------------------------------------------------------------------------|
| account_id                  | string        | 是  | 账号ID                                                                                                                 |
| region                      | string        | 是  | 地域                                                                                                                   |
| zone                        | string        | 是  | 可用区                                                                                                                  |
| name                        | string        | 是  | 名称                                                                                                                   |
| instance_type               | string        | 是  | 实例类型                                                                                                                 |
| cloud_image_id              | string        | 是  | 云镜像ID                                                                                                                |
| cloud_vpc_id                | string        | 是  | 云VpcID                                                                                                               |
| cloud_subnet_id             | string        | 是  | 云子网ID                                                                                                                |
| public_ip_assigned          | string        | 否  | 分配的公网IP                                                                                                              |
| cloud_security_group_ids    | string  array | 是  | 云安全组ID                                                                                                               |
| system_disk                 | object        | 是  | 系统盘                                                                                                                  |
| data_disk                   | object  array | 否  | 数据盘                                                                                                                  |
| password                    | string        | 是  | 密码                                                                                                                   |
| confirmed_password          | string        | 是  | 确认密码                                                                                                                 |
| instance_charge_type        | string        | 是  | 实例计费模式（PREPAID：表示预付费，即包年包月、POSTPAID_BY_HOUR：表示后付费，即按量计费、CDHPAID：专用宿主机付费，即只对专用宿主机计费，不对专用宿主机上的实例计费。、SPOTPAID：表示竞价实例付费） |
| instance_charge_paid_period | int64         | 是  | 实例计费支付周期                                                                                                             |
| internet_charge_type        | string        | 否  | (v1.6.7+) 网络计费类型，见下文。                                                                                                |
| bandwidth_package_id        | string        | 否  | (v1.6.7+) 带宽包id                                                                                                      |
| auto_renew                  | bool          | 是  | 是否自动续订                                                                                                               |
| required_count              | int64         | 是  | 需要数量                                                                                                                 |
| memo                        | string        | 否  | 备注                                                                                                                   |

#### internet_charge_type 网络计费类型 取值范围

- BANDWIDTH_PREPAID：预付费按带宽结算
- TRAFFIC_POSTPAID_BY_HOUR：流量按小时后付费
- BANDWIDTH_POSTPAID_BY_HOUR：带宽按小时后付费
- BANDWIDTH_PACKAGE：带宽包用户

##### internet_charge_type 默认取值：

非带宽包用户默认与子机付费类型保持一致，比如子机付费类型为预付费，网络计费类型默认为预付费；子机付费类型为后付费，网络计费类型默认为后付费。

#### system_disk

| 参数名称         | 参数类型   | 必选 | 描述                                                                             |
|--------------|--------|----|--------------------------------------------------------------------------------|
| disk_type    | string | 是  | 云盘类型（枚举值：LOCAL_BASIC、LOCAL_SSD、CLOUD_BASIC、CLOUD_SSD、CLOUD_PREMIUM、CLOUD_BSSD） |
| disk_size_gb | int64  | 是  | 云盘大小                                                                           |

#### data_disk

| 参数名称         | 参数类型   | 必选 | 描述                                                                             |
|--------------|--------|----|--------------------------------------------------------------------------------|
| disk_type    | string | 是  | 云盘类型（枚举值：LOCAL_BASIC、LOCAL_SSD、CLOUD_BASIC、CLOUD_SSD、CLOUD_PREMIUM、CLOUD_BSSD） |
| disk_size_gb | int64  | 是  | 云盘大小                                                                           |
| disk_count   | int64  | 是  | 云盘数量                                                                           |

#### aws

| 参数名称                     | 参数类型          | 必选 | 描述      |
|--------------------------|---------------|----|---------|
| account_id               | string        | 是  | 账号ID    |
| region                   | string        | 是  | 地域      |
| zone                     | string        | 是  | 可用区     |
| name                     | string        | 是  | 名称      |
| instance_type            | string        | 是  | 实例类型    |
| cloud_image_id           | string        | 是  | 云镜像ID   |
| cloud_vpc_id             | string        | 是  | 云VpcID  |
| cloud_subnet_id          | string        | 是  | 云子网ID   |
| public_ip_assigned       | string        | 否  | 分配的公网IP |
| cloud_security_group_ids | string  array | 是  | 云安全组ID  |
| system_disk              | object        | 是  | 系统盘     |
| data_disk                | object  array | 否  | 数据盘     |
| password                 | string        | 是  | 密码      |
| confirmed_password       | string        | 是  | 确认密码    |
| required_count           | int64         | 是  | 需要数量    |
| memo                     | string        | 否  | 备注      |

#### system_disk

| 参数名称         | 参数类型   | 必选 | 描述                                          |
|--------------|--------|----|---------------------------------------------|
| disk_type    | string | 是  | 云盘类型（枚举值：standard、io1、io2、gp2、sc1、st1、gp3 ） |
| disk_size_gb | int64  | 是  | 云盘大小                                        |

#### data_disk

| 参数名称         | 参数类型   | 必选 | 描述                                          |
|--------------|--------|----|---------------------------------------------|
| disk_type    | string | 是  | 云盘类型（枚举值：standard、io1、io2、gp2、sc1、st1、gp3 ） |
| disk_size_gb | int64  | 是  | 云盘大小                                        |
| disk_count   | int64  | 是  | 云盘数量                                        |

#### huawei

| 参数名称                        | 参数类型          | 必选 | 描述                                                                                                                   |
|-----------------------------|---------------|----|----------------------------------------------------------------------------------------------------------------------|
| account_id                  | string        | 是  | 账号ID                                                                                                                 |
| region                      | string        | 是  | 地域                                                                                                                   |
| zone                        | string        | 是  | 可用区                                                                                                                  |
| name                        | string        | 是  | 名称                                                                                                                   |
| instance_type               | string        | 是  | 实例类型                                                                                                                 |
| cloud_image_id              | string        | 是  | 云镜像ID                                                                                                                |
| cloud_vpc_id                | string        | 是  | 云VpcID                                                                                                               |
| cloud_subnet_id             | string        | 是  | 云子网ID                                                                                                                |
| public_ip_assigned          | string        | 否  | 分配的公网IP                                                                                                              |
| cloud_security_group_ids    | string  array | 是  | 云安全组ID                                                                                                               |
| system_disk                 | object        | 是  | 系统盘                                                                                                                  |
| data_disk                   | object  array | 否  | 数据盘                                                                                                                  |
| password                    | string        | 是  | 密码                                                                                                                   |
| confirmed_password          | string        | 是  | 确认密码                                                                                                                 |
| instance_charge_type        | string        | 是  | 实例计费模式（PREPAID：表示预付费，即包年包月、POSTPAID_BY_HOUR：表示后付费，即按量计费、CDHPAID：专用宿主机付费，即只对专用宿主机计费，不对专用宿主机上的实例计费。、SPOTPAID：表示竞价实例付费） |
| instance_charge_paid_period | int64         | 是  | 实例计费支付周期                                                                                                             |
| auto_renew                  | bool          | 是  | 是否自动续订                                                                                                               |
| required_count              | int64         | 是  | 需要数量                                                                                                                 |
| memo                        | string        | 否  | 备注                                                                                                                   |

#### system_disk

| 参数名称         | 参数类型   | 必选 | 描述                                |
|--------------|--------|----|-----------------------------------|
| disk_type    | string | 是  | 云盘类型（枚举值：SATA、SAS、GPSSD、SSD、ESSD） |
| disk_size_gb | int64  | 是  | 云盘大小                              |

#### data_disk

| 参数名称         | 参数类型   | 必选 | 描述                                |
|--------------|--------|----|-----------------------------------|
| disk_type    | string | 是  | 云盘类型（枚举值：SATA、SAS、GPSSD、SSD、ESSD） |
| disk_size_gb | int64  | 是  | 云盘大小                              |
| disk_count   | int64  | 是  | 云盘数量                              |

#### azure

| 参数名称                     | 参数类型          | 必选 | 描述      |
|--------------------------|---------------|----|---------|
| account_id               | string        | 是  | 账号ID    |
| resource_group_name      | string        | 是  | 资源组名称   |
| region                   | string        | 是  | 地域      |
| zone                     | string        | 是  | 可用区     |
| name                     | string        | 是  | 名称      |
| instance_type            | string        | 是  | 实例类型    |
| cloud_image_id           | string        | 是  | 云镜像ID   |
| cloud_vpc_id             | string        | 是  | 云VpcID  |
| cloud_subnet_id          | string        | 是  | 云子网ID   |
| public_ip_assigned       | string        | 否  | 分配的公网IP |
| cloud_security_group_ids | string  array | 是  | 云安全组ID  |
| system_disk              | object        | 是  | 系统盘     |
| data_disk                | object  array | 否  | 数据盘     |
| username                 | string        | 是  | 用户名     |
| password                 | string        | 是  | 密码      |
| confirmed_password       | string        | 是  | 确认密码    |
| required_count           | int64         | 是  | 需要数量    |
| memo                     | string        | 否  | 备注      |

#### system_disk

| 参数名称         | 参数类型   | 必选 | 描述                                                                                           |
|--------------|--------|----|----------------------------------------------------------------------------------------------|
| disk_type    | string | 是  | 云盘类型（枚举值：Premium_LRS、PremiumV2_LRS、Premium_ZRS、Standard_LRS、StandardSSD_LRS、StandardSSD_ZRS） |
| disk_size_gb | int64  | 是  | 云盘大小                                                                                         |

#### data_disk

| 参数名称         | 参数类型   | 必选 | 描述                                                                                                        |
|--------------|--------|----|-----------------------------------------------------------------------------------------------------------|
| disk_type    | string | 是  | 云盘类型（枚举值：Premium_LRS、PremiumV2_LRS、Premium_ZRS、Standard_LRS、StandardSSD_LRS、StandardSSD_ZRS、UltraSSD_LRS） |
| disk_size_gb | int64  | 是  | 云盘大小                                                                                                      |
| disk_count   | int64  | 是  | 云盘数量                                                                                                      |

#### gcp

| 参数名称            | 参数类型          | 必选 | 描述     |
|-----------------|---------------|----|--------|
| account_id      | string        | 是  | 账号ID   |
| region          | string        | 是  | 地域     |
| zone            | string        | 是  | 可用区    |
| name            | string        | 是  | 名称     |
| instance_type   | string        | 是  | 实例类型   |
| cloud_image_id  | string        | 是  | 云镜像ID  |
| cloud_vpc_id    | string        | 是  | 云VpcID |
| cloud_subnet_id | string        | 是  | 云子网ID  |
| system_disk     | object        | 是  | 系统盘    |
| data_disk       | object  array | 否  | 数据盘    |
| password        | string        | 是  | 密码     |
| required_count  | int64         | 是  | 需要数量   |
| memo            | string        | 否  | 备注     |

#### system_disk

| 参数名称         | 参数类型   | 必选 | 描述                                                  |
|--------------|--------|----|-----------------------------------------------------|
| disk_type    | string | 是  | 云盘类型（枚举值：pd-standard、pd-balanced、pd-ssd、pd-extreme） |
| disk_size_gb | int64  | 是  | 云盘大小                                                |

#### data_disk

| 参数名称         | 参数类型   | 必选 | 描述                                                  |
|--------------|--------|----|-----------------------------------------------------|
| disk_type    | string | 是  | 云盘类型（枚举值：pd-standard、pd-balanced、pd-ssd、pd-extreme） |
| disk_size_gb | int64  | 是  | 云盘大小                                                |
| disk_count   | int64  | 是  | 云盘数量                                                |
| mode         | string | 是  | 模式（枚举值：READ_ONLY、READ_WRITE）                        |
| auto_delete  | bool   | 是  | 是否自动删除                                              |

### 调用示例

#### tcloud

```json
{
  "account_id": "0000001",
  "region": "ap-hk",
  "zone": "ap-hk-1",
  "name": "xxx",
  "instance_type": "cvm",
  "cloud_image_id": "image-123",
  "cloud_vpc_id": "vpc-123",
  "cloud_subnet_id": "subnet-123",
  "cloud_security_group_ids": [
    "1001",
    "1002"
  ],
  "system_disk": {
    "disk_type": "LOCAL_BASIC",
    "disk_size_gb": 50
  },
  "data_disk": {
    "disk_type": "LOCAL_BASIC",
    "disk_size_gb": 50,
    "disk_count": 1
  },
  "password": "xxxxxx",
  "confirmed_password": "xxxxxx",
  "instance_charge_type": "PREPAID",
  "instance_charge_paid_period": 1,
  "auto_renew": true,
  "required_count": 1,
  "memo": ""
}
```

#### huawei

```json
{
  "account_id": "0000001",
  "region": "ap-hk",
  "zone": "ap-hk-1",
  "name": "xxx",
  "instance_type": "cvm",
  "cloud_image_id": "image-123",
  "cloud_vpc_id": "vpc-123",
  "cloud_subnet_id": "subnet-123",
  "cloud_security_group_ids": [
    "1001",
    "1002"
  ],
  "system_disk": {
    "disk_type": "SATA",
    "disk_size_gb": 50
  },
  "data_disk": {
    "disk_type": "SATA",
    "disk_size_gb": 50,
    "disk_count": 1
  },
  "password": "xxxxxx",
  "confirmed_password": "xxxxxx",
  "instance_charge_type": "PREPAID",
  "instance_charge_paid_period": 1,
  "auto_renew": true,
  "required_count": 1,
  "memo": ""
}
```

#### gcp

```json
{
  "account_id": "0000001",
  "region": "ap-hk",
  "zone": "ap-hk-1",
  "name": "xxx",
  "instance_type": "cvm",
  "cloud_image_id": "image-123",
  "cloud_vpc_id": "vpc-123",
  "cloud_subnet_id": "subnet-123",
  "system_disk": {
    "disk_type": "pd-standard",
    "disk_size_gb": 50
  },
  "data_disk": {
    "disk_type": "pd-standard",
    "disk_size_gb": 50,
    "disk_count": 1,
    "mode": "READ_WRITE",
    "auto_delete": true
  },
  "password": "xxxxxx",
  "required_count": 1,
  "memo": ""
}
```

#### azure

```json
{
  "account_id": "0000001",
  "resource_group_name": "test",
  "region": "ap-hk",
  "zone": "ap-hk-1",
  "name": "xxx",
  "instance_type": "cvm",
  "cloud_image_id": "image-123",
  "cloud_vpc_id": "vpc-123",
  "cloud_subnet_id": "subnet-123",
  "cloud_security_group_ids": [
    "1001",
    "1002"
  ],
  "system_disk": {
    "disk_type": "Premium_LRS",
    "disk_size_gb": 50
  },
  "data_disk": {
    "disk_type": "Premium_LRS",
    "disk_size_gb": 50,
    "disk_count": 1
  },
  "username": "xxxxxx",
  "password": "xxxxxx",
  "confirmed_password": "xxxxxx",
  "required_count": 1,
  "memo": ""
}
```

#### aws

```json
{
  "account_id": "0000001",
  "region": "ap-hk",
  "zone": "ap-hk-1",
  "name": "xxx",
  "instance_type": "cvm",
  "cloud_image_id": "image-123",
  "cloud_vpc_id": "vpc-123",
  "cloud_subnet_id": "subnet-123",
  "cloud_security_group_ids": [
    "1001",
    "1002"
  ],
  "system_disk": {
    "disk_type": "standard",
    "disk_size_gb": 50
  },
  "data_disk": {
    "disk_type": "standard",
    "disk_size_gb": 50,
    "disk_count": 1
  },
  "password": "xxxxxx",
  "confirmed_password": "xxxxxx",
  "required_count": 1,
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

| 参数名称              | 参数类型   | 描述           |
|-------------------|--------|--------------|
| unknown_cloud_ids | string | 未知创建状态的云主机ID |
| success_cloud_ids | string | 成功创建的云主机ID   |
| failed_cloud_ids  | string | 创建失败的云主机ID   |
| failed_message    | string | 失败原因         |
