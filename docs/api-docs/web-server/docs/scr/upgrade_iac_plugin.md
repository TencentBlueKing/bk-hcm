### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-全局配置。
- 该接口功能描述：触发 http 类型 IaC 插件的升级检查或失败恢复。接口同步执行，向插件探活获取最新版本，发现新版本时自动注册并同步 Schema，当前版本注册失败时可重新同步恢复；暂不支持版本回滚。`tf_provider` 插件不支持该接口。

### URL

POST /api/v1/cloud/iac_plugins/upgrade

### 输入参数

| 参数名称        | 参数类型   | 必选 | 描述   |
|-------------|--------|----|------|
| plugin_name | string | 是  | 插件名称 |

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

| 参数名称           | 参数类型   | 描述                             |
|----------------|--------|--------------------------------|
| id             | string | 升级后最新版本的记录 ID                  |
| plugin_name    | string | 插件名称                           |
| plugin_version | string | 升级后的插件版本                       |
| state          | string | 升级后的状态（枚举值：enabled、disabled）   |
