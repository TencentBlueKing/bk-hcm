### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务-IaaS资源操作。
- 该接口功能描述：主机批量关联安全组(仅支持: tcloud、aws)。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/cvms/{cvm_id}/security_groups/batch_associate

### 输入参数

| 参数名称               | 参数类型         | 必选 | 描述                                                    |
|--------------------|--------------|----|-------------------------------------------------------|
| bk_biz_id          | int64        | 是  | 业务ID                                                  |
| cvm_id             | string       | 是  | 主机ID                                                  |
| security_group_ids | string array | 是  | 安全组ID, 前端排序后的安全组顺序, 云上接口会根据这个顺序覆盖更新, 最小传入1个, 最大传入500个 |

### 调用示例

```json
{
  "security_group_ids": ["sg-123", "sg-456"]
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

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| code     | int32    | 状态码   |
| message  | string   | 请求信息 |
