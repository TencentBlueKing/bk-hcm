### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询业务下应用了指定策略库的权限模版列表, 全量返回，不分页。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/permission_policy_libraries/{id}/permission_templates

### 路径参数

| 参数名称      | 参数类型   | 必选 | 描述                        |
|-----------|--------|----|---------------------------|
| bk_biz_id | int64  | 是  | 业务ID                      |
| vendor    | string | 是  | 云厂商（枚举值：tcloud） |
| id        | string | 是  | 策略库ID                     |

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "id": "00000001",
        "cloud_id": "policy-12345678",
        "name": "example-policy",
        "vendor": "tcloud",
        "account_id": "00000001",
        "policy_library_id": "00000001",
        "policy_library_version": 2,
        "policy_library_sync_time": "2025-01-01T00:00:00Z",
        "policy_document": "{\"version\":\"2.0\",\"statement\":[{\"effect\":\"allow\",\"action\":[\"cvm:*\"],\"resource\":[\"*\"]}]}",
        "policy_hash": "a1b2c3d4e5f6...",
        "memo": "示例权限模版",
        "creator": "admin",
        "reviser": "admin",
        "created_at": "2025-01-01T00:00:00Z",
        "updated_at": "2025-01-01T00:00:00Z",
        "extension": {
          "cloud_type": 1
        }
      },
      {
        "id": "00000002",
        "cloud_id": "policy-87654321",
        "name": "example-policy-2",
        "vendor": "tcloud",
        "account_id": "00000002",
        "policy_library_id": "00000001",
        "policy_library_version": 1,
        "policy_library_sync_time": "2025-01-01T00:00:00Z",
        "policy_document": "{\"version\":\"2.0\",\"statement\":[{\"effect\":\"allow\",\"action\":[\"cos:*\"],\"resource\":[\"*\"]}]}",
        "policy_hash": "f6e5d4c3b2a1...",
        "memo": "示例权限模版2",
        "creator": "admin",
        "reviser": "admin",
        "created_at": "2025-01-01T00:00:00Z",
        "updated_at": "2025-01-01T00:00:00Z",
        "extension": {
          "cloud_type": 1
        }
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

| 参数名称    | 参数类型          | 描述                              |
|---------|---------------|---------------------------------|
| details | detail array  | 权限模版列表（全量返回，不分页，仅包含管理业务为当前业务的账号下的模版） |

#### detail

| 参数名称                       | 参数类型   | 描述                                                                      |
|----------------------------|--------|-------------------------------------------------------------------------|
| id                         | string | 权限模版ID                                                                  |
| cloud_id                   | string | 云上策略ID                                                                  |
| name                       | string | 模板名称                                                                    |
| vendor                     | string | 云厂商                                                                     |
| account_id                 | string | 所属二级账号ID                                                                |
| policy_library_id          | string | 来源权限策略库ID                                                               |
| policy_library_version     | int    | 应用时的策略库版本                                                               |
| policy_library_sync_time   | string | 策略库同步时间                                                                 |
| policy_document            | string | 策略JSON内容                                                                |
| policy_hash                | string | 策略内容哈希值                                                                 |
| memo                       | string | 描述                                                                      |
| creator                    | string | 创建者                                                                     |
| reviser                    | string | 更新者                                                                     |
| created_at                 | string | 创建时间                                                                    |
| updated_at                 | string | 更新时间                                                                    |
| extension                  | object | 云厂商扩展字段                                                                 |

##### extension[tcloud]

| 参数名称       | 参数类型 | 描述                                                  |
|------------|------|-----------------------------------------------------|
| cloud_type | int  | 策略类型（枚举值：1-自定义策略，2-预设策略） |
