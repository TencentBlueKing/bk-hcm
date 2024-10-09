### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询任务详情状态数量。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/task_details/state/count

### 输入参数

| 参数名称       | 参数类型   | 必选 | 描述              |
|------------|--------|----|-----------------|
| ids        | string array    | 是  | 任务id列表，最大长度为100 |

### 调用示例

```json
{
  "ids": ["0000001","0000002"]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "details": [
      {
        "id": "0000001",
        "success": 1,
        "failed": 1,
        "init": 1,
        "running": 1,
        "cancel": 0,
        "total": 4
      },
      {
        "id": "0000002",
        "success": 1,
        "failed": 1,
        "init": 1,
        "running": 1,
        "cancel": 0,
        "total": 4
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称    | 参数类型   | 描述             |
|---------|--------|----------------|
| details | array  | 查询返回的数据        |

#### data.details[n]

| 参数名称    | 参数类型   | 描述    |
|---------|--------|-------|
| id      | string | 任务ID  |
| success | int    | 成功数量  |
| failed  | int    | 失败数量  |
| init    | int    | 待执行数量 |
| running | int    | 运行中数量 |
| cancel  | int    | 取消数量  |
| total   | int    | 总数    |
