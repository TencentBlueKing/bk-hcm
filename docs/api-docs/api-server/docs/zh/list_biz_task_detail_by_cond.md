### 描述

- 该接口提供版本：v1.8.2+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询指定条件的任务详情(内部接口，只给特定业务使用，不对外公开)。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/task_details/list_by_cond

### 输入参数

| 参数名称             | 参数类型       | 必选 | 描述                   |
|---------------------|--------------|------|-----------------------|
| bk_biz_id           | int64        | 是   | 业务ID                 |
| task_management_ids | string array | 是   | 任务管理ID，最大数量为100 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "task_management_ids": ["xxxxxx"]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 0,
    "details": [
      {
        "id": "00000001",
        "task_management_id": "00000001",
        "flow_id": "00000001",
        "task_action_ids": ["00000001"],
        "operation": "binding_layer4_rs",
        "param": {
          "rs_ip": "127.0.0.1",
          "status": "executable",
          "weight": 10,
          "rs_port": [1001],
          "protocol": "TCP",
          "inst_type": "CVM",
          "region_id": "ap-nanjing",
          "user_remark": "",
          "cloud_clb_id": "lb-xxxxxx",
          "listener_port": [1001],
          "clb_vip_domain": "127.0.0.1",
          "validate_result": []
        },
        "state": "failed",
        "reason": "the listener already exists",
        "extension": {},
        "creator": "Jim",
        "reviser": "Jim",
        "created_at": "2023-02-12T14:47:39Z",
        "updated_at": "2023-02-12T14:55:40Z"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述    |
|---------|--------|---------|
| code    | int32  | 状态码   |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称    | 参数类型   | 描述             |
|---------|--------|----------------|
| count   | uint64 | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据        |

#### data.details[n]

| 参数名称            | 参数类型       | 描述                                                                                |
|--------------------|--------------|-------------------------------------------------------------------------------------|
| id                 | string       | 任务详情数据ID                                                                        |
| task_management_id | string       | 关联的任务管理数据的唯一标识                                                             |
| flow_id            | string       | 关联的异步任务flow的唯一标识                                                            |
| task_action_ids    | string array | 关联的异步任务task的aciton id数组                                                       |
| operation          | string       | 操作（创建4层监听器：create_layer4_listener、创建7层监听器：create_layer7_listener、绑定4层RS：binding_layer4_rs、绑定7层RS：binding_layer7_rs、创建URL规则：create_layer7_rule、解绑4层RS：listener_layer4_unbind_rs、解绑7层RS：listener_layer7_unbind_rs、调整4层权重：listener_layer4_rs_weight、调整7层权重：listener_layer7_rs_weight、删除监听器：listener_delete、开机：start_cvm、关机：stop_cvm、重启：reboot_cvm、重装：cvm_reset_system） |
| param              | json         | 任务详情数据                                                                          |
| result             | json         | 任务详情执行结果                                                                       |
| state              | string       | 任务状态，如：init（待执行）、running（运行）、failed（失败）、success（成功）、cancel（取消） |
| reason             | string       | 失败原因                                                                              |
| extension          | object       | 扩展字段                                                                              |
| creator            | string       | 创建者                                                                                |
| reviser            | string       | 修改者                                                                                |
| created_at         | string       | 创建时间，标准格式：2006-01-02T15:04:05Z                                                |
| updated_at         | string       | 修改时间，标准格式：2006-01-02T15:04:05Z                                                |
