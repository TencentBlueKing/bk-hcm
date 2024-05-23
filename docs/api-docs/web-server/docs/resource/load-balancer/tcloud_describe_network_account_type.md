### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询用户网络类型。腾讯云代理接口 DescribeNetworkAccountType

### URL

GET /api/v1/cloud/vendors/tcloud/accounts/{account_id}/network_type

### 输入参数

| 参数名称       | 参数类型   | 必选 | 描述    |
|------------|--------|----|-------|
| account_id | string | 是  | 云账户id |

### 响应参数说明

| 参数名称    | 参数类型            | 描述   |
|---------|-----------------|------|
| code    | int32           | 状态码  |
| message | string          | 请求信息 |
| data    | AccountTypeInfo |      |

### AccountTypeResp

| 参数名称               | 参数类型   | 描述                                  |
|--------------------|--------|-------------------------------------|
| NetworkAccountType | string | 用户账号的网络类型，STANDARD为标准用户，LEGACY为传统用户 |