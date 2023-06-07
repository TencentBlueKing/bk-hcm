### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源分配。
- 该接口功能描述：快速分配一个账号下的资源到业务下。

### URL

POST /api/v1/cloud/resources/assign/bizs

### 输入参数

| 参数名称            | 参数类型         | 必选  | 描述                                           |
|-----------------|--------------|-----|----------------------------------------------|
| account_id      | string       | 是   | 账号ID                                         |
| bk_biz_id       | int64        | 是   | 业务ID                                         |
| res_types       | string array | 否   | 要分配的资源类型，res_types和is_all_res_type有且只有一个必填   |
| is_all_res_type | boolean      | 否   | 是否分配全部资源类型，res_types和is_all_res_type有且只有一个必填 |

### 调用示例

```json
{
  "account_id": "00000001",
  "bk_biz_id": 3,
  "is_all_res_type": true
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
