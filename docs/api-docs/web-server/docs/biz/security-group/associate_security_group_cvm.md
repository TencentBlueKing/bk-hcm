### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务-IaaS资源操作。
- 该接口功能描述：安全组关联主机（仅支持：tcloud、aws、huawei）。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/security_groups/associate/cvms

### 输入参数

| 参数名称              | 参数类型   | 必选    | 描述      |
|-------------------|--------|-------|---------|
| bk_biz_id         | int64  | 是     | 业务ID    |
| security_group_id | string | 是     | 安全组ID   |
| cvm_id            | string | 是     | 主机ID    |

### 调用示例

```json
{
  "security_group_id": "00001111",
  "cvm_id": "00001112"
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
