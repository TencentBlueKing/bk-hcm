### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：IaaS资源创建。
- 该接口功能描述：创建用于创建Aws虚拟机的申请。

### URL

POST /api/v1/cloud/vendors/aws/applications/types/create_cvm

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
| cloud_security_group_ids    | string  array | 是  | 云安全组ID                                                                                                               |
| system_disk                 | object        | 是  | 系统盘                                                                                                                  |
| data_disk                   | object  array | 否  | 数据盘                                                                                                                  |
| password                    | string        | 是  | 密码                                                                                                                   |
| confirmed_password          | string        | 是  | 确认密码                                                                                                                 |
| required_count              | int64         | 是  | 需要数量                                                                                                                 |
| memo                        | string        | 否  | 备注                                                                                                                   |

#### system_disk
| 参数名称             | 参数类型    | 必选  | 描述                                          |
|------------------|---------|-----|---------------------------------------------|
| disk_type        | string  | 是   | 云盘类型（枚举值：standard、io1、io2、gp2、sc1、st1、gp3 ） |
| disk_size_gb     | int64   | 是   | 云盘大小                                        |

#### data_disk
| 参数名称         | 参数类型    | 必选  | 描述                                           |
|--------------|---------|-----|----------------------------------------------|
| disk_type    | string  | 是   | 云盘类型（枚举值：standard、io1、io2、gp2、sc1、st1、gp3 ）  |
| disk_size_gb | int64   | 是   | 云盘大小                                         |
| disk_count   | int64   | 是   | 云盘数量                                         |

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
  "cloud_security_group_ids": ["1001","1002"],
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

| 参数名称 | 参数类型   | 描述   |
|------|--------|------|
| id   | string | 单据ID |
