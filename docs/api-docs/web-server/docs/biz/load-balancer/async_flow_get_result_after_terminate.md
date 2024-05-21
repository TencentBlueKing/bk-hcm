### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：获取任务终止后rs的状态，仅支持任务终止后五分钟内

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/{lb_id}/async_tasks/result

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
  "data": [
    {
      "task_id": "0000001",
      "target_group_id": "xxxxx",
      "target_list": [
        {
          "account_id": "xx",
          "inst_type": "CVM",
          "cloud_inst_id": "yyy",
          "port": 80,
          "weight": 10
        }
      ]
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述   |
|---------|--------------|------|
| code    | int          | 状态码  |
| message | string       | 请求信息 |
| data    | detail array | 结果信息 |

#### detail

| 参数名称            | 参数类型         | 描述                |
|-----------------|--------------|-------------------|
| task_id         | string       | task id           |
| status          | string       | 结果 succeed/failed |
| target_group_id | string       | 目标组ID             |
| target_list     | target array | 目标详情              |

#### target

| 参数名称          | 参数类型   | 描述    |
|---------------|--------|-------|
| account_id    | string | 账号ID  |
| inst_type     | string | 实例类型  |
| cloud_inst_id | string | 云实例ID |
| port          | int    | 端口    |
| weight        | int    | 权重    |
