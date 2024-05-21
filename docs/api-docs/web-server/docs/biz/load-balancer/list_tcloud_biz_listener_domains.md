### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：业务下腾讯云监听器域名列表。注意：域名非实体，查询条件是规则的。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/tcloud/listeners/{lbl_id}/domains/list

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述    |
|-----------|--------|----|-------|
| bk_biz_id | int64  | 是  | 业务ID  |
| lbl_id    | string | 是  | 监听器id |


### 调用示例
```json
{}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "default_domain": "www.qq.com.cn",
    "domain_list": [
      {
        "domain": "www.qq.com.cn",
        "url_count": 3
      },
      {
        "domain": "www.weixin.com",
        "url_count": 1
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

| 参数名称           | 参数类型   | 描述     |
|----------------|--------|--------|
| default_domain | string | 默认域名   |
| domain_list    | array  | 域名信息列表 |

#### data.domain_list[n]

| 参数名称      | 参数类型   | 描述    |
|-----------|--------|-------|
| domain    | string | 监听的域名 |
| url_count | int    | url数量 |
