### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：平台-CVM机型。
- 该接口功能描述：CVM机型配置信息创建。

### URL

POST /api/v1/woa/config/createmany/config/cvm/device

### 输入参数

| 参数名称       | 参数类型       | 必选 | 描述          |
|---------------|--------------|------|--------------|
| device_types	| object array | 是	  | 机型配置列表，最多100个 |

#### device_types 数组元素说明

| 参数名称          | 参数类型                | 必选 | 描述          |
|------------------|-----------------------|------|--------------|
| device_type       | string                | 是   | 机型，最大长度64 |
| device_class      | string                | 是   | 机型分类，最大长度64 |
| device_family     | string                | 是   | 机型族，最大长度64 |
| core_type         | string                | 是   | 核心类型，枚举值：小核心、中核心、大核心 |
| cpu_core          | int64                 | 是   | CPU核心数，单位：核，>=0 |
| memory            | int64                 | 是   | 内存大小，单位：GB，>=0 |
| device_type_class | string                | 是   | 通/专用机型，枚举值：SpecialType（专用）、CommonType（通用） |
| technical_class   | string                | 是   | 技术分类，最大长度64 |
| region            | string                | 是   | 地域，最大长度64 |
| zone              | string                | 是   | 可用区，最大长度64 |
| disable           | bool                  | 否   | 是否不使用 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "device_types": [
    {
      "device_type": "S2.LARGE16",
      "device_class": "标准型",
      "device_family": "标准型",
      "core_type": "大核心",
      "cpu_core": 4,
      "memory": 16,
      "device_type_class": "CommonType",
      "technical_class": "标准型",
      "region": "ap-shanghai",
      "zone": "ap-shanghai-2"
    }
  ]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "ids": ["00000001", "00000002"]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| result     | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| data	     | object       | 请求返回的数据        |

#### data 字段说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| ids        | string array | 创建的机型ID列表    |
