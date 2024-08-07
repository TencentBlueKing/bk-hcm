### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：。
- 该接口功能描述：发送邮箱验证码。

### URL

POST /api/v1/cloud/mail/send_code

### 输入参数

| 参数名称   | 参数类型    | 必选 | 描述        |
|--------|---------|----|-----------|
| mail	  | string	 | 是	 | 邮箱号       |
| scene	 | string	  | 是	 | 验证的场景     |
| info	  | json	 | 是	 | 邮件内容的部分信息 |

### 调用示例

```json
{
  "mail": "xxx@xxx.com",
  "scene": "SecondAccountApplication",
  "info": {
    "vendor": "tcloud",
    "account_name": "测试账号111"
  }
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
| data | null   | 响应数据 |