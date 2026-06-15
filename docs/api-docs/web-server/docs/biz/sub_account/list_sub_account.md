### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询三级账号列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/sub_accounts/list

### 输入参数

| 参数名称      | 参数类型 | 必选 | 描述        |
|-----------|------|----|-----------|
| bk_biz_id | int64 | 是 | 业务ID      |
| filter | object | 是  | 查询过滤条件 |
| page   | object | 是  | 分页设置    |

#### filter

| 参数名称  | 参数类型      | 必选 | 描述                                                                                          |
|-------|-------------|----|---------------------------------------------------------------------------------------------|
| op    | enum string | 是  | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是  | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。                 |

#### rules[n] （详情请看 rules 表达式说明）

| 参数名称 | 参数类型      | 必选 | 描述                                                              |
|------|-------------|----|-----------------------------------------------------------------|
| field | string      | 是  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍 |
| op    | enum string | 是  | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis）           |
| value | 可变类型     | 是  | 查询条件Value值                                                     |

##### rules 表达式说明：

##### 1. 操作符

| 操作符 | 描述                                              | 操作符的value支持的数据类型                               |
|-----|-------------------------------------------------|------------------------------------------------------|
| eq  | 等于。不能为空字符串                                | boolean, numeric, string                             |
| neq | 不等。不能为空字符串                                | boolean, numeric, string                             |
| gt  | 大于                                             | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte | 大于等于                                          | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lt  | 小于                                             | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lte | 小于等于                                          | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in  | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string |
| nin | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string |
| cs  | 模糊查询，区分大小写                                | string                                               |
| cis | 模糊查询，不区分大小写                              | string                                               |

##### 2. 协议示例

查询 vendor 是 "tcloud" 且 name 为 "Jim" 的数据。

```json
{
  "op": "and",
  "rules": [
    {
      "field": "vendor",
      "op": "eq",
      "value": "tcloud"
    },
    {
      "field": "name",
      "op": "eq",
      "value": "Jim"
    }
  ]
}
```

#### page

| 参数名称 | 参数类型 | 必选 | 描述                                                                                                                                                                                                         |
|------|------|----|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                                                                     |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                                                                    |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                                                                   |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                                                                    |

### 调用示例

#### 获取详细信息请求参数示例

查询 vendor 为 tcloud 的三级账号列表。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "vendor",
        "op": "eq",
        "value": "tcloud"
      }
    ]
  },
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
  }
}
```

#### 获取数量请求参数示例

查询 vendor 为 tcloud 的三级账号数量。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "vendor",
        "op": "eq",
        "value": "tcloud"
      }
    ]
  },
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
        "id": "00000031",
        "cloud_id": "13943695",
        "name": "Jim",
        "vendor": "tcloud",
        "site": "china",
        "account_id": "00000003",
        "account_type": "current_account",
        "account_name": "account1",
        "operable": true,
        "managers": [],
        "bk_biz_ids": [
          310
        ],
        "memo": "",
        "email": "sub@example.com",
        "phone_num": "13800000000",
        "country_code": "86",
        "cloud_created_at": "2024-01-01T12:00:00Z",
        "sub_account_secret_count": 10,
		"permission_template_ids": ["00000001"],
		"permission_templates": [
		  {
		  "id": "00000001",
		  "name": "template1"
		   }
		],
        "creator": "Jim",
        "reviser": "Jim",
        "created_at": "2024-01-01T12:00:00Z",
        "updated_at": "2024-01-02T12:00:00Z",
        "extension": {
          "login_flag": "token",
          "action_flag": "stoken",
          "console_login": 1
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
    "count": 1,
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

| 参数名称    | 参数类型         | 描述             |
|---------|--------------|----------------|
| count   | uint64       | 当前规则能匹配到的总记录条数 |
| details | object array | 查询返回的数据        |

#### data.details[n]

| 参数名称                     | 参数类型         | 描述                                   |
|--------------------------|--------------|--------------------------------------|
| id                       | string       | 三级账号HCM本地ID                          |
| vendor                   | string       | 云厂商（枚举值：tcloud、aws、azure、gcp、huawei） |
| name                     | string       | 名称                                   |
| cloud_id                 | string       | 三级账号云ID                              |
| account_id               | string       | 三级账号所属二级账号HCM本地ID                    |
| account_type             | string       | 三级账号所属二级账号类型                         |
| account_name             | string       | 三级账号所属二级账号名称                         |
| operable                 | bool         | 当前业务是否可操作该三级账号                       |
| managers                 | string array | 账号管理者                                |
| bk_biz_ids               | int64 array  | 使用业务id                               |
| site                     | string       | 站点（枚举值：china:中国站、international:国际站）  |
| memo                     | string       | 备注                                   |
| email                    | string       | 邮箱                                   |
| phone_num                | string       | 手机号                                  |
| country_code             | string       | 手机区域代码                               |
| cloud_created_at         | string       | 云上创建时间，标准格式：2006-01-02T15:04:05Z     |
| sub_account_secret_count | int64        | 三级账号密钥数                              |
| permission_template_ids  | string array | 三级账号关联权限模版ID列表                       |
| permission_templates     | object array | 三级账号关联权限模版信息列表                       |
| creator                  | string       | 创建者                                  |
| reviser                  | string       | 更新者                                  |
| created_at               | string       | 创建时间，标准格式：2006-01-02T15:04:05Z       |
| updated_at               | string       | 更新时间，标准格式：2006-01-02T15:04:05Z       |
| extension                | object       | 扩展字段                                 |

##### extension[tcloud]

| 参数名称                  | 参数类型    | 描述                                                                                                        |
|-----------------------|---------|-----------------------------------------------------------------------------------------------------------|
| login_flag            | string  | 登录保护设置，枚举值：phone(安全手机)、token(硬token)、stoken(MFA字段)、wechat(微信)、custom(自定义)、mail(邮箱)、u2FToken(u2f硬件token)   |
| action_flag           | string  | 敏感操作保护设置，枚举值：phone(安全手机)、token(硬token)、stoken(MFA字段)、wechat(微信)、custom(自定义)、mail(邮箱)、u2FToken(u2f硬件token) |
| console_login         | int64   | 枚举值：0（编程账号，无法登录控制台）、1（控制台账号，可登录控制台）                                                                       |
| cloud_main_account_id | string  | 三级账号所属的二级账号云ID                                                                                            |

##### permission_templates[n]

| 参数名称  | 参数类型   | 描述        |
|-------|--------|-----------|
| id    | string | 权限模版本地ID  |
| name  | string | 权限模版名称    |
