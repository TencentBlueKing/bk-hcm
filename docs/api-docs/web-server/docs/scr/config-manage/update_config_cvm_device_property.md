### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：平台-CVM机型。
- 该接口功能描述：CVM机型配置信息更新。

### URL

PUT /api/v1/woa/config/updatemany/config/cvm/device/property

### 输入参数

| 参数名称     | 参数类型                | 必选 | 描述          |
|-------------|-----------------------|------|--------------|
| device_types| object array         | 是   | 更新的机型配置列表，最多100个 |

#### device_types 数组元素说明

| 参数名称          | 参数类型                | 必选 | 描述          |
|------------------|-----------------------|------|--------------|
| id                | string                | 是   | 机型唯一ID，最大长度64 |
| device_type       | string                | 否   | 机型，最大长度64 |
| device_class      | string                | 否   | 机型分类，最大长度64 |
| device_family     | string                | 否   | 机型族，最大长度64 |
| core_type         | string                | 否   | 核心类型，枚举值：小核心、中核心、大核心 |
| cpu_core          | int64                 | 否   | CPU核心数，单位：核，>=0 |
| memory            | int64                 | 否   | 内存大小，单位：GB，>=0 |
| device_type_class | string                | 否   | 通/专用机型，枚举值：SpecialType（专用）、CommonType（通用） |
| technical_class   | string                | 否   | 技术分类，最大长度64 |
| region            | string                | 否   | 地域，最大长度64 |
| zone              | string                | 否   | 可用区，最大长度64 |
| disable           | bool                  | 否   | 是否不使用 |
| source            | string                | 否   | 机型来源 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "device_types": [
    {
      "id": "00000001",
      "disable": true
    },
    {
      "id": "00000002",
      "cpu_core": 8,
      "memory": 32
    }
  ]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| result     | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| data	     | object       | 请求返回的数据        |
