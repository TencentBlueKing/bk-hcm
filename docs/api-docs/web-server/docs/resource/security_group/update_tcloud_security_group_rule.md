### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：IaaS资源操作。
- 该接口功能描述：更新TCloud安全组规则，只支持覆盖更新。

### URL

PUT /api/v1/cloud/vendors/tcloud/security_groups/{security_group_id}/rules/{id}

### 输入参数

| 参数名称                           | 参数类型   | 必选  | 描述                                                          |
|--------------------------------|--------|-----|-------------------------------------------------------------|
| id                             | string | 是   | 安全组规则ID                                                     |
| security_group_id              | string | 是   | 安全组规则所属安全组ID                                                |
| protocol                       | string | 是   | 协议, 取值: TCP,UDP,ICMP,ICMPv6,ALL |
| port                           | string | 是   | 端口(all, 离散port, range)。 说明：如果Protocol设置为ALL，则Port也需要设置为all。 |
| cloud_service_id               | string | 否   | 协议端口云ID(与 protocol、port、cloud_service_group_id 互斥)。|
| cloud_service_group_id         | string | 否   | 协议端口组云ID(与 protocol、port、cloud_service_id 互斥)。    |
| ipv4_cidr                      | string | 是   | IPv4网段(与 ipv6_cidr、cloud_address_id、cloud_address_group_id 互斥)。 |
| ipv6_cidr                      | string | 是   | IPv6网段(与 ipv4_cidr、cloud_address_id、cloud_address_group_id 互斥)。 |
| cloud_address_id               | string | 否   | IP参数模版云ID(与 ipv4_cidr、ipv6_cidr、cloud_address_group_id 互斥)。|
| cloud_address_group_id         | string | 否   | IP参数模版集合云ID(与 ipv4_cidr、ipv6_cidr、cloud_address_id 互斥)。   |
| cloud_target_security_group_id | string | 是   | 下一跳安全组实例云ID，例如：sg-ohuuioma。                                 |
| action                         | string | 是   | ACCEPT 或 DROP。                                              |
| memo                           | string | 是   | 备注。                                                         |
注：为空时不要传递该字段，对字段为""敏感。

### 调用示例

更新腾讯云出站规则。

```json
{
  "protocol": "TCP",
  "port": "8080",
  "ipv4_cidr": "0.0.0.0/0",
  "action": "ACCEPT",
  "memo": "create egress rule"
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
