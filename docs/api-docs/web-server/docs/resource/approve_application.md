### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：
- 该接口功能描述：Itsm回调接口。

### URL

POST /api/v1/cloud/applications/approve

### 输入参数

| 参数名称           | 参数类型     | 必选   | 描述    |
|----------------|----------|------|-------|
| sn             | string   | 是    | 序列号   |
| current_status | string   | 是    | 当前状态  |
| approve_result | bool     | 是    | 批准结果  |
| token          | string   | 是    | Token |

### 调用示例
```json
{
  "sn": "",
  "current_status": "pending",
  "approve_result": true,
  "token": "xxxxxxxxxx"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": ""
}
```

### 响应参数说明
| 参数名称       | 参数类型   | 描述   |
|------------|--------|------|
| code       | int32  | 状态码  |
| message    | string | 请求信息 |
