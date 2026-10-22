### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-全局配置。
- 该接口功能描述：升级 http 类型的 IaC 插件，用于插件方发版后触发升级检查，或在插件处于 `failed` 时触发恢复。同步执行，完成后才响应。以该插件名下版本最高的记录 L 为准，按 L 的 `api_path` 探活取版本 V：
  - V 高于 L：新增 V 的记录（`api_path`、加密配置、备注沿用 L）并同步 Schema，成功后继承 L 的状态（L 为 `disabled` 则为 `disabled`，否则为 `enabled`），失败为 `failed`；L 保持原样；
  - V 等于 L 且 L 为 `init` / `failed`：L 重新同步，成功后按上一版本继承状态；
  - V 等于 L 且 L 为 `enabled` / `disabled`：无需处理，直接返回当前状态；
  - V 低于 L：返回"暂不支持回滚"。
- tf_provider 插件不支持本接口，需通过新增版本目录并滚动重启 hc-service 升级。

### URL

POST /api/v1/cloud/iac_plugins/upgrade

### 输入参数

| 参数名称        | 参数类型   | 必选 | 描述                                              |
|-------------|--------|----|-------------------------------------------------|
| plugin_name | string | 是  | 插件名称。按该名称找到版本最高的记录再探活，不使用记录 ID               |

### 调用示例

```json
{
  "plugin_name": "bk-hcm-plugin-cmdb"
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
| id             | string | 升级后版本最高记录的 ID（新增版本时为新记录 ID）             |
| plugin_name    | string | 插件名称                                     |
| plugin_version | string | 升级后版本最高记录的版本                             |
| state          | string | 升级后的状态（枚举值：enabled、disabled）             |

### 错误码与提示说明

| 错误码     | 场景与提示说明 |
|---------|--------------|
| 2000001 | 参数错误：<br>1. `plugin_name` 为空或不合法<br>2. tf_provider 插件不支持本接口：`iac plugin <name> is a tf_provider plugin, upgrade it by adding a new version directory and rolling restart hc-service` |
| 2000003 | 记录不存在：`plugin_name` 下没有插件记录 |
| 2000006 | 业务处理异常：<br>1. 插件正在注册、升级或变更中（拿不到插件名锁）：`iac plugin <name> is being registered, upgraded or modified, please retry later`<br>2. 探活版本低于当前版本：`iac plugin <name> probed version <v> is lower than the current version <v>, rollback is not supported yet`<br>3. 探活失败（超时、非 200、重定向、插件返回错误、缺少 `plugin_version`），信息带插件返回的 `message`<br>4. Schema 同步失败（新版本或重入的记录置为 `failed`，旧版本不变）：`sync iac plugin <name>@<version> schema failed, err: <原因>` |
| 2000007 | 鉴权服务异常：无法判断调用者是否有权限，`authorize failed` |
| 2030403 | 无权限：调用者缺少平台全局配置权限，`no permission` |
