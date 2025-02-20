### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：无。
- 该接口功能描述：创建 “创建二级账号”的申请单。

### URL

POST /api/v1/cloud/applications/types/create_main_account

## 请求参数
| 参数名称          | 参数类型         | 必选 | 描述                                      |
|---------------|--------------|----|-----------------------------------------|
| vendor        | string       | 是  | 云厂商，aws/gcp/azure/huawei/zenlayer/kaopu |
| site          | string       | 是  | 站点类型 international/china                |
| email         | string       | 是  | 邮箱地址，提交前需要通过邮箱验证                        |
| op_product_id | int          | 是  | 运营产品ID                                  |
| business_type | string       | 是  | 业务使用范围 international/china              |
| managers      | string array | 是  | 负责人列表，英文逗号分隔                            |
| bak_managers  | string array | 是  | 备份负责人列表，英文逗号分隔                          |
| dept_id       | int          | 是  | 部门ID,要求是3级部门（不含公司，BG作为1级开始数）            |
| bk_biz_id     | int          | 是  | 业务ID                                    |
| memo          | string       | 是  | 账号用途,512个字符                             |
| extension     | object       | 是  | 各云厂商区别处理                                |

### extension

各云厂商区别处理，根据云厂商不同，需要提供不同的参数。

#### AWS

| 参数名称                    | 参数类型   | 必选 | 描述   |
|-------------------------|--------|----|------|
| cloud_main_account_name | string | 是  | 主账号名 |


#### GCP

| 参数名称               | 参数类型   | 必选 | 描述  |
|--------------------|--------|----|-----|
| cloud_project_name | string | 是  | 项目名 |


#### Huawei

| 参数名称                    | 参数类型   | 必选 | 描述   |
|-------------------------|--------|----|------|
| cloud_main_account_name | string | 是  | 主账号名 |


#### Azure

| 参数名称                    | 参数类型   | 必选 | 描述  |
|-------------------------|--------|----|-----|
| cloud_subscription_name | string | 是  | 订阅名 |


#### Zenlayer

| 参数名称                    | 参数类型   | 必选 | 描述   |
|-------------------------|--------|----|------|
| cloud_main_account_name | string | 是  | 主账号名 |


#### Kaopu

| 参数名称                    | 参数类型   | 必选 | 描述   |
|-------------------------|--------|----|------|
| cloud_main_account_name | string | 是  | 主账号名 |



### 响应数据
```
{
    "code": 0,
    "message": "",
    "data": {
        "id": "xxxxxx"              // string, 海垒申请单ID，非ITSM申请单ID
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
