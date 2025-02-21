### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：批量查询安全组使用业务负责人列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/security_groups/usage_biz_maintainers/list

### 输入参数

| 参数名称               | 参数类型         | 必选 | 描述              |
|--------------------|--------------|----|-----------------|
| bk_biz_id          | int64        | 是  | 业务ID            |
| security_group_ids | array string | 是  | 安全组ID,最大可传入500个 |

### 请求示例

```json
{
  "security_group_ids": ["00000001", "00000002"]
}
```

- 已分配的安全组无法查询

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": "00000001",
      "managers": "zhangsan",
      "details": [
        {
          "bk_biz_id": 1,
          "bk_biz_name": "业务1",
          "bk_biz_maintainer": "zhangsan"
        },
        {
          "bk_biz_id": 2,
          "bk_biz_name": "业务2",
          "bk_biz_maintainer": "zhangsan"
        }
      ]
    },
    {
      "id": "00000002",
      "managers": "zhangsan",
      "details": [
        {
          "bk_biz_id": 1,
          "bk_biz_name": "业务1",
          "bk_biz_maintainer": "zhangsan"
        }
      ]
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | array  | 响应数据 |

#### data[n]

| 参数名称    | 参数类型     | 描述      |
|---------|----------|---------|
| id      | string   | 安全组ID   |
| manager | []string | 账号负责人列表 |
| details | array    | 使用业务详情  |

#### details[n]

| 参数名称              | 参数类型   | 描述   |
|-------------------|--------|------|
| bk_biz_id         | int32  | 使用业务 |
| bk_biz_name       | string | 业务名称 |
| bk_biz_maintainer | string | 运维人员 |

