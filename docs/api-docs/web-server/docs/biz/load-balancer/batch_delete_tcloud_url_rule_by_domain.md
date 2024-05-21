### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下按域名删除腾讯云URL规则

### URL

DELETE /api/v1/cloud/vendors/tcloud/listeners/{lbl_id}/rules/by/domains/batch

### 输入参数

| 参数名称               | 参数类型         | 必选 | 描述                      |
|--------------------|--------------|----|-------------------------|
| bk_biz_id          | int          | 是  | 业务ID                    |
| lbl_id             | string       | 是  | 监听器id                   |
| domains            | string array | 是  | 按域名删除数组                 |
| new_default_domain | string       | 否  | 新默认域名,删除的域名是默认域名的时候需要指定 |

### 调用示例

```json
{
  "domains": [
    "qweqwe.com",
    "kkkk.com"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
