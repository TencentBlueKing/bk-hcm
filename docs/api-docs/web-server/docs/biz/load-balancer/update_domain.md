### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：域名更新。
- 该接口功能描述：业务下更新域名。

### URL

PATCH /api/v1/cloud/bizs/{bk_biz_id}/listeners/{lbl_id}/domains

### 输入参数

| 参数名称          | 参数类型       | 必选 | 描述              |
|------------------|--------------|------|------------------|
| bk_biz_id        | int          | 是   | 业务ID            |
| lbl_id           | string       | 是   | 监听器ID          |
| domain           | string       | 是   | 旧域名            |
| new_domain       | string       | 是   | 新域名            |

### 调用示例

```json
{
  "domain": "www.old.com",
  "new_domain": "www.new.com"
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

| 参数名称  | 参数类型  | 描述    |
|---------|----------|---------|
| code    | int      | 状态码   |
| message | string   | 请求信息 |
