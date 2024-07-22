### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下批量修改RS权重。

### URL

PATCH /api/v1/cloud/bizs/{bk_biz_id}/sops/target_groups/targets/weight

### 输入参数

| 参数名称          | 参数类型       | 必选 | 描述                    |
|------------------|--------------|------|------------------------|
| bk_biz_id        | int          | 是   | 业务ID                  |
| account_id       | string       | 是   | 账号ID                  |
| rule_query_list  | object array | 是   | 规则查询列表，单次最多10个 |
| rs_weight        | int          | 是   | 新权重,取值范围：[0, 100] |

#### rule_query_list

| 参数名称     | 参数类型  | 必选 | 描述                     |
|----------|-------|----|------------------------|
| region   | string | 是  | 区域                     |
| vip      | []string | 否  | 负载均衡的VIP               |
| vport    | []int | 否  | 监听器的默认端口               |
| rs_ip    | []string      | 是  | real server的ip         |
| rs_type  | string | 是  | real server的类型         |
| protocol | string | 否  | 协议(UDP、TCP、HTTP、HTTPS) |
| domain   | string | 否  | 域名                     |


### 调用示例

```json
{
  "account_id": "xxxxxxxx",
  "rule_query_list": [
    {
      "region": "ap-nanjing",
      "rs_ip": [
        "xxx.xxx.xxx.xxx",
        "zzz.zzz.zzz.zzz"
      ],
      "rs_type": "CVM",
      "protocol": "HTTPS",
      "domain": "www.xxxx.com"
    }
  ],
  "rs_weight": 66
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
    },
    {
      "flow_id": "xxxxxxxx"
    }
  ]
}
```

### 响应参数说明

| 参数名称  | 参数类型      | 描述    |
|---------|-----------|---------|
| code    | int       | 状态码   |
| message | string    | 请求信息 |
| data    | object array  | 响应数据 |

#### data[n]

| 参数名称    | 参数类型 | 描述   |
|---------|--------|------|
| flow_id | string | 任务id |
