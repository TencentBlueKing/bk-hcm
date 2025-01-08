### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：一级账号管理。
- 该接口功能描述：修改一级账号。

### URL

PATCH /api/v1/account/root_accounts/{account_id}

### 请求参数

| 参数名称         | 是否必选 | 类型           | 描述              |
|--------------|------|--------------|-----------------|
| account_id   | 是    | string       | 一级账号ID          |
| name         | 否    | string       | 名称              |
| managers     | 否    | string array | 主账号管理者列表，最大5个   |
| bak_managers | 否    | string array | 主账号备份管理者列表，最大5个 |
| memo         | 否    | string       | 备注              |
| dept_id      | 否    | int64        | 部门ID            |
| extension    | 否    | object       | 扩展字段            |

#### extension字段说明

##### aws

| 参数名称               | 参数类型   | 必选 | 描述      |
|--------------------|--------|----|---------|
| cloud_iam_username | string | 是  | 云IAM用户名 |
| cloud_secret_id    | string | 是  | 云密钥ID   |
| cloud_secret_key   | string | 是  | 云密钥KEY  |

##### gcp

| 参数名称                       | 参数类型   | 必选 | 描述       |
|----------------------------|--------|----|----------|
| cloud_project_name         | string | 是  | 云项目名     |
| cloud_service_account_id   | string | 是  | 云服务账号ID  |
| cloud_service_account_name | string | 是  | 云服务账号名   |
| cloud_service_secret_id    | string | 是  | 云服务密钥ID  |
| cloud_service_secret_key   | string | 是  | 云服务密钥KEY |

##### azure
| 参数名称                    | 参数类型   | 必选 | 描述     |
|-------------------------|--------|----|--------|
| cloud_tenant_id         | string | 是  | 云租户ID  |
| cloud_subscription_name | string | 是  | 云订阅名   |
| cloud_application_id    | string | 是  | 云应用ID  |
| cloud_application_name  | string | 是  | 云应用名   |
| cloud_client_secret_key | string | 是  | 云客户端密钥 |


##### huawei
| 参数名称                   | 参数类型   | 必选 | 描述       |
|------------------------|--------|----|----------|
| cloud_sub_account_name | string | 是  | 二级账号名    |
| cloud_secret_id        | string | 是  | 云密钥ID    |
| cloud_secret_key       | string | 是  | 云密钥KEY   |
| cloud_iam_user_id      | string | 是  | 云IAM用户ID |
| cloud_iam_username     | string | 是  | 云IAM用户名  |

##### zenlayer/kaopu

null,无需传值


### 响应数据
```
{
    "code": 0,
    "message": "",
    "data": {}
}

```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |
