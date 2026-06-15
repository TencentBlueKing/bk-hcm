### 描述

- 该接口提供版本：v1.8.11+。
- 该接口所需权限：资源接入-云资源-云厂商配置。
- 该接口功能描述：删除权限策略库。若关联了云权限模板，则不允许删除。

### URL

DELETE /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}

### 路径参数

| 参数名称   | 参数类型   | 必选 | 描述                                    |
|--------|--------|----|---------------------------------------|
| vendor | string | 是  | 云厂商（枚举值：tcloud） |
| id     | string | 是  | 策略库ID                                 |

### 调用示例

```
DELETE /api/v1/cloud/vendors/tcloud/permission_policy_libraries/00000001
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述       |
|---------|--------|----------|
| code    | int32  | 状态码      |
| message | string | 请求信息     |
| data    | object | 响应数据（为空） |
