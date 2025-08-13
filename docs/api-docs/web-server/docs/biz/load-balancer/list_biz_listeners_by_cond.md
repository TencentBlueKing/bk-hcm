### 描述

- 该接口提供版本：v1.8.5+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询业务下符合指定条件的监听器列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/listeners/list_by_cond

### 输入参数

| 参数名称         | 参数类型       | 必选 | 描述      |
|-----------------|--------------|-----|-----------|
| bk_biz_id       | int          | 是  | 业务ID     |
| vendor          | string       | 是  | 云厂商     |
| account_id      | string       | 是  | 账号ID     |
| rule_query_list | object array | 是  | 规则查询条件数组，最大支持50个 |

#### rule_query_list

| 参数名称         | 参数类型       | 必选 | 描述                            |
|-----------------|--------------|------|--------------------------------|
| protocol        | string       | 否   | 协议                            |
| region          | string       | 是   | 地域                            |
| clb_vip_domains | string array | 是   | 负载均衡VIP或域名数组，最大支持50个 |
| cloud_lb_ids    | string array | 是   | 负载均衡云ID数组，最大支持50个      |
| rs_ips          | string array | 是   | RSIP数组，最大支持500个           |
| rs_ports        | int  array   | 否   | RS端口数组，最大支持500个          |

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
        "port": 1001
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
| details       | array object | 查询返回的数据 |

#### data.details[n]

| 参数名称        | 参数类型       | 描述            |
|----------------|--------------|-----------------|
| clb_id         | string       | 负载均衡实例ID    |
| cloud_lb_id    | string       | 负载均衡云实例ID  |
| clb_vip_domain | string       | 负载均衡VIP或域名 |
| bk_biz_id      | int          | 业务ID           |
| region         | string       | 地域             |
| vendor         | string       | 云厂商           |
| lbl_id         | string       | 监听器ID         |
| cloud_lbl_id   | string       | 监听器云ID        |
| protocol       | string       | 监听器协议        |
| port           | int          | 监听器端口        |
