### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下移除标准运维中指定的负载均衡规则。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/sops/rule/offline

### 输入参数

| 参数名称          | 参数类型       | 必选 | 描述     |
|------------------|--------------|------|--------|
| bk_biz_id        | int          | 是   | 业务ID   |
| account_id       | string       | 是   | 账号ID   |
| rule_query_list  | object array | 是   | 规则查询列表 |

#### rule_query_list

| 参数名称     | 参数类型    | 必选   | 描述    |
|----------|---------|------|-------|
| region   | string  | 否    | 区域    |
| vip      | string  | 否    | 负载均衡IP |
| vport    | string  | 否    | 监听器端口 |
| rs_ip    | string  | 是    | 后端实例IP |
| rs_type  | string  | 是    | 后端实例端口 |           
| protocol | string  | 否    | 协议    |
| domain   | string  | 否    | 域名    |


### 调用示例

```json
{
  "account_id": "xxx",
  "rule_query_list": [
    {
      "region": "xxx",
      "vip": "xxx.xxx.xxx.xxx",
      "vport": "5565",
      "rs_ip": "xxx.xxx.xxx.xxx",
      "rs_type": "CVM",
      "protocol": "HTTPS",
      "domain": "www.xxx.com"
    }
  ]
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

| 参数名称  | 参数类型  | 描述    |
|---------|----------|---------|
| code    | int      | 状态码   |
| message | string   | 请求信息 |
| data    | object   | 响应数据 |

#### data

| 参数名称               | 参数类型         | 描述    |
|--------------------|--------------|---------|
| flow_state_results | object array | 任务id   |

#### flow_state_results

| 参数名称    | 参数类型   | 描述      |
|---------|--------|---------|
| flow_id | string | 异步任务流ID |
| state   | string | 异步任务流状态 |
