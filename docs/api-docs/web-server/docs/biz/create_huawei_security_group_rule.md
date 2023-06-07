### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务-IaaS资源创建。
- 该接口功能描述：创建HuaWei安全组规则。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/huawei/security_groups/{security_group_id}/rules/create

### 输入参数

| 参数名称              | 参数类型                      | 必选     | 描述                                             |
|-------------------|---------------------------|--------|------------------------------------------------|
| bk_biz_id         | int64                     | 是      | 业务ID                                           |
| security_group_id | string                    | 是      | 安全组规则所属安全组ID                                   |
| egress_rule_set   | security_group_rule array | 否      | 出站规则集。EgressRuleSet和IngressRuleSet，单次请求仅可使用一个。 |
| ingress_rule_set  | security_group_rule array | 否      | 入站规则集。EgressRuleSet和IngressRuleSet，单次请求仅可使用一个。 |

#### security_group_rule

| 参数名称                  | 参数类型   | 描述  | 描述                                                                                                                       |
|-----------------------|--------|-----|--------------------------------------------------------------------------------------------------------------------------|
| protocol              | string | 否   | 协议类型取值范围：icmp、tcp、udp、icmpv6或IP协议号约束：为空表示支持所有协议协议，为icmpv6时，网络类型应该为IPv6协议为icmp时，网络类型应该为IPv4。                              |
| ethertype             | string | 否   | IP地址协议类型取值范围。（枚举值：IPv4、IPv6）                                                                                             |
| cloud_remote_group_id | string | 否   | 远端安全组ID，表示该安全组内的流量允许或拒绝取值范围：与remote_ip_prefix，remote_address_group_id功能互斥。                                               |
| remote_ip_prefix      | string | 否   | 远端IP地址，当type是egress时，为虚拟机访问端的地址。当type是ingress时，为访问虚拟机的地址取值范围：IP地址，或者cidr格式约束：与remote_group_id、remote_address_group_id互斥。 |
| port                  | string | 是   | 端口取值范围取值范围：支持单端口(80)，连续端口(1-30)以及不连续端口(22,3389,80)。                                                                      |
| priority              | uint32 | 是   | 功能说明：优先级取值范围：1~100，1代表最高优先级。                                                                                             |
| action                | string | 是   | 安全组规则生效策略。取值范围：allow表示允许，deny表示拒绝。                                                                                       |
| memo                  | string | 否   | 备注。                                                                                                                      |
注：为空是不要传递该字段，对字段为""铭感。

### 调用示例

创建HuaWei出站规则。

```json
{
  "egress_rule_set": [
    {
      "protocol": "icmp",
      "ethertype": "IPv4",
      "remote_ip_prefix": "0.0.0.0/0",
      "port": "8080",
      "priority": 50,
      "action": "allow",
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
