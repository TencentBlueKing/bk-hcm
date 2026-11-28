### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-全局配置。
- 该接口功能描述：注册 http 类型的 IaC 插件。接口同步执行，完成后返回注册结果；插件版本号由服务端探活插件健康检查接口获取，无需请求传入。

### URL

POST /api/v1/cloud/iac_plugins/register

### 输入参数

| 参数名称          | 参数类型   | 必选 | 描述                                                                                         |
|---------------|--------|----|--------------------------------------------------------------------------------------------|
| plugin_name   | string | 是  | 插件名称，字母或数字开头，只含字母、数字、`_`、`-`、`.`，最长 64 个字符                                            |
| api_path      | string | 是  | 插件在蓝鲸 API 网关上的路径后缀，以 `/` 开头、不以 `/` 结尾，不能含 scheme、host、`..` 段、查询参数，最长 255 个字符（如 `/api/bk-hcm-plugin-cmdb/prod`） |
| crypto_type   | string | 是  | 加密类型（枚举值：RSA）                                                                              |
| crypto_config | object | 是  | 加密配置                                                                                       |
| memo          | string | 否  | 备注，最长 255 个字符                                                                              |

#### crypto_config

| 参数名称       | 参数类型   | 必选 | 描述                                          |
|------------|--------|----|---------------------------------------------|
| public_key | string | 是  | PEM 编码的 SPKI 格式（`PUBLIC KEY`）RSA 公钥，长度不低于 3072 位 |

### 调用示例

```json
{
  "plugin_name": "bk-hcm-plugin-cmdb",
  "api_path": "/api/bk-hcm-plugin-cmdb/prod",
  "crypto_type": "RSA",
  "crypto_config": {
    "public_key": "-----BEGIN PUBLIC KEY-----\nMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEA...\n-----END PUBLIC KEY-----\n"
  },
  "memo": "cmdb 插件"
}
```

### 响应示例

#### 成功响应

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000001",
    "plugin_name": "bk-hcm-plugin-cmdb",
    "plugin_version": "1.2.0",
    "state": "enabled"
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

| 参数名称           | 参数类型   | 描述                           |
|----------------|--------|------------------------------|
| id             | string | 本次注册的记录 ID                   |
| plugin_name    | string | 插件名称                         |
| plugin_version | string | 插件版本（探活返回）                   |
| state          | string | 注册后的状态（枚举值：enabled、disabled） |
