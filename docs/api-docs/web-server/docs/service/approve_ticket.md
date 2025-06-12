### 描述

- 该接口提供版本：v1.1.23+。
- 该接口所需权限：。
- 该接口功能描述：处理Itsm单据。

### URL

POST /api/v1/web/tickets/approve

### 输入参数

| 参数名称      | 参数类型    | 必选 | 描述                       |
|-----------|---------|----|--------------------------|
| sn	       | string	 | 是	 | 单号                       |
| state_id	 | int64	  | 是	 | 单据所处流程ID                 |
| action	   | string	 | 是	 | 审批动作。（pass:通过，refuse:拒绝） |
| memo	     | string	 | 否	 | 审批意见                     |

### 调用示例

```json
{
  "sn": "REQ20230725000000",
  "state_id": 4270,
  "action": "pass",
  "memo": "test"
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": ""
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
