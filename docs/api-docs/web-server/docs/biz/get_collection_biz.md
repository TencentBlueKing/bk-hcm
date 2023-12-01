### 描述

- 该接口提供版本：v1.2.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询用户收藏的业务ID列表。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/collections/bizs

### 输入参数

### 调用示例

#### 获取详细信息请求参数示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    1,
    2,
    3,
    4
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述   |
|---------|------------|------|
| code    | int        | 状态码  |
| message | string     | 请求信息 |
| data    | int object | 响应数据 |
