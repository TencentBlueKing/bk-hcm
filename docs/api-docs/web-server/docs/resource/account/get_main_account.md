### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：二级账号查看权限。
- 该接口功能描述：获取二级账号。

### URL

POST /api/v1/account/main_accounts/{account_id}

## 请求参数
| 参数名称       | 参数类型   | 必选 | 描述      |
|------------|--------|----|---------|
| account_id | string | 是  | 二级账号ID。 |


### 响应数据
```
{
    "code": 0,
    "message": "",
    "data": {
        "id": "xxxx",                           // id
        "vendor": "aws",                        // string,云厂商
        "email": "xxxx@tencent.com",            // 邮箱
        "cloud_id": "xxx",						  // 云账号id
        "parent_account_name": "xxx",           // 所属一级账号名
        "parent_account_id": "xxxx",            // 所属一级账号id
        "site": "international",                // string,站点
        "business_type": "internal",            // string,业务类型
        "managers": ["xxx","xxx"],        // string,负责人
        "bak_managers": ["xxx","xxx"],    // string,备份负责人
        "dept_id": 1234,                        // int,组织架构ID
        "op_product_id": 1234,                  // int,运营产品ID
        "bk_biz_id": 1312,                      // int,业务ID
        "status": "xxxx",                       // string,账号状态
        "memo": "xxxxx",                        // string,备忘录
        "creator": "xx",                        // string,创建者
        "reviser": "",                          // string,修改者
        "created_at": "",                       // string,创建时间
        "updated_at": "",                        // string,修改时间
        "extension": {}						  // 混合云差异字段，见extension说明
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

| 参数名称                | 参数类型         | 描述         |
|---------------------|--------------|------------|
| id                  | string       | 资源ID       |
| vendor              | string       | 云厂商        |
| email               | string       | 邮箱         |
| cloud_id            | string       | 云ID        |
| parent_account_name | string       | 父账号名称      |
| parent_account_id   | string       | 父账号ID      |
| site                | string       | 站点         |
| business_type       | string       | 业务类型       |
| status              | string       | 状态         |
| managers            | string array | 主账号管理者列表   |
| bak_managers        | string array | 主账号备份管理者列表 |
| dept_id             | int          | 部门ID       |
| op_product_id       | int          | 运营产品ID     |
| bk_biz_id           | int          | 业务ID       |
| status              | string       | 状态         |
| memo                | string       | 备注         |
| creator             | string       | 创建者        |
| reviser             | string       | 修改者        |
| created_at          | string       | 创建时间       |
| updated_at          | string       | 修改时间       |
| extension           | object       | 扩展字段       |

#### extension

##### aws

| 参数名称                    | 参数类型   | 描述     |
|-------------------------|--------|--------|
| cloud_main_account_name | string | 二级账号名  |
| cloud_main_account_id   | string | 二级账号ID |

##### gcp

| 参数名称               | 参数类型   | 描述    |
|--------------------|--------|-------|
| cloud_project_name | string | 云项目名  |
| cloud_project_id   | string | 云项目ID |

##### azure

| 参数名称                    | 参数类型   | 描述   |
|-------------------------|--------|------|
| cloud_subscription_name | string | 订阅名  |
| cloud_subscription_id   | string | 订阅ID |

##### huawei/zenlayer/kaopu
| 参数名称                    | 参数类型   | 描述     |
|-------------------------|--------|--------|
| cloud_main_account_name | string | 二级账号名  |
| cloud_main_account_id   | string | 二级账号ID |

