### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下移除标准运维中指定的负载均衡规则。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/sops/rule/offline

### 输入参数

| 参数名称            | 参数类型         | 必选 | 描述                |
|-----------------|--------------|----|-------------------|
| bk_biz_id       | int          | 是  | 业务ID              |
| account_id      | string       | 是  | 账号ID              |
| rule_query_list | object array | 是  | 规则查询列表，最少1个，最多10个 |

#### rule_query_list

| 参数名称     | 参数类型     | 必选 | 描述                        |
|----------|----------|----|---------------------------|
| region   | string   | 是  | 地域，必填                     |
| vip      | string[] | 否  | 负载均衡的VIP                  |
| vport    | int[]    | 否  | 监听器的默认端口                  |
| rs_ip    | string[] | 否  | RS IP，长度必须大于0             |
| rs_type  | string   | 否  | RS类型，仅支持 CVM 或 ENI        |
| protocol | string[] | 否  | 协议，仅支持 UDP、TCP、HTTP、HTTPS |
| domain   | string[] | 否  | 域名，七层协议下必填，非七层协议下不可填写     |
| url      | string[] | 否  | URL，七层协议下必填，非七层协议下不可填写    |

### 调用示例

```json
{
  "account_id": "xxxxxxxx",
  "rule_query_list": [
    {
      "region": "ap-nanjing",
      "protocol": ["HTTPS"],
      "domain": ["www.example.com"],
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
  "data": {
    "flow_state_results": [
      {
        "flow_id": "xxx",
        "state": "xxx"
      }
    ]
  }
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
