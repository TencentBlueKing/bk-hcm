### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：IaaS资源创建。
- 该接口功能描述：创建用于创建腾讯云虚拟机的申请。

### URL

POST /api/v1/cloud/vendors/tcloud/applications/types/create_cvm

### 输入参数

| 参数名称                        | 参数类型          | 必选 | 描述                                                                                                                   |
|-----------------------------|---------------|----|----------------------------------------------------------------------------------------------------------------------|
| bk_biz_id                   | int64         | 是  | 业务ID                                                                                                                 |
| account_id                  | string        | 是  | 账号ID                                                                                                                 |
| region                      | string        | 是  | 地域                                                                                                                   |
| zone                        | string        | 是  | 可用区                                                                                                                  |
| name                        | string        | 是  | 名称                                                                                                                   |
| instance_type               | string        | 是  | 实例类型                                                                                                                 |
| cloud_image_id              | string        | 是  | 云镜像ID                                                                                                                |
| cloud_vpc_id                | string        | 是  | 云VpcID                                                                                                               |
| cloud_subnet_id             | string        | 是  | 云子网ID                                                                                                                |
| public_ip_assigned          | string        | 否  | 分配的公网IP                                                                                                              |
| internet_max_bandwidth_out  | string        | 否  | 网络带宽，public_ip_assigned=true时必填                                                                                      |
| cloud_security_group_ids    | string  array | 是  | 云安全组ID                                                                                                               |
| system_disk                 | object        | 是  | 系统盘                                                                                                                  |
| data_disk                   | object  array | 否  | 数据盘                                                                                                                  |
| password                    | string        | 是  | 密码                                                                                                                   |
| confirmed_password          | string        | 是  | 确认密码                                                                                                                 |
| instance_charge_type        | string        | 是  | 实例计费模式（PREPAID：表示预付费，即包年包月、POSTPAID_BY_HOUR：表示后付费，即按量计费、CDHPAID：专用宿主机付费，即只对专用宿主机计费，不对专用宿主机上的实例计费。、SPOTPAID：表示竞价实例付费） |
| internet_charge_type        | string        | 否  | (v1.6.7+)网络计费类型，见下文。                                                                                                 |
| bandwidth_package_id        | string        | 否  | (v1.6.7+)带宽包id                                                                                                       |
| instance_charge_paid_period | int64         | 是  | 实例计费支付周期                                                                                                             |
| auto_renew                  | bool          | 是  | 是否自动续订                                                                                                               |
| required_count              | int64         | 是  | 需要数量                                                                                                                 |
| memo                        | string        | 否  | 备注                                                                                                                   |
| remark                      | string        | 否  | 单据备注                                                                                                                 |

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

### 调用示例

```json
{
  "bk_biz_id": 100,
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
