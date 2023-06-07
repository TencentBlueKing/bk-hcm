### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询主机绑定安全组列表。

### URL

GET /api/v1/cloud/security_groups/cvms/{cvm_id}

### 输入参数

| 参数名称   | 参数类型   | 必选  | 描述   |
|--------|--------|-----|------|
| cvm_id | string | 是   | 主机ID |

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
      "cvm_id": "0000000x",
      "rel_creator": "Jim",
      "rel_created_at": "2023-02-27T19:31:33Z"
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

#### data[n]

| 参数名称           | 参数类型         | 描述                      |
|----------------|--------------|-------------------------|
| id             | string | 安全组ID                   |
| vendor         | string | 云厂商                     |
| cloud_id       | string | 安全组云ID                  |
| bk_biz_id      | int64  | 业务ID, -1代表未分配业务         |
| region         | string | 地域                      |
| name           | string | 安全组名称                   |
| memo           | string | 安全组备注                   |
| account_id     | string | 安全组账号ID                 |
| creator        | string | 安全组创建者                  |
| reviser        | string | 安全组最后一次修改的修改者           |
| created_at     | string | 安全组创建时间，标准格式：2006-01-02T15:04:05Z                 |
| updated_at     | string | 安全组最后一次修改时间，标准格式：2006-01-02T15:04:05Z             |
| cvm_id         | string | 主机ID                    |
| rel_creator    | string | 主机和安全组绑定操作人（azure该字段为空） |
| rel_created_at | string | 主机和安全组绑定时间（azure该字段为空，标准格式：2006-01-02T15:04:05Z）   |
