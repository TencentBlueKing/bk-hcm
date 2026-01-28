### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：创建用于创建三级账号的申请。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/add_sub_account

### 输入参数

| 参数名称         | 参数类型   | 必选 | 描述     |
|--------------|--------|----|--------|
| bk_biz_id    | int64  | 是  | 业务ID   |
| vendor    | string       | 是  | 云厂商, 枚举值：tcloud |
| sub_accounts | object | 是  | 三级账号列表，长度限制100 |

#### sub_accounts
| 参数名称        | 参数类型   | 必选 | 描述     |
|-------------|--------|----|--------|
| account_id    | string      | 是  | 资源账号ID   |
| name          | string      | 是  | 三级账号名称   |
| receive_email | string      | 是  | 账号开通接收邮箱 |
| email         | string      | 否  | 三级账号邮箱   |
| phone_num     | string      | 否  | 手机号      |
| country_code  | string    | 否  | 手机区域代码   |
| managers      | string array | 否  | 账号管理者    |
| memo          | string      | 否  | 备注       |

### 调用示例

#### TCloud

```json
{
  "sub_accounts": [
    {
      "account_id": "00000001",
      "name": "sub-account-01",
      "receive_email": "sub@example.com",
      "email": "sub@example.com",
      "phone_num": "13800000000",
      "country_code": "86",
      "managers": [
        "hcm"
      ],
      "memo": "create sub account"
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "ids": ["00000001"]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述   |
|------|------|------|
| code | int32 | 状态码  |
| message | string | 请求信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型          | 描述    |
|------|---------------|-------|
| ids  | string array  | 单据ID数组 |
