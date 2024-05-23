### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询异步任务的操作记录指定子任务的详情。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/audits/async_task/list

### 输入参数

| 参数名称   | 参数类型 | 必选 | 描述            |
|-----------|--------|------|----------------|
| bk_biz_id | int    | 是   | 业务ID          |
| audit_id  | int    | 是   | 操作记录ID       |
| flow_id   | string | 是   | 任务ID          |
| action_id | string | 是   | 子任务ID        |

### 调用示例

```json
{
    "audit_id": 1001,
    "flow_id": "00000001",
    "action_id": 1
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
      "flow": {
        "id": "00000001",
        "name": "add_rs",
        "state": "success",
        "reason": {
            "message": "some tasks failed to be executed"
        },
        "share_data": {
            "lb_id": "00000001"
        },
        "creator": "Jim",
        "reviser": "Jim",
        "created_at": "2023-02-12T14:47:39Z",
        "updated_at": "2023-02-12T14:55:40Z"
      },
      "tasks": [
        {
          "id": "00000004",
          "flow_id": "00000001",
          "flow_name": "add_rs",
          "action_name": "add_rs",
          "params": {
            "vendor": "tcloud",
            "targets": []
          },
          "state": "failed",
          "reason": {
              "message": "some tasks failed to be executed"
          },
          "creator": "sync-timing-admin",
          "reviser": "sync-timing-admin",
          "created_at": "2023-02-25T18:28:46Z",
          "updated_at": "2023-02-27T19:14:33Z"
        }
      ]
    }
}
```

### 响应参数说明

| 参数名称 | 参数类型       | 描述   |
|---------|--------------|--------|
| code    | int          | 状态码  |
| message | string       | 请求信息 |
| data    | object       | 响应数据 |

#### data.flow

| 参数名称         | 参数类型 | 描述                                               |
|----------------|---------|----------------------------------------------------|
| id             | string  | 任务ID                                              |
| name           | string  | 任务名称                                            |
| state          | string  | 任务状态                                            |
| reason         | json    | 任务失败原因                                         |
| creator        | string  | 任务创建者                                           |
| reviser        | string  | 任务最后一次修改的修改者                                |
| created_at     | string  | 任务创建时间，标准格式：2006-01-02T15:04:05Z            |
| updated_at     | string  | 任务最后一次修改时间，标准格式：2006-01-02T15:04:05Z     |

#### data.tasks

| 参数名称         | 参数类型 | 描述                                                 |
|----------------|---------|------------------------------------------------------|
| id             | string  | 子任务ID                                              |
| action_name    | string  | 子任务名称                                             |
| flow_id        | string  | 任务ID                                                |
| flow_name      | string  | 任务名称                                               |
| params         | json    | 子任务参数                                             |
| state          | string  | 子任务状态                                             |
| reason         | json    | 子任务失败原因                                          |
| creator        | string  | 子任务创建者                                           |
| reviser        | string  | 子任务最后一次修改的修改者                                |
| created_at     | string  | 子任务创建时间，标准格式：2006-01-02T15:04:05Z            |
| updated_at     | string  | 子任务最后一次修改时间，标准格式：2006-01-02T15:04:05Z     |

#### data.flow.reason

| 参数名称   | 参数类型 | 描述    |
|----------|---------|---------|
| message  | string  | 任务失败原因 |

#### data.flow.share_data

| 参数名称   | 参数类型 | 描述      |
|----------|---------|-----------|
| lb_id    | string  | 负载均衡ID |

#### data.tasks[n].reason

| 参数名称   | 参数类型 | 描述         |
|----------|---------|--------------|
| message  | string  | 子任务失败原因 |


#### data.tasks[n].params（针对负载均衡操作的参数）

| 参数名称 | 参数类型         | 描述   |
|---------|----------------|-------|
| vendor  | string         | 云厂商 |
| targets | object array   | RS列表 |


#### data.tasks[n].params[n].targets（针对负载均衡操作的参数）

| 参数名称              | 参数类型       | 描述                      |
|----------------------|--------------|--------------------------|
| account_id           | string       | 账号ID                    |
| inst_type            | string       | 实例类型                   |
| inst_name            | string       | 实例名称                   |
| cloud_inst_id        | string       | 云实例ID                   |
| port                 | int          | 端口                       |
| weight               | int          | 权重                       |
| new_port             | int          | 新端口，仅限批量修改端口时有值 |
| new_weight           | int          | 新权重，仅限批量修改权重时有值 |
| target_group_id      | string       | 目标组ID                   |
| private_ip_addresses | string array | 内网IP地址                  |
| public_ip_addresses  | string array | 外网IP地址                  |
| cloud_vpc_ids        | string array | 云端VpcID                  |
| zone                 | string       | 可用区                     |
