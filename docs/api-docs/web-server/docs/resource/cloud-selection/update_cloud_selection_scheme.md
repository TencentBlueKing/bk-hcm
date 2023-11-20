### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源选型-方案编辑。
- 该接口功能描述：更新云选型方案。

### URL

PATCH /api/v1/cloud/selections/schemes/{id}

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述   |
|-----------|--------|----|------|
| id        | string | 是  | 方案ID |
| name      | string | 否  | 名称   |
| bk_biz_id | int    | 否  | 业务ID |

### 调用示例

```json
{
  "name": "tcloud_account",
  "bk_biz_id": 310
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
| code    | int32  | 状态码  |
| message | string | 请求信息 |
