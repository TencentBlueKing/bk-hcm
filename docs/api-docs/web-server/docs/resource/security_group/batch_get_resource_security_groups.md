### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询指定资源绑定的安全组列表。

### URL

POST /api/v1/cloud/security_groups/res/{res_type}/batch

### 输入参数

| 参数名称     | 参数类型         | 必选                           | 描述                 |
|----------|--------------|------------------------------|--------------------|
| res_type | string       | 资源类型, 可选值：cvm, load_balancer |
| res_ids  | string array | 是                            | 资源ID列表, 最多传入500个ID |


###  调用示例

```json

{
  "res_ids": [
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
      "res_id": "000001",
      "security_groups": [
        {
          "id": "Xxxxx",
          "cloud_id": "Xxxxx",
          "name": "default"
        }
      ]
    },
    {
      "res_id": "000002",
      "security_groups": []
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
| res_id          | string       | 资源ID  |
| security_groups | object array | 安全组列表 |

#### security_groups[i] 参数说明

| 参数名称     | 参数类型   | 描述    |
|----------|--------|-------|
| id       | string | 安全组ID |
| cloud_id | string | 云ID   |
| name     | string | 安全组名称 |
