### 描述

- 该接口提供版本：v1.2.1+。
- 该接口所需权限：账号录入。
- 该接口功能描述：创建用于创建账号的申请。

### URL

POST /api/v1/cloud/applications/types/add_account

### 输入参数

| 参数名称       | 参数类型         | 必选 | 描述                                                               |
|------------|--------------|----|------------------------------------------------------------------|
| vendor     | string       | 是  | 云厂商（枚举值：tcloud、aws、huawei、gcp、azure）                             |
| name       | string       | 是  | 名称                                                               |
| managers   | string array | 是  | 账号管理者                                                            |
| type       | string       | 是  | 账号类型 (枚举值：resource:资源账号、registration:登记账号、security_audit:安全审计账号) |
| site       | string       | 是  | 站点（枚举值：china:中国站、international:国际站）                              |
| memo       | string       | 否  | 备注                                                               |
| bk_biz_ids | int64 array  | 否  | 账号关联的业务ID列表，账号类型为资源账号时必填                                         |
| extension  | object       | 是  | 混合云差异字段                                                          |
| remark     | string       | 否  | 单据备注                                                             |

##### extension[tcloud]

| 参数名称                  | 参数类型   | 描述     |
|-----------------------|--------|--------|
| cloud_main_account_id | string | 云主账户ID |
| cloud_sub_account_id  | string | 云子账户ID |
| cloud_secret_id       | string | 云加密ID  |
| cloud_secret_key      | string | 云密钥    |

##### extension[aws]

| 参数名称               | 参数类型   | 必选 | 描述      |
|--------------------|--------|----|---------|
| cloud_account_id   | string | 是  | 云账户ID   |
| cloud_iam_username | string | 是  | 云iam用户名 |
| cloud_secret_id    | string | 否  | 云加密ID   |
| cloud_secret_key   | string | 否  | 云密钥     |

##### extension[huawei]

| 参数名称                   | 参数类型   | 必选 | 描述       |
|------------------------|--------|----|----------|
| cloud_sub_account_id   | string | 是  | 云子账户ID   |
| cloud_sub_account_name | string | 是  | 云子账户名称   |
| cloud_iam_user_id      | string | 是  | 云iam用户ID |
| cloud_iam_username     | string | 是  | 云iam用户名  |
| cloud_secret_id        | string | 否  | 云加密ID    |
| cloud_secret_key       | string | 否  | 云密钥      |

##### extension[gcp]

| 参数名称                       | 参数类型   | 必选 | 描述      |
|----------------------------|--------|----|---------|
| Email                      | string | 否  | 邮箱地址    |
| cloud_project_id           | string | 是  | 云项目ID   |
| cloud_project_name         | string | 是  | 云项目名称   |
| cloud_service_account_id   | string | 否  | 云服务账户ID |
| cloud_service_account_name | string | 否  | 云服务账户名称 |
| cloud_service_secret_id    | string | 否  | 云服务加密ID |
| cloud_service_secret_key   | string | 否  | 云服务密钥   |

##### extension[azure]

| 参数名称                    | 参数类型   | 必选 | 描述     |
|-------------------------|--------|----|--------|
| display_name_name       | string | 否  | 展示名称   |
| cloud_tenant_id         | string | 是  | 云租户ID  |
| cloud_subscription_id   | string | 是  | 云订阅ID  |
| cloud_subscription_name | string | 是  | 云订阅名称  |
| cloud_application_id    | string | 否  | 云应用ID  |
| cloud_application_name  | string | 否  | 云应用名称  |
| cloud_client_secret_key | string | 否  | 云客户端密钥 |

### 调用示例

#### TCloud

```json
{
  "vendor": "tcloud",
  "name": "jim",
  "managers": [
    "hcm"
  ],
  "type": "resource",
  "site": "china",
  "bk_biz_ids": [
    1010011010
  ],
  "extension": {
    "cloud_main_account_id": "main-xxxxxx",
    "cloud_sub_account_id": "sub-xxxxxx",
    "cloud_secret_id": "xxxxx",
    "cloud_secret_key": "xxxxxxxx"
  },
  "memo": ""
}
```

#### Aws

```json
{
  "vendor": "tcloud",
  "name": "jim",
  "managers": [
    "hcm"
  ],
  "type": "resource",
  "site": "china",
  "bk_biz_ids": [
    1010011010
  ],
  "extension": {
    "cloud_account_id": "main-xxxxxx",
    "cloud_iam_username": "sub-xxxxxx",
    "cloud_secret_id": "xxxxx",
    "cloud_secret_key": "xxxxxxxx"
  },
  "memo": ""
}
```

#### HuaWei

```json
{
  "vendor": "tcloud",
  "name": "jim",
  "managers": [
    "hcm"
  ],
  "type": "resource",
  "site": "china",
  "bk_biz_ids": [
    1010011010
  ],
  "extension": {
    "cloud_main_account_name": "main-xxxxxx",
    "cloud_sub_account_id": "sub-xxxxxx",
    "cloud_sub_account_name": "xxxxx",
    "cloud_iam_user_id": "xxxxxxxx",
    "cloud_iam_username": "xxxxxxxx",
    "cloud_secret_id": "xxxxxxxx",
    "cloud_secret_key": "xxxxxxxx"
  },
  "memo": ""
}
```

#### Gcp

```json
{
  "vendor": "tcloud",
  "name": "jim",
  "managers": [
    "hcm"
  ],
  "type": "resource",
  "site": "china",
  "bk_biz_ids": [
    1010011010
  ],
  "extension": {
    "cloud_project_id": "main-xxxxxx",
    "cloud_project_name": "sub-xxxxxx",
    "cloud_service_account_id": "xxxxx",
    "cloud_service_account_name": "xxxxxxxx",
    "cloud_service_secret_id": "xxxxxxxx",
    "cloud_service_secret_key": "xxxxxxxx"
  },
  "memo": ""
}
```

#### Azure

```json
{
  "vendor": "tcloud",
  "name": "jim",
  "managers": [
    "hcm"
  ],
  "type": "resource",
  "site": "china",
  "bk_biz_ids": [
    1010011010
  ],
  "extension": {
    "cloud_tenant_id": "main-xxxxxx",
    "cloud_subscription_id": "sub-xxxxxx",
    "cloud_subscription_name": "xxxxx",
    "cloud_application_id": "xxxxxxxx",
    "cloud_application_name": "xxxxxxxx",
    "cloud_client_secret_id": "xxxxxxxx",
    "cloud_client_secret_key": "xxxxxxxx"
  },
  "memo": ""
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

| 参数名称 | 参数类型   | 描述   |
|------|--------|------|
| id   | string | 单据ID |
