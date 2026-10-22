### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-全局配置。
- 该接口功能描述：注册 http 类型的 IaC 插件。同步执行，注册完成后才响应。插件版本取自插件探活接口返回的 `plugin_version`，不由请求传入。流程：持插件名锁 → 检查插件名未被 tf_provider 插件占用 → 地址与加密配置校验 → 经蓝鲸 API 网关请求 `{网关地址前缀}{api_path}/health` 探活 → 写入 `init` 记录 → 同步 Schema → 成功为 `enabled`（上一版本为 `disabled` 时为 `disabled`），失败为 `failed`。地址、加密配置校验失败或探活失败时不写注册表。

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

| 参数名称           | 参数类型   | 描述                                       |
|----------------|--------|------------------------------------------|
| id             | string | 本次注册的记录 ID                               |
| plugin_name    | string | 插件名称                                     |
| plugin_version | string | 插件版本（探活返回）                               |
| state          | string | 注册后的状态（枚举值：enabled、disabled）             |

### 错误码与提示说明

| 错误码     | 场景与提示说明 |
|---------|--------------|
| 2000001 | 参数错误：<br>1. 必填字段缺失、`plugin_name` / `api_path` / `memo` 不合法，或 `api_path` 拼接后不在蓝鲸 API 网关内<br>2. `crypto_type` 不支持、`public_key` 无法解析、不是 RSA 或长度不足 3072 位<br>3. 插件名已被 tf_provider 插件占用：`plugin name <name> is already used by a tf_provider plugin` |
| 2000006 | 业务处理异常：<br>1. 插件正在注册、升级或变更中（拿不到插件名锁）：`iac plugin <name> is being registered, upgraded or modified, please retry later`<br>2. 已注册（版本最高记录为 `enabled` / `disabled`）：`iac plugin <name> is already registered, current version: <version>, state: <state>`<br>3. 探活版本低于已有记录：`iac plugin <name> probed version <v> is lower than the current version <v>, rollback is not supported yet`<br>4. 探活失败（超时、非 200、重定向、插件返回错误、缺少 `plugin_version`），信息带插件返回的 `message`<br>5. Schema 同步失败（记录置为 `failed`，`message` 写明原因）：`sync iac plugin <name>@<version> schema failed, err: <原因>`<br>6. 收尾时记录已被其他流程改动（条件更新冲突） |
| 2000007 | 鉴权服务异常：无法判断调用者是否有权限，`authorize failed` |
| 2030403 | 无权限：调用者缺少平台全局配置权限，`no permission` |
