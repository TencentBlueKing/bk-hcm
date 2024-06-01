### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下腾讯云规则更新。

### URL

PATCH /api/v1/cloud/bizs/{bk_biz_id}/vendors/tcloud/listeners/{lbl_id}/rules/{rule_id}

### 输入参数

| 参数名称                | 参数类型   | 必选 | 描述                |
|---------------------|--------|----|-------------------|
| bk_biz_id           | int64  | 是  | 业务ID              |
| lbl_id              | string | 是  | 监听器id             |
| rule_id             | string | 是  | 规则id              |
| url                 | string | 否  | 监听的url            |
| session_expire_time | int    | 否  | 会话过期时间            |
| scheduler           | string | 否  | 均衡方式              |
| forward_type        | string | 否  | 转发类型              |
| default_server      | bool   | 否  | 默认服务              |
| http2               | bool   | 否  | http2             |
| target_type         | string | 否  | 目标类型              |
| quic                | bool   | 否  | quic  开关          |
| trpc_func           | string | 否  | trpc函数            |
| trpc_callee         | string | 否  | trpc 调用者          |
| health_check        | object | 否  | 健康检查信息            |
| certificate         | object | 否  | 证书信息，当协议为HTTPS时必传 |

### health_check

| 参数名称              | 参数类型   | 描述                                                                |
|-------------------|--------|-------------------------------------------------------------------|
| health_switch     | int    | 是否开启健康检查：1（开启）、0（关闭）                                              |
| time_out          | int    | 健康检查的响应超时时间，可选值：2~60，单位：秒                                         |
| interval_time     | int    | 健康检查探测间隔时间                                                        |
| health_num        | int    | 健康阈值                                                              |
| un_health_num     | int    | 不健康阈值                                                             |
| check_port        | int    | 自定义探测相关参数。健康检查端口，默认为后端服务的端口                                       |
| check_type        | string | 健康检查使用的协议。取值 TCP/HTTP/HTTPS/GRPC/PING/CUSTOM                      |
| http_code         | string | 健康检查类型                                                            |
| http_version      | string | HTTP版本                                                            |
| http_check_path   | string | 健康检查路径（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）                      |
| http_check_domain | string | 健康检查域名                                                            |
| http_check_method | string | 健康检查方法（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式），默认值：HEAD，可选值HEAD或GET |
| source_ip_type    | string | 健康检查源IP类型：0（使用LB的VIP作为源IP），1（使用100.64网段IP作为源IP）                   |
| context_type      | string | 健康检查的输入格式，可取值：HEX或TEXT；                                           |

### http_code 取值说明

| 值  | 说明                |
|----|-------------------|
| 1  | 表示探测后返回值 1xx 代表健康 |
| 2  | 表示返回 2xx 代表健康     |
| 4  | 表示返回 3xx 代表健康     |
| 8  | 表示返回 4xx 代表健康，    |
| 16 | 表示返回 5xx 代表健康。    |

若希望多种返回码都可代表健康，则将相应的值相加。

### 调用示例

```json
{
  "url": "/new/url",
  "scheduler": "IP_HASH",
  "session_type": "NORMAL",
  "session_expire": 300,
  "health_check": {
    "http_check_path": "/healthz",
    "http_check_domain": "www.updatedomain.com",
    "http_check_method": "HEAD",
    "source_ip_type": 1
  }
}
```

### 响应示例

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
