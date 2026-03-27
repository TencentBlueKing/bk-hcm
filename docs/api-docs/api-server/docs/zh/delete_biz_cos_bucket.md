### 描述

- 该接口提供版本：v1.8.10.5+。
- 该接口所需权限：业务-COS桶删除。
- 该接口功能描述：删除指定业务的存储桶接口。

### URL

DELETE /api/v1/cloud/bizs/{bk_biz_id}/cos/buckets/delete

### 输入参数

#### tcloud-ziyan

| 参数名称       | 参数类型   | 必选 | 描述    |
|------------|--------|----|-------|
| account_id | string | 是  | 账号ID  |
| region     | string | 是  | 地域    |
| cloud_name | string | 是  | 存储桶名称 |

### 调用示例

#### tcloud

```json
{
  "account_id": "0000001",
  "region": "ap-nanjing",
  "cloud_name": "xxx"
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
