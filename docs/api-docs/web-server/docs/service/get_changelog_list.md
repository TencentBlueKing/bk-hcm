### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：无。
- 该接口功能描述：获取版本日志列表。

### URL

GET /api/v1/web/changelogs


### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "version": "v3.10.10",
      "time": "2021-12-24",
      "is_current": false
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型     | 描述   |
|---------|----------|------|
| code    | int32    | 状态码  |
| message | string   | 请求信息 |
| data    | []object | 响应数据 |

#### data[n] 字段说明
| 参数名称       | 参数类型    | 描述      |
|------------|---------|---------|
| version    | string  | 版本信息    |
| time       | string  | 版本发布时间  |
| is_current | boolean | 是否为当前版本 |
