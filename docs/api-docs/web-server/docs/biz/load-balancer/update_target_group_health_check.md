### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：目标组更新。
- 该接口功能描述：业务下更新目标组健康检查。

### URL

PATCH /api/v1/cloud/bizs/{bk_biz_id}/target_groups/{id}/health_check

### 输入参数

#### tcloud

| 参数名称         | 参数类型         | 必选 | 描述     |
|--------------|--------------|----|--------|
| id           | string       | 是  | 目标组ID  |
| health_check | health_check | 是  | 健康检查信息 |

### health_check

| 参数名称              | 参数类型   | 描述                                                                                |
|-------------------|--------|-----------------------------------------------------------------------------------|
| health_switch     | int    | 是否开启健康检查：1（开启）、0（关闭）                                                              |
| time_out          | int    | 健康检查的响应超时时间，可选值：2~60，单位：秒                                                         |
| interval_time     | int    | 健康检查探测间隔时间                                                                        |
| health_num        | int    | 健康阈值                                                                              |
| un_health_num     | int    | 不健康阈值                                                                             |
| check_port        | int    | 自定义探测相关参数。健康检查端口，默认为后端服务的端口                                                       |
| check_type        | string | 健康检查使用的协议。取值 TCP/HTTP/HTTPS/GRPC/PING/CUSTOM                                      |
| http_code         | int    | 健康检查类型 （仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）                                     |
| http_version      | string | HTTP版本 健康检查协议CheckType的值取HTTP时，必传此字段，代表后端服务的HTTP版本：HTTP/1.0、HTTP/1.1；（仅适用于TCP监听器） |
| http_check_path   | string | 健康检查路径（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）                                      |
| http_check_domain | string | 健康检查域名                                                                            |
| http_check_method | string | 健康检查方法（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式），默认值：HEAD，可选值HEAD或GET                 |
| source_ip_type    | string | 健康检查源IP类型：0（使用LB的VIP作为源IP），1（使用100.64网段IP作为源IP）                                   |
| context_type      | string | 健康检查的输入格式，可取值：HEX或TEXT；                                                           |
| send_context      | string | 自定义探测相关参数。健康检查协议CheckType的值取CUSTOM时，必填此字段,代表健康检查返回的结果，只允许ASCII可见字符，最大长度限制500。     |
| recv_context      | string | 自定义探测相关参数。健康检查协议CheckType的值取CUSTOM时，必填此字段，代表健康检查返回的结果，只允许ASCII可见字符，最大长度限制500      |
| extended_code     | string | GRPC健康检查状态码（仅适用于后端转发协议为GRPC的规则）.默认值为 12，可输入值为数值、多个数值, 或者范围，例如 20 或 20,25 或 0-99   |

#### http_code 取值范围

可选值：1~31，默认 31。

1 表示探测后返回值 1xx 代表健康
2 表示返回 2xx 代表健康
4 表示返回 3xx 代表健康
8 表示返回 4xx 代表健康
16 表示返回 5xx 代表健康。

若希望多种返回码都可代表健康,则将相应的值相加。

### 调用示例

#### tcloud

```json
{
  "account_id": "0000001",
  "name": "xxx",
  "protocol": "TCP",
  "port": 22,
  "region": "ap-hk",
  "cloud_vpc_id": [
    "xxxx",
    "xxxx"
  ],
  "memo": ""
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
| code    | int    | 状态码  |
| message | string | 请求信息 |

