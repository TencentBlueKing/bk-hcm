### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询负载均衡详情。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/{id}

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述     |
|-----------|--------|----|--------|
| bk_biz_id | string | 是  | 业务id   |
| id        | string | 是  | 负载均衡id |

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000001",
    "cloud_id": "lb-asdfefe",
    "name": "test",
    "vendor": "tcloud",
    "account_id": "00000001",
    "bk_biz_id": 1234,
    "ip_version": "ipv4",
    "lb_type": "OPEN",
    "region": "ap-guangzhou",
    "zones": [
      "ap-guangzhou-1"
    ],
    "backup_zones": [],
    "vpc_id": "00000001",
    "cloud_vpc_id": "vpc-abcdef",
    "subnet_id": "",
    "cloud_subnet_id": "",
    "private_ipv4_addresses": [],
    "private_ipv6_addresses": [],
    "public_ipv4_addresses": [
      "1.1.1.1"
    ],
    "public_ipv6_addresses": [],
    "domain": "",
    "status": "1",
    "cloud_created_time": "2024-01-02 15:04:05",
    "cloud_status_time": "2024-01-02 15:04:05",
    "cloud_expired_time": "",
    "memo": null,
    "creator": "admin",
    "reviser": "admin",
    "created_at": "2024-01-02T15:04:05Z",
    "updated_at": "2024-01-02T15:04:05Z",
    "extension": {
      "vip_isp": "BGP"
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

| 参数名称                   | 参数类型         | 描述                                   |
|------------------------|--------------|--------------------------------------|
| id                     | string       | 资源ID                                 |
| cloud_id               | string       | 云资源ID                                |
| name                   | string       | 名称                                   |
| vendor                 | string       | 供应商（枚举值：tcloud、aws、azure、gcp、huawei） |
| bk_biz_id              | int64        | 业务ID                                 |
| account_id             | string       | 账号ID                                 |
| region                 | string       | 地域                                   |
| zones                  | string       | 主可用区                                 |
| backup_zones           | string       | 备可用区                                 |
| cloud_vpc_id           | string       | 云vpcID                               |
| vpc_id                 | string       | vpcID                                |
| lb_type                | string       | 负载均衡类型                               |
| ip_version             | string       | 负载均衡网络版本                             |
| memo                   | string       | 备注                                   |
| status                 | string       | 状态                                   |
| domain                 | string       | 域名                                   |
| private_ipv4_addresses | string array | 内网ipv4地址                             |
| private_ipv6_addresses | string array | 内网ipv6地址                             |
| public_ipv4_addresses  | string array | 外网ipv4地址                             |
| public_ipv6_addresses  | string array | 外网ipv6地址                             |
| cloud_created_time     | string       | lb在云上创建时间，标准格式：2006-01-02T15:04:05Z  |
| cloud_status_time      | string       | lb状态变更时间，标准格式：2006-01-02T15:04:05Z   |
| cloud_expired_time     | string       | lb过期时间，标准格式：2006-01-02T15:04:05Z     |
| extension              | object       | 拓展                                   |
| creator                | string       | 创建者                                  |
| reviser                | string       | 修改者                                  |
| created_at             | string       | 创建时间，标准格式：2006-01-02T15:04:05Z       |
| updated_at             | string       | 修改时间，标准格式：2006-01-02T15:04:05Z       |

##### TCloud status 状态含义：

| 状态值 | 含义   |
|-----|------|
| 0   | 创建中  |
| 1   | 正常运行 |

#### data.extension[tcloud]

腾讯云拓展字段

| 参数名称                         | 参数类型   | 描述                                          |
|------------------------------|--------|---------------------------------------------|
| sla_type                     | string | 性能容量型规格。                                    |
| vip_isp                      | string | 运营商类型。                                      |
| load_balancer_pass_to_target | string | Target是否放通来自CLB的流量。                         |
| internet_max_bandwidth_out   | string | 最大出带宽，单位Mbps，                               |
| internet_charge_type         | string | 计费模式                                        |
| bandwidthpkg_sub_type        | string | 带宽包的类型                                      |
| bandwidth_package_id         | string | 带宽包ID                                       |
| ipv6_mode                    | string | IP地址版本为ipv6时此字段有意义， IPv6Nat64/IPv6FullChain |
| snat                         | string | snat                                        |
| snat_pro                     | string | 是否开启SnatPro。                                |
| snat_ips                     | string | 开启SnatPro负载均衡后，SnatIp列表。                    |
| target_region                | string | 开启跨域1.0后，返回目标地域信息                           |
| target_vpc                   | string | 开启跨域1.0后，返回目标VPC云上ID，返回0表示基础网络              |
| delete_protect               | string | 删除保护                                        |
| egress                       | string | 网络出口                                        |
| mix_ip_target                | string | 双栈混绑                                        |

