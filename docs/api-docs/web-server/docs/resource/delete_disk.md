### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：IaaS资源删除。
- 该接口功能描述：删除云盘。

### URL

DELETE /api/v1/cloud/disks/{id}

### 输入参数

| 参数名称 | 参数类型    | 必选 | 描述     |
|------|---------|----|--------|
| id   | string  | 是  | 云盘 ID  |

### 调用示例

如删除云厂商是 tcloud , ID 是 00000002 的云盘信息

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明

| 参数名称     | 参数类型   | 描述   |
|----------|--------|------|
| code     | int    | 状态码  |
| message  | string | 请求信息 |
