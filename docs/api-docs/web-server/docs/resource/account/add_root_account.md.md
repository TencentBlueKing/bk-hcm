### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：一级账号管理。
- 该接口功能描述：一级账号录入。

### URL

POST /api/v1/account/root_accounts/add

### 请求参数
| 参数名称         | 参数类型         | 必选 | 描述     |
|--------------|--------------|----|--------|
| name         | string       | 是  | 名字     |
| vendor       | string       | 是  | 云厂商    |
| email        | string       | 是  | 邮箱     |
| managers     | string array | 是  | 负责人    |
| bak_managers | string array | 是  | 备份负责人  |
| site         | string       | 是  | 站点     |
| dept_id      | int          | 是  | 组织架构ID |
| memo         | string       | 否  | 备忘录    |
| extension    | object       | 否  | 扩展字段   |

#### extension字段说明

##### aws

| 参数名称               | 参数类型   | 必选 | 描述      |
|--------------------|--------|----|---------|
| cloud_account_id   | string | 是  | 云账号ID   |
| cloud_iam_username | string | 是  | 云IAM用户名 |
| cloud_secret_id    | string | 是  | 云密钥ID   |
| cloud_secret_key   | string | 是  | 云密钥KEY  |

##### gcp

| 参数名称                       | 参数类型   | 必选 | 描述       |
|----------------------------|--------|----|----------|
| cloud_project_name         | string | 是  | 云项目名     |
| cloud_project_id           | string | 是  | 云项目ID    |
| cloud_service_account_id   | string | 是  | 云服务账号ID  |
| cloud_service_account_name | string | 是  | 云服务账号名   |
| cloud_service_secret_id    | string | 是  | 云服务密钥ID  |
| cloud_service_secret_key   | string | 是  | 云服务密钥KEY |

##### azure
| 参数名称                    | 参数类型   | 必选 | 描述     |
|-------------------------|--------|----|--------|
| display_name_name       | string | 是  | 显示名称名  |
| cloud_tenant_id         | string | 是  | 云租户ID  |
| cloud_subscription_id   | string | 是  | 云订阅ID  |
| cloud_subscription_name | string | 是  | 云订阅名   |
| cloud_application_id    | string | 是  | 云应用ID  |
| cloud_application_name  | string | 是  | 云应用名   |
| cloud_client_secret_key | string | 是  | 云客户端密钥 |


##### huawei
| 参数名称                   | 参数类型   | 必选 | 描述       |
|------------------------|--------|----|----------|
| cloud_sub_account_name | string | 是  | 二级账号名    |
| cloud_sub_account_id   | string | 是  | 二级账号ID   |
| cloud_secret_id        | string | 是  | 云密钥ID    |
| cloud_secret_key       | string | 是  | 云密钥KEY   |
| cloud_iam_user_id      | string | 是  | 云IAM用户ID |
| cloud_iam_username     | string | 是  | 云IAM用户名  |

##### zenlayer/kaopu
| 参数名称             | 参数类型   | 必选 | 描述    |
|------------------|--------|----|-------|
| cloud_account_id | string | 是  | 云账号ID |



### 响应数据
```
{
    "code": 0,
    "message": "",
    "data": {
        "id": "xxxx"           
    }
}

```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | array  | 响应数据 |

#### data

| 参数名称                | 参数类型         | 描述        |
|---------------------|--------------|-----------|
| id                  | string       | AccountID |
