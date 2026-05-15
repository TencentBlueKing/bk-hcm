### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：根据资源类型查询业务下关联资源的二级账号元数据信息，仅返回对应资源有权限的二级账号信息

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/accounts/list/by/res_type

### 输入参数

| 参数名称        | 参数类型         | 必选 | 描述                                                              |
|-------------|--------------|----|-----------------------------------------------------------------|
| bk_biz_id   | string       | 是  | 业务ID                                                            |
| vendor      | string       | 是  | 云厂商                                                             |
| ids         | string array | 是  | 二级账号ID列表，最大支持为100个                                              |
| res_type    | string       | 是  | 资源类型(目前支持的类型sub_account、sub_account_secret、permission_template、permission_policy_library) |

### 调用示例

```json
{
  "ids": ["000000001"],
  "res_type": "sub_account"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "id": "0000001",
        "name": "二级账号名字",
        "bk_biz_id": 123,
		"vendor": "tcloud",
        "usage_biz_ids": [
          123,
          1234
        ],
        "managers": [
          "person1",
          "person2"
        ],
        "extension": {
          "cloud_main_account_id": "123456"
        }
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称              | 参数类型         | 描述       |
|-------------------|--------------|----------|
| code              | int32        | 状态码      |
| message           | string       | 请求信息     |
| data              | object       | 账号信息     |

#### data
| 参数名称    | 参数类型         | 描述       |
|---------|--------------|----------|
| details | object array | 二级账号信息列表 |

#### details[n]
| 参数名称          | 参数类型         | 描述         |
|---------------|--------------|------------|
| id            | string       | 账号ID       |
| name          | string       | 二级账号名称     |
| vendor        | string       | 云厂商        |
| bk_biz_id     | int64        | 二级账号管理业务ID |
| usage_biz_ids | int64  array | 使用业务ID列表   |
| managers      | string array | 负责人列表      |
| extension     | string       | 云厂商扩展字段    | 

###### extension[tcloud]

| 参数名称                  | 参数类型    | 描述       |
|-----------------------|---------|----------|
| cloud_main_account_id | string  | 二级账号云上ID |





