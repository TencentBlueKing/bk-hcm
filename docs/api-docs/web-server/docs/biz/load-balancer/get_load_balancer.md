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
    "sync_time": "2025-06-30T09:03:06Z",
    "band_width": 0,
    "isp": "BGP",
    "memo": null,
    "creator": "admin",
    "reviser": "admin",
    "created_at": "2024-01-02T15:04:05Z",
    "updated_at": "2024-01-02T15:04:05Z",
    "extension": {
      "vip_isp": "BGP",
      "exclusive": 1,
      "clusters": [
        {
          "cloud_cluster_id": "tgw-xxxxxxxx",
          "cluster_id": "00000001",
          "cluster_name": "cluster-l4-1",
          "cluster_tag": "cluster_l4",
          "cluster_type": "TGW"
        },
        {
          "cloud_cluster_id": "stgw-xxxxxxxx",
          "cluster_id": "00000002",
          "cluster_name": "cluster-l7-1",
          "cluster_tag": "cluster_l7",
          "cluster_type": "STGW"
        }
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
| band_width             | int          | 带宽，单位Mbps                            |
| isp                    | string       | 运营商类型（枚举值：BGP、CMCC、CUCC、CTCC）        |
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
| sync_time              | string       | 数据同步时间，标准格式：2006-01-02T15:04:05Z     |
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
| load_balancer_pass_to_target | string | Target是否放通来自CLB的流量。                         |
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
| exclusive                    | int    | 是否独占型实例：1是、0否                     |
| clusters                     | array  | 独占集群信息列表，非独占型实例为空数组                       |

说明：

- 详情页「实例规格」由 `exclusive` 与 `sla_type` 合成，后端只返回原始值：`exclusive` 为 1 展示「独占型」；否则 `sla_type` 非空展示对应档位名；否则展示「共享型」。
- `exclusive`、`clusters` 均为云上同步回来的结果，可能与提单时的入参不完全一致（如七层标签落到共享集群）。
- 申请单选择「随机分配」时单据 `content` 内没有实际集群与 IP，需从本接口取实际落地值。

##### data.extension.clusters[n]

负载均衡关联的独占集群列表。集群名称与本地ID由服务端用云上集群ID关联本地独占集群表得到，前端无需再查集群接口。通过 `cluster_type` 区分四层（TGW）与七层（STGW）。

| 参数名称             | 参数类型   | 描述                          |
|------------------|--------|-----------------------------|
| cloud_cluster_id | string | 集群云上ID，如tgw-xxxxxxxx        |
| cluster_id       | string | 集群本地ID，本地表未同步到该集群时为空字符串     |
| cluster_name     | string | 集群名称，本地表未同步到该集群时为空字符串       |
| cluster_tag      | string | 集群标签                        |
| cluster_type     | string | 集群类型（枚举值：TGW、STGW）          |

说明：

- 非独占型实例（`exclusive` 为 0）`clusters` 为空数组。
- 仅使用四层或七层独占集群时，数组中只返回对应类型的元素。
- 七层集群由云侧调度决定，若云上未返回具体的 STGW 集群ID，则该元素仅 `cluster_tag`、`cluster_type` 有值，其余字段为空字符串。
- 集群在本地表被删除或尚未同步时，`cluster_id`、`cluster_name` 为空字符串，`cloud_cluster_id` 仍返回。

