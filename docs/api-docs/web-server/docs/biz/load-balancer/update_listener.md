### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：监听器更新。
- 该接口功能描述：业务下更新监听器。

### URL

PATCH /api/v1/cloud/bizs/{bk_biz_id}/listeners/{id}

### 输入参数

| 参数名称       | 参数类型   | 必选 | 描述   |
|------------|--------|----|------|
| bk_biz_id  | int    | 是  | 业务ID |
| account_id | string | 是  | 账号ID |
| name       | string | 是  | 名称   |
| extension  | object | 否  | 拓展信息 |

#### extension

| 参数名称        | 参数类型   | 必选 | 描述                       |
|-------------|--------|----|--------------------------|
| certificate | object | 否  | 证书信息, 非SNI类型HTTPS监听器可以修改 |

### certificate

| 参数名称           | 参数类型         | 描述                                   |
|----------------|--------------|--------------------------------------|
| ssl_mode       | string       | 认证类型，UNIDIRECTIONAL：单向认证，MUTUAL：双向认证 |
| ca_cloud_id    | string       | CA证书的云ID                             |
| cert_cloud_ids | string array | 服务端证书的云ID                            |

### 调用示例

```json
{
  "account_id": "0000001",
  "name": "xxx"
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

