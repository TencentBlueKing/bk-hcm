### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：COS桶创建。
- 该接口功能描述：创建存储桶代理接口。

### URL

POST /api/v1/cloud/cos/buckets/create

### 输入参数

#### tcloud
代理接口文档地址：https://cloud.tencent.com/document/product/436/7738

| 参数名称                                   | 参数类型   | 必选 | 描述                                                       |
|----------------------------------------|--------|----|----------------------------------------------------------|
| account_id                             | string | 是  | 账号ID                                                     |
| region                                 | string | 是  | 地域                                                       |
| name                                   | string | 是  | 存储桶名称                                                    |
| x_cos_acl                              | string | 否  | 定义存储桶的访问控制列表（ACL）属性                                      |
| x_cos_grant_read                       | string | 否  | 赋予被授权者读取存储桶的权限                                           |
| x_cos_grant_write                      | string | 否  | 赋予被授权者写入存储桶的权限                                           |
| x_cos_grant_full_control               | string | 否  | 赋予被授权者操作存储桶的所有权限                                         |
| x_cos_grant_read_acp                   | string | 否  | 赋予被授权者读取存储桶的访问控制列表（ACL）和存储桶策略（Policy）的权限                 |
| x_cos_grant_write_acp                  | string | 否  | 赋予被授权者写入存储桶的访问控制列表（ACL）和存储桶策略（Policy）的权限                 |
| x_cos_tagging                          | string | 否  | 在创建存储桶的同时，为存储桶添加标签，最多可设置50个标签。例如 key1=value1&key2=value2 |
| create_bucket_configuration            | object | 否  | 包含操作的所有请求信息                                              |

#### create_bucket_configuration
| 参数名称                                  | 参数类型   | 必选 | 描述                                |
|---------------------------------------|--------|----|-----------------------------------|
| bucket_az_config                      | string | 是  | 存储桶 AZ 配置，指定为 MAZ 以创建多 AZ 存储桶     |

### 调用示例

#### tcloud

```json
{
  "account_id": "0000001",
  "region": "ap-hk",
  "name": "xxx",
  "create_bucket_configuration":{
    "bucket_az_config": "MAZ"
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
| code    | int32  | 状态码  |
| message | string | 请求信息 |

