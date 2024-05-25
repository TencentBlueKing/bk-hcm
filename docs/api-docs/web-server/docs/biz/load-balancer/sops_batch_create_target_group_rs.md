### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下给标准运维插件中指定的目标组批量添加RS。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/sops/target_groups/targets/create

### 输入参数

| 参数名称          | 参数类型       | 必选 | 描述                     |
|------------------|--------------|------|-------------------------|
| bk_biz_id        | int          | 是   | 业务ID                   |
| account_id       | string       | 是   | 账号ID                   |
| rule_query_list  | object array | 是   | 规则查询列表，单次最多10个  |
| rs_ip            | string array | 是   | RS IP数组                |
| rs_port          | string array | 是   | RS 端口数组               |
| rs_weight        | int          | 是   | RS权重，取值范围：[0, 100] |
| rs_type          | string       | 是   | RS类型(枚举值：CVM)       |

#### rule_query_list

| 参数名称   | 参数类型   | 必选 | 描述                       |
|-----------|----------|------|---------------------------|
| region    | string   | 否   | 区域                       |
| protocol  | string   | 否   | 协议(UDP、TCP、HTTP、HTTPS) |
| domain    | string   | 否   | 域名                       |
| vip       | string   | 否   | 负载均衡的VIP               |
| vport     | string   | 否   | 监听器的默认端口             |

### 调用示例

```json
{
  "account_id": "xxxxxxxx",
  "rule_query_list": [
    {
      "region": "ap-nanjing",
      "protocol": "HTTPS",
      "domain": "www.xxxx.com"
    }
  ],
  "rs_ip": ["xxxxxx", "xxxxxx"],
  "rs_port": ["1001", "1002"],
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
    "flow_id": "xxxxxxxx"
  }
}
```

### 响应参数说明

| 参数名称  | 参数类型  | 描述    |
|---------|----------|---------|
| code    | int      | 状态码   |
| message | string   | 请求信息 |
| data    | object   | 响应数据 |

#### data

| 参数名称  | 参数类型 | 描述    |
|----------|--------|---------|
| flow_id  | string | 任务id   |
