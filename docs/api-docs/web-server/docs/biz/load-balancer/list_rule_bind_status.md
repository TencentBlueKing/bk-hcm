### 描述

- 该接口提供版本：v1.7.4+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询规则绑定目标组状态接口。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/listeners/{lbl_id}/rules/binding_status/list

### 输入参数

| 参数名称      | 参数类型    | 必选 | 描述              |
|-----------|---------|----|-----------------|
| bk_biz_id | int     | 是  | 业务ID            |
| vendor    | string  | 是  | 供应商（枚举值：tcloud） |
| lbl_id    | string  | 是  | 监听器id           |
| rule_ids  | string array | 是  | 规则id，列表最大为100   |

### 调用示例
```json
{
  "rule_ids": ["1111", "2222"]
}
```

### 响应示例

```json
{
  "result": true,
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "rule_id": "1111",
        "binding_status": "success"
      },
      {
        "rule_id": "2222",
        "binding_status": "failed"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |


#### data参数说明

| 参数名称       | 参数类型   | 描述     |
|------------|--------|--------|
| details    | object | 规则绑定状态 |

#### details[0]
| 参数名称       | 参数类型   | 描述     |
|------------|--------|--------|
| rule_id    | string | 规则id   |
| binding_status    | string | 规则绑定状态绑定状态(success:成功 failed:失败 binding:绑定中) |

