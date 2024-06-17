### 描述

- 该接口提供版本：v9.9.9+。
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
| sla_type                   | string       | 否  | 性能容量型规格, 留空为共享型                                               |
| auto_renew                 | boolean      | 否  | 按月付费自动续费                                                      |
| require_count	             | int          | 是  | 购买数量                                                          |
| memo                       | string       | 否  | 备注                                                            |
| remark                     | string       | 否  | 单据备注                                                          |

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

| 参数名称 | 参数类型   | 描述   |
|------|--------|------|
| id   | string | 单据ID |
