### 描述

- 该接口提供版本：v1.2.1+
- 该接口所需权限：账号录入。
- 该接口功能描述：通过秘钥获取账号信息。 禁用主账号重复检查需要版本v1.4.4+

### URL

POST /api/v1/cloud/vendors/{vendor}/accounts/secret

#### 路径参数说明

| 参数名称   | 参数类型   | 必选 | 描述                                   |
|--------|--------|----|--------------------------------------|
| vendor | string | 是  | 云厂商（枚举值：tcloud、aws、huawei、gcp、azure） |

### 输入参数

#### TCloud

| 参数名称             | 参数类型   | 必选 | 描述        |
|------------------|--------|----|-----------|
| cloud_secret_id  | string | 是  | 云加密ID     |
| cloud_secret_key | string | 是  | 云密钥       |
| disable_check    | bool   | 否  | 禁用主账号重复检查 |

#### AWS

| 参数名称             | 参数类型   | 必选 | 描述        |
|------------------|--------|----|-----------|
| cloud_secret_id  | string | 是  | 云加密ID     |
| cloud_secret_key | string | 是  | 云密钥       |
| disable_check    | bool   | 否  | 禁用主账号重复检查 |

#### Azure

| 参数名称                    | 参数类型   | 必选 | 描述        |
|-------------------------|--------|----|-----------|
| cloud_tenant_id         | string | 是  | 云租户ID     |
| cloud_application_id    | string | 是  | 云应用ID     |
| cloud_client_secret_key | string | 是  | 云客户端密钥    |
| disable_check           | bool   | 否  | 禁用主账号重复检查 |

#### GCP

| 参数名称                     | 参数类型   | 必选 | 描述        |
|--------------------------|--------|----|-----------|
| cloud_service_secret_key | string | 是  | 云服务密钥     |
| disable_check            | bool   | 否  | 禁用主账号重复检查 |

#### Huawei

| 参数名称             | 参数类型   | 必选 | 描述        |
|------------------|--------|----|-----------|
| cloud_secret_id  | string | 是  | 云加密ID     |
| cloud_secret_key | string | 是  | 云密钥       |
| disable_check    | bool   | 否  | 禁用主账号重复检查 |

### 调用示例

#### TCloud

```json
{
  "disable_check": true,
  "cloud_secret_id": "xxxx",
  "cloud_secret_key": "xxxx"
}
```

#### Aws

```json
{
  "cloud_secret_id": "xxxx",
  "cloud_secret_key": "xxxx"
}
```

#### Azure

```json
{
  "cloud_tenant_id": "0000000",
  "cloud_application_id": "xxxxxx",
  "cloud_client_secret_key": "xxxxxx"
}
```

#### Gcp

```json
{
  "cloud_service_secret_key": "{xxxx:xxx}"
}
```

#### HuaWei

```json
{
  "cloud_secret_id": "xxxx",
  "cloud_secret_key": "xxxx"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "cloud_main_account_id": "00000001",
    "cloud_sub_account_id": "xxxx"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data[tcloud]

| 参数名称                  | 参数类型   | 描述     |
|-----------------------|--------|--------|
| cloud_main_account_id | string | 云主账户ID |
| cloud_sub_account_id  | string | 云子账户ID |

#### data[aws]

| 参数名称               | 参数类型   | 描述      |
|--------------------|--------|---------|
| cloud_account_id   | string | 云账户ID   |
| cloud_iam_username | string | 云iam用户名 |

#### data[huawei]

| 参数名称                   | 参数类型   | 描述       |
|------------------------|--------|----------|
| cloud_sub_account_id   | string | 云子账户ID   |
| cloud_sub_account_name | string | 云子账户名称   |
| cloud_iam_user_id      | string | 云iam用户ID |
| cloud_iam_username     | string | 云iam用户名  |

#### data[gcp]

| 参数名称                       | 参数类型   | 描述      |
|----------------------------|--------|---------|
| Email                      | string | 邮箱地址    |
| cloud_project_id           | string | 云项目ID   |
| cloud_project_name         | string | 云项目名称   |
| cloud_service_account_id   | string | 云服务账户ID |
| cloud_service_account_name | string | 云服务账户名称 |
| cloud_service_secret_id    | string | 云服务秘钥ID |

#### data[azure]

| 参数名称                    | 参数类型   | 描述    |
|-------------------------|--------|-------|
| cloud_subscription_id   | string | 云订阅ID |
| cloud_subscription_name | string | 云订阅名称 |
| cloud_application_name  | string | 云应用名称 |

