### 描述

- 该接口提供版本：v1.2.1+。
- 该接口所需权限：回收站操作。
- 该接口功能描述：批量删除回收站中的虚拟机。

### URL

DELETE /api/v1/cloud/bizs/{bk_biz_id}/recycled/cvms/batch

### 输入参数

| 参数名称       | 参数类型         | 必选 | 描述     |
|------------|--------------|----|--------|
| bk_biz_id  | int64        | 是  | 业务ID   |
| record_ids | string array | 是  | 回收记录ID |

### 调用示例

```json
{
  "record_ids": [
    "00000001",
    "00000002"
  ]
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
