### 描述

- 该接口提供版本：v1.8.5.6+。
- 该接口所需权限：平台管理-单据管理。
- 该接口功能描述：资源预测申请单据审核-部门审批阶段。

### URL

POST /api/v1/woa/plans/resources/sub_tickets/{sub_ticket_id}/approve_admin_node

### 输入参数

| 参数名称               | 参数类型 | 必选 | 描述          |
|--------------------|------|----|-------------|
| approval	          | bool | 是  | 是否通过        |
| use_transfer_pool	 | bool | 是  | 是否使用中转池额度   |
| operate_info   | string | 否  | 审批意见,最多100字 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "approval": true,
  "use_transfer_pool": false,
  "operate_info": ""
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述                         |
|---------|--------------|----------------------------|
| result  | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code    | int          | 错误编码。 0表示success，>0表示失败错误  |
| message | string       | 请求失败返回的错误信息                |
| data	   | object array | 响应数据                       |

#### data

无
