### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：删除收藏的业务。

### URL

DELETE /api/v1/cloud/bizs/{bk_biz_id}/collections/bizs

### 输入参数

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "bk_biz_id": 1
}
```

### 响应示例

```json
{
  "code": 0,
  "message": ""
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
