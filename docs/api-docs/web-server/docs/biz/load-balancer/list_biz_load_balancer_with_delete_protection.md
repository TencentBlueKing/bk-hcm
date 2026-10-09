### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询业务下的负载均衡列表，额外返回删除保护、实例规格等拓展字段。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/with/delete_protection/list

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述     |
|-----------|--------|----|--------|
| bk_biz_id | int64  | 是  | 业务ID   |
| filter    | object | 是  | 查询过滤条件 |
| page      | object | 是  | 分页设置   |

#### filter

| 参数名称  | 参数类型        | 必选 | 描述                                                              |
|-------|-------------|----|-----------------------------------------------------------------|
| op    | enum string | 是  | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是  | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选 | 描述                                                  |
|-------|-------------|----|-----------------------------------------------------|
| field | string      | 是  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍          |
| op    | enum string | 是  | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis、json_eq） |
| value | 可变类型        | 是  | 查询条件Value值                                          |

##### rules 表达式说明：

##### 1. 操作符

| 操作符     | 描述                                        | 操作符的value支持的数据类型                              |
|---------|-------------------------------------------|-----------------------------------------------|
| eq      | 等于。不能为空字符串                                | boolean, numeric, string                      |
| neq     | 不等。不能为空字符串                                | boolean, numeric, string                      |
| gt      | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte     | 大于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lt      | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lte     | 小于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in      | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string                      |
| nin     | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string                      |
| cs      | 模糊查询，区分大小写                                | string                                        |
| cis     | 模糊查询，不区分大小写                               | string                                        |
| json_eq | JSON字段等于，用于 extension 内字段的查询              | boolean, numeric, string                      |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

#### 查询参数介绍：

| 参数名称                | 参数类型   | 描述                                                |
|---------------------|--------|---------------------------------------------------|
| extension.exclusive | int    | 是否独占型实例（1是、0否），仅tcloud，需使用 `json_eq` 操作符          |
| extension.sla_type  | string | 性能容量型规格档位，空字符串表示非性能容量型，仅tcloud，需使用 `json_eq` 操作符  |

#### 实例规格筛选说明：

| 前端「实例规格」筛选项 | 传入的过滤规则                                                           |
|--------------|-------------------------------------------------------------------|
| 独占型          | `extension.exclusive json_eq 1`                                   |
| 共享型          | `extension.exclusive json_eq 0` 且 `extension.sla_type json_eq ""` |
| 性能容量型（指定档位）  | `extension.sla_type json_eq "clb.c2.medium"`（其它档位同理）              |

### 调用示例

#### 获取详细信息请求参数示例

查询业务下的独占型负载均衡列表。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "extension.exclusive",
        "op": "json_eq",
        "value": 1
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

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "extension.exclusive",
        "op": "json_eq",
        "value": 1
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
        "id": "00000001",
        "cloud_id": "lb-123",
        "name": "lb-test",
        "vendor": "tcloud",
        "bk_biz_id": 213,
        "account_id": "00000001",
        "region": "ap-guangzhou",
        "zones": [
          "ap-guangzhou-3"
        ],
        "backup_zones": [],
        "cloud_vpc_id": "vpc-123",
        "vpc_id": "00000002",
        "lb_type": "OPEN",
        "ip_version": "ipv4",
        "domain": "",
        "memo": "lb test",
        "band_width": 10,
        "isp": "BGP",
        "status": "1",
        "private_ipv4_addresses": [],
        "private_ipv6_addresses": [],
        "public_ipv4_addresses": [
          "1.1.1.1"
        ],
        "public_ipv6_addresses": [],
        "cloud_created_time": "2026-09-15T14:47:39Z",
        "cloud_status_time": "2026-09-15T14:47:39Z",
        "cloud_expired_time": "",
        "sync_time": "2026-09-15T09:03:06Z",
        "creator": "admin",
        "reviser": "admin",
        "created_at": "2026-09-15T14:47:39Z",
        "updated_at": "2026-09-15T14:55:40Z",
        "delete_protect": false,
        "exclusive": 1,
        "sla_type": ""
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

| 参数名称           | 参数类型    | 描述                                          |
|----------------|---------|---------------------------------------------|
| delete_protect | boolean | 是否开启删除保护，仅tcloud，非tcloud固定为false            |
| exclusive      | int     | 是否独占型实例（1是、0否），仅tcloud         |
| sla_type       | string  | 性能容量型规格档位，空字符串表示非性能容量型，仅tcloud |

#### 实例规格展示说明：

前端按以下优先级合成「实例规格」一列，后端只返回原始值、不做中文映射：

1. `exclusive` 为 1，展示「独占型」；
2. 否则 `sla_type` 非空，展示对应档位名（标准型、高阶型1/2、超强型1/2/3/4）；
3. 否则展示「共享型」。
