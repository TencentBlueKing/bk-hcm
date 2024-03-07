### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询URL规则详情。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/url_rules/{id}

### 输入参数

| 参数名称   | 参数类型 | 必选 | 描述     |
|-----------|--------|-----|----------|
| bk_biz_id | int64  | 是  | 业务ID    |
| id        | string | 是  | URL规则ID |

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
    "cloud_id": "listener-123",
    "name": "listener-name",
    "lbl_id": "00000001",
    "lbl_name": "listener-name",
    "cloud_lbl_id": "listener-123",
    "lb_id": "xxxx",
    "cloud_lb_id": "lb-xxxx",
    "protocol": "HTTP",
    "port": 8080,
    "domain_num": 100
    "url_num": 100,
    "scheduler": "WRR",
    "session_type": "NORMAL",
    "session_expire": 0,
    "health_check": {
      "HealthSwitch": 1,
      "TimeOut": 2,
      "IntervalTime": 5,
      "HealthNum": 3,
      "UnHealthNum": 3,
      "CheckPort": 80,
      "CheckType": "HTTP",
      "HttpVersion": "HTTP/1.0",
      "HttpCheckPath": "/",
      "HttpCheckDomain": "www.weixin.com",
      "HttpCheckMethod": "GET",
      "SourceIpType": 1
    },
    "certificate": {
      "SSLMode": "MUTUAL",
      "CertId": "cert-001",
      "CertCaId": "ca-001",
      "ExtCertIds": ["ext-001"]
    },
    "memo": "memo-test",
    "creator": "Jim",
    "reviser": "Jim",
    "created_at": "2023-02-12T14:47:39Z",
    "updated_at": "2023-02-12T14:55:40Z"
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述    |
|---------|--------|---------|
| code    | int    | 状态码   |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称                | 参数类型         | 描述                                   |
|------------------------|----------------|---------------------------------------|
| id                     | uint64         | 资源ID                                 |
| cloud_id               | string         | 云资源ID                                |
| name                   | string         | 名称                                   |
| lbl_id                 | int            | 监听器ID                               |
| lbl_name               | string         | 监听器名称                              |
| cloud_lbl_id           | string         | 云监听器ID                             |
| lb_id                  | string         | 负载均衡ID                              |
| cloud_lb_id            | string         | 云负载均衡ID                            |
| protocol               | string         | 协议                                   |
| port                   | int            | 端口                                   |
| domain_num             | int            | 域名数量                                |
| url_num                | int            | URL数量                                |
| scheduler              | string         | 负载均衡方式                             |
| session_type           | string         | 会话保持类型                             |
| session_expire         | int            | 会话保持时间，0为关闭                     |
| health_check           | object         | 健康检查                                |
| certificate            | object         | 证书信息                                |
| memo                   | string         | 备注                                    |
| creator                | string         | 创建者                                  |
| reviser                | string         | 修改者                                  |
| created_at             | string         | 创建时间，标准格式：2006-01-02T15:04:05Z   |
| updated_at             | string         | 修改时间，标准格式：2006-01-02T15:04:05Z   |

### health_check

| 参数名称          | 参数类型 | 描述        |
|------------------|--------|-------------|
| HealthSwitch     | int    | 是否开启健康检查：1（开启）、0（关闭）  |
| TimeOut          | int    | 健康检查的响应超时时间，可选值：2~60，单位：秒 |
| IntervalTime     | int    | 健康检查探测间隔时间 |
| HealthNum        | int    | 健康阈值 |
| UnHealthNum      | int    | 不健康阈值 |
| CheckPort        | int    | 自定义探测相关参数。健康检查端口，默认为后端服务的端口 |
| CheckType        | string | 健康检查使用的协议。取值 TCP | HTTP | HTTPS | GRPC | PING | CUSTOM  |
| HttpVersion      | string | HTTP版本  |
| HttpCheckPath    | string | 健康检查路径（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式） |
| HttpCheckDomain  | string | 健康检查域名 |
| HttpCheckMethod  | string | 健康检查方法（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式），默认值：HEAD，可选值HEAD或GET |
| SourceIpType     | string | 健康检查源IP类型：0（使用LB的VIP作为源IP），1（使用100.64网段IP作为源IP） |

### certificate

| 参数名称     | 参数类型       | 描述          |
|-------------|--------------|---------------|
| SSLMode     | string       | 认证类型，UNIDIRECTIONAL：单向认证，MUTUAL：双向认证  |
| CertId      | string       | 服务端证书的ID  |
| CertCaId    | string       | 客户端证书的 ID |
| ExtCertIds  | string array | 多本服务器证书场景扩展的服务器证书ID |

