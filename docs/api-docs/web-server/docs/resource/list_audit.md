### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务审计查看。
- 该接口功能描述：查询审计列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/audits/list

### 输入参数

| 参数名称         | 参数类型      | 必选  | 描述     |
|--------------|-----------|-----|--------|
| bk_biz_id    | string    | 是   | 业务ID   |
| filter       | object    | 是   | 查询过滤条件 |
| page         | object    | 是   | 分页设置   |

#### filter

| 参数名称  | 参数类型        | 必选  | 描述                                                              |
|-------|-------------|-----|-----------------------------------------------------------------|
| op    | enum string | 是   | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是   | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选  | 描述                                          |
|-------|-------------|-----|---------------------------------------------|
| field | string      | 是   | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| op    | enum string | 是   | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis） |
| value | 可变类型        | 是   | 查询条件Value值                                  |

##### rules 表达式说明：

##### 1. 操作符

| 操作符 | 描述                                        | 操作符的value支持的数据类型                             |
|-----|-------------------------------------------|----------------------------------------------|
| eq  | 等于。不能为空字符串                                | boolean, numeric, string                     |
| neq | 不等。不能为空字符串                                | boolean, numeric, string                     |
| gt  | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte | 大于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lt  | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lte | 小于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in  | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string                     |
| nin | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string                     |
| cs  | 模糊查询，区分大小写                                | string                                       |
| cis | 模糊查询，不区分大小写                               | string                                       |

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

#### page

| 参数名称  | 参数类型   | 必选  | 描述                                                                                                                                                  |
|-------|--------|-----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是   | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否   | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否   | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否   | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否   | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

#### 查询参数介绍：

| 参数名称       | 参数类型   | 描述              |
|------------|--------|-----------------|
| id                      | uint64 | 审计ID                                            |
| res_id                  | string | 资源ID                                            |
| cloud_res_id            | string | 云资源ID                                           |
| res_name                | string | 资源名称                                            |
| res_type                | string | 资源类型                                            |
| associated_res_id       | string | 关联资源ID                                          |
| associated_cloud_res_id | string | 关联云资源ID                                         |
| associated_res_name     | string | 关联资源名称                                          |
| associated_res_type     | string | 关联资源类型                                          |
| action                  | string | 动作（枚举值：create、update、delete）                    |
| bk_biz_id               | string | 业务ID                                            |
| vendor                  | string | 供应商（枚举值：tcloud、aws、azure、gcp、huawei）            |
| account_id              | string | 账号ID                                            |
| operator                | string | 操作者                                             |
| source                  | string | 请求来源（枚举值：api_call[API调用]、background_sync[后台同步]） |
| rid                     | string | 请求ID                                            |
| app_code                | string | 应用代码                                            |
| created_at              | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                            |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 获取详细信息请求参数示例

如操作者为Jim的审计列表。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "operator",
        "op": "eq",
        "value": "Jim"
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

如操作者为Jim的审计数量。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "operator",
        "op": "eq",
        "value": "Jim"
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
        "id": 1,
        "res_id": "00000001",
        "cloud_res_id": "sg-xxxxxx",
        "res_name": "test-update",
        "res_type": "security_group",
        "associated_res_id": "",
        "associated_cloud_res_id": "",
        "associated_res_name": "",
        "associated_res_type": "",
        "action": "update",
        "bk_biz_id": -1,
        "vendor": "tcloud",
        "account_id": "00000001",
        "operator": "Jim",
        "source": "api_call",
        "rid": "xxxxxx",
        "app_code": "xxxxxx",
        "created_at": "2023-02-05T15:29:15Z"
      }
    ]
  }
}
```

#### 获取数量返回结果示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "count": 1
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

| 参数名称                    | 参数类型   | 描述                                                                                                             |
|-------------------------|--------|----------------------------------------------------------------------------------------------------------------|
| id                      | uint64 | 审计ID                                                                                                           |
| res_id                  | string | 资源ID                                                                                                           |
| cloud_res_id            | string | 云资源ID                                                                                                          |
| res_name                | string | 资源名称                                                                                                           |
| res_type                | string | 资源类型（枚举值：account、security_group、vpc、subnet、disk、cvm、route_table、eip、gcp_firewall_rule、image、network_interface） |
| associated_res_id       | string | 关联资源ID                                                                                                         |
| associated_cloud_res_id | string | 关联云资源ID                                                                                                        |
| associated_res_name     | string | 关联资源名称                                                                                                         |
| associated_res_type     | string | 关联资源类型                                                                                                         |
| action                  | string | 动作（枚举值：create、update、delete、assign、recycle、recover、reboot、start、stop、reset_pwd、associate、disassociate、bind、deliver）   |
| bk_biz_id               | string | 业务ID                                                                                                           |
| vendor                  | string | 供应商（枚举值：tcloud、aws、azure、gcp、huawei）                                                                           |
| account_id              | string | 账号ID                                                                                                           |
| operator                | string | 操作者                                                                                                            |
| source                  | string | 请求来源（枚举值：api_call[API调用]、background_sync[后台同步]）                                                                |
| rid                     | string | 请求ID                                                                                                           |
| app_code                | string | 应用代码                                                                                                           |
| created_at              | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                                                                 |

#### detail

| 参数名称     | 参数类型   | 描述                                  |
|----------|--------|-------------------------------------|
| data     | object | 创建资源信息/资源更新前信息/资源删除前信息，且不同资源审计该字段不同 |
| changed  | object | 资源更新信息                              |
