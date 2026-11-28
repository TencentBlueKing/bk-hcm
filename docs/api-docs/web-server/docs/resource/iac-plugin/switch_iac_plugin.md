### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-全局配置。
- 该接口功能描述：启用或停用指定的 IaC 插件。

### URL

POST /api/v1/cloud/iac_plugins/switch

### 输入参数

| 参数名称        | 参数类型   | 必选 | 描述                                  |
|-------------|--------|----|-------------------------------------|
| plugin_name | string | 是  | 插件名称（如 tencentcloud）               |
| state       | string | 是  | 目标状态（枚举值：enabled、disabled）         |

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
