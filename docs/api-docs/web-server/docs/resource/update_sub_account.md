### 描述

- 该接口提供版本：v1.2.1+。
- 该接口所需权限：账号编辑。
- 该接口功能描述：更新子账号。

### URL

PATCH /api/v1/cloud/sub_accounts/{id}

### 输入参数

| 参数名称       | 参数类型        | 必选 | 描述    |
|------------|-------------|----|-------|
| id         | string      | 是  | 子账号ID |
| managers   | string      | 否  | 账号管理者 |
| bk_biz_ids | int64 array | 否  | 业务ID  |
| memo       | string      | 否  | 备注    |

### 调用示例

```json
{
  "managers": [
    "hcm"
  ],
  "memo": "account update",
  "bk_biz_ids": [
    310
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |
