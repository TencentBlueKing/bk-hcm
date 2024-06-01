### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询异步任务Flow详情。

### URL

GET /api/v1/cloud/async_task/flows/{id}

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述      |
|---------|--------|------|----------|
| id      | string | 是   | 异步任务ID |

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000001",
    "name": "xxxxxx",
    "state": "failed",
    "reason": null,
    "creator": "admin",
    "reviser": "admin",
    "created_at": "2024-01-01T19:31:58Z",
    "updated_at": "2024-01-01T19:32:40Z"
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

| 参数名称     | 参数类型  | 描述                                   |
|------------|----------|----------------------------------------|
| id         | string   | 异步任务ID                              |
| name       | string   | 异步任务名称                             |
| state      | string   | 任务状态(初始状态:init 等待中:pending 待调度:scheduled 执行中:running 已取消:canceled 成功:success 失败:failed) |
| reason     | string   | 任务失败原因                             |
| creator    | string   | 创建者                                  |
| reviser    | string   | 修改者                                  |
| created_at | string   | 创建时间，标准格式：2006-01-02T15:04:05Z   |
| updated_at | string   | 修改时间，标准格式：2006-01-02T15:04:05Z   |