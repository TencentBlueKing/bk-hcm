### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务-IaaS资源创建。
- 该接口功能描述：创建Aws安全组规则。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/aws/security_groups/{security_group_id}/rules/create

### 输入参数

| 参数名称              | 参数类型                       | 必选    | 描述                                             |
|-------------------|----------------------------|-------|------------------------------------------------|
| bk_biz_id         | int64                      | 是     | 业务ID                                           |
| security_group_id | string                     | 是     | 安全组规则所属安全组ID                                   |
| egress_rule_set   | security_group_rule array  | 否     | 出站规则集。EgressRuleSet和IngressRuleSet，单次请求仅可使用一个。 |
| ingress_rule_set  | security_group_rule  array | 否     | 入站规则集。EgressRuleSet和IngressRuleSet，单次请求仅可使用一个。 |

#### security_group_rule

| 参数名称                           | 参数类型   | 描述  | 描述                                                                                                                                        |
|--------------------------------|--------|-----|-------------------------------------------------------------------------------------------------------------------------------------------|
| protocol                       | string | 是   | 协议, 取值: `tcp`, `udp`, `icmp`, `icmpv6`,用数字 `-1` 代表所有协议 。                                                                                  |
| from_port                      | uint32 | 是   | 起始端口，与 to_port 配合使用。-1代表所有端口。<br />port: 8080 (from_port: 8080, to_port: 8080) <br />port_range: 8080-9000(from_port: 8080, to_port:9000) |
| to_port                        | uint32 | 是   | 结束端口，与from_port配合使用。-1代表所有端口。                                                                                                                      |
| ipv4_cidr                      | string | 否   | IPv4网段。                                                                                                                                   |
| ipv6_cidr                      | string | 否   | IPv4网段。                                                                                                                                   |
| cloud_target_security_group_id | string | 否   | 下一跳安全组实例云ID，例如：sg-ohuuioma。                                                                                                               |
| memo                           | string | 否   | 备注。                                                                                                                                       |
注：为空是不要传递该字段，对字段为""铭感。

### 调用示例

创建AWS出站规则。

```json
{
  "egress_rule_set": [
    {
      "protocol": "tcp",
      "from_port": 8080,
      "to_port": 8080,
      "ipv4_cidr": "0.0.0.0/0",
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
