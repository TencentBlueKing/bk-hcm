### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：监听器删除。
- 该接口功能描述：业务下删除监听器, 支持跨CLB, 但不支持跨账号。

### URL

DELETE /api/v1/cloud/bizs/{bk_biz_id}/listeners/batch

### 输入参数

| 参数名称       | 参数类型         | 必选 | 描述                     |
|------------|--------------|----|------------------------|
| bk_biz_id  | int          | 是  | 业务ID                   |
| account_id | string       | 是  | 账号ID                   |
| ids        | string array | 是  | 监听器ID数组, 最大可传入1000个监听器 |

### 调用示例

```json
{
  "account_id": "00000001",
  "ids": [
    "00000001",
    "00000002"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "task_management_id": "xxxxxx"
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述    |
|---------|--------|---------|
| code    | int    | 状态码   |
| message | string | 请求信息 |
| data    | object | 响应数据 |


#### data参数说明

| 参数名称    | 参数类型   | 描述     |
|---------|--------|--------|
| task_management_id | string | 任务管理id |
