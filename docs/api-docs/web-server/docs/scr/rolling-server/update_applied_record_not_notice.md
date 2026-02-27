### 描述

- 该接口提供版本：v1.8.9.1+。
- 该接口所需权限：平台管理-滚服管理。
- 该接口功能描述：屏蔽滚服申请记录通知。

### URL

POST /api/v1/woa/rolling_servers/applied_records/notice/disabled/update

### 输入参数

| 参数名称 | 参数类型         | 必选    | 描述                 |
|------|--------------|---------|--------------------|
| ids  | string array | 是      | 滚服申请记录id列表，长度限制100 |

### 调用示例

```json
{
  "ids": ["0000001", "0000002"]
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

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| result     | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| data	     | object       | 响应数据             |
