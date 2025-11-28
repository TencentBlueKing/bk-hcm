### 描述

- 该接口提供版本：v1.8.7.10+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询设备负载利用率信息。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/device/load_usage

### 输入参数

| 参数名称 | 参数类型   | 必选 | 描述                   |
|------|--------|----|----------------------|
| date | string | 是  | 查询日期，格式如"2024-01-01" |

### 调用示例

#### 获取详细信息请求参数示例


### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "threshold": 30,
    "cpu_usage": 31.01,
    "achieved_kpi": true,
    "empty_load_cpu_core": 123.45,
    "empty_load_os": 15.0,
    "low_load_cpu_core": 256.78,
    "low_load_os": 28.0
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                         |
|---------|--------|----------------------------|
| code    | int    | 错误编码。 0表示success，>0表示失败错误  |
| message | string | 请求失败返回的错误信息                |
| data    | object | 响应数据                       |

#### data

| 参数名称                | 参数类型  | 描述        |
|---------------------|-------|-----------|
| threshold           | int   | 利用率达标阈值   |
| cpu_usage           | float | 当天的CPU利用率 |
| achieved_kpi        | bool  | 是否达标      |
| empty_load_cpu_core | float | 空负载核心数    |
| empty_load_os       | float | 空负载OS数    |
| low_load_cpu_core   | float | 低负载核心数    |
| low_load_os         | float | 低负载OS数    |

