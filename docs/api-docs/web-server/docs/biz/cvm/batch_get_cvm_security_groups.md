### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询虚拟机绑定的安全组列表。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/cvms/security_groups/batch/list

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
  "data": [
    {
      "cvm_id": "xxxxx",
      "security_groups": [
        {
          "id": "Xxxxx",
          "cloud_id": "Xxxxx",
          "name": "default"
        }
      ]
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述   |
|---------|--------------|------|
| code    | int32        | 状态码  |
| message | string       | 请求信息 |
| data    | object array | 响应数据 |

#### data[i] 参数说明

| 参数名称            | 参数类型         | 描述    |
|-----------------|--------------|-------|
| cvm_id          | string       | 云主机ID |
| security_groups | object array | 安全组列表 |

#### security_groups[i] 参数说明

| 参数名称     | 参数类型   | 描述    |
|----------|--------|-------|
| id       | string | 安全组ID |
| cloud_id | string | 云ID   |
| name     | string | 安全组名称 |
