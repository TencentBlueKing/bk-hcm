### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询目标组详情。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/target_groups/{id}

### 输入参数

| 参数名称   | 参数类型 | 必选 | 描述    |
|-----------|--------|-----|---------|
| bk_biz_id | int    | 是  | 业务ID   |
| id        | string | 是  | 目标组ID |

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
    "cloud_id": "clb-123",
    "name": "clb-test",
    "vendor": "tcloud",
    "account_id": "0000001",
    "bk_biz_id": -1,
    "target_group_type": "local",
    "region": "ap-hk",
    "protocol": "TCP",
    "port": 22,
    "weight": 22,
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
    "target_list": [
      {
        "id": "tg-xxxx",
        "account_id": "0000001",
        "inst_id": "inst-0000001",
        "inst_name": "inst-xxxx",
        "cloud_inst_id": "cloud-inst-0000001",
        "inst_type": "cvm",
        "target_group_id": "0000001",
        "cloud_target_group_id": "cloud-tg-0000001",
        "port": 80,
        "weight": 80,
        "private_ip_address": [],
        "public_ip_address": [],
        "zone": ""
      }
    ],
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

| 参数名称                | 参数类型         | 描述                                  |
|------------------------|----------------|---------------------------------------|
| id                     | int            | 目标组ID                               |
| cloud_id               | string         | 云目标组ID                             |
| name                   | string         | 目标组名称                              |
| vendor                 | string         | 供应商（枚举值：tcloud）                 |
| account_id             | string         | 账号ID                                 |
| bk_biz_id              | int            | 业务ID                                 |
| target_group_type      | string         | 目标组类型                              |
| region                 | string         | 地域                                   |
| protocol               | string         | 协议                                   |
| port                   | int            | 端口                                   |
| weight                 | int            | 权重                                   |
| vpc_id                 | string array   | vpcID数组                              |
| health_check           | object         | 健康检查                                |
| target_list            | object array   | 目标列表                                |
| memo                   | string         | 备注                                   |
| creator                | string         | 创建者                                  |
| reviser                | string         | 修改者                                  |
| created_at             | string         | 创建时间，标准格式：2006-01-02T15:04:05Z   |
| updated_at             | string         | 修改时间，标准格式：2006-01-02T15:04:05Z   |

### data.health_check

| 参数名称           | 参数类型 | 描述        |
|-------------------|--------|-------------|
| health_switch     | int    | 是否开启健康检查：1（开启）、0（关闭）  |
| time_out          | int    | 健康检查的响应超时时间，可选值：2~60，单位：秒 |
| interval_time     | int    | 健康检查探测间隔时间 |
| health_num        | int    | 健康阈值 |
| un_health_num     | int    | 不健康阈值 |
| check_port        | int    | 自定义探测相关参数。健康检查端口，默认为后端服务的端口 |
| check_type        | string | 健康检查使用的协议。取值 TCP | HTTP | HTTPS | GRPC | PING | CUSTOM  |
| http_version      | string | HTTP版本  |
| http_check_path   | string | 健康检查路径（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式） |
| http_check_domain | string | 健康检查域名 |
| http_check_method | string | 健康检查方法（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式），默认值：HEAD，可选值HEAD或GET |
| source_ip_type    | string | 健康检查源IP类型：0（使用LB的VIP作为源IP），1（使用100.64网段IP作为源IP） |

### data.target_list

| 参数名称                | 参数类型       | 描述      |
|------------------------|--------------|-----------|
| id                     | string       | 目标ID    |
| account_id             | string       | 账号ID    |
| inst_id                | string       | 实例ID    |
| inst_name              | string       | 实例名称   |
| cloud_inst_id          | string       | 云实例ID   |
| inst_type              | string       | 实例类型   |
| target_group_id        | string       | 目标组ID   |
| cloud_target_group_id  | string       | 云目标组ID  |
| port                   | int          | 端口       |
| weight                 | int          | 权重       |
| private_ip_address     | string array | 私有IP数组  |
| public_ip_address      | string array | 公有IP数组  |
| zone                   | string       | 可用区      |