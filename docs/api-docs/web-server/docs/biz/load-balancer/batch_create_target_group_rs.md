### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：业务下给指定目标组批量添加RS。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/target_groups/targets/create

### 输入参数

| 参数名称          | 参数类型       | 必选 | 描述                   |
|------------------|--------------|------|-----------------------|
| bk_biz_id        | int          | 是   | 业务ID                 |
| account_id       | string       | 是   | 账号ID                 |
| target_groups    | object array | 是   | 目标组列表，单次最多10个  |

#### target_groups

| 参数名称          | 参数类型       | 必选 | 描述                   |
|------------------|--------------|------|-----------------------|
| target_group_id  | string       | 是   | 目标组ID                |
| targets          | object array | 是   | RS实例列表，单次最多100个 |

#### targets

| 参数名称          | 参数类型       | 必选 | 描述                               |
|------------------|--------------|------|-----------------------------------|
| inst_type        | string       | 是   | 实例类型(CVM:云服务器)               |
| cloud_inst_id    | string       | 是   | 云实例ID                           |
| port             | int          | 是   | 端口                               |
| weight           | int          | 是   | 权重,取值范围：[0, 100]              |

### 调用示例

```json
{
  "account_id": "00000001",
  "target_groups": [
    {
      "target_group_id": "0000000g",
        "targets": [
          {
            "inst_type": "CVM",
            "cloud_inst_id": "cvm-xxxxxx",
            "port": 8000,
            "weight": 10
          }
        ]
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "flow_id": "xxxxxxxx"
  }
}
```

### 响应参数说明

| 参数名称  | 参数类型  | 描述    |
|---------|----------|---------|
| code    | int      | 状态码   |
| message | string   | 请求信息 |
| data    | object   | 响应数据 |

#### data

| 参数名称  | 参数类型 | 描述    |
|----------|--------|---------|
| flow_id  | string | 任务id   |
