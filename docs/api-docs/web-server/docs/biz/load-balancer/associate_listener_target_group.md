### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：监听器操作。
- 该接口功能描述：给指定的监听器，关联目标组（仅支持：tcloud）。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/listeners/associate/target_group

### 输入参数

| 参数名称            | 参数类型  | 必选 | 描述          |
|--------------------|---------|------|--------------|
| bk_biz_id          | int     | 是   | 业务ID        |
| listener_id        | string  | 是   | 监听器的ID     |
| listener_rule_id   | string  | 是   | 监听器规则的ID  |
| target_group_id    | string  | 是   | 目标组的ID     |

### 调用示例

```json
{
  "listener_id": "00000001",
  "listener_rule_id": "00000002",
  "target_group_id": "00001112"
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
