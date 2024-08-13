### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：。
- 该接口功能描述：验证邮箱验证码。

### URL

POST /api/v1/cloud/mail/verify_code

### 输入参数

| 参数名称   | 参数类型    | 必选 | 描述           |
|--------|---------|----|--------------|
| mail	  | string	 | 是	 | 邮箱号          |
| scene	 | string	 | 是	 | 验证的场景        |
| verify_code	 | string	 | 是	 | 验证码          |
| delete_after_verify	 | bool	   | 否	 | 是否验证通过后删除验证码 |

### 调用示例

```json
{
  "mail": "xxx@xxx.com",
  "scene": "SecondAccountApplication",
  "verify_code": "469766",
  "delete_after_verify": false
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": true
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述     |
|---------|--------|--------|
| code    | int32  | 状态码    |
| message | string | 请求信息   |
| data | bool   | 验证是否通过 |