### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：IaaS资源操作。
- 该接口功能描述：业务下更新参数模版，只支持覆盖更新。

### URL

PUT /api/v1/cloud/bizs/{bk_biz_id}/argument_templates/{id}

### 输入参数

| 参数名称         | 参数类型              | 必选 | 描述                                          |
|-----------------|---------------------|------|----------------------------------------------|
| bk_biz_id       | int64               | 是   | 业务ID                                        |
| vendor          | string              | 是   | 云厂商（枚举值：tcloud，当前版本暂只支持tcloud）   |
| account_id      | string              | 是   | 账号ID                                        |
| name            | string              | 否   | 参数模版名称                                   |
| templates       | address_info array  | 否   | "IP地址"、"协议端口"参数模版的参数数组（互斥，templates 和 group_templates必须传其中一个）       |
| group_templates | string array        | 否   | "IP地址组"、"协议端口组"参数模版的参数数组（互斥，templates 和 group_templates必须传其中一个）    |

#### address_info

| 参数名称     | 参数类型 | 描述 | 描述                          |
|-------------|--------|------|------------------------------|
| address     | string | 是   | 地址信息, 支持 IP、CIDR、IP 范围 |
| description | string | 否   | 备注                          |

### 调用示例

更新腾讯云模版规则。

```json
{
  "vendor": "tcloud",
  "account_id": "00000001",
  "name": "test-template",
  "type": "address",
  "templates": [
    {"address":"127.0.0.1", "description":"test1"},
    {"address":"127.0.0.2", "description":"test2"}
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

| 参数名称  | 参数类型 | 描述    |
|---------|---------|---------|
| code    | int     | 状态码   |
| message | string  | 请求信息 |
