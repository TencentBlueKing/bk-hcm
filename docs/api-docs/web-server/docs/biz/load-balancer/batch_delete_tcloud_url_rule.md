### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下删除腾讯云URL规则, 支持指定规则id或域名

### URL

DELETE /api/v1/cloud/vendors/tcloud/listeners/{lbl_id}/rules/batch

### 输入参数

| 参数名称               | 参数类型         | 必选 | 描述                      |
|--------------------|--------------|----|-------------------------|
| bk_biz_id          | int          | 是  | 业务ID                    |
| lbl_id             | string       | 是  | 监听器id                   |
| rule_ids           | string array | 否  | URL规则ID数组               |
| domain             | string       | 否  | 按域名删除, 没有指定规则id的时候必填    |
| new_default_domain | string       | 否  | 新默认域名,删除的域名是默认域名的时候需要指定 |

### 调用示例

```json
{
  "rule_ids": [
    "00000001",
    "00000002"
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
