
### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：账单管理。
- 该接口功能描述：确认调账明细，确认后不可再修改、删除。

### URL

POST /api/v1/account/bills/adjustment_items/confirm

### 输入参数

| 参数名称 | 参数类型         | 必选 | 描述     |
|------|--------------|----|--------|
| ids  | string array | 是  | 调账明细列表 |

### 调用示例

```json
{
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
  "message": "ok"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |