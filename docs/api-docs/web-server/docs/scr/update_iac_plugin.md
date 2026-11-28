### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-全局配置。
- 该接口功能描述：修改 http 类型 IaC 插件的地址、加密配置或备注，同步执行。仅 `enabled` 和 `disabled` 状态的插件可修改；至少传入一个待修改字段，未传字段保持不变；插件名称不可修改。若修改了 `api_path`，服务端会探活新地址并同步 Schema（发现更高版本时按升级处理，暂不支持回滚）。`tf_provider` 插件不支持该接口。

### URL

POST /api/v1/cloud/iac_plugins/update

### 输入参数

| 参数名称          | 参数类型   | 必选 | 描述                                                         |
|---------------|--------|----|------------------------------------------------------------|
| plugin_name   | string | 是  | 插件名称，不可修改                                                  |
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

| 参数名称           | 参数类型   | 描述                             |
|----------------|--------|--------------------------------|
| id             | string | 修改后最新版本的记录 ID                  |
| plugin_name    | string | 插件名称                           |
| plugin_version | string | 修改后的插件版本                       |
| state          | string | 修改后的状态（枚举值：enabled、disabled）   |
