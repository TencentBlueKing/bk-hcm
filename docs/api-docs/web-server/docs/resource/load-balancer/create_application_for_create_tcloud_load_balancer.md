### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：负载均衡创建。
- 该接口功能描述：业务下创建负载均衡申请。

### URL

POST /api/v1/cloud/vendors/tcloud/applications/types/create_load_balancer

### 输入参数

#### tcloud

| 参数名称                       | 参数类型         | 必选 | 描述                                                            |
|----------------------------|--------------|----|---------------------------------------------------------------|
| bk_biz_id                  | int64        | 是  | 业务ID                                                          |
| account_id                 | string       | 是  | 账号ID                                                          |
| region                     | string       | 是  | 地域                                                            |
| load_balancer_type         | string       | 是  | 网络类型  公网 OPEN，内网 INTERNAL                                     |
| name                       | string       | 是  | 名称                                                            |
| zones                      | string array | 否  | 主可用区，仅限公网型                                                    |
| backup_zones               | string array | 否  | 备可用区，目前仅广州、上海、南京、北京、中国香港、首尔地域的 IPv4 版本的 CLB 支持主备可用区。          |
| address_ip_version         | string       | 否  | ip版本，IPV4,IPV6(ipv6 nat64),IPv6FullChain(ipv6)                |
| cloud_vpc_id               | string       | 是  | 云VpcID                                                        |
| cloud_subnet_id            | string       | 否  | 云子网ID ，内网型必填                                                  |
| vip                        | string       | 否  | 绑定已有eip的ip地址，，ipv6 nat64 不支持                                  |
| cloud_eip_id               | string       | 否  | 绑定eip id                                                      |
| vip_isp                    | string       | 否  | 运营商类型仅公网，枚举值：CMCC,CUCC,CTCC,BGP。通过TCloudDescribeResource 接口确定 |
| internet_charge_type       | string       | 否  | 网络计费模式                                                        |
| internet_max_bandwidth_out | int64        | 否  | 最大出带宽，单位Mbps                                                  |
| bandwidthpkg_sub_type      | string       | 否  | 带宽包的类型，如SINGLEISP（单线）、BGP（多线）。                                |
| bandwidth_package_id       | string       | 否  | 带宽包id，计费模式为带宽包计费时必填                                           |
| egress                     | string       | 否  | 网络出口                                                          |
| sla_type                   | string       | 否  | 性能容量型规格, 留空为共享型                                               |
| exclusive                  | int          | 否  | 是否独占型：1是、0否，默认0，详见独占型规格说明                                     |
| cloud_cluster_ids          | string array | 否  | 四层（TGW）独占集群的云上ID列表，取自标签聚合查询接口返回的 `cloud_cluster_id`，详见独占型规格说明 |
| cluster_tag                | string       | 否  | 七层独占集群标签，详见独占型规格说明                                    |
| auto_renew                 | boolean      | 否  | 按月付费自动续费                                                      |
| require_count	             | int          | 是  | 购买数量                                                          |
| memo                       | string       | 否  | 备注                                                            |
| remark                     | string       | 否  | 单据备注                                                          |
| load_balancer_pass_to_target | boolean      | 是  | 1次校验-仅校验CLB上的安全组,传入true; 2次校验-同时校验CLB和RS上的安全组, 传入false        |

#### 网络计费模式取值范围：

- `TRAFFIC_POSTPAID_BY_HOUR` 按流量按小时后计费
- `BANDWIDTH_POSTPAID_BY_HOUR` 按带宽按小时后计费
- `BANDWIDTH_PACKAGE` 带宽包计费

#### sla_type 性能容量型规格取值范围：

- `clb.c2.medium` 标准型规格
- `clb.c3.small` 高阶型1规格
- `clb.c3.medium` 高阶型2规格
- `clb.c4.small` 超强型1规格
- `clb.c4.medium` 超强型2规格
- `clb.c4.large` 超强型3规格
- `clb.c4.xlarge` 超强型4规格

#### 独占型规格说明：

独占型仅支持业务视角购买、仅支持公网（`load_balancer_type` 为 `OPEN`），三种规格类型的参数组合如下：

| 规格类型  | exclusive | sla_type    | cluster_tag | cloud_cluster_ids |
|-------|-----------|-------------|-------------|-----------------|
| 共享型    | 0 或不传     | 留空          | 不传          | 不传              |
| 性能容量型 | 0 或不传     | `clb.c*` 档位 | 不传          | 不传              |
| 独占型    | 1         | 必须留空        |  七层独占集群填          | 四层独占集群填，填一个:指定集群；填多个:多个里面随机挑一个 |

参数取值与校验：

- `exclusive` 为 1 时 `cluster_tag` 必填且 `sla_type` 必须留空；`exclusive` 为 0 或不传时 `cluster_tag`、`cloud_cluster_ids` 必须为空。否则返回 `InvalidParameter`。缺少该字段时，「选了独占型但未选标签」与「选共享型」的报文完全相同，服务端无法区分。
- `cluster_tag` 必须属于当前业务已分配的集群，否则返回 `PermissionDenied`。
- `cloud_cluster_ids` 填多个云上集群ID，云上会在这个中间随机挑选，用户在页面选择随机匹配集群时，需要把后端返回的标签对应集群云上ID都填进来；集群ID随机匹配不支持指定`vip`。
- `vip` 指定 `vip` 时 `cloud_cluster_ids` 必填且唯一，且后端创建前会复核该 VIP 仍闲置，否则返回 `InvalidParameter`。
- 计费方式为 `BANDWIDTH_PACKAGE` 时，所选带宽包的网络出口必须与独占集群一致，否则返回 `InvalidParameter`。

