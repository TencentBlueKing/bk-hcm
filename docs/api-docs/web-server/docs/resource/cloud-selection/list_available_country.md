### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源选型-选型推荐。
- 该接口功能描述：查询云选型数据支持的国家。

### URL

POST /api/v1/cloud/selections/countries/list

### 输入参数
无


### 调用示例
```
{}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
     "country_1","country_2","country_3","country_4"
    ]
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

| 参数名称    | 参数类型   | 描述             |
|---------|--------|----------------|
| details | string array  | 查询返回的数据        |


