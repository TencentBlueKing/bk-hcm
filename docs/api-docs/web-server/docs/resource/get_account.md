### 描述

- 该接口提供版本：v1.2.1+。
- 该接口所需权限：账号查看。
- 该接口功能描述：查询指定账号。

### URL

GET /api/v1/cloud/accounts/{account_id}

### 输入参数

| 参数名称       | 参数类型   | 必选 | 描述   |
|------------|--------|----|------|
| account_id | string | 是  | 账号ID |

### 调用示例

```json
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
    "usage_biz_ids": [],
    "creator": "Jim",
    "reviser": "Jim",
    "created_at": "2022-12-25T23:42:15Z",
    "updated_at": "2023-02-15T08:46:59Z",
    "extension": {
      "cloud_main_account_id": "main-xxxxxx",
      "cloud_sub_account_id": "sub-xxxxxx",
      "cloud_secret_id": "xxxxx"
    }
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

| 参数名称                 | 参数类型         | 描述                                                               |
|----------------------|--------------|------------------------------------------------------------------|
| id                   | string       | 账号ID                                                             |
| vendor               | string       | 供应商（枚举值：tcloud、aws、azure、gcp、huawei、other）                       |
| name                 | string       | 名称                                                               |
| managers             | string array | 账号管理者                                                            |
| type                 | string       | 账号类型 (枚举值：resource:资源账号、registration:登记账号、security_audit:安全审计账号) |
| site                 | string       | 站点（枚举值：china:中国站、international:国际站）                              |
| price                | string       | 余额                                                               |
| price_unit           | string       | 余额单位                                                             |
| memo                 | string       | 备注                                                               |
| bk_biz_id            | int64        | 管理业务                                                             |
| usage_biz_ids        | int64 array  | 使用业务                                                             |
| bk_biz_ids           | int64 array  | 旧的业务字段，用于兼容旧的api，值与使用业务的完全相同，不推荐使用                               |
| recycle_reserve_time | int          | 回收站资源的保留时长，单位小时                                                  |
| sync_status          | string       | 资源同步状态                                                           |
| sync_failed_reason   | string       | 资源同步失败原因                                                         |
| creator              | string       | 创建者                                                              |
| reviser              | string       | 更新者                                                              |
| created_at           | string       | 创建时间，标准格式：2006-01-02T15:04:05Z                                   |
| updated_at           | string       | 更新时间，标准格式：2006-01-02T15:04:05Z                                   |
| extension            | object       | 混合云差异字段                                                          |

##### extension[tcloud]

| 参数名称                  | 参数类型   | 描述     |
|-----------------------|--------|--------|
| cloud_main_account_id | string | 云主账户ID |
| cloud_sub_account_id  | string | 云子账户ID |
| cloud_secret_id       | string | 云加密ID  |
| cloud_secret_key      | string | 云密钥    |

##### extension[aws]

| 参数名称               | 参数类型   | 描述      |
|--------------------|--------|---------|
| cloud_account_id   | string | 云账户ID   |
| cloud_iam_username | string | 云iam用户名 |
| cloud_secret_id    | string | 云加密ID   |
| cloud_secret_key   | string | 云密钥     |

##### extension[huawei]

| 参数名称                    | 参数类型   | 描述       |
|-------------------------|--------|----------|
| cloud_main_account_name | string | 云主账户名称   |
| cloud_sub_account_id    | string | 云子账户ID   |
| cloud_sub_account_name  | string | 云子账户名称   |
| cloud_iam_user_id       | string | 云iam用户ID |
| cloud_iam_username      | string | 云iam用户名  |
| cloud_secret_id         | string | 云加密ID    |
| cloud_secret_key        | string | 云密钥      |

##### extension[gcp]

| 参数名称                       | 参数类型   | 描述      |
|----------------------------|--------|---------|
| Email                      | string | 邮箱地址    |
| cloud_project_id           | string | 云项目ID   |
| cloud_project_name         | string | 云项目名称   |
| cloud_service_account_id   | string | 云服务账户ID |
| cloud_service_account_name | string | 云服务账户名称 |
| cloud_service_secret_id    | string | 云服务加密ID |
| cloud_service_secret_key   | string | 云服务密钥   |

##### extension[azure]

| 参数名称                    | 参数类型   | 描述       |
|-------------------------|--------|----------|
| display_name_name       | string | 展示名称     |
| cloud_tenant_id         | string | 云租户ID    |
| cloud_subscription_id   | string | 云订阅ID    |
| cloud_subscription_name | string | 云订阅名称    |
| cloud_application_id    | string | 云应用ID    |
| cloud_application_name  | string | 云应用名称    |
| cloud_client_secret_id  | string | 云客户端加密ID |
| cloud_client_secret_key | string | 云客户端密钥   |

#####  extension[other]

其他云厂商的extension目前为空
返回值为空对象
```json
{
  "extension": {}
}
```