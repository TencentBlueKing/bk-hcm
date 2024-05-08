### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询申请。

### URL

GET /api/v1/cloud/applications/{application_id}

### 输入参数

| 参数名称           | 参数类型   | 必选 | 描述   |
|----------------|--------|----|------|
| application_id | string | 是  | 申请ID |

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 0,
    "details": [
      {
        "id": "1",
        "sn": "1",
        "type": "1",
        "status": "1",
        "applicant": "xxxxxxx",
        "content": "",
        "delivery_detail": "1",
        "memo": "1",
        "creator": "xxxxxxx",
        "reviser": "xxxxxxx",
        "created_at": "2023-06-09T11:00:08Z",
        "updated_at": "2023-06-09T11:01:10Z",
        "ticket_url": "xxxxxxxxxxxxxxxxxx"
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
| count   | uint64 | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据        |

#### data.details[n]

| 参数名称            | 参数类型   | 描述                                                                                           |
|-----------------|--------|----------------------------------------------------------------------------------------------|
| id              | string | 申请ID                                                                                         |
| source          | string | 来源（枚举值：itsm,bpaas)                                                                           |
| sn              | string | 序列号                                                                                          |
| type            | string | 申请类型（枚举值：add_account、create_cvm、create_vpc、create_disk）                                      |
| status          | string | 申请状态（枚举值：pending、pass、rejected、cancelled、delivering、completed、deliver_partial、deliver_error） |
| applicant       | string | 申请人                                                                                          |
| content         | string | 申请内容                                                                                         |
| delivery_detail | string | 交付详情                                                                                         |
| memo            | string | 备注                                                                                           |
| creator         | string | 创建者                                                                                          |
| reviser         | string | 更新者                                                                                          |
| created_at      | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                                               |
| updated_at      | string | 更新时间，标准格式：2006-01-02T15:04:05Z                                                               |
| ticket_url      | string | 门票地址                                                                                         |
