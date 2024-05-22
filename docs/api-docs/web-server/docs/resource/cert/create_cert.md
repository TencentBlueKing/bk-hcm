### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：证书创建。
- 该接口功能描述：资源下上传证书。

### URL

POST /api/v1/cloud/certs/create

### 输入参数

输入参数由接口通用参数和vendor对应的云厂商差异参数组成。

#### 接口通用参数

| 参数名称      | 参数类型 | 必选 | 描述                                          |
|--------------|--------|------|----------------------------------------------|
| vendor       | string | 是   | 云厂商（枚举值：tcloud、aws、gcp、azure、huawei） |
| account_id   | string | 是   | 账号ID                                        |
| name         | string | 是   | 证书名称                                       |
| memo         | string | 否   | 备注                                          |

#### 云厂商差异参数[tcloud]

| 参数名称      | 参数类型 | 必选 | 描述                                          |
|--------------|--------|------|----------------------------------------------|
| cert_type    | string | 是   | 证书类型（CA:客户端证书，SVR:服务器证书）          |
| public_key   | string | 是   | 证书内容，需要做base64编码                       |
| private_key  | string | 否   | 私钥内容，需要做base64编码，CA证书可不传该参数      |

### 腾讯云调用示例

```json
{
  "vendor": "tcloud",
  "account_id": "00000001",
  "name": "test-cert",
  "memo": "test cert",
  "cert_type": "CA",
  "public_key": "xxxxxx",
  "private_key": "xxxxxx"
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

| 参数名称 | 参数类型 | 描述    |
|---------|--------|---------|
| code    | int    | 状态码   |
| message | string | 请求信息 |
| data    | object | 调用数据 |

#### data

| 参数名称 | 参数类型 | 描述        |
|---------|--------|------------|
| id      | string | 证书ID      |
