### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务-IaaS资源操作。
- 该接口功能描述：更新安全组（aws不支持更新）。

### URL

PATCH /api/v1/cloud/bizs/{bk_biz_id}/security_groups/{id}

### 输入参数

| 参数名称       | 参数类型   | 必选    | 描述                             |
|------------|--------|-------|--------------------------------|
| bk_biz_id  | int64  | 是     | 业务ID                           |
| id         | string | 是     | 安全组ID                          |
| name       | string | 否     | 安全组名称(资源所属供应商如果为azure，名称不允许修改) |
| memo       | string | 否     | 备注                             |

### 调用示例

```json
{
  "name": "update",
  "memo": "update security group"
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
