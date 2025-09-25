### 描述

- 该接口提供版本：v1.8.6+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下创建腾讯云规则。

说明：仅创建规则, 健康检查默认开启, 使用云上默认的检查规则, 绑定目标组时会将目标组的健康检查设置同步到规则上。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/listeners/{lbl_id}/rule/create

### 输入参数

| 参数名称                | 参数类型         | 必选 | 描述                |
|---------------------|--------------|----|-------------------|
| bk_biz_id           | int64        | 是  | 业务ID              |
| lbl_id              | string       | 是  | 监听器id             |
| url                 | string       | 是  | 监听的url            |
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



```json
{
  "domains": [
    "aaa.aaa.com"
  ],
  "url": "/123321",
  "scheduler": "WRR",
  "session_expire_time": 30,
  "forward_type": "HTTP",
  "default_server": true,
  "http2": true
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |
