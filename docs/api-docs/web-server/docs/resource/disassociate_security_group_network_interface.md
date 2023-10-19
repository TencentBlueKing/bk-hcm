### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：IaaS资源操作。
- 该接口功能描述：取消安全组和网络接口的关联（仅支持：azure）。

### URL

POST /api/v1/cloud/security_groups/disassociate/network_interfaces

### 输入参数

| 参数名称                 | 参数类型      | 必选  | 描述    |
|----------------------|-----------|-----|-------|
| security_group_id    | string    | 是   | 安全组ID |
| network_interface_id | string    | 是   | 网络接口ID  |

### 调用示例

```json
{
  "security_group_id": "00001111",
  "network_interface_id": "00001112"
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
