### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源接入-云资源-云厂商配置。
- 该接口功能描述：创建权限策略库。

### URL

POST /api/v1/cloud/vendors/{vendor}/permission_policy_libraries/create

### 路径参数

| 参数名称   | 参数类型   | 必选 | 描述                                    |
|--------|--------|----|---------------------------------------|
| vendor | string | 是  | 云厂商（枚举值：tcloud） |

### 输入参数

| 参数名称            | 参数类型        | 必选 | 描述                    |
|-----------------|-------------|----|------------------------|
| name            | string      | 是  | 策略库名称 |
| policy_document | string      | 是  | 权限策略JSON内容             |
| bk_biz_ids      | int64 array | 是  | 允许使用的业务ID列表            |
| memo            | string      | 是  | 描述           |

### 调用示例

```json
{
  "name": "ReadOnlyPolicy",
  "policy_document": "{\"version\":\"2.0\",\"statement\":[{\"effect\":\"allow\",\"action\":[\"cos:GetObject\"],\"resource\":[\"*\"]}]}",
  "bk_biz_ids": [2, 3],
  "memo": "只读权限策略"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000001"
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

| 参数名称 | 参数类型   | 描述    |
|------|--------|-------|
| id   | string | 策略库ID |
