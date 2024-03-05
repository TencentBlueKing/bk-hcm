### 描述

- 该接口提供版本：v1.4.0+。
- 该接口所需权限：IaaS资源创建。
- 该接口功能描述：业务下创建参数模版。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/argument_templates/create

### 输入参数

| 参数名称            | 参数类型               | 必选 | 描述                                                                      |
|-----------------|--------------------|----|-------------------------------------------------------------------------|
| bk_biz_id       | int64              | 是  | 业务ID                                                                    |
| vendor          | string             | 是  | 云厂商（枚举值：tcloud，当前版本暂只支持tcloud）                                          |
| account_id      | string             | 是  | 账号ID                                                                    |
| name            | string             | 是  | 参数模版名称                                                                  |
| type            | string             | 是  | 模版类型（address:IP地址、address_group:IP地址组、service:协议端口、service_group:协议端口组） |
| templates       | address_info array | 否  | "IP地址"、"协议端口"参数模版的参数数组（互斥，templates 和 group_templates必须传其中一个）           |
| group_templates | string array       | 否  | "IP地址组"、"协议端口组"参数模版的参数数组（互斥，templates 和 group_templates必须传其中一个）         |

#### address_info

| 参数名称        | 参数类型   | 描述 | 描述                                            |
|-------------|--------|----|-----------------------------------------------|
| address     | string | 是  | 地址信息, 支持 IP、CIDR、IP 范围、IP地址模板ID、协议端口、协议端口模板ID |
| description | string | 否  | 备注                                            |

### 腾讯云调用示例

```json
{
  "vendor": "tcloud",
  "account_id": "00000001",
  "name": "test-template",
  "type": "address",
  "templates": [
    {
      "address": "127.0.0.1",
      "description": "test1"
    },
    {
      "address": "127.0.0.2",
      "description": "test2"
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "00000001"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 调用数据 |

#### data

| 参数名称 | 参数类型   | 描述     |
|------|--------|--------|
| id   | string | 参数模版ID |
