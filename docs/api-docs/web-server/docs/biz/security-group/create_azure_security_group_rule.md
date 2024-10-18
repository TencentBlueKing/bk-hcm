### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务-IaaS资源创建。
- 该接口功能描述：创建Azure安全组规则。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/azure/security_groups/{security_group_id}/rules/create

### 输入参数

| 参数名称              | 参数类型                      | 必选      | 描述                                             |
|-------------------|---------------------------|---------|------------------------------------------------|
| bk_biz_id         | int64                     | 是       | 业务ID                                           |
| security_group_id | string                    | 是       | 安全组规则所属安全组ID                                   |
| egress_rule_set   | security_group_rule array | 否       | 出站规则集。EgressRuleSet和IngressRuleSet，单次请求仅可使用一个。 |
| ingress_rule_set  | security_group_rule array | 否       | 入站规则集。EgressRuleSet和IngressRuleSet，单次请求仅可使用一个。 |

#### security_group_rule

| 参数名称                                 | 参数类型   | 描述  | 描述                                                                                                           |
|--------------------------------------|--------|-----|--------------------------------------------------------------------------------------------------------------|
| name                                 | string | 是   | 资源组中唯一的资源名称。此名称可用于访问资源。                                                                                      |
| memo                                 | string | 否   | 备注。                                                                                                          |
| destination_address_prefix           | string | 否   | 目的地址前缀。CIDR或目标IP范围。星号‘*’也可用于匹配所有源IP。也可以使用‘VirtualNetwork’、‘AzureLoadBalancer’和‘Internet’等默认标签。               |
| destination_address_prefixes         | string | 否   | 目的地址带有前缀。CIDR或目标IP范围。                                                                                        |
| destination_port_range               | string | 否   | 目标端口或范围。介于0和65535之间的整数或范围。星号‘*’也可用于匹配所有端口。                                                                   |
| destination_port_ranges              | string | 否   | 目的端口范围。                                                                                                      |
| protocol                             | string | 是   | 网络协议。（枚举值：*、Ah、Esp、Icmp、Tcp、Udp）                                                                             |
| source_address_prefix                | string | 否   | CIDR或来源IP范围。星号‘*’也可用于匹配所有源IP。也可以使用‘VirtualNetwork’、‘AzureLoadBalancer’和‘Internet’等默认标签。如果这是入口规则，则指定网络流量源自何处。 |
| source_address_prefixes              | string | 否   | CIDR或来源IP范围。                                                                                                 |
| source_port_range                    | string | 否   | 源端口或范围。介于0和65535之间的整数或范围。星号‘*’也可用于匹配所有端口。                                                                    |
| source_port_ranges                   | string | 否   | 源端口范围。                                                                                                       |
| priority                             | uint32 | 是   | 规则的优先级。该值可以介于100和4096之间。对于集合中的每个规则，优先级编号必须是唯一的。优先级数字越小，规则的优先级越高。                                             |
| access                               | string | 是   | 允许或拒绝网络流量。（枚举值：Allow、Deny）                                                                                   |
注：为空是不要传递该字段，对字段为""铭感。

### 调用示例

创建Azure出站规则。

```json
{
  "egress_rule_set": [
    {
      "name": "HTTP",
      "protocol": "TCP",
      "source_port_range": "*",
      "destination_port_range": "80",
      "source_address_prefix": "*",
      "destination_address_prefix": "*",
      "access": "allow",
      "priority": 300,
      "memo": "create egress rule"
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "ids": [
      "1234567889"
    ]
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

| 参数名称 | 参数类型         | 描述        |
|------|--------------|-----------|
| ids  | string array | 安全组规则ID列表 |
