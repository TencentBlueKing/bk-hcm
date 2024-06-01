### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询指定的目标组绑定的负载均衡下的端口健康信息。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/target_groups/{target_group_id}/targets/health

### 输入参数

| 参数名称         | 参数类型      | 必选 | 描述           |
|-----------------|--------------|-----|---------------|
| bk_biz_id       | int          | 是  | 业务ID         |
| target_group_id | string       | 是  | 目标组ID        |
| cloud_lb_ids    | string array | 是  | 云负载均衡ID数组 |

### 调用示例

```json
{
    "cloud_lb_ids": ["lb-xxxxxx"]
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
        "cloud_lb_id": "lb-prcn8tfk",
        "listeners": [
          {
            "cloud_lbl_id": "lbl-eu8ct24u",
            "protocol": "HTTP",
            "listener_name": "lt-test-001",
            "rules": [
              {
                "cloud_rule_id": "loc-g06bng5g",
                "health_check": {
                  "health_num": 0,
                  "un_health_num": 4,
                }
              }
            ]
          }
        ]
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述    |
|---------|--------|---------|
| code    | int    | 状态码   |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述                    |
|---------|--------|-------------------------|
| count   | int    | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据             |

#### details

| 参数名称     | 参数类型 | 描述               |
|-------------|--------|--------------------|
| cloud_lb_id | string | 云负载均衡ID        |
| listeners   | array  | 查询返回的监听器数组  |

#### listeners

| 参数名称        | 参数类型       | 描述              |
|----------------|--------------|-------------------|
| cloud_lbl_id   | string       | 云监听器ID         |
| listener_name  | string       | 云监听器名称        |
| protocol       | string       | 云监听器协议        |
| health_check   | object       | 4层监听器的健康检查  |
| rules          | array        | 7层规则的数组       |

#### rules

| 参数名称        | 参数类型       | 描述            |
|----------------|--------------|-----------------|
| cloud_rule_id  | string       | 云规则ID         |
| health_check   | object       | 7层规则的健康检查  |

#### health_check

| 参数名称        | 参数类型 | 描述      |
|----------------|--------|-----------|
| health_num     | int    | 健康阈值   |
| un_health_num  | int    | 不健康阈值 |
