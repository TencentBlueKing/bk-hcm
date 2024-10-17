### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询Azure默认安全组规则。

### URL

GET /api/v1/cloud/vendors/azure/default/security_groups/rules/{type}

### 输入参数

| 参数名称  | 参数类型      | 必选                         | 描述    |
|-------|-----------|----------------------------|-------|
| type  | string ｜是 | 规则类型。（枚举值：egress、ingress）  |

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
  "data": [
    {
      "name": "AllowVnetOutBound",
      "memo": "Allow outbound traffic from all VMs to all VMs in VNET",
      "destination_address_prefix": "VirtualNetwork",
      "destination_address_prefixes": null,
      "cloud_destination_app_security_group_ids": null,
      "destination_port_range": "*",
      "destination_port_ranges": null,
      "protocol": "*",
      "provisioning_state": "",
      "source_address_prefix": "VirtualNetwork",
      "source_address_prefixes": null,
      "cloud_source_app_security_group_ids": null,
      "source_port_range": "*",
      "source_port_ranges": null,
      "priority": 65000,
      "type": "egress",
      "access": "Allow"
    },
    {
      "name": "AllowInternetOutBound",
      "memo": "Allow outbound traffic from all VMs to Internet",
      "destination_address_prefix": "Internet",
      "destination_address_prefixes": null,
      "cloud_destination_app_security_group_ids": null,
      "destination_port_range": "*",
      "destination_port_ranges": null,
      "protocol": "*",
      "provisioning_state": "",
      "source_address_prefix": "*",
      "source_address_prefixes": null,
      "cloud_source_app_security_group_ids": null,
      "source_port_range": "*",
      "source_port_ranges": null,
      "priority": 65001,
      "type": "egress",
      "access": "Allow"
    },
    {
      "name": "DenyAllOutBound",
      "memo": "Deny all outbound traffic",
      "destination_address_prefix": "*",
      "destination_address_prefixes": null,
      "cloud_destination_app_security_group_ids": null,
      "destination_port_range": "*",
      "destination_port_ranges": null,
      "protocol": "*",
      "provisioning_state": "",
      "source_address_prefix": "*",
      "source_address_prefixes": null,
      "cloud_source_app_security_group_ids": null,
      "source_port_range": "*",
      "source_port_ranges": null,
      "priority": 65500,
      "type": "egress",
      "access": "Deny"
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data[n]

| 参数名称                                     | 参数类型   | 描述                                                                                                           |
|------------------------------------------|--------|--------------------------------------------------------------------------------------------------------------|
| name                                     | string | 资源组中唯一的资源名称。此名称可用于访问资源。                                                                                      |
| memo                                     | string | 备注。                                                                                                          |
| destination_address_prefix               | string | 目的地址前缀。CIDR或目标IP范围。星号‘*’也可用于匹配所有源IP。也可以使用‘VirtualNetwork’、‘AzureLoadBalancer’和‘Internet’等默认标签。               |
| destination_address_prefixes             | string | 目的地址带有前缀。CIDR或目标IP范围。                                                                                        |
| cloud_destination_app_security_group_ids | string | 目标应用安全组云ID列表。                                                                                                |
| destination_port_range                   | string | 目标端口或范围。介于0和65535之间的整数或范围。星号‘*’也可用于匹配所有端口。                                                                   |
| destination_port_ranges                  | string | 目的端口范围。                                                                                                      |
| protocol                                 | string | 网络协议。（枚举值：*、Ah、Esp、Icmp、Tcp、Udp）                                                                             |
| provisioning_state                       | string | 调度状态。（枚举值：Deleting、Failed、Succeeded、Updating）                                                                |
| source_address_prefix                    | string | CIDR或来源IP范围。星号‘*’也可用于匹配所有源IP。也可以使用‘VirtualNetwork’、‘AzureLoadBalancer’和‘Internet’等默认标签。如果这是入口规则，则指定网络流量源自何处。 |
| source_address_prefixes                  | string | CIDR或来源IP范围。                                                                                                 |
| cloud_source_app_security_group_ids      | string | 源应用安全组云ID列表。                                                                                                 |
| source_port_range                        | string | 源端口或范围。介于0和65535之间的整数或范围。星号‘*’也可用于匹配所有端口。                                                                    |
| source_port_ranges                       | string | 源端口范围。                                                                                                       |
| priority                                 | uint32 | 规则的优先级。该值可以介于100和4096之间。对于集合中的每个规则，优先级编号必须是唯一的。优先级数字越小，规则的优先级越高。                                             |
| type                                     | string | 规则类型。（枚举值：egress、ingress）                                                                                    |
| access                                   | string | 允许或拒绝网络流量。（枚举值：Allow、Deny）                                                                                   |
