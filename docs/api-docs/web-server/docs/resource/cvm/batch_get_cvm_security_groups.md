### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询虚拟机绑定的安全组列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/cvms/security_groups/batch

### 输入参数

| 参数名称      | 参数类型         | 必选 | 描述                  |
|-----------|--------------|----|---------------------|
| bk_biz_id | int64        | 是  | 业务ID                |
| cvm_ids   | string array | 是  | 云主机ID列表, 最多传入500个ID |


###  调用示例

```json

{
  "cvm_ids": [
    "xxxxx",
    "xxxxxx"
  ]
}

```
### 响应示例
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "xxxxx": [
      {
        "id": "Xxxxx",
        "cloud_id": "Xxxxx",
        "name": "default"
      }
    ],
    "xxxxxx": [
      {
        "id": "Xxxxx",
        "cloud_id": "Xxxxx",
        "name": "default2"
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

#### data参数说明

| 参数名称  | 参数类型   | 描述    |
|-------|--------|-------|
| key   | string | 云主机ID |
| value | array  | 安全组列表 |

#### data[n].value参数说明

| 参数名称     | 参数类型   | 描述    |
|----------|--------|-------|
| id       | string | 安全组ID |
| cloud_id | string | 云ID   |
| name     | string | 安全组名称 |
