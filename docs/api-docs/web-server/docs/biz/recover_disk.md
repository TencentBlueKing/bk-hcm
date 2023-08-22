### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：回收站操作。
- 该接口功能描述：从回收站恢复硬盘。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/disks/recover

### 输入参数

| 参数名称 | 参数类型         | 必选  | 描述        |
|------|--------------|-----|-----------|
| bk_biz_id | int64        | 是   | 业务的ID     |
| record_ids | string array | 是   | 回收记录ID |

### 调用示例

```json
{
  "record_ids": [
    "000000001"
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
