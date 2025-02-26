### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：一级账号管理。
- 该接口功能描述：获取一级账号。

### URL

GET /api/v1/account/root_accounts/{account_id}

### 请求参数

| 参数名称       | 是否必选 | 类型     | 描述     |
|------------|------|--------|--------|
| account_id | 是    | string | 一级账号ID |


### 响应数据
```
{
    "code": 0,
    "message": "",
    "data": {
        "id": "xxxx",                           // id
        "name": "xxx",                          // 名字
        "vendor": "aws",                        // string,云厂商
        "cloud_id": "xxxx",                     // string,云ID
        "email": "xxxx@tencent.com",            // 邮箱
        "managers": ["xxx","xxx"],    // string,负责人，最大5个
        "bak_managers": ["xxx","xxx"],// string,备份负责人
        "site": "international",                // string,站点
        "dept_id": 1234,                        // int,组织架构ID
        "memo": "xxxxx",                        // string,备忘录
        "creator": "xx",                        // string,创建者
        "reviser": "",                          // string,修改者
        "created_at": "",                       // string,创建时间
        "updated_at": ""                        // string,修改时间
        "extension": {}
    }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述               |
|---------|--------|------------------|
| code    | int32  | 状态码              |
| message | string | 请求信息             |
| data    | object | 不同的vendor会有不同的响应 |

#### vendor=aws, data说明

| 参数名称         | 参数类型         | 描述         |
|--------------|--------------|------------|
| id           | string       | 资源ID       |
| name         | string       | 名称         |
| vendor       | string       | 云厂商        |
| cloud_id     | string       | 云ID        |
| email        | string       | 邮箱         |
| managers     | string array | 主账号管理者列表   |
| bak_managers | string array | 主账号备份管理者列表 |
| site         | string       | 站点         |
| dept_id      | int          | 部门ID       |
| memo         | string       | 备注         |
| creator      | string       | 创建者        |
| reviser      | string       | 修改者        |
| created_at   | string       | 创建时间       |
| updated_at   | string       | 修改时间       |
| extension    | object       | 扩展信息       |

##### extension说明

| 参数名称               | 参数类型   | 描述      |
|--------------------|--------|---------|
| cloud_account_id   | string | 云账号ID   |
| cloud_iam_username | string | 云IAM用户名 |
| cloud_secret_id    | string | 云密钥ID   |
| cloud_secret_key   | string | 云密钥KEY  |

#### vendor=gcp

| 参数名称         | 参数类型         | 描述         |
|--------------|--------------|------------|
| id           | string       | 资源ID       |
| name         | string       | 名称         |
| vendor       | string       | 云厂商        |
| cloud_id     | string       | 云ID        |
| email        | string       | 邮箱         |
| managers     | string array | 主账号管理者列表   |
| bak_managers | string array | 主账号备份管理者列表 |
| site         | string       | 站点         |
| dept_id      | int          | 部门ID       |
| memo         | string       | 备注         |
| creator      | string       | 创建者        |
| reviser      | string       | 修改者        |
| created_at   | string       | 创建时间       |
| updated_at   | string       | 修改时间       |
| extension    | object       | 扩展信息       |

##### extension说明

| 参数名称                       | 参数类型   | 描述       |
|----------------------------|--------|----------|
| cloud_project_name         | string | 云项目名     |
| cloud_project_id           | string | 云项目ID    |
| cloud_service_account_id   | string | 云服务账号ID  |
| cloud_service_account_name | string | 云服务账号名   |
| cloud_service_secret_id    | string | 云服务密钥ID  |
| cloud_service_secret_key   | string | 云服务密钥KEY |
| cloud_billing_account      | string | 云账单账号    |
| cloud_organization         | string | 云组织      |


#### vendor=huawei
| 参数名称         | 参数类型         | 描述         |
|--------------|--------------|------------|
| id           | string       | 资源ID       |
| name         | string       | 名称         |
| vendor       | string       | 云厂商        |
| cloud_id     | string       | 云ID        |
| email        | string       | 邮箱         |
| managers     | string array | 主账号管理者列表   |
| bak_managers | string array | 主账号备份管理者列表 |
| site         | string       | 站点         |
| dept_id      | int          | 部门ID       |
| memo         | string       | 备注         |
| creator      | string       | 创建者        |
| reviser      | string       | 修改者        |
| created_at   | string       | 创建时间       |
| updated_at   | string       | 修改时间       |
| extension    | object       | 扩展信息       |

##### extension说明

| 参数名称                   | 参数类型   | 描述       |
|------------------------|--------|----------|
| cloud_sub_account_name | string | 二级账号名    |
| cloud_sub_account_id   | string | 二级账号ID   |
| cloud_secret_id        | string | 云密钥ID    |
| cloud_secret_key       | string | 云密钥KEY   |
| cloud_iam_user_id      | string | 云IAM用户ID |
| cloud_iam_username     | string | 云IAM用户名  |

#### vendor=azure

| 参数名称         | 参数类型         | 描述         |
|--------------|--------------|------------|
| id           | string       | 资源ID       |
| name         | string       | 名称         |
| vendor       | string       | 云厂商        |
| cloud_id     | string       | 云ID        |
| email        | string       | 邮箱         |
| managers     | string array | 主账号管理者列表   |
| bak_managers | string array | 主账号备份管理者列表 |
| site         | string       | 站点         |
| dept_id      | int          | 部门ID       |
| memo         | string       | 备注         |
| creator      | string       | 创建者        |
| reviser      | string       | 修改者        |
| created_at   | string       | 创建时间       |
| updated_at   | string       | 修改时间       |
| extension    | object       | 扩展信息       |

##### extension说明

| 参数名称                    | 参数类型   | 描述       |
|-------------------------|--------|----------|
| display_name_name       | string | 显示名称名    |
| cloud_tenant_id         | string | 云租户ID    |
| cloud_subscription_id   | string | 云订阅ID    |
| cloud_subscription_name | string | 云订阅名     |
| cloud_application_id    | string | 云应用ID    |
| cloud_application_name  | string | 云应用名     |
| cloud_client_secret_id  | string | 云客户端密钥ID |
| cloud_client_secret_key | string | 云客户端密钥   |

#### vendor=kaopu/zenlayer

| 参数名称         | 参数类型         | 描述         |
|--------------|--------------|------------|
| id           | string       | 资源ID       |
| name         | string       | 名称         |
| vendor       | string       | 云厂商        |
| cloud_id     | string       | 云ID        |
| email        | string       | 邮箱         |
| managers     | string array | 主账号管理者列表   |
| bak_managers | string array | 主账号备份管理者列表 |
| site         | string       | 站点         |
| dept_id      | int          | 部门ID       |
| memo         | string       | 备注         |
| creator      | string       | 创建者        |
| reviser      | string       | 修改者        |
| created_at   | string       | 创建时间       |
| updated_at   | string       | 修改时间       |
| extension    | object       | 扩展信息       |

##### extension说明

| 参数名称             | 参数类型   | 描述    |
|------------------|--------|-------|
| cloud_account_id | string | 云账号ID |
