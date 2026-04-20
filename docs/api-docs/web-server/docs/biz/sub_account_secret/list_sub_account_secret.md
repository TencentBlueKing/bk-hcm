### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询三级账号密钥列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/sub_account_secrets/list

### 输入参数

| 参数名称                  | 参数类型           | 必选   | 描述                                 |
|-----------------------|----------------|------|------------------------------------|
| bk_biz_id             | int64          | 是    | 业务ID                               |
| vendor                | string         | 是    | 云厂商, 枚举值：tcloud                    |
| status                | string array   | 否    | 密钥状态, 枚举值：enabled(启用)、disabled(禁用) |
| extension             | object[vendor] | 否    | 云产商扩展字段                            |
| ids                   | string array   | 否    | 密钥id列表，长度限制500                   |
| account_ids           | string array   | 否    | 二级账号id列表，长度限制500                   |
| sub_account_ids       | string array   | 否    | 三级账号id列表，长度限制500                   |
| account_managers      | string array   | 否    | 二级账号负责人列表，长度限制500                  |
| sub_account_managers  | string array   | 否    | 三级账号负责人列表，长度限制500                  |
| page                  | object         | 是    | 分页设置                               |

#### extension [tcloud]
| 参数名称                    | 参数类型           | 必选    | 描述                       |
|-------------------------|----------------|-------|--------------------------|
| cloud_secret_ids        | string array   | 否     | 云密钥id列表，长度限制500          |
| cloud_main_account_ids  | string array   | 否     | 云二级账号id列表，长度限制500        |
| cloud_sub_account_ids   | string array   | 否     | 云三级账号id列表，长度限制500        |

#### page

| 参数名称   | 参数类型     | 必选   | 描述                                                                                                                                                  |
|--------|----------|------|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count  | bool     | 是    | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start  | uint32   | 否    | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit  | uint32   | 否    | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort   | string   | 否    | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order  | string   | 否    | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

### 调用示例

#### 获取详细信息请求参数示例

#### TCloud
```json
{
  "status": ["enabled"],
  "account_ids": ["test"],
  "ids":["00000001"],
  "sub_account_ids": ["test"],
  "extension":{
    "cloud_secret_ids": ["test"],
    "cloud_main_account_ids": ["test"],
    "cloud_sub_account_ids": ["test"]
  },
  "sub_account_managers": ["test"],
  "account_managers": ["test"],
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
  }
}
```

#### 获取数量请求参数示例

查询 vendor 为 tcloud 的密钥数量。

```json
{
  "cloud_secret_ids": ["test"],
  "status": ["enabled"],
  "account_ids": ["test"],
  "sub_account_ids": ["test"],
  "ids":["00000001"],
  "extension":{
    "cloud_secret_ids": ["test"],
    "cloud_main_account_ids": ["test"],
    "cloud_sub_account_ids": ["test"]
  },
  "sub_account_managers": ["test"],
  "account_managers": ["test"],
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
        "vendor": "tcloud",
        "status": "enabled",
        "account_id": "0000001",
        "sub_account_id": "0000001",
        "operable": true,
        "extension": {
          "cloud_secret_id": "xxxx",
          "cloud_main_account_id": "main-xxxx",
          "cloud_sub_account_id": "sub-xxxx",
          "console_login": 1
        },
        "sub_account_managers": ["test"],
        "account_managers": ["test"],
        "cloud_created_at": "2024-01-01T12:00:00Z",
        "disabled_time": "2024-01-03T12:00:00Z",
        "last_used_time": "2024-01-03T12:00:00Z",
        "creator": "Jim",
        "reviser": "Jim",
        "created_at": "2024-01-01T12:00:00Z",
        "updated_at": "2024-01-02T12:00:00Z"
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
    "count": 1,
    "details": null
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型    | 描述   |
|---------|---------|------|
| code    | int32   | 状态码  |
| message | string  | 请求信息 |
| data    | object  | 响应数据 |

#### data

| 参数名称    | 参数类型         | 描述             |
|---------|--------------|----------------|
| count   | uint64       | 当前规则能匹配到的总记录条数 |
| details | object array | 查询返回的数据        |

#### data.details[n]

| 参数名称                | 参数类型         | 描述                                                    |
|---------------------|--------------|-------------------------------------------------------|
| id                  | string       | 密钥ID                                                  |
| vendor              | string       | 云厂商, 枚举值：tcloud                                       |
| status              | string       | 密钥状态, 枚举值：enabled(启用)、disabled(禁用)                    |
| account_id          | string       | 账号id                                                  |
| operable            | bool         | 当前业务是否可操作该密钥（关联账号的 bk_biz_id 与路径 bk_biz_id 一致时为 true） |
| sub_account_id      | string       | 子账号id                                                 |
| extension           | object       | 云厂商差异扩展字段                                             |
| cloud_created_at    | string       | 云上创建时间，标准格式：2006-01-02T15:04:05Z                      |
| disabled_time       | string       | 本地禁用时间，标准格式：2006-01-02T15:04:05Z                      |
| last_used_time      | string       | 密钥上次调用时间，标准格式：2006-01-02T15:04:05Z                    |
| creator             | string       | 创建者                                                   |
| reviser             | string       | 更新者                                                   |
| created_at          | string       | 创建时间，标准格式：2006-01-02T15:04:05Z                        |
| updated_at          | string       | 更新时间，标准格式：2006-01-02T15:04:05Z                        |
| sub_account_manager | string array | 三级账号负责人列表                                             |
| account_manager     | string array | 二级账号负责人列表                                             |

##### extension[tcloud]

| 参数名称                  | 参数类型    | 描述                                  |
|-----------------------|---------|-------------------------------------|
| cloud_secret_id       | string  | 云密钥id                               |
| cloud_main_account_id | string  | 云二级账号id                             |
| cloud_sub_account_id  | string  | 云三级账号id                             |
| console_login         | int64   | 枚举值：0（编程账号，无法登录控制台）、1（控制台账号，可登录控制台） |
