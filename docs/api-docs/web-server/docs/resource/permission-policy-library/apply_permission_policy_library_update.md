### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源接入-云资源-云厂商配置。
- 该接口功能描述：应用权限策略库（更新）。将指定权限策略库的最新策略内容同步更新到已应用的目标权限模版，逐个执行，返回每个模版的执行结果（成功/失败+原因）。

### URL

PUT /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}/apply

### 路径参数

| 参数名称   | 参数类型   | 必选 | 描述                        |
|--------|--------|----|---------------------------|
| vendor | string | 是  | 云厂商（枚举值：tcloud） |
| id     | string | 是  | 权限策略库ID                   |

### 输入参数

| 参数名称                    | 参数类型         | 必选 | 描述                              |
|-------------------------|--------------|----|---------------------------------|
| permission_template_ids | string array | 是  | 目标权限模版ID列表，长度限制100              |

### 调用示例

```json
{
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
    "results": [
      {
        "permission_template_id": "00000001",
        "status": "success"
      },
      {
        "permission_template_id": "00000002",
        "status": "success"
      },
      {
        "permission_template_id": "00000003",
        "status": "failed",
        "reason": "该权限模版不存在"
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
| results | object array | 每个权限模版的执行结果列表  |

#### result[n]

| 参数名称                    | 参数类型   | 描述                                    |
|-------------------------|--------|---------------------------------------|
| permission_template_id  | string | 权限模版ID                                |
| status                  | string | 执行状态（枚举值：success-成功，failed-失败）        |
| reason                  | string | 失败原因，仅当 status 为 failed 时返回           |
