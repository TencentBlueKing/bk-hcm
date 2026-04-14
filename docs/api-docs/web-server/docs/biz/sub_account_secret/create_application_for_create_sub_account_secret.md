### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：三级账号密钥操作。
- 该接口功能描述：创建用于新增三级账号密钥的申请。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/applications/types/add_sub_account_secret

### 输入参数

| 参数名称                | 参数类型     | 必选 | 描述       |
|---------------------|----------|----|----------|
| bk_biz_id           | int64    | 是  | 业务ID     |
| vendor              | string   | 是  | 云厂商, 枚举值：tcloud |
| sub_account_secrets | object   | 是  | 三级账号密钥列表，长度限制100 |

#### sub_account_secrets

| 参数名称         | 参数类型   | 必选 | 描述                                       |
|--------------|--------|----|------------------------------------------|
| sub_account_id | string | 是  | 三级账号ID                                   |


### 调用示例

#### TCloud

```json
{
  "sub_account_secrets": [
    {
      "sub_account_id": "00000001"
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

| 参数名称 | 参数类型          | 描述     |
|------|---------------|--------|
| ids  | string array  | 单据ID数组 |
