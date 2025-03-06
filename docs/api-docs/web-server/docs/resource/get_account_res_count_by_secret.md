### 描述

- 该接口提供版本：v1.2.1+
- 该接口所需权限：账号录入
- 该接口功能描述：通过秘钥获取资源数量

### URL

POST /api/v1/cloud/vendors/{vendor}/accounts/res_counts/by_secrets

#### 路径参数说明

| 参数名称   | 参数类型   | 必选 | 描述                                   |
|--------|--------|----|--------------------------------------|
| vendor | string | 是  | 云厂商（枚举值：tcloud、aws、huawei、gcp、azure） |

### 输入参数

#### TCloud

| 参数名称             | 参数类型   | 必选 | 描述    |
|------------------|--------|----|-------|
| cloud_secret_id  | string | 是  | 云加密ID |
| cloud_secret_key | string | 是  | 云密钥   |

#### AWS

| 参数名称             | 参数类型   | 必选 | 描述                                  |
|------------------|--------|----|-------------------------------------|
| cloud_secret_id  | string | 是  | 云加密ID                               |
| cloud_secret_key | string | 是  | 云密钥                                 |
| site             | string | 是  | 站点（枚举值：china:中国站、international:国际站） |

#### Azure

| 参数名称                    | 参数类型   | 必选 | 描述     |
|-------------------------|--------|----|--------|
| cloud_tenant_id         | string | 是  | 云租户ID  |
| cloud_subscription_id   | string | 是  | 云订阅ID  |
| cloud_application_id    | string | 是  | 云应用ID  |
| cloud_client_secret_key | string | 是  | 云客户端密钥 |

#### GCP

| 参数名称                     | 参数类型   | 必选 | 描述    |
|--------------------------|--------|----|-------|
| cloud_project_id         | string | 是  | 云项目ID |
| cloud_service_secret_key | string | 是  | 云服务密钥 |

#### Huawei

| 参数名称             | 参数类型   | 必选 | 描述    |
|------------------|--------|----|-------|
| cloud_secret_id  | string | 是  | 云加密ID |
| cloud_secret_key | string | 是  | 云密钥   |

### 调用示例

#### TCloud

```json
{
  "cloud_secret_id": "xxxx",
  "cloud_secret_key": "xxxx"
}
```

#### Aws

```json
{
  "cloud_secret_id": "xxxx",
  "cloud_secret_key": "xxxx"
}
```

#### Azure

```json
{
  "cloud_tenant_id": "0000000",
  "cloud_subscription_id": "xxxxx",
  "cloud_application_id": "xxxxxx",
  "cloud_client_secret_key": "xxxxxx"
}
```

#### Gcp

```json
{
  "cloud_project_id": "",
  "cloud_service_secret_key": "{xxxx:xxx}"
}
```

#### HuaWei

```json
{
  "cloud_secret_id": "xxxx",
  "cloud_secret_key": "xxxx"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "items": [
      {
        "type": "cvm",
        "count": 6
      },
      {
        "type": "disk",
        "count": 17
      },
      {
        "type": "vpc",
        "count": 18
      },
      {
        "type": "eip",
        "count": 6
      },
      {
        "type": "security_group",
        "count": 29
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称  | 参数类型   | 描述     |
|-------|--------|--------|
| items | object | 资源详情列表 |

##### items[0]

| 参数名称  | 参数类型   | 描述   |
|-------|--------|------|
| type  | string | 资源类型 |
| count | int    | 资源数量 |
