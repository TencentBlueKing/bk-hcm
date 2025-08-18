### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下批量移除RS, 支持跨CLB, 但不支持跨账号。

### URL

DELETE /api/v1/cloud/bizs/{bk_biz_id}/target_groups/targets/batch

### 输入参数

| 参数名称          | 参数类型         | 必选 | 描述                  |
|---------------|--------------|----|---------------------|
| bk_biz_id     | int          | 是  | 业务ID                |
| account_id       | string       | 是   | 账号ID                |
| target_ids | string array | 是  | RS列表, 最大限制传入5000个rs |



### 调用示例

```json
{
  "account_id": "00000001",
  "target_ids": ["00000001"]
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
