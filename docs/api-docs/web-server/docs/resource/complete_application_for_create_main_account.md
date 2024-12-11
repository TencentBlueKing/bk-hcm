### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：一级账号管理。
- 该接口功能描述：单据信息填写。

### URL

POST /api/v1/cloud/applications/types/complete_main_account

## 请求参数
| 参数名称            | 参数类型   | 必选 | 描述                                      |
|-----------------|--------|----|-----------------------------------------|
| sn              | string | 是  | 单据编号                                    |
| id              | string | 是  | 海垒单据编号                                  |
| vendor          | string | 是  | 云厂商，aws/gcp/azure/huawei/zenlayer/kaopu |
| root_account_id | string | 是  | 一级账号的id                                 |
| extension       | object | 是  | 各云厂商传递的参数不一致                            |

### extension

#### aws/gcp

null,自动创建流程不需要该字段

#### azure

| 参数名称                    | 参数类型   | 必选 | 描述   |
|-------------------------|--------|----|------|
| cloud_subscription_name | string | 是  | 订阅名  |
| cloud_subscription_id   | string | 是  | 订阅ID |
| cloud_init_password     | string | 是  | 初始密码 |


#### huawei
| 参数名称                    | 参数类型   | 必选 | 描述     |
|-------------------------|--------|----|--------|
| cloud_main_account_name | string | 是  | 二级账号名  |
| cloud_main_account_id   | string | 是  | 二级账号ID |
| cloud_init_password     | string | 是  | 初始密码   |

#### zenlayer/kaopu
| 参数名称                    | 参数类型   | 必选 | 描述     |
|-------------------------|--------|----|--------|
| cloud_main_account_name | string | 是  | 二级账号名  |
| cloud_main_account_id   | string | 是  | 二级账号ID |
| cloud_init_password     | string | 是  | 初始密码   |


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
