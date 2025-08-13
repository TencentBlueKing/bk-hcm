### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下创建腾讯云规则。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/tcloud/listeners/{lbl_id}/rules/create

### 输入参数

| 参数名称                | 参数类型         | 必选 | 描述                |
|---------------------|--------------|----|-------------------|
| bk_biz_id           | int64        | 是  | 业务ID              |
| lbl_id              | string       | 是  | 监听器id             |
| url                 | string       | 是  | 监听的url            |
| target_group_id     | string       | 是  | 目标组id             |
| domains             | string array | 否  | 域名                |
| session_expire_time | int          | 否  | 会话过期时间            |
| scheduler           | string       | 否  | 均衡方式              |
| forward_type        | string       | 否  | 转发类型              |
| default_server      | bool         | 否  | 默认服务              |
| http2               | bool         | 否  | http2             |
| target_type         | string       | 否  | 目标类型              |
| quic                | bool         | 否  | quic  开关          |
| trpc_func           | string       | 否  | trpc函数            |
| trpc_callee         | string       | 否  | trpc 调用者          |
| certificate         | object       | 否  | 证书信息，当协议为HTTPS时必传 |

### certificate

| 参数名称           | 参数类型         | 描述                                   |
|----------------|--------------|--------------------------------------|
| ssl_mode       | string       | 认证类型，UNIDIRECTIONAL：单向认证，MUTUAL：双向认证 |
| ca_cloud_id    | string       | ca证书的云ID                             |
| cert_cloud_ids | string array | 服务端证书的云ID                            |

### 调用示例

#### 获取详细信息请求参数示例

查询创建者是Jim的监听器列表。

```json
{
  "rules": [
    {
      "url": "/testcreate0",
      "domains": [
        "domain1.com"
      ],
      "session_expire_time": 30,
      "scheduler": "WRR",
      "forward_type": "HTTP",
      "default_server": true,
      "http2": true
    },
    {
      "url": "/testcreate1",
      "domains": [
        "domain1.com"
      ],
      "session_expire_time": 30,
      "scheduler": "WRR",
      "forward_type": "HTTP",
      "default_server": true,
      "http2": true
    }
  ]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "unknown_cloud_ids": [],
    "success_cloud_ids": [
      "loc-oxanwx1q"
    ],
    "failed_cloud_ids": [],
    "failed_message": "",
    "success_ids": [""]
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

| 参数名称              | 参数类型         | 描述      |
|-------------------|--------------|---------|
| unknown_cloud_ids | string array | 未知云id列表 |
| success_cloud_ids | string array | 成功云id列表 |
| failed_cloud_ids  | string array | 失败云id列表 |
| failed_message    | string       | 失败信息    |
| success_ids       | string array | 成功id列表  |
