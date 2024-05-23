### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：复制flow参数重新执行（仅支持：tcloud）。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/{lb_id}/async_flows/clone

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述      |
|-----------|--------|----|---------|
| bk_biz_id | int64  | 是  | 业务ID    |
| lb_id     | string | 是  | 负载均衡id  |
| flow_id   | string | 是  | flow id |

### 调用示例

```json
{
  "flow_id": "00001112"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "flow_id": "cccddd"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 结果信息 |

#### data

| 参数名称    | 参数类型   | 描述          |
|---------|--------|-------------|
| flow_id | string | 新生成的flow_id |

