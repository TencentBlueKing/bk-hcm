### 描述

- 该接口提供版本：v9.9.9。
- 该接口所需权限：无。
- 该接口功能描述：获取预测需要审批的 itsm 单据。

### URL

POST /api/v1/woa/plans/resources/itsm/ticket/list

### 输入参数

| 参数名称         | 参数类型   | 必选 | 描述               |
|--------------|--------|----|------------------|
| submitted_at | object | 是  | 提单时间范围           |

#### submitted_at

| 参数名称  | 参数类型   | 必选 | 描述                                    |
|-------|--------|----|---------------------------------------|
| start | string | 是  | 查询提单时间大于等于该时间的单据                      |
| end   | string | 否  | 查询提单时间小于等于该时间的单据，不提供时只使用 start 条件 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "submitted_at": {
    "start": "2022-11-14T01:57:41.159Z",
    "end": "2022-11-15T01:57:41.159Z"
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "tickets": [
      {
        "id": "111",
        "ticket_id": "00000001",
        "url": "http://www.test.com/111",
        "user": "admin",
        "approval_state": "leader_approval",
        "submitted_at": "2022-11-14T01:57:41.159Z"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                          |
|---------|--------|-----------------------------|
| result  | bool   | 请求成功与否。true:请求成功；false 请求失败 |
| code    | int    | 错误编码。0 表示 success，>0 表示失败错误 |
| message | string | 请求失败返回的错误信息                 |
| data    | object | 响应数据                        |

#### data

| 参数名称    | 参数类型         | 描述   |
|---------|--------------|------|
| tickets | object array | 单据列表 |

#### tickets[0]

| 参数名称           | 参数类型   | 描述                                                                   |
|----------------|--------|----------------------------------------------------------------------|
| id             | string | itsm 单据 id                                                         |
| ticket_id      | string | HCM 侧的单据 id（主单 id）                                                |
| url            | string | itsm 链接                                                             |
| user           | string | 提单人                                                                |
| approval_state | string | 审批状态，"leader_approval"(直属 leader 审批)、"hcm_admin_approval"(hcm 管理员审批) |
| submitted_at   | string | 单据提单时间                                                           |


