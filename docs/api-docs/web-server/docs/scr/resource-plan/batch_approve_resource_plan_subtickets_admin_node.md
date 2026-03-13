### 描述

- 该接口提供版本：v1.8.10.1。
- 该接口所需权限：平台-单据管理。
- 该接口功能描述：批量审批处于“部门审批”阶段的资源预测子单（管理员审批）。

### URL

POST /api/v1/woa/plans/resources/sub_tickets/approve_admin_node/batch

### 输入参数

| 参数名称              | 参数类型         | 必选 | 描述                                                   |
|-------------------|--------------|----|------------------------------------------------------|
| sub_ticket_ids    | string array | 是  | 需要进行审批的子单 ID 列表（只能包含当前处于部门审批的子单）       |
| approval          | bool         | 是  | 审批结果，true=同意，false=拒绝                |
| use_transfer_pool | bool         | 是  | 是否使用中转池额度                  |
| operate_info      | string       | 否  | 审批意见/理由，最多 100 字           |

### 调用示例

#### 请求

```json
{
  "sub_ticket_ids": [
    "00000001",
    "00000002",
    "00000003"
  ],
  "approval": true,
  "use_transfer_pool": true,
  "operate_info": "审批意见"
}
```

### 响应示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "handled_count": 3
  }
}
```

### 响应参数说明

| 参数名称               | 参数类型    | 描述                              |
|----------------------|---------|---------------------------------|
| result  | bool    | 请求成功与否。true:请求成功；false请求失败      |
| code    | int     | 错误编码。 0表示success，>0表示失败错误       |
| message | string  | 描述信息                            |
| data    | object  | 响应数据                            |

#### data

| 参数名称               | 参数类型   | 描述                        |
|--------------------|--------|---------------------------|
| handled_count      | int    | 实际成功处理的子单数量               |