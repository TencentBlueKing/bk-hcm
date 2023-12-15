### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：x。
- 该接口功能描述：业务下创建负载均衡。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/clbs/create

### 输入参数

#### tcloud

| 参数名称                       | 参数类型         | 必选 | 描述                                                                                                      |
|----------------------------|--------------|----|---------------------------------------------------------------------------------------------------------|
| bk_biz_id                  | int64        | 是  | 业务ID                                                                                                    |
| account_id                 | string       | 是  | 账号ID                                                                                                    |
| region                     | string       | 是  | 地域                                                                                                      |
| load_balance_type          | string       | 是  | 网络类型  公网 OPEN，内网 INTERNAL                                                                               |
| name                       | string       | 是  | 名称                                                                                                      |
| zone                       | string       | 是  | 主可用区                                                                                                    |
| backup_zones               | string array | 否  | 备可用区                                                                                                    |
| address_ip_version         | string       | 否  | ip版本，IPV4,IPV6(ipv6 nat64),IPv6FullChain(ipv6)                                                          |
| cloud_vpc_id               | string       | 否  | 云VpcID                                                                                                  |
| cloud_subnet_id            | string       | 否  | 云子网ID ，内网型必填                                                                                            |
| vip                        | string       | 否  | 指定vip，ipv6 nat64 不支持                                                                                    |
| vip_isp                    | string       | 否  | 仅公网                                                                                                     |
| charge_type                | string       | 否  | 计费模式 TRAFFIC_POSTPAID_BY_HOUR 按流量按小时后计费 ; BANDWIDTH_POSTPAID_BY_HOUR 按带宽按小时后计费; BANDWIDTH_PACKAGE 带宽包计费 |
| internet_max_bandwidth_out | int64        | 否  | 最大出带宽，单位Mbps                                                                                            |
| sla_type                   | string       | 否  | 性能容量型规格, 留空为共享型                                                                                         |
| auto_renew                 | boolean      | 否  | 按月付费是自动续费                                                                                               |
| require_count	             | int          | 是  | 购买数量                                                                                                    |
| memo                       | string       | 否  | 备注                                                                                                      |

### 调用示例

#### tcloud

```json
{
  "account_id": "0000001",
  "region": "ap-hk",
  "zone": "ap-hk-1",
  "backup_zones": [],
  "name": "xxx",
  "load_balance_type": "INTERNAL",
  "cloud_vpc_id": "vpc-123",
  "cloud_subnet_id": "subnet-123",
  "address_ip_version": "IPV4",
  "vip": "1.2.3.4",
  "vip_isp": "BGP",
  "charge_type": "TRAFFIC_POSTPAID_BY_HOUR",
  "sla_type": "clb.c2.medium",
  "internet_max_bandwidth_out": 10,
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

| 参数名称              | 参数类型   | 描述            |
|-------------------|--------|---------------|
| unknown_cloud_ids | string | 未知创建状态的clb id |
| success_cloud_ids | string | 成功创建的clb id   |
| failed_cloud_ids  | string | 创建失败的clb id   |
| failed_message    | string | 失败原因          |
