### 描述

- 该接口提供版本：v9.9.9。
- 该接口所需权限：资源-IaaS资源操作。
- 该接口功能描述：全量覆盖TCloud安全组规则（出站和入站），云上原有的所有规则将被完整替换为请求体中的内容，版本号依次累计叠加。

### URL

PUT /api/v1/cloud/vendors/{vendor}/security_groups/{security_group_id}/rules/batch/overwrite

### 输入参数

| 参数名称              | 参数类型   | 必选 | 描述                      |
|-------------------|--------|----|-------------------------|
| vendor            | string | 是  | 云厂商，当前仅支持 tcloud        |
| security_group_id | string | 是  | 安全组规则所属安全组ID            |
| request body      | object | 是  | 请求体                     |

#### 请求体参数说明

| 参数名称             | 参数类型  | 必选 | 描述                             |
|------------------|-------|----|--------------------------------|
| egress_rule_set  | array | 是  | 出方向规则列表，至少1条，最多100条           |
| ingress_rule_set | array | 是  | 入方向规则列表，至少1条，最多100条           |

**注意：**
- `egress_rule_set` 和 `ingress_rule_set` **均为必填**，且各自至少包含 1 条、最多 100 条规则。
- 本接口为**全量覆盖**：调用成功后，安全组云上所有出站和入站规则将被替换为本次请求体中的内容，原有规则全部清除。

#### egress_rule_set[n] & ingress_rule_set[n] 参数说明

| 参数名称                           | 参数类型   | 必选 | 描述                                                              |
|--------------------------------|--------|----|-----------------------------------------------------------------|
| protocol                       | string | 是  | 协议, 取值: TCP,UDP,ICMP,ICMPv6,ALL                                 |
| port                           | string | 是  | 端口(all, 离散port, range)。 说明：如果Protocol设置为ALL，则Port也需要设置为all。     |
| cloud_service_id               | string | 否  | 协议端口云ID(与 protocol、port、cloud_service_group_id 互斥)。             |
| cloud_service_group_id         | string | 否  | 协议端口组云ID(与 protocol、port、cloud_service_id 互斥)。                  |
| ipv4_cidr                      | string | 否  | IPv4网段(与 ipv6_cidr、cloud_address_id、cloud_address_group_id 互斥)。 |
| ipv6_cidr                      | string | 否  | IPv6网段(与 ipv4_cidr、cloud_address_id、cloud_address_group_id 互斥)。 |
| cloud_address_id               | string | 否  | IP参数模版云ID(与 ipv4_cidr、ipv6_cidr、cloud_address_group_id 互斥)。     |
| cloud_address_group_id         | string | 否  | IP参数模版集合云ID(与 ipv4_cidr、ipv6_cidr、cloud_address_id 互斥)。         |
| cloud_target_security_group_id | string | 否  | 下一跳安全组实例云ID，例如：sg-ohuuioma。                                     |
| action                         | string | 是  | ACCEPT 或 DROP。                                                  |
| memo                           | string | 否  | 备注。                                                             |

注：为空时不要传递该字段，对字段为""敏感。

### 调用示例

全量覆盖腾讯云安全组规则（同时指定出站和入站规则）。

```json
{
  "egress_rule_set": [
    {
      "protocol": "ALL",
      "port": "ALL",
      "ipv4_cidr": "0.0.0.0/0",
      "action": "ACCEPT",
      "memo": "allow all egress"
    }
  ],
  "ingress_rule_set": [
    {
      "protocol": "TCP",
      "port": "80",
      "ipv4_cidr": "0.0.0.0/0",
      "action": "ACCEPT",
      "memo": "allow http ingress"
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
