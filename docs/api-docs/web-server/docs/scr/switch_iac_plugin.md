### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-全局配置。
- 该接口功能描述：启用或停用 IaC 插件。接口以该插件名下版本最高的记录（当前版本）为准判定状态规则，并对该插件名下所有处于 `enabled` / `failed`（目标 `disabled`）或 `disabled`（目标 `enabled`）的版本批量生效（`init` 版本保持不变）。接口只修改注册表状态并立即返回。

### URL

POST /api/v1/cloud/iac_plugins/switch

### 输入参数

| 参数名称        | 参数类型   | 必选 | 描述                                    |
|-------------|--------|----|-----------------------------------------|
| plugin_name | string | 是  | 插件名称（如 tencentcloud）             |
| state       | string | 是  | 目标状态（枚举值：enabled、disabled）   |

### 调用示例

```json
{
  "plugin_name": "tencentcloud",
  "state": "disabled"
}
```

### 响应示例

#### 成功响应

```json
{
  "code": 0,
  "message": "",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述       |
|---------|--------|----------|
| code    | int32  | 状态码      |
| message | string | 请求信息     |
| data    | object | 响应数据（为空） |

### 错误码与提示说明

| 错误码    | 场景与提示说明 |
|---------|--------------|
| 2000001 | 参数错误：`plugin_name` 为空，或 `state` 不是 `enabled` / `disabled` |
| 2000003 | 记录不存在：`plugin_name` 对应的插件记录不存在 |
| 2000006 | 业务冲突或状态不允许：<br>1. 插件正在注册、升级或变更中（拿不到插件名锁）：`iac plugin <name> is being registered, upgraded or modified, please retry later`<br>2. 当前版本处于 `init`（注册或升级未完成）：`iac plugin <name>@<version> registration or upgrade is not finished`<br>3. 当前版本处于 `failed` 时请求启用（需先恢复）：`iac plugin <name>@<version> is in failed state (<message>), please recover first`<br>4. 当前版本处于 `disabled` 且 Schema 不存在时请求启用（需先恢复）：`iac plugin <name>@<version> schema does not exist, please recover first` |
| 2000007 | 鉴权服务异常：无法判断调用者是否有权限，`authorize failed` |
| 2030403 | 无权限：调用者缺少平台全局配置权限，`no permission` |
