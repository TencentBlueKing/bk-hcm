### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：账号查看。
- 该接口功能描述：查询指定业务的账号列表。

### URL

GET /api/v1/cloud/accounts/bizs/{bk_biz_id}

### 输入参数

| 参数名称         | 参数类型   | 必选 | 描述     |
|--------------|--------|----|--------|
| bk_biz_id    | int64  | 是  | 使用业务ID |
| account_type | string | 否  | 账户类型   |

### 调用示例

```json
{
  "account_type": "resource"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000003",
    "vendor": "tcloud",
    "name": "Jim_account",
    "managers": [
      "hcm"
    ],
    "type": "resource",
    "site": "china",
    "price": "",
    "price_unit": "",
    "memo": "account create",
    "bk_biz_id": 13,
    "usage_biz_ids": [13,1111],
    "bk_biz_ids": [13,1111],
    "sync_status": "success",
    "sync_failed_reason":"",
    "creator": "Jim",
    "reviser": "Jim",
    "created_at": "2022-12-25T23:42:15Z",
    "updated_at": "2023-02-15T08:46:59Z",
    "rel_usage_biz_id": 1111,
    "rel_creator": "Jim",
    "rel_created_at": "2022-12-25T23:42:15Z"
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
| 参数名称               | 参数类型         | 描述                                                               |
|--------------------|--------------|------------------------------------------------------------------|
| id                 | string       | 账号ID                                                             |
| vendor             | string       | 供应商（枚举值：tcloud、aws、azure、gcp、huawei）                             |
| name               | string       | 名称                                                               |
| managers           | string array | 账号管理者                                                            |
| type               | string       | 账号类型 (枚举值：resource:资源账号、registration:登记账号、security_audit:安全审计账号) |
| site               | string       | 站点（枚举值：china:中国站、international:国际站）                              |
| price              | string       | 余额                                                               |
| price_unit         | string       | 余额单位                                                             |
| memo               | string       | 备注                                                               |
| bk_biz_id          | int64        | 管理业务                                                             |
| usage_biz_ids      | int64 array  | 使用业务                                                             |
| bk_biz_ids         | int64 array  | 旧的业务字段，用于兼容旧的api，值与使用业务的完全相同，不推荐使用                               |
| creator            | string       | 创建者                                                              |
| reviser            | string       | 更新者                                                              |
| sync_status        | string       | 资源同步状态                                                           |
| sync_failed_reason | string       | 资源同步失败原因                                                         |
| created_at         | string       | 创建时间，标准格式：2006-01-02T15:04:05Z                                   |
| updated_at         | string       | 更新时间，标准格式：2006-01-02T15:04:05Z                                   |
| rel_bk_biz_id      | int64        | 关联的使用业务ID                                                        |
| rel_creator        | string       | 关联创建者                                                            |
| rel_created_at     | string       | 关联创建时间，标准格式：2006-01-02T15:04:05Z                                 |
