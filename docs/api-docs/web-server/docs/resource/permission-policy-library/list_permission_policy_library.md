### 描述

- 该接口提供版本：v1.8.11+。
- 该接口所需权限：资源接入-云资源-云厂商配置。
- 该接口功能描述：查询权限策略库列表。

### URL

POST /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/list

### 路径参数

| 参数名称   | 参数类型   | 必选 | 描述                                    |
|--------|--------|----|---------------------------------------|
| vendor | string | 是  | 云厂商（枚举值：tcloud） |

### 输入参数

| 参数名称   | 参数类型   | 必选 | 描述     |
|--------|--------|----|--------|
| filter | object | 是  | 查询过滤条件 |
| page   | object | 是  | 分页设置   |

#### filter

| 参数名称  | 参数类型        | 必选 | 描述                                                              |
|-------|-------------|----|-----------------------------------------------------------------|
| op    | enum string | 是  | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是  | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选 | 描述                                          |
|-------|-------------|----|---------------------------------------------|
| field | string      | 是  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| op    | enum string | 是  | 操作符（枚举值：eq、neq、gt、gte、lt、lte、in、nin、cs、cis） |
| value | 可变类型        | 是  | 查询条件Value值                                  |

##### rules 表达式说明：

##### 1. 操作符

| 操作符 | 描述                                        | 操作符的value支持的数据类型                              |
|-----|-------------------------------------------|-----------------------------------------------|
| eq  | 等于。不能为空字符串                                | boolean, numeric, string                      |
| neq | 不等。不能为空字符串                                | boolean, numeric, string                      |
| gt  | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte | 大于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lt  | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lte | 小于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in  | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string                      |
| nin | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string                      |
| cs  | 模糊查询，区分大小写                                | string                                        |
| cis | 模糊查询，不区分大小写                               | string                                        |

##### 2. 协议示例

查询 name 包含 "ReadOnly" 的数据。

```json
{
  "op": "and",
  "rules": [
    {
      "field": "name",
      "op": "cs",
      "value": "ReadOnly"
    }
  ]
}
```

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

#### 查询参数介绍：

| 参数名称       | 参数类型        | 描述                                   |
|------------|-------------|--------------------------------------|
| id         | string      | 策略库ID   |
| name       | string      | 策略库名称  |
| version    | int         | 策略库版本号                               |
| bk_biz_ids | int64 array | 允许使用的业务ID列表                          |
| memo       | string      | 描述                                   |
| creator    | string      | 创建者                                  |
| reviser    | string      | 更新者                                  |
| created_at | string      | 创建时间，标准格式：2006-01-02T15:04:05Z       |
| updated_at | string      | 更新时间，标准格式：2006-01-02T15:04:05Z       |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 获取详细信息请求参数示例

如查询名称包含 "ReadOnly" 的权限策略库列表。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "name",
        "op": "cs",
        "value": "ReadOnly"
      }
    ]
  },
  "page": {
    "count": false,
    "start": 0,
    "limit": 20
  }
}
```

#### 获取数量请求参数示例

```json
{
  "filter": {
    "op": "and",
    "rules": []
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
        "id": "00000001",
        "vendor": "tcloud",
        "name": "ReadOnlyPolicy",
        "policy_document": "{\"version\":\"2.0\",\"statement\":[{\"effect\":\"allow\",\"action\":[\"cos:GetObject\"],\"resource\":[\"*\"]}]}",
        "policy_hash": "a1b2c3d4e5f67890abcdef1234567890abcdef1234567890abcdef1234567890",
        "version": 2,
        "bk_biz_ids": [2, 3],
        "memo": "只读权限策略",
        "associated_account_count": 5,
        "creator": "admin",
        "reviser": "admin",
        "created_at": "2026-03-01T10:00:00Z",
        "updated_at": "2026-03-10T15:30:00Z"
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
    "count": 10,
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

| 参数名称                     | 参数类型        | 描述                                                       |
|--------------------------|-------------|----------------------------------------------------------|
| id                       | string      | 策略库ID                                                    |
| name                     | string      | 策略库名称                                                    |
| policy_document          | string      | 当前版本的权限策略JSON内容                                          |
| policy_hash              | string      | 策略内容SHA256哈希值                                            |
| version                  | int         | 当前版本号，从1开始递增，策略内容变更时自动加1                                |
| bk_biz_ids               | int64 array | 允许使用的业务ID列表                                              |
| memo                     | string      | 描述                                                       |
| associated_account_count | int         | 关联二级账号数 |
| creator                  | string      | 创建者                                                      |
| reviser                  | string      | 更新者                                                      |
| created_at               | string      | 创建时间，标准格式：2006-01-02T15:04:05Z                           |
| updated_at               | string      | 更新时间，标准格式：2006-01-02T15:04:05Z                           |
