### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：
- 该接口功能描述：Itsm回调接口。

### URL

POST /api/v1/cloud/applications/approve

### 输入参数

| 参数名称           | 参数类型   | 必选 | 描述    |
|----------------|--------|----|-------|
| callback_token | string | 是  | token |
| ticket         | object | 是  | 审批单信息 |

#### ticket

| 参数名称           | 参数类型   | 必选 | 描述                                                          |
|----------------|--------|----|-------------------------------------------------------------|
| workflow_id    | string | 是  | 流程ID                                                        |
| id             | string | 是  | 单据ID                                                        |
| title          | string | 是  | 单据标题                                                        |
| approve_result | bool   | 是  | 批准结果                                                        |
| status         | string | 是  | 单据状态，枚举值：draft/running/finished/suspend/termination/revoked |
| end_at         | string | 否  | 单据结束时间，格式为 YYYY-MM-DD HH:MM:SS                              |
| sn             | string | 否  | 暂无实际用途                                                      |

### 调用示例

```json
{
  "callback_token": "xxxxxxxxxx",
  "ticket": {
    "workflow_id": "20250605160200001901",
    "id": "102025061220360800005902",
    "title": "测试单据",
    "approve_result": true,
    "status": "finished",
    "end_at": "2024-10-31 11:19:12",
    "sn": "HCM2025061200000001"
  }
}
```

### 响应示例

```json
{
  "result": true,
  "code": 0,
  "message": ""
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述     |
|---------|--------|--------|
| result  | bool   | 是否请求成功 |
| code    | int32  | 状态码    |
| message | string | 请求信息   |
