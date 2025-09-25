### 描述

- 该接口提供版本：v1.8.6+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询业务下符合指定条件的RS列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/targets/list_by_cond

### 输入参数

| 参数名称            | 参数类型         | 必选 | 描述               |
|-----------------|--------------|----|------------------|
| bk_biz_id       | int          | 是  | 业务ID             |
| vendor          | string       | 是  | 云厂商（枚举值：tcloud）  |
| account_id      | string       | 是  | 账号ID             |
| rule_query_list | object array | 是  | 规则查询条件数组，最大支持50个 |

#### rule_query_list

| 参数名称            | 参数类型         | 必选 | 描述                   |
|-----------------|--------------|----|----------------------|
| protocol        | string       | 是  | 协议                   |
| region          | string       | 是  | 地域                   |
| clb_vip_domains | string array | 是  | 负载均衡VIP或域名数组，最大支持50个 |
| cloud_lb_ids    | string array | 是  | 负载均衡云ID数组，最大支持50个    |
| listener_ports  | int array    | 否  | 监听器端口数组，最大支持50个      |
| rs_ips          | string array | 否  | RSIP数组，最大支持500个      |
| rs_ports        | int  array   | 否  | RS端口数组，最大支持500个      |
| domains         | string array | 否  | 域名                   |
| urls            | string array | 否  | URL                  |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "bk_biz_id": 123,
  "vendor": "tcloud",
  "account_id": "xxxxxx",
  "rule_query_list":[
    {
      "protocol": "TCP",
      "region": "ap-nanjing",
      "clb_vip_domains": ["1.1.1.1"],
      "cloud_lb_ids": ["lb-xxxxxx"],
      "listener_ports": [100],
      "rs_ips": ["1.1.1.2"],
      "rs_ports": [101]
    },
    {
      "protocol": "HTTP",
      "region": "ap-nanjing",
      "clb_vip_domains": ["1.1.1.1"],
      "cloud_lb_ids": ["lb-xxxxxx"],
      "listener_ports": [100],
      "domains": ["example.com"],
      "urls": ["/123"],
      "rs_ips": ["1.1.1.2"],
      "rs_ports": [101]
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
    "details": [
      {
        "clb_id": "00000001",
        "cloud_lb_id": "lb-00000001",
        "clb_vip_domain": "1.1.1.1",
        "bk_biz_id": 123,
        "region": "ap-nanjing",
        "vendor": "tcloud",
        "lbl_id": "00000002",
        "cloud_lbl_id": "lbl-00000001",
        "protocol": "HTTPS",
        "domain": "www.qq.com",
        "url": "/",
        "port": 1001,
        "inst_type": "CVM",
        "rs_ip": "10.10.10.10",
        "rs_port": 8000,
        "rs_weight": 50
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述   |
|---------|--------|--------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称       | 参数类型 | 描述               |
|---------------|--------|--------------------|
| details       | array object | 查询返回的RS列表数据 |

#### data.details[n]

| 参数名称           | 参数类型   | 描述                   |
|----------------|--------|----------------------|
| clb_id         | string | 负载均衡实例ID             |
| cloud_lb_id    | string | 负载均衡云实例ID            |
| clb_vip_domain | string | CLB（负载均衡）的VIP域名或IP地址 |
| bk_biz_id      | int    | 业务ID                 |
| region         | string | 地域                   |
| vendor         | string | 云厂商                  |
| lbl_id         | string | 监听器ID                |
| cloud_lbl_id   | string | 监听器云ID               |
| protocol       | string | 协议类型                 |
| domain         | string | 域名, 四层监听器不会返回        |
| url            | string | URL路径, 四层监听器不会返回     |
| port           | int    | 端口号                  |
| inst_type      | string | RS类型,枚举值：CVM、ENI     |
| rs_ip          | string | RS IP地址              |
| rs_port        | int    | RS的端口号               |
| rs_weight      | int    | RS的权重值               |
