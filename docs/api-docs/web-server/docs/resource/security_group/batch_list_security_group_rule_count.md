### 描述

- 该接口提供版本：v1.8.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：批量查询安全组规则数量。

### URL

POST /api/v1/cloud/security_groups/rules/count

### 输入参数

| 参数名称               | 参数类型         | 必选 | 描述                  |
|--------------------|--------------|----|---------------------|
| security_group_ids | array string | 是  | 安全组ID, 单次请求限制500个id |

### 请求示例

```json
{
  "security_group_ids": ["00000001", "00000002"]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "00000001": 12,
    "00000002": 0
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型           | 描述   |
|---------|----------------|------|
| code    | int32          | 状态码  |
| message | string         | 请求信息 |
| data    | map[string]int | 响应数据 |

#### data map键值说明

key: 安全组ID
value: 安全组规则数量
