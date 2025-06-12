### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：。
- 该接口功能描述：租户初始化接口。仅限内部后台调用，通过X-Bk-Tenant-Id头指定租户id。

### URL

POST /api/v1/cloud/admin/system/tenant/init

### 输入参数

### 调用示例

```json
 
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "message": "tenant create success, 00000001"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述    |
|---------|--------|-------|
| code    | int32  | 状态码   |
| message | string | 请求信息  |
| data    | result | 初始化结果 |

#### result

| 参数名称    | 参数类型   | 描述     |
|---------|--------|--------|
| message | string | 结果提示信息 |
