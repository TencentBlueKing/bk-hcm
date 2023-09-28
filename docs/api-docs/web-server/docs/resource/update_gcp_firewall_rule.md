### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：IaaS资源操作。
- 该接口功能描述：更新Gcp防火墙规则，只支持覆盖更新。

### URL

PUT /api/v1/cloud/vendors/gcp/firewalls/rules/{id}

### 输入参数

| 参数名称               | 参数类型               | 必选 | 描述                                      |
|--------------------|--------------------|----|-----------------------------------------|
| id                 | string             | 是  | Gcp防火墙规则ID                              |
| memo               | string             | 是  | 备注                                      |
| priority           | uint64             | 是  | 优先级。0-65535                             |
| source_ranges      | string array       | 否  | 源网段列表                                   |
| destination_ranges | string array       | 否  | 目标网段列表                                  |
| source_tags        | string array       | 否  | 源标记列表                                   |
| target_tags        | string array       | 否  | 目标标记列表                                  |
| denied             | protocol_set array | 是  | 防火墙指定的拒绝规则列表。每个规则都指定描述拒绝连接的协议和端口范围元组。   |
| allowed            | protocol_set array | 是  | 防火墙指定的允许规则列表。每个规则都指定了描述允许的连接的协议和端口范围元组。 |
| disabled           | boolean            | 是  | 是否已禁用。                                  |

#### protocol_set

| 参数名称     | 参数类型         | 描述                                           |
|----------|--------------|----------------------------------------------|
| protocol | string       | 协议。（枚举值：tcp, udp, icmp, esp, ah, ipip, sctp） |
| port     | string array | 端口列表。                                        |

### 调用示例

更新id为1的防火墙规则为允许从端口443访问HTTPS规则，且目标带有 https-server 的标记。

```json
{
  "priority": 100,
  "memo": "firewall list",
  "source_ranges": [
    "0.0.0.0/0"
  ],
  "destination_ranges": [],
  "source_tags": [],
  "target_tags": [
    "https-server"
  ],
  "denied": [],
  "allowed": [
    {
      "protocol": "tcp",
      "port": [
        "443"
      ]
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
