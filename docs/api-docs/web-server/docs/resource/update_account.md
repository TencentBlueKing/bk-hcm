### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：账号编辑。
- 该接口功能描述：更新指定账号。

### URL

PATCH /api/v1/cloud/accounts/{account_id}

### 输入参数

| 参数名称                | 参数类型        | 必选 | 描述                        |
|---------------------|-------------|----|---------------------------|
| account_id          | string      | 是  | 账号ID                      |
| name                | string      | 否  | 名称                        |
| managers            | string      | 否  | 账号管理者                     |
| memo                | string      | 否  | 备注                        |
| recycle_reserve_time | int         | 否  | 回收站资源的保留时长，单位小时           |
| bk_biz_id            | int64       | 否  | 管理业务，非资源账号不允许传递该参数        |
| usage_biz_ids       | int64 array | 否  | 使用业务，非资源账号若填写该字段则数组长度必须为1 |
| extension           | object      | 否  | 混合云差异字段                   |

#### extension[tcloud]

| 参数名称                 | 参数类型   | 必选 | 描述     |
|----------------------|--------|----|--------|
| cloud_sub_account_id | string | 是  | 云子账户ID |
| cloud_secret_id      | string | 否  | 云加密ID  |
| cloud_secret_key     | string | 否  | 云密钥    |

#### extension[aws]

| 参数名称               | 参数类型   | 必选 | 描述      |
|--------------------|--------|----|---------|
| cloud_iam_username | string | 是  | 云iam用户名 |
| cloud_secret_id    | string | 否  | 云加密ID   |
| cloud_secret_key   | string | 否  | 云密钥     |

#### extension[huawei]

| 参数名称               | 参数类型   | 必选 | 描述       |
|--------------------|--------|----|----------|
| cloud_iam_user_id  | string | 是  | 云iam用户ID |
| cloud_iam_username | string | 是  | 云iam用户名  |
| cloud_secret_id    | string | 否  | 云加密ID    |
| cloud_secret_key   | string | 否  | 云密钥      |

#### extension[gcp]

| 参数名称                       | 参数类型   | 必选 | 描述      |
|----------------------------|--------|----|---------|
| Email                      | string | 否  | 邮箱地址    |
| cloud_service_account_id   | string | 否  | 云服务账户ID |
| cloud_service_account_name | string | 否  | 云服务账户名称 |
| cloud_service_secret_id    | string | 否  | 云服务加密ID |
| cloud_service_secret_key   | string | 否  | 云服务密钥   |

#### extension[azure]

| 参数名称                    | 参数类型   | 必选 | 描述       |
|-------------------------|--------|----|----------|
| display_name_name       | string | 否  | 展示名称     |
| cloud_application_id    | string | 否  | 云应用ID    |
| cloud_application_name  | string | 否  | 云应用名称    |
| cloud_client_secret_id  | string | 否  | 云客户端加密ID |
| cloud_client_secret_key | string | 否  | 云客户端密钥   |

### 调用示例

```json
{
  "name": "tcloud_account",
  "managers": [
    "hcm"
  ],
  "memo": "account update",
  "bk_biz_id": 310,
  "usage_biz_ids": [
    310
  ],
  "extension": {
    "cloud_sub_account_id": "sub-xxxxxx",
    "cloud_secret_id": "xxxxx",
    "cloud_secret_key": "xxxxxxxx"
  }
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

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |
