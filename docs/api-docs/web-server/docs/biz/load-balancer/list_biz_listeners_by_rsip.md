### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：列出业务下的指定RSIP的监听器列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/listeners/list/by/rsip

### 输入参数

| 参数名称      | 参数类型       | 必选 | 描述      |
|--------------|--------------|-----|-----------|
| bk_biz_id    | int          | 是  | 业务ID     |
| vendor       | string       | 是  | 云厂商     |
| account_id   | string       | 是  | 账号ID     |
| region       | string       | 是  | 云地域     |
| ip_addresses | string array | 是  | 负载均衡VIP或域名数组 |
| cloud_lb_ids | string array | 是  | 负载均衡ID数组       |
| protocol     | string       | 否  | 协议     |
| port         | int          | 否  | 端口     |
| rule_type    | string       | 是  | 监听器类型(枚举值:layer_4:四层监听器 layer_7:七层监听器) |
| domain       | string       | 否  | 域名     |
| url          | string       | 否  | URL      |
| inst_type    | string       | 是  | 后端类型(枚举值:CVM、ENI) |
| rs_ip        | string       | 是  | RSIP     |
| rs_port      | int          | 否  | RS端口    |
| page         | object       | 是  | 分页设置   |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "bk_biz_id": 123,
  "vendor": "tcloud",
  "account_id": "xxxxxx",
  "region": "ap-nanjing",
  "ip_addresses": ["1.1.1.1"],
  "cloud_lb_ids": ["lb-xxxxxx"],
  "protocol": "",
  "port": 0,
  "rule_type": "layer_7",
  "domain": "",
  "url": "",
  "inst_type": "CVM",
  "rs_ip": "1.1.1.2",
  "rs_port": "",
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "lbl_total_num": 6,
    "batch_num": 3,
    "details": [
      {
        "cloud_lb_id": "lb-00000001",
        "cloud_lbl_id": "lbl-00000001",
        "protocol": "HTTPS",
        "port": 1001,
        "rs_num": 1
      },
      {
        "cloud_lb_id": "lb-00000002",
        "cloud_lbl_id": "lbl-00000002",
        "protocol": "HTTPS",
        "port": 1002,
        "rs_num": 2
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

| 参数名称       | 参数类型 | 描述                            |
|---------------|--------|---------------------------------|
| lbl_total_num | int    | 当前规则能匹配到的监听器总数量       |
| batch_num     | int    | 想获取所有监听器需要拉取的批次数      |
| details       | array object | 查询返回的数据               |

#### data.details[n]

| 参数名称        | 参数类型  | 描述            |
|----------------|---------|-----------------|
| cloud_lb_id    | string  | 负载均衡云实例ID  |
| cloud_lbl_id   | string  | 监听器云ID       |
| protocol       | string  | 监听器协议       |
| port           | int     | 监听器端口       |
| rs_num         | int     | 绑定的RS数量     |
