### 描述

- 该接口提供版本：v1.2.1+
- 该接口所需权限：
- 该接口功能描述：查询任务流列表

### URL

POST /api/v1/task/async/flows/list

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

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

### 调用示例

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "state",
        "op": "eq",
        "value": "success"
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

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "id": "0000000n",
      "name": "first_test",
      "state": "success",
      "tasks": [
        {
          "id": "0000002h",
          "flow_id": "0000000n",
          "flow_name": "first_test",
          "action_name": "test_CreateSG",
          "params": "{}",
          "retry_count": 0,
          "timeout_secs": 10,
          "depend_on": [],
          "state": "success",
          "memo": "",
          "reason": "{}"
        },
        {
          "id": "0000002i",
          "flow_id": "0000000n",
          "flow_name": "first_test",
          "action_name": "test_CreateSubnet",
          "params": "{}",
          "retry_count": 0,
          "timeout_secs": 10,
          "depend_on": [
            "0000002h"
          ],
          "state": "success",
          "memo": "",
          "reason": "{}"
        },
        {
          "id": "0000002j",
          "flow_id": "0000000n",
          "flow_name": "first_test",
          "action_name": "test_CreateVpc",
          "params": "{}",
          "retry_count": 0,
          "timeout_secs": 10,
          "depend_on": [
            "0000002h"
          ],
          "state": "success",
          "memo": "",
          "reason": "{}"
        },
        {
          "id": "0000002k",
          "flow_id": "0000000n",
          "flow_name": "first_test",
          "action_name": "test_CreateCvm",
          "params": "{}",
          "retry_count": 0,
          "timeout_secs": 10,
          "depend_on": [
            "0000002i",
            "0000002j"
          ],
          "state": "success",
          "memo": "",
          "reason": "{}"
        }
      ],
      "memo": "",
      "reason": "{}",
      "creator": "hcm-backend-async",
      "reviser": "hcm-backend-async",
      "created_at": "2023-08-28 15:46:23 +0000 UTC",
      "updated_at": "2023-08-28 16:18:20 +0000 UTC"
    },
    {
      "id": "0000000o",
      "name": "first_test",
      "state": "success",
      "tasks": [
        {
          "id": "0000002l",
          "flow_id": "0000000o",
          "flow_name": "first_test",
          "action_name": "test_CreateSG",
          "params": "{}",
          "retry_count": 0,
          "timeout_secs": 10,
          "depend_on": [],
          "state": "success",
          "memo": "",
          "reason": "{}"
        },
        {
          "id": "0000002m",
          "flow_id": "0000000o",
          "flow_name": "first_test",
          "action_name": "test_CreateSubnet",
          "params": "{}",
          "retry_count": 0,
          "timeout_secs": 10,
          "depend_on": [
            "0000002l"
          ],
          "state": "success",
          "memo": "",
          "reason": "{}"
        },
        {
          "id": "0000002n",
          "flow_id": "0000000o",
          "flow_name": "first_test",
          "action_name": "test_CreateVpc",
          "params": "{}",
          "retry_count": 0,
          "timeout_secs": 10,
          "depend_on": [
            "0000002l"
          ],
          "state": "success",
          "memo": "",
          "reason": "{}"
        },
        {
          "id": "0000002o",
          "flow_id": "0000000o",
          "flow_name": "first_test",
          "action_name": "test_CreateCvm",
          "params": "{}",
          "retry_count": 0,
          "timeout_secs": 10,
          "depend_on": [
            "0000002m",
            "0000002n"
          ],
          "state": "success",
          "memo": "",
          "reason": "{}"
        }
      ],
      "memo": "",
      "reason": "{}",
      "creator": "hcm-backend-async",
      "reviser": "hcm-backend-async",
      "created_at": "2023-08-28 15:46:33 +0000 UTC",
      "updated_at": "2023-08-28 15:54:07 +0000 UTC"
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data[n]

| 参数名称       | 参数类型         | 描述                             |
|------------|--------------|--------------------------------|
| id         | string       | 任务流ID                          |
| name       | string       | 任务流名称                          |
| state      | string       | 任务流状态                          |
| tasks      | object array | 任务集合                           |
| memo       | string       | 备注                             |
| reason     | string       | 失败等原因                          |
| creator    | string       | 创建者                            |
| reviser    | string       | 更新者                            |
| created_at | string       | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at | string       | 更新时间，标准格式：2006-01-02T15:04:05Z |

#### tasks[n]

| 参数名称         | 参数类型         | 描述     |
|--------------|--------------|--------|
| id           | string       | 任务ID   |
| flow_id      | string       | 任务流ID  |
| flow_name    | string       | 任务流名称  |
| action_name  | string       | 执行动作名称 |
| state        | string       | 任务流状态  |
| params       | object       | 参数信息   |
| retry_count  | int          | 重试次数   |
| timeout_secs | int          | 超时时间   |
| depend_on    | string array | 依赖任务集合 |
| memo         | string       | 备注     |
| reason       | string       | 失败等原因  |