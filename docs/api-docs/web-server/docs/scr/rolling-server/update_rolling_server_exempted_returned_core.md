### 描述

- 该接口提供版本：v9.9.9。
- 该接口所需权限：平台管理-滚服管理。
- 该接口功能描述：单据配置-批量更新减免退还核心数。

### URL

PATCH /api/v1/woa/rolling_servers/applied_records/exempted_returned_core

### 输入参数

| 参数名称                   | 参数类型         | 必选 | 描述                 |
|------------------------|--------------|----|--------------------|
| ids                    | string array | 是  | 滚服申请记录id列表，长度限制100 |
| exempted_returned_core | int          | 是  | 减免退还核心数,大于等于0      |

### 调用示例

```json
{
  "ids": ["1", "2"],
  "exempted_returned_core": 500
}
```

### 响应示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                         |
|---------|--------|----------------------------|
| result  | bool   | 请求成功与否。true:请求成功；false请求失败 |
| code    | int    | 错误编码。 0表示success，>0表示失败错误  |
| message | string | 请求失败返回的错误信息                |
| data	   | object | 响应数据                       |