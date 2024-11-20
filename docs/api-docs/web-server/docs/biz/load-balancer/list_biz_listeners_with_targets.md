### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询业务下符合指定条件的监听器列表及监听器绑定的RS列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/listeners/with/targets/list

### 输入参数

| 参数名称         | 参数类型       | 必选 | 描述      |
|-----------------|--------------|-----|-----------|
| bk_biz_id       | int          | 是  | 业务ID     |
| vendor          | string       | 是  | 云厂商     |
| account_id      | string       | 是  | 账号ID     |
| region          | string       | 是  | 云地域     |
| rule_query_list | object array | 是  | 规则查询条件数组 |

#### rule_query_list

| 参数名称         | 参数类型       | 必选 | 描述                                                                                                                                                  |
|-----------------|--------------|------|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| region          | string       | 是   | 地域   |
| clb_vip_domains | string array | 是   | 负载均衡VIP或域名数组 |
| cloud_lb_ids    | string array | 是   | 负载均衡云ID数组      |
| protocol        | string       | 否   | 协议     |
| ports           | int array    | 否   | 端口数组  |
| rule_type       | string       | 是   | 监听器类型(枚举值:layer_4:四层监听器 layer_7:七层监听器) |
| domain          | string       | 否   | 域名      |
| url             | string       | 否   | URL      |
| inst_type       | string       | 是   | 后端类型(枚举值:CVM、ENI) |
| rs_ips          | string array | 是   | RSIP数组     |
| rs_ports        | int  array   | 否   | RS端口数组    |
| rs_weights      | int array    | 否   | RS权重数组    |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "bk_biz_id": 123,
  "vendor": "tcloud",
  "account_id": "xxxxxx",
  "rule_query_list":[
    {
      "region": "ap-nanjing",
      "clb_vip_domains": ["1.1.1.1"],
      "cloud_lb_ids": ["lb-xxxxxx"],
      "protocol": "",
      "ports": [111],
      "rule_type": "layer_7",
      "domain": "",
      "url": "",
      "inst_type": "CVM",
      "rs_ips": ["1.1.1.2"],
      "rs_ports": [101],
      "rs_weights": []
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
        "port": 1001,
        "rs_list": [
            {
                "id": "00000003",
                "ip": "1.1.1.1",
                "port": 202,
                "weight": 50,
                "inst_type": "CVM",
                "cloud_inst_id": "ins-00000004",
                "inst_name": "test-name",
                "target_group_id": "000000005",
                "rule_id": "00000007",
                "cloud_rule_id": "lbl-xxxxxxxx",
                "rule_type": "layer_4",
                "domain": "www.xxx.com",
                "url": "/path"
            }
        ]
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
| rs_list        | object array | 绑定的RS列表 |

#### data.details[n].rs_list

| 参数名称         | 参数类型       | 描述            |
|-----------------|--------------|-----------------|
| id              | string       | 目标ID           |
| ip              | string       | 目标IP           |
| port            | string       | 目标端口          |
| weight          | int          | 目标当前的权重     |
| inst_type       | string       | 后端类型(枚举值:CVM、ENI) |
| cloud_inst_id   | string       | 目标云实例ID       |
| inst_name       | string       | 云实例名称         |
| target_group_id | string       | 目标组ID          |
| rule_id         | string       | 规则ID            |
| cloud_rule_id   | string       | 云规则ID(4层为云监听器ID 7层为云规则ID)  |
| rule_type       | string       | 监听器类型(枚举值:layer_4:四层监听器 layer_7:七层监听器) |
| domain          | string       | 域名(仅7层监听器)   |
| url             | string       | URL(仅7层监听器)   |
