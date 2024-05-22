### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询负载均衡绑定安全组列表。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/security_groups/res/{res_type}/{res_id}

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述   |
|-----------|--------|----|------|
| bk_biz_id | int64  | 是  | 业务ID |
| res_id    | string | 是  | 资源ID |
| res_type  | string | 是  | 资源类型 |

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "id": "00000004",
      "vendor": "tcloud",
      "cloud_id": "sg-xxxxx",
      "region": "ap-guangzhou",
      "name": "security-group",
      "memo": "security group",
      "account_id": "00000003",
      "bk_biz_id": -1,
      "creator": "sync-timing-admin",
      "reviser": "sync-timing-admin",
      "created_at": "2023-02-25T18:28:46Z",
      "updated_at": "2023-02-27T19:14:33Z",
      "res_id": "0000000x",
      "res_type": "load_balancer",
      "priority": 1,
      "rel_creator": "Jim",
      "rel_created_at": "2023-02-27T19:31:33Z"
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述   |
|---------|--------------|------|
| code    | int          | 状态码  |
| message | string       | 请求信息 |
| data    | object array | 响应数据 |

#### data[n]

| 参数名称           | 参数类型   | 描述                                                 |
|----------------|--------|----------------------------------------------------|
| id             | string | 安全组ID                                              |
| vendor         | string | 云厂商                                                |
| cloud_id       | string | 安全组云ID                                             |
| bk_biz_id      | int64  | 业务ID, -1代表未分配业务                                    |
| region         | string | 地域                                                 |
| name           | string | 安全组名称                                              |
| memo           | string | 安全组备注                                              |
| account_id     | string | 安全组账号ID                                            |
| creator        | string | 安全组创建者                                             |
| reviser        | string | 安全组最后一次修改的修改者                                      |
| created_at     | string | 安全组创建时间，标准格式：2006-01-02T15:04:05Z                  |
| updated_at     | string | 安全组最后一次修改时间，标准格式：2006-01-02T15:04:05Z              |
| res_id         | string | 资源ID                                               |
| res_type       | string | 资源类型                                               |
| priority       | int64  | 安全组排序ID                                            |
| rel_creator    | string | 负载均衡和安全组绑定操作人（azure该字段为空）                          |
| rel_created_at | string | 负载均衡和安全组绑定时间（azure该字段为空，标准格式：2006-01-02T15:04:05Z） |
