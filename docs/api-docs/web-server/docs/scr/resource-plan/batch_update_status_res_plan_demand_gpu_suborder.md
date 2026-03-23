### 描述

- 该接口提供版本：v1.8.10.3+。
- 该接口所需权限：平台-GPU需求。
- 该接口功能描述：资源下GPU需求提报子单批量更新状态接口。

### URL

POST /api/v1/woa/plans/resources/gpu/demands/suborders/batch/status

### 输入参数

| 参数名称        | 参数类型      | 必选 | 描述                                     |
|---------------|--------------|------|-----------------------------------------|
| suborder_ids  | string array | 是   | GPU需求提报子单ID列表，最大数量限制1000条     |
| status        | string       | 是   | 需求子单状态（DONE:评审通过 REJECT:评审驳回） |
| comment       | json array   | 否   | 评审意见                                  |

### 调用示例

#### 请求参数示例

```json
{
  "suborder_ids": ["suborderid-1001", "suborderid-1002"],
  "status": "DONE",
  "comment": ["评审通过"]
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

| 参数名称 | 参数类型 | 描述                               |
|---------|--------|------------------------------------|
| code    | int    | 错误编码。 0表示success，>0表示失败错误 |
| message | string | 请求失败返回的错误信息                 |
| data    | object | 响应数据                             |

#### data

无