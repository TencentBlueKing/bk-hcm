### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：COS桶查询。
- 该接口功能描述：查询存储桶代理接口。

### URL

POST /api/v1/cloud/cos/buckets/list

### 输入参数

#### tcloud
代理接口文档地址：https://cloud.tencent.com/document/product/436/8291

| 参数名称        | 参数类型   | 必选 | 描述                                                                                                                                          |
|-------------|--------|----|---------------------------------------------------------------------------------------------------------------------------------------------|
| account_id  | string | 是  | 账号ID                                                                                                                                        |
| region      | string | 否  | 地域                                                                                                                                          |
| tag_key     | string | 否  | 支持根据存储桶标签（由标签键 tag_key 和标签值 tag_value 组成）过滤存储桶，仅支持传入一个存储桶标签，tag_key 用于传入标签键。如需根据存储桶标签查询存储桶，则 tag_key 和 tag_value 为必填项                       |
| tag_value   | string | 否  | 支持根据存储桶标签（由标签键 tag_key 和标签值 tag_value 组成）过滤存储桶，仅支持传入一个存储桶标签，tag_value 用于传入标签值。如需根据存储桶标签查询存储桶，则 tag_key 和 tag_value 为必填项                     |
| max_keys    | string | 否  | 单次返回最大的条目数量，默认值为2000，最大为2000。如果单次响应中未列出所有存储桶，COS 会返回next_marker节点，其值作为下次请求的marker参数                                                         |
| marker      | string | 否  | 起始标记，从该标记之后（不含）按照 UTF-8 字典序返回存储桶条目                                                                                                          |
| range       | string | 否  | 和 create_time 参数一起使用，支持根据创建时间过滤存储桶，支持枚举值 lt（创建时间早于 create_time）、gt（创建时间晚于 create_time）、lte（创建时间早于或等于 create_time）、gte（创建时间晚于或等于create_time） |
| create_time | string | 否  | GMT 时间戳，和 range 参数一起使用，支持根据创建时间过滤存储桶，例如 create_time=1642662645                                                                              |

### 调用示例

#### tcloud

```json
{
  "account_id": "0000001",
  "region": "ap-hk"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "owner": {
      "uin": "xxx",
      "id": "xxx",
      "display_name": "xxx"
    },
    "buckets": [
      {
        "name": "bucket-test-1",
        "region": "ap-hk",
        "creation_date": "2022-09-14T03:17:34Z",
        "bucket_type": "cos"
      },
      {
        "name": "bucket-test-2",
        "region": "ap-hk",
        "creation_date": "2022-09-14T02:37:05Z",
        "bucket_type": "cos"
      }
    ],
    "marker": "",
    "next_marker": "",
    "is_truncated": false
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

| 参数名称          | 参数类型         | 描述                                                 |
|---------------|--------------|----------------------------------------------------|
| owner         | object       | 存储桶持有者信息                                           |
| buckets       | object array | 存储桶列表                                              |
| marker        | string       | 表示本次请求的起点                                          |
| next_marker   | string       | 未返回所有结果时，作为下次请求的marker参数                           |
| is_truncated  | bool         | 是否所有的结果都已经返回。true：表示本次没有返回全部结果。false：表示本次已经返回了全部结果 |

#### data.owner

| 参数名称           | 参数类型    | 描述          |
|----------------|---------|-------------|
| uin            | string  | 存储桶持有者uin   |
| id             | string  | 存储桶持有者的完整ID |
| display_name   | string  | 存储桶持有者的名字   |

#### data.buckets[0]

| 参数名称            | 参数类型    | 描述                                            |
|-----------------|---------|-----------------------------------------------|
| name            | string  | 存储桶的名称                                        |
| region          | string  | 存储桶的地域                                        |
| creation_date   | string  | 存储桶的创建时间，为 ISO8601 格式，例如 2019-05-24T10:56:40Z |
| bucket_type     | string  | 存储桶类型                                         |