### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：业务视角下查看申请单详情。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/applications/{application_id}

### 输入参数

| 参数名称           | 参数类型   | 必选 | 描述    |
|----------------|--------|----|-------|
| bk_biz_id      | int64  | 是  | 业务ID  |
| application_id | string | 是  | 申请单ID |

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000001",
    "source": "itsm",
    "sn": "REQ20260401000001",
    "type": "create_cvm",
    "status": "completed",
    "applicant": "admin",
    "content": "{...}",
    "delivery_detail": "{...}",
    "memo": "申请云主机",
    "creator": "admin",
    "reviser": "admin",
    "created_at": "2026-04-01T10:00:00Z",
    "updated_at": "2026-04-01T10:30:00Z",
    "ticket_url": "https://itsm.example.com/ticket/xxx"
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

| 参数名称            | 参数类型   | 描述                                                                                           |
|-----------------|--------|----------------------------------------------------------------------------------------------|
| id              | string | 申请ID                                                                                         |
| source          | string | 来源（枚举值：itsm、bpaas）                                                                          |
| sn              | string | 序列号                                                                                          |
| type            | string | 申请类型（枚举值：add_account、create_cvm、create_vpc、create_disk）                                      |
| status          | string | 申请状态（枚举值：pending、pass、rejected、cancelled、delivering、completed、deliver_partial、deliver_error） |
| applicant       | string | 申请人                                                                                          |
| content         | string | 申请内容（已脱敏）                                                                                    |
| delivery_detail | string | 交付详情                                                                                         |
| memo            | string | 备注                                                                                           |
| creator         | string | 创建者                                                                                          |
| reviser         | string | 更新者                                                                                          |
| created_at      | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                                               |
| updated_at      | string | 更新时间，标准格式：2006-01-02T15:04:05Z                                                               |
| ticket_url      | string | ITSM审批链接                                                                                     |

### 错误码

| 错误码            | 描述                                   |
|----------------|--------------------------------------|
| RecordNotFound | 申请单不存在、用户无业务访问权限、或申请单不归属当前业务时均返回此错误 |
