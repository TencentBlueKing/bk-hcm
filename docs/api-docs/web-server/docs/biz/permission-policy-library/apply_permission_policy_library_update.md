### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务-权限策略库操作。
- 该接口功能描述：应用权限策略库（更新）。为指定权限模版创建审批单，审批通过后将权限策略库的最新策略内容同步更新到权限模板。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/apply_permission_policy_library_update

### 路径参数

| 参数名称      | 参数类型   | 必选 | 描述                        |
|-----------|--------|----|---------------------------|
| bk_biz_id | int64  | 是  | 业务ID                      |
| vendor    | string | 是  | 云厂商（枚举值：tcloud） |

### 输入参数

| 参数名称                    | 参数类型         | 必选 | 描述                                  |
|-------------------------|--------------|----|------------------------------------|
| policy_library_id       | string       | 是  | 权限策略库ID                             |
| permission_template_ids | string array | 是  | 目标权限模版ID列表，需已应用当前策略库，长度限制100        |

### 调用示例

```json
{
  "policy_library_id": "00000001",
  "permission_template_ids": [
    "00000001",
    "00000002",
    "00000003"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "ids": ["00000001", "00000002", "00000003"]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称 | 参数类型         | 描述      |
|------|--------------|---------|
| ids  | string array | 审批单据ID数组 |
