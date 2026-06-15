### 描述

- 该接口提供版本：v1.8.0+。
- 该接口所需权限：业务访问权限。
- 该接口功能描述：业务视角查看申请单据明细。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/applications/{application_id}

### 输入参数

| 参数名称           | 参数类型   | 必选 | 描述     |
|----------------|--------|----|--------|
| bk_biz_id      | int64  | 是  | 业务ID   |
| application_id | string | 是  | 申请单据ID |

### 调用示例

```bash
curl -X GET "http://your-host/api/v1/cloud/bizs/123/applications/abc-123-def"
```

### 响应示例

#### 成功响应

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "abc-123-def",
    "source": "itsm",
    "sn": "SN20260324001",
    "type": "create_cvm",
    "status": "pending",
    "applicant": "user1",
    "content": "{\"region\":\"ap-guangzhou\",\"instance_type\":\"S5.MEDIUM2\"}",
    "delivery_detail": "",
    "memo": "申请云主机",
    "creator": "user1",
    "reviser": "user1",
    "created_at": "2026-03-24T10:00:00Z",
    "updated_at": "2026-03-24T10:00:00Z",
    "ticket_url": "https://itsm.example.com/ticket/12345"
  }
}
```

#### 错误响应（单据不存在/无权限/不归属该业务）

```json
{
  "code": 404,
  "message": "application not found",
  "data": null
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
| id              | string | 申请单据ID                                                                                       |
| source          | string | 来源（枚举值：itsm、bpaas）                                                                          |
| sn              | string | 序列号                                                                                          |
| type            | string | 申请类型（枚举值：add_account、create_cvm、create_vpc、create_disk、create_load_balancer 等）               |
| status          | string | 申请状态（枚举值：pending、pass、rejected、cancelled、delivering、completed、deliver_partial、deliver_error） |
| applicant       | string | 申请人                                                                                          |
| content         | string | 申请内容（敏感字段如 password 已脱敏）                                                                     |
| delivery_detail | string | 交付详情                                                                                         |
| memo            | string | 备注                                                                                           |
| creator         | string | 创建者                                                                                          |
| reviser         | string | 更新者                                                                                          |
| created_at      | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                                               |
| updated_at      | string | 更新时间，标准格式：2006-01-02T15:04:05Z                                                               |
| ticket_url      | string | ITSM 审批链接（仅 source=itsm 时返回）                                                                 |

### 权限说明

- 该接口使用**业务访问权限**（`meta.Biz.Access`）进行鉴权
- 用户必须拥有对应业务的访问权限
- 单据的 `bk_biz_ids` 必须包含请求路径中的 `bk_biz_id`

### 错误码说明

| 错误码 | 说明                                   |
|-----|--------------------------------------|
| 404 | 单据不存在 / 用户无业务访问权限 / 单据不归属该业务（统一返回） |

### 注意事项

1. 该接口是业务视角的单据查看接口，与资源视角接口 `GET /api/v1/cloud/applications/{application_id}` 不同
2. 为了安全考虑，所有错误场景（无权限、不归属、不存在）统一返回 `NotFound`，避免泄露单据存在性
3. 单据可能归属多个业务，只要请求的 `bk_biz_id` 在单据的 `bk_biz_ids` 列表中即可查看
4. 敏感字段（如密码）会被自动脱敏
