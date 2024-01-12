### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务-IaaS资源创建。
- 该接口功能描述：创建TCloud安全组规则。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/tcloud/security_groups/{security_group_id}/rules/create

### 输入参数

| 参数名称              | 参数类型                       | 必选   | 描述                                             |
|-------------------|----------------------------|------|------------------------------------------------|
| bk_biz_id         | int64                      | 是    | 业务ID                                           |
| security_group_id | string                     | 是    | 安全组规则所属安全组ID                                   |
| egress_rule_set   | security_group_rule array  | 否    | 出站规则集。EgressRuleSet和IngressRuleSet，单次请求仅可使用一个。 |
| ingress_rule_set  | security_group_rule array  | 否    | 入站规则集。EgressRuleSet和IngressRuleSet，单次请求仅可使用一个。 |

#### security_group_rule

| 参数名称                           | 参数类型   | 描述  | 描述                                                          |
|--------------------------------|--------|-----|-------------------------------------------------------------|
| protocol                       | string | 是   | 协议, 取值: TCP,UDP,ICMP,ICMPv6,ALL |
| port                           | string | 是   | 端口(all, 离散port, range)。 说明：如果Protocol设置为ALL，则Port也需要设置为all。 |
| cloud_service_id               | string | 否   | 协议端口云ID(与 protocol、port、cloud_service_group_id 互斥)。|
| cloud_service_group_id         | string | 否   | 协议端口组云ID(与 protocol、port、cloud_service_id 互斥)。   |
| ipv4_cidr                      | string | 否   | IPv4网段(与 ipv6_cidr、cloud_address_id、cloud_address_group_id 互斥)。 |
| ipv6_cidr                      | string | 否   | IPv6网段(与 ipv4_cidr、cloud_address_id、cloud_address_group_id 互斥)。 |
| cloud_address_id               | string | 否   | IP参数模版云ID(与 ipv4_cidr、ipv6_cidr、cloud_address_group_id 互斥)。|
| cloud_address_group_id         | string | 否   | IP参数模版集合云ID(与 ipv4_cidr、ipv6_cidr、cloud_address_id 互斥)。   |
| cloud_target_security_group_id | string | 否   | 下一跳安全组实例云ID，例如：sg-ohuuioma。                                 |
| action                         | string | 是   | ACCEPT 或 DROP。                                              |
| memo                           | string | 否   | 备注。                                                         |
注：为空时不要传递该字段，对字段为""敏感。

### 调用示例

创建腾讯云出站规则。

```json
{
  "egress_rule_set": [
    {
      "protocol": "TCP",
      "port": "8080",
      "ipv4_cidr": "0.0.0.0/0",
      "action": "ACCEPT",
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
