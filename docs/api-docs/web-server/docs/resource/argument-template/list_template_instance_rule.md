### 描述

- 该接口提供版本：v1.4.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询参数模版绑定的实例数及规则数列表。

### URL

POST /api/v1/cloud/argument_templates/instance/rule/list

### 输入参数

| 参数名称 | 参数类型         | 必选 | 描述                      |
|------|--------------|----|-------------------------|
| ids  | string array | 是  | 要查询的参数模版ID，最多数量不能超过100个 |

### 调用示例

#### 获取详细信息请求参数示例

查询模版id是00000001绑定的实例数及规则数列表。

```json
{
  "ids": [
    "00000001"
  ]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "id": "00000001",
        "instance_num": 5,
        "rule_num": 6
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称    | 参数类型  | 描述             |
|---------|-------|----------------|
| count   | int   | 当前规则能匹配到的总记录条数 |
| details | array | 查询返回的数据        |

#### data.details[n]

| 参数名称         | 参数类型   | 描述      |
|--------------|--------|---------|
| id           | string | 参数模版ID  |
| instance_num | string | 绑定的实例数量 |
| rule_num     | string | 绑定的规则数量 |
