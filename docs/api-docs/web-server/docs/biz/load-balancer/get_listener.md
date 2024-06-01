### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询监听器详情。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/listeners/{id}

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述    |
|-----------|--------|----|-------|
| bk_biz_id | int64  | 是  | 业务ID  |
| id        | string | 是  | 监听器ID |

### 调用示例

#### 获取详细信息请求参数示例

```json
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000001",
    "name": "listener-name",
    "cloud_id": "listener-123",
    "vendor": "tcloud",
    "account_id": "0000001",
    "bk_biz_id": -1,
    "lb_id": "xxxx",
    "cloud_lb_id": "lb-xxxx",
    "protocol": "HTTP",
    "port": 8080,
    "target_group_id": "tg-001",
    "target_group_name": "tg-name",
    "cloud_target_group_id": "cloud-tg-001",
    "scheduler": "WRR",
    "session_type": "NORMAL",
    "session_expire": 0,
    "health_check": {
      "health_switch": 1,
      "time_out": 2,
      "interval_time": 5,
      "health_num": 3,
      "un_health_num": 3,
      "check_port": 80,
      "check_type": "HTTP",
      "http_version": "HTTP/1.0",
      "http_check_path": "/",
      "http_check_domain": "www.weixin.com",
      "http_check_method": "GET",
      "source_ip_type": 1
    },
    "certificate": {
      "ssl_mode": "MUTUAL",
      "cert_id": "cert-001",
      "cert_ca_id": "ca-001",
      "ext_cert_ids": [
        "ext-001"
      ]
    },
    "domain_num": 50,
    "url_num": 100,
    "memo": "memo-test",
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
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称                  | 参数类型   | 描述                             |
|-----------------------|--------|--------------------------------|
| id                    | int    | 监听器ID                          |
| name                  | string | 监听器名称                          |
| cloud_id              | string | 云监听器ID                         |
| vendor                | string | 供应商（枚举值：tcloud）                |
| account_id            | string | 账号ID                           |
| bk_biz_id             | int64  | 业务ID                           |
| lb_id                 | string | 负载均衡ID                         |
| cloud_lb_id           | string | 云负载均衡ID                        |
| protocol              | string | 协议                             |
| port                  | int    | 端口                             |
| target_group_id       | string | 目标组ID                          |
| target_group_name     | string | 目标组名称                          |
| cloud_target_group_id | string | 云目标组ID                         |
| scheduler             | string | 负载均衡方式                         |
| session_type          | string | 会话保持类型                         |
| session_expire        | int    | 会话保持时间，0为关闭                    |
| health_check          | object | 健康检查                           |
| certificate           | object | 证书信息                           |
| domain_num            | int    | 域名数量                           |
| url_num               | int    | URL数量                          |
| memo                  | string | 备注                             |
| creator               | string | 创建者                            |
| reviser               | string | 修改者                            |
| created_at            | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at            | string | 修改时间，标准格式：2006-01-02T15:04:05Z |

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
| http_version      | string | HTTP版本                                                            |
| http_check_path   | string | 健康检查路径（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）                      |
| http_check_domain | string | 健康检查域名                                                            |
| http_check_method | string | 健康检查方法（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式），默认值：HEAD，可选值HEAD或GET |
| source_ip_type    | string | 健康检查源IP类型：0（使用LB的VIP作为源IP），1（使用100.64网段IP作为源IP）                   |

### certificate

| 参数名称           | 参数类型         | 描述                                   |
|----------------|--------------|--------------------------------------|
| ssl_mode       | string       | 认证类型，UNIDIRECTIONAL：单向认证，MUTUAL：双向认证 |
| ca_cloud_id    | string       | CA证书的云ID                             |
| cert_cloud_ids | string array | 服务端证书的云ID                            |