### 调用示例

#### tcloud

```json
{
  "account_id": "0000001",
  "region": "ap-hk",
  "zone": "ap-hk-1",
  "backup_zones": [],
  "name": "xxx",
  "load_balancer_type": "INTERNAL",
  "cloud_vpc_id": "vpc-123",
  "cloud_subnet_id": "subnet-123",
  "address_ip_version": "IPV4",
  "vip": "1.2.3.4",
  "vip_isp": "BGP",
  "charge_type": "TRAFFIC_POSTPAID_BY_HOUR",
  "sla_type": "clb.c2.medium",
  "internet_max_bandwidth_out": 10,
  "auto_renew": true,
  "require_count": 1,
  "memo": "",
  "load_balancer_pass_to_target": true
}
```

#### tcloud 独占型（指定四层集群与IP）

```json
{
  "bk_biz_id": 213,
  "account_id": "0000001",
  "region": "ap-guangzhou",
  "zones": [
    "ap-guangzhou-3"
  ],
  "backup_zones": [],
  "name": "xxx",
  "load_balancer_type": "OPEN",
  "cloud_vpc_id": "vpc-123",
  "address_ip_version": "IPV4",
  "vip_isp": "BGP",
  "internet_charge_type": "BANDWIDTH_PACKAGE",
  "bandwidth_package_id": "bwp-1234556",
  "bandwidthpkg_sub_type": "BGP",
  "internet_max_bandwidth_out": 10,
  "cloud_cluster_ids": [
    "tgw-xxxxxxxx"
  ],
  "sla_type":"", // sla_type为“”+exclusive 表示独占型
  "exclusive": 1,
  "vip": "1.1.1.1",
  "require_count": 1,
  "memo": "",
  "load_balancer_pass_to_target": true
}
```

#### tcloud 独占型（同时指定四层集群+指定IP+指定七层集群）

```json
{
  "bk_biz_id": 213,
  "account_id": "0000001",
  "region": "ap-guangzhou",
  "zones": [
    "ap-guangzhou-3"
  ],
  "backup_zones": [],
  "name": "xxx",
  "load_balancer_type": "OPEN",
  "cloud_vpc_id": "vpc-123",
  "address_ip_version": "IPV4",
  "vip_isp": "BGP",
  "internet_charge_type": "BANDWIDTH_PACKAGE",
  "bandwidth_package_id": "bwp-1234556",
  "bandwidthpkg_sub_type": "BGP",
  "internet_max_bandwidth_out": 10,
  "cluster_tag": "cluster_l7", // 七层集群标签
  "cloud_cluster_ids": [
    "tgw-xxxxxxxx"
  ],
  "sla_type":"", // sla_type为“”+exclusive 表示独占型
  "exclusive": 1,
  "vip": "1.1.1.1",
  "require_count": 1,
  "memo": "",
  "load_balancer_pass_to_target": true
}
```

#### tcloud 独占型（同时指定四层集群+随机集群+随机IP）

```json
{
  "bk_biz_id": 213,
  "account_id": "0000001",
  "region": "ap-guangzhou",
  "zones": [
    "ap-guangzhou-3"
  ],
  "backup_zones": [],
  "name": "xxx",
  "load_balancer_type": "OPEN",
  "cloud_vpc_id": "vpc-123",
  "address_ip_version": "IPV4",
  "vip_isp": "BGP",
  "internet_charge_type": "BANDWIDTH_PACKAGE",
  "bandwidth_package_id": "bwp-1234556",
  "bandwidthpkg_sub_type": "BGP",
  "internet_max_bandwidth_out": 10,
  "cluster_tag": "cluster_l7", // 七层集群标签
  "cloud_cluster_ids": [ // 列表是通过四层集群标签过滤出来的
    "tgw-xxxxxxxx",
    "tgw-yyyyyyyy",
    "tgw-zzzzzzzz",
  ],
  "sla_type":"", // sla_type为“”+exclusive 表示独占型
  "exclusive": 1,
  "vip": "",
  "require_count": 1,
  "memo": "",
  "load_balancer_pass_to_target": true
}
```

#### tcloud 独占型（仅指定七层独占集群）

```json
{
  "bk_biz_id": 213,
  "account_id": "0000001",
  "region": "ap-guangzhou",
  "zones": [
    "ap-guangzhou-3"
  ],
  "backup_zones": [],
  "name": "xxx",
  "load_balancer_type": "OPEN",
  "cloud_vpc_id": "vpc-123",
  "address_ip_version": "IPV4",
  "vip_isp": "BGP",
  "internet_charge_type": "BANDWIDTH_PACKAGE",
  "bandwidth_package_id": "bwp-1234556",
  "bandwidthpkg_sub_type": "BGP",
  "internet_max_bandwidth_out": 10,
  "cluster_tag": "cluster_l7", // 七层集群标签
  "cloud_cluster_ids": null,
  "sla_type":"", // sla_type为“”+exclusive 表示独占型
  "exclusive": 1,
  "vip": "",
  "require_count": 1,
  "memo": "",
  "load_balancer_pass_to_target": true
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
