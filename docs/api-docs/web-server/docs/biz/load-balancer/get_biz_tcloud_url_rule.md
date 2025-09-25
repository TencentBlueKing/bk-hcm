### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：获取业务下腾讯云规则详情。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/vendors/tcloud/listeners/{lbl_id}/rules/{rule_id}

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述     |
|-----------|--------|----|--------|
| bk_biz_id | int64  | 是  | 业务ID   |
| lbl_id    | string | 是  | 监听器id  |

### 调用示例
```json
{}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000005",
    "cloud_id": "loc-abcde",
    "name": "loc-005",
    "rule_type": "layer_7",
    "lb_id": "00000001",
    "cloud_lb_id": "lb-123456",
    "lbl_id": "00000002",
    "cloud_lbl_id": "lbl-xyz",
    "target_group_id": "00000003",
    "cloud_target_group_id": "lbtg-xxxx",
    "domain": "www.qq.com",
    "url": "/test",
    "scheduler": "WRR",
    "sni_switch": 0,
    "session_type": "NORMAL",
    "session_expire": 0,
    "account_id": "admin",
    "bk_biz_id": 0,
    "creator": "Jim",
    "reviser": "Jim",
    "created_at": "2023-02-12T14:47:39Z",
    "updated_at": "2023-02-12T14:55:40Z"
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

| 参数名称           | 参数类型   | 描述                            |
|----------------|--------|-------------------------------|
| id             | string | 资源ID                          |
| cloud_id       | string | 云资源ID                         |
| name           | string | 名称                            |
| lb_id          | string | 负载均衡id                        |
| cloud_lb_id    | string | 云上负载均衡id                      |
| lbl_id         | string | 所属监听器id                       |
| cloud_lbl_id   | string | 所属监听器云上id                     |
| domain         | string | 监听的域名                         |
| url            | string | 监听的url                        |
| scheduler      | string | 调度器                           |
| sni_switch     | int    | sni开关                         |
| session_type   | string | 会话保持类型                        |
| session_expire | string | 会话过期时间                        |
| bk_biz_id      | int64  | 业务ID                          |
| account_id     | string | 账号ID                          |
| creator        | string | 创建者                           |
| reviser        | string | 修改者                           |
| created_at     | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at     | string | 修改时间，标准格式：2006-01-02T15:04:05Z |