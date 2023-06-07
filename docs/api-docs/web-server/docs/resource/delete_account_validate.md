### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：账号删除。
- 该接口功能描述：校验账号是否可以删除。

### URL

POST /api/v1/cloud/accounts/{account_id}/delete/validate

### 输入参数

| 参数名称       | 参数类型   | 必选  | 描述   |
|------------|--------|-----|------|
| account_id | string | 是   | 账号ID |

### 调用示例

```json

```

### 响应示例

```json
{
    "code": 2000000,
    "message": "account: 00000001 has some cloud resource, that can not delete",
    "data": {
        "azure_resource_group": 0,
        "cvm": 0,
        "disk": 0,
        "eip": 0,
        "gcp_firewall_rule": 0,
        "network_interface": 0,
        "route_table": 0,
        "security_group": 0,
        "subnet": 1,
        "vpc": 19
    }
}
```

```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
