### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：COS桶删除。
- 该接口功能描述：删除存储桶代理接口。

### URL

Delete /api/v1/cloud/cos/buckets/delete

### 输入参数

#### tcloud
代理接口文档地址：https://cloud.tencent.com/document/api/436/7732

| 参数名称       | 参数类型   | 必选 | 描述    |
|------------|--------|----|-------|
| account_id | string | 是  | 账号ID  |
| region     | string | 是  | 地域    |
| name       | string | 是  | 存储桶名称 |

### 调用示例

#### tcloud

```json
{
  "account_id": "0000001",
  "region": "ap-hk",
  "name": "xxx"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
