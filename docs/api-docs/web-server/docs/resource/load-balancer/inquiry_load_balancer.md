### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：负载均衡创建。
- 该接口功能描述：查询负载均衡价格。

### URL

POST /api/v1/cloud/load_balancer/prices/inquiry

### 输入参数

#### tcloud

| 参数名称                       | 参数类型         | 必选 | 描述                                                            |
|----------------------------|--------------|----|---------------------------------------------------------------|
| account_id                 | string       | 是  | 账号ID                                                          |
| region                     | string       | 是  | 地域                                                            |
| load_balancer_type         | string       | 是  | 网络类型  公网 OPEN，内网 INTERNAL                                     |
| name                       | string       | 是  | 名称                                                            |
| zones                      | string array | 否  | 主可用区,，仅限公网型                                                   |
| backup_zones               | string array | 否  | 备可用区，目前仅广州、上海、南京、北京、中国香港、首尔地域的 IPv4 版本的 CLB 支持主备可用区。          |
| address_ip_version         | string       | 否  | ip版本，IPV4,IPV6(ipv6 nat64),IPv6FullChain(ipv6)                |
| cloud_vpc_id               | string       | 是  | 云VpcID                                                        |
| cloud_subnet_id            | string       | 否  | 云子网ID ，内网型必填                                                  |
| vip                        | string       | 否  | 绑定已有eip的ip地址，，ipv6 nat64 不支持                                  |
| cloud_eip_id               | string       | 否  | 绑定eip id                                                      |
| vip_isp                    | string       | 否  | 运营商类型仅公网，枚举值：CMCC,CUCC,CTCC,BGP。通过TCloudDescribeResource 接口确定 |
| internet_charge_type       | string       | 否  | 网络计费模式                                                        |
| internet_max_bandwidth_out | int64        | 否  | 最大出带宽，单位Mbps                                                  |
| bandwidth_package_id       | string       | 否  | 带宽包id，计费模式为带宽包计费时必填                                           |
| sla_type                   | string       | 否  | 性能容量型规格, 留空为共享型                                               |
| require_count              | int          | 是  | 购买数量                                                          |
| memo                       | string       | 否  | 备注                                                            |

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
    "bandwidth_price": null,
    "instance_price": {
      "charge_unit": "HOUR",
      "discount": 1.2,
      "discount_price": null,
      "original_price": null,
      "unit_price": 3.4,
      "unit_price_discount": 5.6
    },
    "lcu_price": null
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data[tcloud]

| 参数名称            | 参数类型       | 描述                            |
|-----------------|------------|-------------------------------|
| bandwidth_price | price_item | 网络价格信息，对于标准账户，网络在cvm上计费，该选项为空 |
| instance_price  | price_item | 实例价格信息                        |
| lcu_price       | price_item | lcu 价格信息                      |

#### price_item

| 参数名称                | 参数类型   | 描述             |
|---------------------|--------|----------------|
| charge_unit         | string | 后续计价单元，HOUR、GB |
| discount            | float  | 折扣 ，如20.0代表2折  |
| discount_price      | float  | 预支费用的折扣价，单位：元  |
| original_price      | float  | 预支费用的原价，单位：元   |
| unit_price          | float  | 后付费单价，单位：元     |
| unit_price_discount | float  | 后付费的折扣单价，单位:元  |

##### 后续计价单元 charge_unit，取值范围：
- HOUR：表示计价单元是按每小时来计算。当前涉及该计价单元的场景有：
  - 实例按小时后付费（POSTPAID_BY_HOUR） 、
  - 带宽按小时后付费（BANDWIDTH_POSTPAID_BY_HOUR）；
- GB：表示计价单元是按每GB来计算。当前涉及该计价单元的场景有：
  - 流量按小时后付费（TRAFFIC_POSTPAID_BY_HOUR）。