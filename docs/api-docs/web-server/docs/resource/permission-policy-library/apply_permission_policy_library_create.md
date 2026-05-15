### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源接入-云资源-云厂商配置。
- 该接口功能描述：应用权限策略库（创建）。将指定权限策略库应用到目标二级账号，逐个执行，返回每个账号的执行结果（成功/失败+原因）。

### URL

POST /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}/apply

### 路径参数

| 参数名称   | 参数类型   | 必选 | 描述                        |
|--------|--------|----|---------------------------|
| vendor | string | 是  | 云厂商（枚举值：tcloud） |
| id     | string | 是  | 权限策略库ID                   |

### 输入参数

| 参数名称       | 参数类型       | 必选 | 描述                              |
|--------------|--------------|----|---------------------------------|
| account_ids  | string array | 是  | 目标二级账号ID列表，长度限制100              |

### 调用示例

```json
{
  "account_ids": [
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
    "results": [
      {
        "account_id": "00000001",
        "status": "success"
      },
      {
        "account_id": "00000002",
        "status": "success"
      },
      {
        "account_id": "00000003",
        "status": "failed",
        "reason": "该二级账号已应用此权限策略库"
      }
    ]
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

| 参数名称    | 参数类型         | 描述             |
|---------|--------------|----------------|
| results | object array | 每个二级账号的执行结果列表  |

#### result[n]

| 参数名称       | 参数类型   | 描述                                    |
|------------|--------|---------------------------------------|
| account_id | string | 二级账号ID                                |
| status     | string | 执行状态（枚举值：success-成功，failed-失败）        |
| reason     | string | 失败原因，仅当 status 为 failed 时返回           |
