### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下给标准运维插件中指定的目标组批量添加RS。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/sops/target_groups/targets/create

### 输入参数

| 参数名称            | 参数类型         | 必选 | 描述                    |
|-----------------|--------------|----|-----------------------|
| bk_biz_id       | int          | 是  | 业务ID                  |
| account_id      | string       | 是  | 账号ID                  |
| rule_query_list | object array | 是  | 规则查询列表，最少1个，最多10个     |
| rs_ip           | string array | 是  | RS IP数组，必填，且长度必须大于0   |
| rs_port         | int array    | 是  | RS 端口数组，必填，且长度必须大于0   |
| rs_weight       | int64        | 是  | RS权重，取值范围：[0, 100]    |
| rs_type         | string       | 是  | RS类型，必填，仅支持 CVM 或 ENI |

#### rule_query_list

| 参数名称     | 参数类型     | 必选 | 描述                                   |
|----------|----------|----|--------------------------------------|
| region   | string   | 是  | 地域，必填                                |
| vip      | string[] | 否  | 负载均衡的VIP                             |
| vport    | int[]    | 否  | 监听器的默认端口                             |
| rs_ip    | string[] | 是  | RS IP，必填，且长度必须大于0                    |
| rs_type  | string   | 是  | RS类型，必填，仅支持 CVM 或 ENI                |
| protocol | string   | 否  | 协议，仅支持 UDP、TCP、HTTP、HTTPS。七层协议下，域名必填 |
| domain   | string   | 否  | 域名，七层协议下必填，非七层协议下不可填写                |
| url      | string[] | 否  | URL，七层协议下必填                          |

### 调用示例

```json
{
  "account_id": "xxxxxxxx",
  "rule_query_list": [
    {
      "region": "ap-nanjing",
      "protocol": "HTTPS",
      "domain": "www.example.com",
      "vip": ["xxx.xxx.xxx.xxx"],
      "vport": [666],
      "rs_ip": ["xxx.xxx.xxx.xxx"],
      "rs_type": "CVM"
    }
  ],
  "rs_ip": ["xxx.xxx.xxx.xxx", "xxx.xxx.xxx.xxx"],
  "rs_port": [666, 666],
  "rs_weight": 10,
  "rs_type": "CVM"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "flow_id": "xxxxxxxx"
    }
  ]
}
```

### 响应参数说明

| 参数名称  | 参数类型         | 描述    |
|---------|--------------|---------|
| code    | int          | 状态码   |
| message | string       | 请求信息 |
| data    | object array | 响应数据 |

#### data

| 参数名称    | 参数类型     | 描述   |
|---------|----------|------|
| flow_id | string   | 任务id |
