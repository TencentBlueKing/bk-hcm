### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下监听器规则绑定目标组。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/rules/target_group/bind

### 输入参数

| 参数名称            | 参数类型   | 必选 | 描述              |
|-----------------|--------|----|-----------------|
| bk_biz_id       | int64  | 是  | 业务ID            |
| vendor          | string | 是  | 供应商（枚举值：tcloud） |
| url_rule_id     | string | 是  | 监听器规则ID         |
| target_group_id | string | 是  | 目标组id           |

### 调用示例


```json
{
    "url_rule_id": "xxxxxxx",
    "target_group_id": "xxxxxxx"
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": "xxxxxx"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述     |
|---------|--------|--------|
| code    | int32  | 状态码    |
| message | string | 请求信息   |
| data    | string | 任务管理ID |
