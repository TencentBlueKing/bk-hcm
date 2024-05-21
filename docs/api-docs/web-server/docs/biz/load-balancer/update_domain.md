### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：域名更新。
- 该接口功能描述：业务下更新域名。

### URL

PATCH /api/v1/cloud/bizs/{bk_biz_id}/listeners/{lbl_id}/domains

### 输入参数

| 参数名称           | 参数类型   | 必选 | 描述                                                       |
|----------------|--------|----|----------------------------------------------------------|
| bk_biz_id      | int    | 是  | 业务ID                                                     |
| lbl_id         | string | 是  | 监听器ID                                                    |
| domain         | string | 是  | 域名                                                       |
| new_domain     | string | 否  | 新域名, 需要修改域名时填此参数                                         |
| default_server | bool   | 否  | 是否设为默认域名，一个监听器下只能设置一个默认域名。如果不想修改请不要传改参数，传false可以取消当前默认域名 |
| certificate    | object | 否  | 证书信息                                                     |

### certificate

| 参数名称           | 参数类型         | 描述                                   |
|----------------|--------------|--------------------------------------|
| ssl_mode       | string       | 认证类型，UNIDIRECTIONAL：单向认证，MUTUAL：双向认证 |
| ca_cloud_id    | string       | CA证书的云ID                             |
| cert_cloud_ids | string array | 服务端证书的云ID                            |

### 调用示例

```json
{
  "domain": "www.old.com",
  "new_domain": "www.new.com",
  "certificate": {
    "ssl_mode": "MUTUAL",
    "ca_cloud_id": "ca-001",
    "cert_cloud_ids": [
      "cert-001"
    ]
  }
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
