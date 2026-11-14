### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-全局配置。
- 该接口功能描述：修改 http 类型 IaC 插件的地址、加密配置或备注，作用于该插件名下版本最高的记录。同步执行，完成后才响应。至少传一个字段，未传的保持不变；插件名称不可修改。仅 `enabled`、`disabled` 的插件可修改。
  - 只改加密配置 / 备注：校验后直接更新，不改状态与版本，不同步 Schema；
  - 改 `api_path`：校验并探活新地址。探活版本高于当前版本时按升级处理（新版本记录使用新地址）；版本相同时用新地址重新同步 Schema，成功后再更新，失败则记录保持原样；版本更低返回"暂不支持回滚"。
- tf_provider 插件不支持本接口，其配置由插件目录和全局配置决定。

### URL

POST /api/v1/cloud/iac_plugins/update

### 输入参数

| 参数名称          | 参数类型   | 必选 | 描述                                                         |
|---------------|--------|----|------------------------------------------------------------|
| plugin_name   | string | 是  | 插件名称。按该名称找到版本最高的记录再修改，不使用记录 ID，名称本身不可修改              |
| api_path      | string | 否  | 插件在蓝鲸 API 网关上的路径后缀，规则同注册接口。除 `plugin_name` 外至少再传一个字段     |
| crypto_type   | string | 否  | 加密类型（枚举值：RSA）                                              |
| crypto_config | object | 否  | 加密配置，结构同注册接口；与 `crypto_type` 合并后必须合法                     |
| memo          | string | 否  | 备注，最长 255 个字符                                              |

### 调用示例

```json
{
  "plugin_name": "bk-hcm-plugin-cmdb",
  "api_path": "/api/bk-hcm-plugin-cmdb/prod-v2",
  "memo": "切换到新网关环境"
}
```

### 响应示例

#### 成功响应

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000002",
    "plugin_name": "bk-hcm-plugin-cmdb",
    "plugin_version": "1.3.0",
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
| id             | string | 修改后版本最高记录的 ID（改地址触发升级时为新记录 ID）          |
| plugin_name    | string | 插件名称                                     |
| plugin_version | string | 修改后版本最高记录的版本                             |
| state          | string | 修改后的状态（枚举值：enabled、disabled）             |

### 错误码与提示说明

| 错误码     | 场景与提示说明 |
|---------|--------------|
| 2000001 | 参数错误：<br>1. `plugin_name` 为空或不合法<br>2. 除 `plugin_name` 外一个字段都没传：`at least one of api_path, crypto_type, crypto_config, memo is required`<br>3. `api_path` / `memo` 不合法，或合并后的加密配置不合法<br>4. tf_provider 插件：`iac plugin <name> is a tf_provider plugin, its config is decided by the plugin directory and the global config` |
| 2000003 | 记录不存在：`plugin_name` 下没有插件记录 |
| 2000006 | 业务处理异常：<br>1. 插件正在注册、升级或变更中（拿不到插件名锁）：`iac plugin <name> is being registered, upgraded or modified, please retry later`<br>2. 当前版本不是 `enabled` / `disabled`：`iac plugin <name>@<version> is in <state> state, only enabled or disabled plugin can be updated`<br>3. 新地址探活版本低于当前版本：`iac plugin <name> probed version <v> is lower than the current version <v>, rollback is not supported yet`<br>4. 新地址探活失败，信息带插件返回的 `message`<br>5. 新地址同步 Schema 失败（记录保持原样）：`sync iac plugin <name>@<version> schema failed, err: <原因>` |
| 2000007 | 鉴权服务异常：无法判断调用者是否有权限，`authorize failed` |
| 2030403 | 无权限：调用者缺少平台全局配置权限，`no permission` |
