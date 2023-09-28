### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询Gcp防火墙规则详情。

### URL

GET /api/v1/cloud/vendors/gcp/firewalls/rules/{id}

### 输入参数

| 参数名称 | 参数类型   | 必选 | 描述         |
|------|--------|----|------------|
| id   | string | 是  | Gcp防火墙规则ID |

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001",
    "cloud_id": "123456",
    "name": "test",
    "priority": 1000,
    "memo": "rule test",
    "cloud_vpc_id": "https://www.googleapis.com/compute/v1/projects/test/global/networks/test",
    "source_ranges": [
      "0.0.0.0/0"
    ],
    "bk_biz_id": -1,
    "vpc_id": "00000001",
    "destination_ranges": null,
    "source_tags": null,
    "target_tags": null,
    "source_service_accounts": null,
    "target_service_accounts": null,
    "denied": null,
    "allowed": [
      {
        "protocol": "tcp",
        "port": [
          "7777"
        ]
      }
    ],
    "type": "INGRESS",
    "log_enable": false,
    "disabled": false,
    "account_id": "00000001",
    "self_link": "https://www.googleapis.com/compute/v1/projects/test/global/firewalls/test",
    "creator": "Jim",
    "reviser": "Jim",
    "created_at": "2023-01-12T19:58:56Z",
    "updated_at": "2023-01-12T19:58:56Z"
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

| 参数名称                    | 参数类型               | 描述                                      |
|-------------------------|--------------------|-----------------------------------------|
| id                      | string             | Gcp防火墙规则ID                              |
| cloud_id                | string             | 云ID                                     |
| name                    | string             | 名称                                      |
| priority                | uint64             | 优先级。0-65535                             |
| cloud_vpc_id            | string             | Vpc云ID                                  |
| bk_biz_id               | int64              | 业务ID, -1代表未分配业务                         |
| vpc_id                  | string             | VpcID                                   |
| memo                    | string             | 备注                                      |
| source_ranges           | string array       | 源网段列表                                   |
| destination_ranges      | string array       | 目标网段列表                                  |
| source_tags             | string array       | 源标记列表                                   |
| target_tags             | string array       | 目标标记列表                                  |
| source_service_accounts | string array       | 源服务账号ID列表                               |
| target_service_accounts | string array       | 目标服务账号ID列表                              |
| denied                  | protocol_set array | 防火墙指定的拒绝规则列表。每个规则都指定描述拒绝连接的协议和端口范围元组。   |
| allowed                 | protocol_set array | 防火墙指定的允许规则列表。每个规则都指定了描述允许的连接的协议和端口范围元组。 |
| type                    | string             | 类型（枚举值：EGRESS、INGRESS）                  |
| disabled                | boolean            | 是否已禁用。                                  |
| log_enable              | boolean            | 防火墙规则日志开关。                              |
| self_link               | string             | 资源的服务器定义的URL。                           |
| account_id              | string             | 账号ID                                    |
| creator                 | string             | 创建者                                     |
| reviser                 | string             | 最后一次修改的修改者                              |
| create_at               | string             | 创建时间，标准格式：2006-01-02T15:04:05Z          |
| update_at               | string             | 最后一次修改时间，标准格式：2006-01-02T15:04:05Z      |

#### protocol_set

| 参数名称     | 参数类型         | 描述                                           |
|----------|--------------|----------------------------------------------|
| protocol | string       | 协议。（枚举值：tcp, udp, icmp, esp, ah, ipip, sctp） |
| port     | string array | 端口列表。                                        |
