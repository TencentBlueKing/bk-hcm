### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询负载均衡状态锁定详情。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/{id}/lock/status

### 输入参数

| 参数名称   | 参数类型 | 必选 | 描述       |
|-----------|--------|------|-----------|
| bk_biz_id | string | 是   | 业务id     |
| id        | string | 是   | 负载均衡id  |

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "res_id": "00000001",
    "res_type": "xxxxxx",
    "flow_id": "xxxxxx",
    "status": "executing"
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

| 参数名称  | 参数类型  | 描述                                     |
|----------|----------|-----------------------------------------|
| res_id   | string   | 当前锁定的资源ID                          |
| res_type | string   | 当前锁定的资源类型(load_balancer:负载均衡)  |
| flow_id  | string   | 当前锁定的任务ID                          |
| status   | string   | 锁定状态(锁定中:executing 未锁定:success)  |
