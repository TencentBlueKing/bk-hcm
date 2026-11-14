### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-全局配置。
- 该接口功能描述：查询 IaC 插件注册表列表（列表与详情共用同一接口，详情以 `id` 过滤）。

### URL

POST /api/v1/cloud/iac_plugins/list

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

#### rules[n]

| 参数名称  | 参数类型        | 必选 | 描述                                          |
|-------|-------------|----|---------------------------------------------|
| field | string      | 是  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| op    | enum string | 是  | 操作符（枚举值：eq、neq、gt、gte、lt、lte、in、nin、cs、cis） |
| value | 可变类型        | 是  | 查询条件Value值                                  |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

#### 查询参数介绍

| 参数名称           | 参数类型   | 描述                                               |
|----------------|----------|----------------------------------------------------|
| id             | string   | 插件记录ID                                         |
| plugin_name    | string   | 插件名称（如 tencentcloud）                        |
| plugin_version | string   | 插件版本号（三段式 semver，如 1.83.34）            |
| protocol       | string   | 协议类型（枚举值：tf_provider、http）              |
| state          | string   | 状态（枚举值：init、enabled、disabled、failed）    |
| memo           | string   | 备注                                               |
| creator        | string   | 创建者                                             |
| reviser        | string   | 更新者                                             |
| created_at     | string   | 创建时间，标准格式：2006-01-02T15:04:05Z           |
| updated_at     | string   | 更新时间，标准格式：2006-01-02T15:04:05Z           |

### 调用示例

#### 按插件名查询列表

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "plugin_name",
        "op": "eq",
        "value": "tencentcloud"
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
        "plugin_name": "tencentcloud",
        "plugin_version": "1.83.34",
        "protocol": "tf_provider",
        "state": "enabled",
        "message": "",
        "extension": {
          "architectures": {
            "linux_amd64": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
          }
        },
        "memo": "",
        "creator": "hcm-backend-admin",
        "reviser": "hcm-backend-admin",
        "created_at": "2026-10-20T10:00:00Z",
        "updated_at": "2026-10-20T10:01:00Z"
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

| 参数名称    | 参数类型   | 描述             |
|---------|--------|----------------|
| count   | uint64 | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据        |

#### data.details[n]

| 参数名称           | 参数类型   | 描述                                               |
|----------------|----------|----------------------------------------------------|
| id             | string   | 插件记录ID                                         |
| plugin_name    | string   | 插件名称                                           |
| plugin_version | string   | 插件版本号                                         |
| protocol       | string   | 协议类型（tf_provider / http）                     |
| state          | string   | 状态（init / enabled / disabled / failed）         |
| message        | string   | 状态说明（failed 时记录失败原因，正常状态为空）    |
| extension      | object   | 协议扩展字段（tf_provider 含 architectures；http 含 api_path、crypto_type、crypto_config） |
| memo           | string   | 备注                                               |
| creator        | string   | 创建者                                             |
| reviser        | string   | 更新者                                             |
| created_at     | string   | 创建时间，标准格式：2006-01-02T15:04:05Z           |
| updated_at     | string   | 更新时间，标准格式：2006-01-02T15:04:05Z           |

### 错误码与提示说明

| 错误码     | 场景与提示说明 |
|---------|--------------|
| 2000001 | 参数错误：`filter` 或 `page` 不合法 |
| 2000004 | 请求体解析失败 |
| 2000007 | 鉴权服务异常：无法判断调用者是否有权限，`authorize failed` |
| 2030403 | 无权限：调用者缺少平台全局配置权限，`no permission` |
