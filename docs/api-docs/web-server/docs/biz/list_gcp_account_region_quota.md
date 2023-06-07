### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询谷歌云账号地域配额。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/gcp/accounts/{account_id}/regions/quotas

### 请求参数

| 参数名称       | 参数类型   | 必选  | 描述   |
|------------|--------|-----|------|
| bk_biz_id  | int64  | 是   | 业务ID |
| account_id | string | 是   | 账号ID |
| vendor     | string | 是   | 供应商  |
| region     | string | 是   | 地域   |

### 调用示例

#### 请求参数示例

```json
{
  "region": "us-central1"
}
```

#### 返回参数示例

查询腾讯云机型列表返回参数。
```json
{
  "code": 0,
  "message": "",
  "data": {
    "instance": {
      "limit": 24000,
      "usage": 16
    }
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述   |
|---------|--------------|------|
| code    | int          | 状态码  |
| message | string       | 请求信息 |
| data    | object array | 响应数据 |

#### data[tcloud]

| 参数名称     | 参数类型   | 描述         |
|----------|--------|------------|
| instance | ResourceQuota | 主机实例的配额信息。 |

#### ResourceQuota

| 参数名称     | 参数类型   | 描述        |
|----------|--------|-----------|
| limit | int32 | 资源最大限制。   |
| usage | int32 | 资源已使用的数量。 |
