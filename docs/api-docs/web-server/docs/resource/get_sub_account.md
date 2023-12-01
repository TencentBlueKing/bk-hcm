### 描述

- 该接口提供版本：v1.2.1+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询指定子账号。

### URL

GET /api/v1/cloud/sub_accounts/{id}

### 输入参数

| 参数名称 | 参数类型   | 必选 | 描述    |
|------|--------|----|-------|
| id   | string | 是  | 子账号ID |

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000042",
    "cloud_id": "13943695",
    "name": "jim",
    "vendor": "tcloud",
    "site": "china",
    "account_id": "00000003",
    "account_type": "current_account",
    "managers": [],
    "bk_biz_ids": [],
    "memo": "",
    "creator": "jim",
    "reviser": "jim",
    "created_at": "2023-08-04T17:36:39Z",
    "updated_at": "2023-08-04T17:36:39Z",
    "extension": {
      "cloud_main_account_id": "main-xxxxxx",
      "uin": 1000269290009,
      "nick_name": "jim",
      "create_time": "2022-08-13 11:03:29"
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

| 参数名称         | 参数类型         | 描述                                             |
|--------------|--------------|------------------------------------------------|
| id           | string       | 账号ID                                           |
| vendor       | string       | 供应商（枚举值：tcloud、aws、azure、gcp、huawei）           |
| name         | string       | 名称                                             |
| cloud_id     | string       | 账号云ID                                          |
| account_id   | string       | 子账号所属资源账号ID                                    |
| account_type | string       | 账号类型 当前账号(current_account)   主账号(main_account) |
| managers     | string array | 账号管理者                                          |
| bk_biz_ids   | int64 array  | 账号关联的业务ID列表                                    |
| site         | string       | 站点（枚举值：china:中国站、international:国际站）            |
| memo         | string       | 备注                                             |
| creator      | string       | 创建者                                            |
| reviser      | string       | 更新者                                            |
| created_at   | string       | 创建时间，标准格式：2006-01-02T15:04:05Z                 |
| updated_at   | string       | 更新时间，标准格式：2006-01-02T15:04:05Z                 |
| extension    | object       | 混合云差异字段                                        |

##### extension[tcloud]

| 参数名称                  | 参数类型   | 描述      |
|-----------------------|--------|---------|
| cloud_main_account_id | string | 云主账户ID  |
| uin                   | string | 云子账户Uin |
| nick_name             | string | 昵称      |
| create_time           | string | 创建时间    |

##### extension[aws]

| 参数名称             | 参数类型   | 描述    |
|------------------|--------|-------|
| cloud_account_id | string | 云账户ID |
| arn              | string | Arn   |
| joined_method    | string | 添加方式  |
| status           | string | 状态    |

##### extension[huawei]

| 参数名称             | 参数类型   | 描述     |
|------------------|--------|--------|
| cloud_account_id | string | 云账户ID  |
| last_project_id  | string | 云子账户ID |
| enabled          | string | 云子账户名称 |

##### extension[gcp]

| 参数名称               | 参数类型   | 描述    |
|--------------------|--------|-------|
| cloud_project_id   | string | 云项目ID |
| cloud_project_name | string | 云项目名称 |

##### extension[azure]

| 参数名称                    | 参数类型   | 描述    |
|-------------------------|--------|-------|
| cloud_tenant_id         | string | 云租户ID |
| cloud_subscription_id   | string | 云订阅ID |
| cloud_subscription_name | string | 云订阅名称 |
| display_name_name       | string | 展示名称  |
| given_name              | string | 名     |
| sur_name                | string | 姓     |
