### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：无。
- 该接口功能描述：CVM子网配置信息查询。

### URL

POST /api/v1/woa/config/findmany/config/cvm/subnet/list

### 输入参数

| 参数名称     | 参数类型  | 必选 | 描述         |
|-------------|---------|------|-------------|
| filter      | object  | 是   | 查询过滤条件   |
| page        | object  | 是   | 分页设置      |

#### filter

| 参数名称 | 参数类型      | 必选 | 描述                                                                                          |
|---------|-------------|------|----------------------------------------------------------------------------------------------|
| op      | enum string | 是   | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules   | array       | 是   | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。                 |

#### rules[n] （详情请看 rules 表达式说明）

| 参数名称 | 参数类型      | 必选 | 描述                                                              |
|---------|-------------|------|------------------------------------------------------------------|
| field   | string      | 是   | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍 |
| op      | enum string | 是   | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis）           |
| value   | 可变类型     | 是   | 查询条件Value值                                                     |

##### rules 表达式说明：

##### 1. 操作符

| 操作符 | 描述                                              | 操作符的value支持的数据类型                               |
|-------|--------------------------------------------------|--------------------------------------------------------|
| eq    | 等于。不能为空字符串                                | boolean, numeric, string                               |
| neq   | 不等。不能为空字符串                                | boolean, numeric, string                               |
| gt    | 大于                                             | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"）|
| gte   | 大于等于                                          | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"）|
| lt    | 小于                                             | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"）|
| lte   | 小于等于                                          | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"）|
| in    | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string                 |
| nin   | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string                |
| cs    | 模糊查询，区分大小写                                | string                                                 |
| cis   | 模糊查询，不区分大小写                              | string                                                  |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                                                                                                                                                                                                         |
|---------|--------|------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| count   | bool   | 否   | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start   | int    | 否   | 记录开始位置，start 起始值为0                                                                                                                                                                                   |
| limit   | int    | 否   | 每页限制条数，最大500，不能为0                                                                                                                                                                                   |
| sort    | string | 否   | 排序字段，返回数据将按该字段进行排序                                                                                                                                                                               |
| order   | string | 否   | 排序顺序（枚举值：ASC、DESC）                                                                                                                                                                                    |

#### 查询参数介绍：

| 参数名称              | 参数类型 | 描述                                          |
|----------------------|--------|----------------------------------------------|
| id                   | string | 资源ID                                        |
| cloud_id             | string | 子网云ID                                       |
| name                 | string | 子网名称                                       |
| region               | string | 地域                                          |
| zone                 | string | 园区                                          |
| cloud_vpc_id         | string | VPC云ID                                       |
| vpc_name             | string | VPC名称                                       |
| extension.enable_cvm | bool   | 是否启用                                       |

**注意：**

- count 如果此标记为true，表示此次请求是获取数量。此时其余字段必须为初始化值，start为0,limit为:0。

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "name",
        "op": "eq",
        "value": "xxxx"
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
        "id":"00000001",
        "region":"ap-shanghai",
        "zone":"ap-shanghai-2",
        "vpc_id":"vpc-2x7lhtse",
        "vpc_name":"VPC-IEG-SH",
        "subnet_id":"subnet-3pxzr6c9",
        "subnet_name":"cvm_use_0",
        "enable":true,
        "comment":""
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
| info    | object array | CVM子网配置信息           |

#### data.info

| 参数名称      | 参数类型  | 描述          |
|--------------|---------|---------------|
| id	       | string	 | 子网自增ID     |
| region	   | string  | 地域           |
| zone	       | string  | 可用区         |
| vpc_id	   | string	 | VPC ID        |
| vpc_name	   | string	 | VPC名         |
| subnet_id	   | string	 | 子网ID         |
| subnet_name  | string	 | 子网名         |
| enable	   | bool	 | 是否启用       |
| comment	   | string	 | 备注          |

**注意：**

- 如果本次请求是查询详细信息那么count为0，如果查询的是数量，那么info为空。
