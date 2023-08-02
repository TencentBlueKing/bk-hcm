### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务-IaaS资源创建。
- 该接口功能描述：创建Gcp防火墙规则。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/gcp/firewalls/rules/create

### 输入参数

| 参数名称               | 参数类型               | 必选 | 描述                                      |
|--------------------|--------------------|----|-----------------------------------------|
| bk_biz_id          | int64              | 是  | 业务ID                                    |
| account_id         | string             | 是  | 账号ID                                    |
| cloud_vpc_id       | string             | 是  | 云VpcID                                  |
| name               | string             | 是  | 名称                                      |
| memo               | string             | 是  | 备注                                      |
| type               | string             | 是  | 类型（EGRESS: 出站、INGRESS：入站）               |
| priority           | uint64             | 是  | 优先级。0-65535                             |
| source_ranges      | string array       | 否  | 源网段列表                                   |
| destination_ranges | string array       | 否  | 目标网段列表                                  |
| source_tags        | string array       | 否  | 源标记列表                                   |
| target_tags        | string array       | 否  | 目标标记列表                                  |
| denied             | protocol_set array | 是  | 防火墙指定的拒绝规则列表。每个规则都指定描述拒绝连接的协议和端口范围元组。   |
| allowed            | protocol_set array | 是  | 防火墙指定的允许规则列表。每个规则都指定了描述允许的连接的协议和端口范围元组。 |
| disabled           | boolean            | 是  | 是否已禁用。                                  |

### 调用示例

```json
{
  "name": "firewall-rule-test",
  "account_id": "00000001",
  "cloud_vpc_id": "https://www.googleapis.com/compute/v1/projects/bk/global/networks/default",
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
      "ports": [
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
  "message": "ok",
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

| 参数名称 | 参数类型   | 描述    |
|------|--------|-------|
| id   | string | 安全组ID |
