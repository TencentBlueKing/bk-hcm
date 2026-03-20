### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：业务下GPU需求提报子单批量终止。将指定子单的状态修改为"已终止"，仅限子单状态为"已驳回"时可操作，单次最多终止100条。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/gpu/demands/suborders/batch/terminate

### 输入参数

| 参数名称       | 参数类型       | 必选 | 描述                                  |
|---------------|--------------|------|--------------------------------------|
| bk_biz_id     | int64        | 是   | 业务ID（路径参数）                      |
| suborder_ids  | string array | 是   | GPU需求提报子单ID列表，数量范围：[1, 100] |

### 调用示例

#### 请求参数示例

```json
{
  "bk_biz_id": 111,
  "suborder_ids": ["suborder-001", "suborder-002"]
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