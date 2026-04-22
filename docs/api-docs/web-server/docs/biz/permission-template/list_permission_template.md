### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询业务下云权限模板列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/permission_templates/list

### 路径参数

| 参数名称      | 参数类型   | 必选 | 描述                        |
|-----------|--------|----|---------------------------|
| bk_biz_id | int64  | 是  | 业务ID                      |
| vendor    | string | 是  | 云厂商（枚举值：tcloud） |

### 输入参数

| 参数名称         | 参数类型   | 必选 | 描述                                                                                     |
|--------------|--------|----|----------------------------------------------------------------------------------------|
| cloud_ids    | string array | 否  | 权限模板云上ID列表，精确匹配，最多500个                                                        |
| names        | string array | 否  | 权限模板名称列表，模糊匹配，最多500个                                                          |
| extension    | object       | 否  | 混合云差异字段                                                   |
| creator      | string | 否  | 创建人，精确匹配                                                                          |
| reviser      | string | 否  | 更新人，精确匹配                                                                          |
| page         | object | 是  | 分页设置                                                                                   |

#### extension[tcloud]

| 参数名称           | 参数类型   | 必选 | 描述                                                                                     |
|----------------|--------|----|----------------------------------------------------------------------------------------|
| cloud_sub_account_ids | string array | 否  | 所属三级账号云上ID列表，筛选关联了这些三级账号的权限模板，最多500个                                      |
| cloud_account_ids     | string array | 否  | 所属二级账号云上ID列表，精确匹配，最多500个                                                      |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

### 调用示例

#### 获取详细信息请求参数示例

查询指定二级账号下名称包含 "ReadOnly" 的云权限模板列表。

```json
{
  "extension": {
    "cloud_account_ids": ["11111"]
  },
  "names": ["ReadOnly"],
  "page": {
    "count": false,
    "start": 0,
    "limit": 10
  }
}
```

#### 获取数量请求参数示例

```json
{
  "page": {
    "count": true
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 0,
    "details": [
      {
        "id": "00000001",
        "cloud_id": "policy-12345678",
        "name": "ReadOnlyAccess",
        "vendor": "tcloud",
        "account_id": "00000001",
        "cloud_account_id": "111111",
        "policy_library_id": "00000010",
        "policy_library_name": "只读权限策略",
        "policy_library_version": 2,
        "policy_library_sync_time": "2026-03-01T10:00:00Z",
        "policy_document": "{\"version\":\"2.0\",\"statement\":[{\"effect\":\"allow\",\"action\":[\"cos:GetObject\"],\"resource\":[\"*\"]}]}",
        "memo": "只读权限模板",
        "associated_sub_account_count": 5,
        "creator": "admin",
        "reviser": "admin",
        "created_at": "2026-03-01T10:00:00Z",
        "updated_at": "2026-03-10T15:30:00Z",
        "extension": {
          "cloud_type": 1
        }
      },
      {
        "id": "00000002",
        "cloud_id": "policy-87654321",
        "name": "FullAccessPolicy",
        "vendor": "tcloud",
        "account_id": "00000001",
        "cloud_account_id": "111111",
        "policy_library_id": "",
        "policy_library_name": "",
        "policy_library_version": 0,
        "policy_library_sync_time": "",
        "policy_document": "{\"version\":\"2.0\",\"statement\":[{\"effect\":\"allow\",\"action\":[\"*\"],\"resource\":[\"*\"]}]}",
        "memo": "云上同步的全量权限策略",
        "associated_sub_account_count": 2,
        "creator": "admin",
        "reviser": "admin",
        "created_at": "2026-02-15T08:00:00Z",
        "updated_at": "2026-03-05T12:00:00Z",
        "extension": {
          "cloud_type": 2
        }
      }
    ]
  }
}
```

#### 获取数量返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 15,
    "details": null
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

| 参数名称    | 参数类型   | 描述             |
|---------|--------|----------------|
| count   | uint64 | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据        |

#### data.details[n]

| 参数名称                       | 参数类型   | 描述                                         |
|----------------------------|--------|--------------------------------------------|
| id                         | string | 本地模板ID                                     |
| cloud_id                   | string | 云上策略ID                                     |
| name                       | string | 模板名称                                       |
| vendor                     | string | 云厂商                                        |
| account_id                 | string | 所属二级账号ID                                   |
| cloud_account_id           | string | 所属二级账号云上ID                                 |
| policy_library_id          | string | 来源权限策略库ID（云上同步的为空）                         |
| policy_library_name        | string | 来源权限策略库名称（云上同步的为空）                         |
| policy_library_version     | int    | 应用时的策略库版本（云上同步的为0）                         |
| policy_library_sync_time   | string | 策略库同步时间（云上同步的为空）                           |
| policy_document            | string | 策略JSON内容                                   |
| memo                       | string | 描述                                         |
| associated_sub_account_count | int  | 关联三级账号数                                    |
| creator                    | string | 创建者                                        |
| reviser                    | string | 更新者                                        |
| created_at                 | string | 创建时间，标准格式：2006-01-02T15:04:05Z             |
| updated_at                 | string | 更新时间，标准格式：2006-01-02T15:04:05Z             |
| extension                  | object | 云厂商扩展字段                                    |

##### extension[tcloud]

| 参数名称       | 参数类型 | 描述                                                  |
|------------|------|-----------------------------------------------------|
| cloud_type | int  | 策略类型（枚举值：1-自定义策略，2-预设策略） |
