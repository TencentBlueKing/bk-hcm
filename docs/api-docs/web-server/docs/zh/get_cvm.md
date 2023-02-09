### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询虚拟机详情。

### 输入参数

| 参数名称 | 参数类型   | 描述    |
|------|--------|-------|
| id   | string | 虚拟机ID |

### 调用示例

#### 获取详细信息请求参数示例

```json
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000001",
    "cloud_id": "cvm-123",
    "name": "cvm-test",
    "vendor": "tcloud",
    "bk_biz_id": -1,
    "bk_cloud_id": 100,
    "account_id": "0000001",
    "region": "ap-hk",
    "zone": "ap-hk-1",
    "cloud_vpc_ids": [
      "vpc-123"
    ],
    "cloud_subnet_ids": [
      "subnet-123"
    ],
    "cloud_image_id": "image-123",
    "os_name": "linux",
    "memo": "cvm test",
    "status": "init",
    "private_ipv4_addresses": [
      "127.0.0.1"
    ],
    "private_ipv6_addresses": [],
    "public_ipv4_addresses": [
      "127.0.0.2"
    ],
    "public_ipv6_addresses": [],
    "machine_type": "s5",
    "cloud_created_time": "2022-01-20",
    "cloud_launched_time": "2022-01-21",
    "cloud_expired_time": "2022-02-22",
    "creator": "Jim",
    "reviser": "Jim",
    "created_at": "2023-02-12T14:47:39Z",
    "updated_at": "2023-02-12T14:55:40Z"
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

| 参数名称                   | 参数类型         | 描述                                   |
|------------------------|--------------|--------------------------------------|
| id                     | uint64       | 资源ID                                 |
| cloud_id               | string       | 云资源ID                                |
| name                   | string       | 名称                                   |
| vendor                 | string       | 供应商（枚举值：tcloud、aws、azure、gcp、huawei） |
| bk_biz_id              | int64        | 业务ID                                 |
| bk_cloud_id            | int64        | 云区域ID                                |
| account_id             | string       | 账号ID                                 |
| region                 | string       | 地域                                   |
| zone                   | string       | 可用区                                  |
| cloud_vpc_ids          | string array | 云VpcID列表                             |
| cloud_subnet_ids       | string array | 云子网ID列表                              |
| cloud_image_id         | string       | 云镜像ID                                |
| os_name                | string       | 操作系统名称                               |
| memo                   | string       | 备注                                   |
| status                 | string       | 状态                                   |
| private_ipv4_addresses | string array | 内网IPv4地址                             |
| private_ipv6_addresses | string array | 内网IPv6地址                             |
| public_ipv4_addresses  | string array | 公网IPv4地址                             |
| public_ipv6_addresses  | string array | 公网IPv6地址                             |
| machine_type           | string       | 设备类型                                 |
| cloud_created_time     | string       | Cvm在云上创建时间                           |
| cloud_launched_time    | string       | Cvm启动时间                              |
| cloud_expired_time     | string       | Cvm过期时间                              |
| creator                | string       | 创建者                                  |
| reviser                | string       | 修改者                                  |
| created_at             | string       | 创建时间                                 |
| updated_at             | string       | 修改时间                                 |
