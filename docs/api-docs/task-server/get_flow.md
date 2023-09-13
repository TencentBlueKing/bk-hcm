### 描述

- 该接口提供版本：v9.9.9+
- 该接口所需权限：
- 该接口功能描述：查询单个flow信息

### URL

GET /api/v1/task/async/flows/{flow_id}

#### 路径参数说明
| 参数名称 | 参数类型     | 必选 | 描述      |
|------|----------|----|---------|
| flow_id   | string   | 是  | flow id  |

### 调用示例
查询ID是0000000p的任务流信息

#### 返回示例
```json
{
    "code": 0,
    "message": "",
    "data": {
        "id": "0000000p",
        "name": "first_test",
        "state": "pending",
        "tasks": [
            {
                "id": "0000002p",
                "flow_id": "0000000p",
                "flow_name": "first_test",
                "action_name": "test_CreateSG",
                "params": "{}",
                "retry_count": 0,
                "timeout_secs": 10,
                "depend_on": [],
                "state": "pending",
                "memo": "",
                "reason": "{}"
            },
            {
                "id": "0000002q",
                "flow_id": "0000000p",
                "flow_name": "first_test",
                "action_name": "test_CreateSubnet",
                "params": "{}",
                "retry_count": 0,
                "timeout_secs": 10,
                "depend_on": [
                    "0000002p"
                ],
                "state": "pending",
                "memo": "",
                "reason": "{}"
            },
            {
                "id": "0000002r",
                "flow_id": "0000000p",
                "flow_name": "first_test",
                "action_name": "test_CreateVpc",
                "params": "{}",
                "retry_count": 0,
                "timeout_secs": 10,
                "depend_on": [
                    "0000002p"
                ],
                "state": "pending",
                "memo": "",
                "reason": "{}"
            },
            {
                "id": "0000002s",
                "flow_id": "0000000p",
                "flow_name": "first_test",
                "action_name": "test_CreateCvm",
                "params": "{}",
                "retry_count": 0,
                "timeout_secs": 10,
                "depend_on": [
                    "0000002q",
                    "0000002r"
                ],
                "state": "pending",
                "memo": "",
                "reason": "{}"
            }
        ],
        "memo": "",
        "reason": "{}",
        "creator": "hcm-backend-async",
        "reviser": "hcm-backend-async",
        "created_at": "2023-08-30 11:34:44 +0000 UTC",
        "updated_at": "2023-08-30 11:34:44 +0000 UTC"
    }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data[n]

| 参数名称           | 参数类型         | 描述                                                               |
|----------------|--------------|------------------------------------------------------------------|
| id             | string       | 任务流ID                                                             |
| name           | string       | 任务流名称                                                               |
| state          | string       | 任务流状态                                                              |
| tasks          | object array | 任务集合                                                            |
| memo           | string       | 备注                                                               |
| reason         | string       | 失败等原因                                                               |
| creator        | string       | 创建者                                                              |
| reviser        | string       | 更新者                                                              |
| created_at     | string       | 创建时间，标准格式：2006-01-02T15:04:05Z                                   |
| updated_at     | string       | 更新时间，标准格式：2006-01-02T15:04:05Z                                   |

#### tasks[n]

| 参数名称           | 参数类型         | 描述                                                               |
|----------------|--------------|------------------------------------------------------------------|
| id             | string       | 任务ID                                                             |
| flow_id        | string       | 任务流ID                                                               |
| flow_name      | string       | 任务流名称                                                               |
| action_name    | string       | 执行动作名称                                                               |
| state          | string       | 任务流状态                                                              |
| params         | object       | 参数信息                                                              |
| retry_count    | int          | 重试次数                                                              |
| timeout_secs   | int          | 超时时间                                                              |
| depend_on      | string array | 依赖任务集合                                                            |
| memo           | string       | 备注                                                               |
| reason         | string       | 失败等原因                                                               |