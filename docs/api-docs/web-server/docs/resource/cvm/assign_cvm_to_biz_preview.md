### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源分配。
- 该接口功能描述：分配虚拟机到业务下预览。

### URL

POST /api/v1/cloud/cvms/assign/bizs/preview

### 输入参数

| 参数名称      | 参数类型         | 必选  | 描述                |
|-----------|--------------|-----|-------------------|
| cvm_ids   | string array | 是   | 虚拟机的ID列表, 单批限制500 |

### 调用示例

```json
{
  "cvm_ids": [
    "00000001",
    "00000002",
    "00000003"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "details": [
      {
        "cvm_id": "00000001",
        "match_type": "auto",
        "bk_cloud_id": 1,
        "bk_biz_id": 1
      },
      {
        "cvm_id": "00000002",
        "match_type": "manual"
      },
      {
        "cvm_id": "00000003",
        "match_type": "no_match"
      }
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
| details | array  | 查询返回的数据        |

#### data.details[n]

| 参数名称        | 参数类型   | 描述                                             |
|-------------|--------|------------------------------------------------|
| cvm_id      | string | 虚拟机的ID                                         |
| match_type  | string | 匹配状态，枚举值：auto(自动匹配)、manual(手动匹配)、no_match(待关联) |
| bk_cloud_id | int64  | 管控区域ID                                         |
| bk_biz_id   | int64  | 业务ID                                           |
