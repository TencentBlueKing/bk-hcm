### 描述

- 该接口提供版本：v1.8.11+。
- 该接口所需权限：资源接入-云资源-云厂商配置。
- 该接口功能描述：更新权限策略库。

### URL

PATCH /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/{id}

### 路径参数

| 参数名称   | 参数类型   | 必选 | 描述                                    |
|--------|--------|----|---------------------------------------|
| vendor | string | 是  | 云厂商（枚举值：tcloud） |
| id     | string | 是  | 策略库ID                                 |

### 输入参数

| 参数名称            | 参数类型        | 必选 | 描述                                  |
|-----------------|-------------|----|-------------------------------------|
| name            | string      | 否  | 策略库名称，最大长度128字符                     |
| policy_document | string      | 否  | 权限策略JSON内容，变更时 version 自动加1         |
| bk_biz_ids      | int64 array | 否  | 允许使用的业务ID列表                         |
| memo            | string      | 否  | 描述，最大长度255字符                        |

### 调用示例

```json
{
  "name": "ReadOnlyPolicyV2",
  "policy_document": "{\"version\":\"2.0\",\"statement\":[{\"effect\":\"allow\",\"action\":[\"cos:GetObject\",\"cos:HeadObject\"],\"resource\":[\"*\"]}]}",
  "bk_biz_ids": [2, 3, 4],
  "memo": "更新后的只读权限策略"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述       |
|---------|--------|----------|
| code    | int32  | 状态码      |
| message | string | 请求信息     |
| data    | object | 响应数据（为空） |
