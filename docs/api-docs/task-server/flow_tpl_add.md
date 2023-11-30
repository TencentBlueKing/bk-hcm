### 描述

- 该接口提供版本：v1.2.1+
- 该接口所需权限：
- 该接口功能描述：根据模板添加异步任务流

### URL

POST /api/v1/task/async/flows/tpls/add

### 输入参数

| 参数名称       | 参数类型          | 必选 | 描述   |
|------------|---------------|----|------|
| flow_name  | string        | 是  | 模板名称 |
| parameters | object  array | 否  | 参数集合 |

### 调用示例

```json
{
  "flow_name": "first_test"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": "0000000p"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述    |
|---------|--------|-------|
| code    | int32  | 状态码   |
| message | string | 请求信息  |
| data    | string | 任务流ID |