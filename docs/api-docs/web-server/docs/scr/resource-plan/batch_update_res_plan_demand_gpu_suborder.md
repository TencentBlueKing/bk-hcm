### 描述

- 该接口提供版本：v1.8.10.3+。
- 该接口所需权限：平台-GPU需求。
- 该接口功能描述：资源下GPU需求提报子单批量编辑接口。

### URL

POST /api/v1/woa/plans/resources/gpu/demands/suborders/batch

### 输入参数

| 参数名称        | 参数类型      | 必选 | 描述                                  |
|---------------|--------------|------|--------------------------------------|
| suborder_data | object array | 是   | GPU需求提报子单数据列表，最大数量限制100条 |

### suborder_data

| 参数名称        | 参数类型      | 必选 | 描述                                                         |
|---------------|--------------|------|-------------------------------------------------------------|
| suborder_id   | string       | 是   | GPU需求提报子单ID                                              |
| status        | string       | 否   | 需求子单状态（子单评审/驳回时，该参数必传；不能跟extension同时传）    |
| comment       | json array   | 否   | 评审意见                                                      |
| demand_year   | int          | 否   | 需求年份                                                      |
| demand_month  | int          | 否   | 需求月份                                                      |
| gpu_num       | int          | 否   | GPU卡数                                                       |
| qpm_max       | int          | 否   | 峰值QPM                                                       |
| extension     | json array   | 否   | GPU需求明细数据（子单编辑时，该参数必传；不能跟status、comment同时传）|

### 调用示例

#### 请求参数示例

```json
{
  "suborder_data": [{
    "suborder_id": "suborderid-1001",
    "status": "DONE",
    "comment": ["评审通过"]
  }]
}

{
  "suborder_data": [{
    "suborder_id": "suborderid-1001",
    "demand_year": 2026,
    "demand_month": 3,
    "gpu_num": 100,
    "extension": [1, "a", 3.5]
  }]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                        |
|---------|--------|---------------------------|
| code    | int    | 错误编码。 0表示success，>0表示失败错误 |
| message | string | 请求失败返回的错误信息               |
| data    | object | 响应数据                      |

#### data

无