### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：账号查看。
- 该接口功能描述：检查指定账号。

### URL

POST /api/v1/cloud/accounts/check

### 输入参数

| 参数名称       | 参数类型   | 必选  | 描述                                             |
|------------|--------|-----|------------------------------------------------|
| vendor     | string | 是   | 云厂商（枚举值：tcloud、aws、huawei、gcp、azure）           |
| type       | string | 是   | 账户类型（枚举值：resource、registration、security_audit） |
| extension  | object | 是   | 混合云差异字段                                        |

#### extension[tcloud]

| 参数名称                  | 参数类型    | 必选 | 描述     |
|-----------------------|---------|----|--------|
| cloud_main_account_id | string  | 是  | 云主账户ID |
| cloud_sub_account_id  | string  | 是  | 云子账户ID |
| cloud_secret_id       | string  | 否  | 云加密ID  |
| cloud_secret_key      | string  | 否  | 云密钥    |

#### extension[aws]

| 参数名称                 | 参数类型    | 必选 | 描述      |
|----------------------|---------|----|---------|
| cloud_account_id     | string  | 是  | 云账户ID   |
| cloud_iam_username   | string  | 是  | 云iam用户名 |
| cloud_secret_id      | string  | 否  | 云加密ID   |
| cloud_secret_key     | string  | 否  | 云密钥     |

#### extension[huawei]

| 参数名称                    | 参数类型    | 必选 | 描述        |
|-------------------------|---------|----|-----------|
| cloud_main_account_name | string  | 是  | 云主账户名称    |
| cloud_sub_account_id    | string  | 是  | 云子账户ID    |
| cloud_sub_account_name  | string  | 是  | 云子账户名称    |
| cloud_iam_user_id       | string  | 是  | 云iam用户ID  |
| cloud_iam_username      | string  | 是  | 云iam用户名   |
| cloud_secret_id         | string  | 否  | 云加密ID     |
| cloud_secret_key        | string  | 否  | 云密钥       |

#### extension[gcp]

| 参数名称                       | 参数类型    | 必选 | 描述       |
|----------------------------|---------|----|----------|
| cloud_project_id           | string  | 是  | 云项目ID    |
| cloud_project_name         | string  | 是  | 云项目名称    |
| cloud_service_account_id   | string  | 否  | 云服务账户ID  |
| cloud_service_account_name | string  | 否  | 云服务账户名称  |
| cloud_service_secret_id    | string  | 否  | 云服务加密ID  |
| cloud_service_secret_key   | string  | 否  | 云服务密钥    |

#### extension[azure]

| 参数名称                     | 参数类型    | 必选 | 描述          |
|--------------------------|---------|----|-------------|
| cloud_tenant_id          | string  | 是  | 云租户ID       |
| cloud_subscription_id    | string  | 是  | 云订阅ID       |
| cloud_subscription_name  | string  | 是  | 云订阅名称       |
| cloud_application_id     | string  | 否  | 云应用ID       |
| cloud_application_name   | string  | 否  | 云应用名称       |
| cloud_client_secret_id   | string  | 否  | 云客户端加密ID    |
| cloud_client_secret_key  | string  | 否  | 云客户端密钥      |

### 调用示例

#### TCloud
```json
{
  "vendor": "tcloud",
  "type": "resource",
  "extension": {
    "cloud_main_account_id": "0000000",
    "cloud_sub_account_id": "0000000"
  }
}
```

#### Aws
```json
{
  "vendor": "aws",
  "type": "resource",
  "extension": {
    "cloud_account_id": "0000000",
    "cloud_iam_username": "0000000"
  }
}
```

#### HuaWei
```json
{
  "vendor": "huawei",
  "type": "resource",
  "extension": {
    "cloud_main_account_name": "xxxxxx",
    "cloud_sub_account_id": "0000000",
    "cloud_sub_account_name": "xxxxxx",
    "cloud_iam_user_id": "0000000",
    "cloud_iam_username": "xxxxxx"
  }
}
```

#### Gcp
```json
{
  "vendor": "gcp",
  "type": "resource",
  "extension": {
    "cloud_project_id": "0000000",
    "cloud_project_name": "xxxxxx"
  }
}
```

#### Azure
```json
{
  "vendor": "azure",
  "type": "resource",
  "extension": {
    "cloud_tenant_id": "0000000",
    "cloud_subscription_id": "xxxxxx",
    "cloud_subscription_name": "xxxxxx"
  }
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
