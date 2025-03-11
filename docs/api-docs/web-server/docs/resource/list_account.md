### 描述

- 该接口提供版本：v1.2.1+。
- 该接口所需权限：账号查看。
- 该接口功能描述：查询账号列表。v1.7.3+起该接口不分页，返回用户有权限访问且符合筛选条件的全量数据。

拥有以下权限的人可以查询账号列表:

1. 有 `资源接入-账号查看` 权限，按账号的实例鉴权
2. 有 账号所属业务的`业务访问` 权限
3. 是 `负责人`

### 输入参数

| 参数名称   | 参数类型   | 必选 | 描述     |
|--------|--------|----|--------|
| filter | object | 是  | 查询过滤条件 |

#### filter

| 参数名称  | 参数类型        | 必选 | 描述                                                              |
|-------|-------------|----|-----------------------------------------------------------------|
| op    | enum string | 是  | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是  | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选 | 描述                                          |
|-------|-------------|----|---------------------------------------------|
| field | string      | 是  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| op    | enum string | 是  | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis） |
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

查询 name 是 "Jim" 且 age 大于18小于30 且 servers 类型是 "api" 或者是 "web" 的数据。

```json
{
  "op": "and",
  "rules": [
    {
      "field": "name",
      "op": "eq",
      "value": "Jim"
    },
    {
      "field": "age",
      "op": "gt",
      "value": 18
    },
    {
      "field": "age",
      "op": "lt",
      "value": 30
    },
    {
      "field": "servers",
      "op": "in",
      "value": [
        "api",
        "web"
      ]
    }
  ]
}
```

#### 查询参数介绍：

| 参数名称       | 参数类型         | 描述                                                               |
|------------|--------------|------------------------------------------------------------------|
| id         | string       | 账号ID                                                             |
| vendor     | string       | 供应商（枚举值：tcloud、aws、azure、gcp、huawei）                             |
| name       | string       | 名称                                                               |
| managers   | string array | 账号负责人                                                            |
| type       | string       | 账号类型 (枚举值：resource:资源账号、registration:登记账号、security_audit:安全审计账号) |
| site       | string       | 站点（枚举值：china:中国站、international:国际站）                              |
| price      | string       | 余额                                                               |
| price_unit | string       | 余额单位                                                             |
| memo       | string       | 备注                                                               |
| creator    | string       | 创建者                                                              |
| reviser    | string       | 更新者                                                              |
| created_at | string       | 创建时间，标准格式：2006-01-02T15:04:05Z                                   |
| updated_at | string       | 更新时间，标准格式：2006-01-02T15:04:05Z                                   |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 获取详细信息请求参数示例

如创建者为Jim的账号列表。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "creator",
        "op": "eq",
        "value": "Jim"
      }
    ]
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
    "count": 1,
    "details": [
      {
        "id": "00000002",
        "vendor": "gcp",
        "name": "gcp_account",
        "managers": [
          "hcm"
        ],
        "type": "resource",
        "site": "china",
        "price": "",
        "price_unit": "",
        "memo": "account create",
        "bk_biz_ids": [
          310
        ],
        "creator": "Jim",
        "reviser": "Jim",
        "created_at": "2022-12-26T07:42:15Z",
        "updated_at": "2023-04-19T19:29:15Z"
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

| 参数名称    | 参数类型   | 描述             |
|---------|--------|----------------|
| count   | uint64 | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据        |

#### data.details[n]

| 参数名称                 | 参数类型         | 描述                                                               |
|----------------------|--------------|------------------------------------------------------------------|
| id                   | string       | 账号ID                                                             |
| vendor               | string       | 供应商（枚举值：tcloud、aws、azure、gcp、huawei）                             |
| name                 | string       | 名称                                                               |
| managers             | string array | 账号管理者                                                            |
| type                 | string       | 账号类型 (枚举值：resource:资源账号、registration:登记账号、security_audit:安全审计账号) |
| site                 | string       | 站点（枚举值：china:中国站、international:国际站）                              |
| price                | string       | 余额                                                               |
| price_unit           | string       | 余额单位                                                             |
| memo                 | string       | 备注                                                               |
| bk_biz_ids           | int64 array  | 账号关联的业务ID列表                                                      |
| sync_status          | string       | 资源同步状态                                                           |
| sync_failed_reason   | string       | 资源同步失败原因                                                         |
| recycle_reserve_time | int          | 回收站资源的保留时长，单位小时                                                  |
| creator              | string       | 创建者                                                              |
| reviser              | string       | 更新者                                                              |
| created_at           | string       | 创建时间，标准格式：2006-01-02T15:04:05Z                                   |
| updated_at           | string       | 更新时间，标准格式：2006-01-02T15:04:05Z                                   |
