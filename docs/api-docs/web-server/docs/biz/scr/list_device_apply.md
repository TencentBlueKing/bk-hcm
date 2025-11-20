### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：已交付设备查询。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/findmany/apply/device

### 输入参数

| 参数名称   | 参数类型 | 必选 | 描述      |
|-----------|--------|-----|-----------|
| filter    | object | 是  | 查询过滤条件 |
| page      | object | 是  | 分页设置    |

#### filter

| 参数名称   | 参数类型      | 必选 | 描述                                                              |
|-----------|-------------|------|------------------------------------------------------------------|
| condition | enum string | 是   | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rule      | array       | 是   | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。 |

#### rule[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型      | 必选 | 描述                                          |
|----------|-------------|----|------------------------------------------------|
| field    | string      | 是  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| operator | enum string | 是  | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis） |
| value    | 可变类型     | 是  | 查询条件Value值                                  |

##### rule 表达式说明：

##### 1. 操作符

| 操作符            | 描述                                        | 操作符的value支持的数据类型                                 |
|------------------|---------------------------------------------|---------------------------------------------------------|
| equal            | 等于。不能为空字符串                           | boolean, numeric, string                                |
| not_equal        | 不等。不能为空字符串                           | boolean, numeric, string                                |
| greater          | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| greater_or_equal | 大于等于                                     | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| less             | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| less_or_equal    | 小于等于                                     | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in               | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string             |
| not_in           | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string            |
| contains         | 模糊查询，区分大小写                           | string                                                   |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                            |
|---------|--------|-----|---------------------------------|
| start   | int    | 否  | 记录开始位置，start 起始值为0       |
| limit   | int    | 是  | 每页限制条数，最大200              |
| sort    | string | 否  | 排序字段，返回数据将按该字段进行排序  |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "filter":{
    "condition":"AND",
    "rules":[
      {
        "field":"order_id",
        "operator":"equal",
        "value":1001
      }
    ]
  },
  "page":{
    "start":0,
    "limit": 20,
    "sort":"suborder_id"
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data":{
    "count":1,
    "info":[
      {
        "order_id":1001,
        "suborder_id":"1001-1",
        "bk_biz_id":213,
        "bk_username":"xxx",
        "ip":"10.0.0.1",
        "asset_id":"TC000000000001",
        "require_type":1,
        "resource_type":"QCLOUDCVM",
        "device_type":"S3ne.4XLARGE64",
        "zone_name":"上海-奉贤",
        "owner_ip": "x.x.x.x",
        "create_at":"2022-04-24T02:29:32.511Z",
        "update_at":"2022-04-24T02:29:32.511Z"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| result     | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| data	     | object array | 响应数据             |

#### data

| 参数名称 | 参数类型       | 描述                    |
|---------|--------------|-------------------------|
| count   | int          | 当前规则能匹配到的总记录条数 |
| info    | object array | 已交付设备列表            |

#### data.info

| 参数名称       | 参数类型    | 描述          |
|---------------|-----------|---------------|
| order_id	    | int       | 资源申请单号    |
| suborder_id   | string    | 资源申请子单号  |
| bk_biz_id	    | int       | 业务ID         |
| bk_username   | string    | 提单人         |
| ip	        | string    | 设备IP         |
| asset_id      | string    | 设备固资号      |
| require_type  | int       | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤 |
| resource_type | string    | 资源类型。"QCLOUDCVM": 腾讯云虚拟机, "IDCPM": IDC物理机, "QCLOUDDVM": Qcloud富容器, "IDCDVM": IDC富容器 |
| device_type   | string    | 机型           |
| zone_name     | string    | 区域           |
| owner_ip      | string    | 所属的母机IP    |
| create_at     | timestamp | 记录创建时间    |
| update_at     | timestamp | 记录更新时间    |
