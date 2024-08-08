### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：。
- 该接口功能描述：查询批量操作记录。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/batch_operations/{batch_operation_id}

### 输入参数

| 参数名称               | 参数类型   | 必选 | 描述     |
|--------------------|--------|----|--------|
| batch_operation_id | string | 是  | 批量操作ID |



### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "batch_operation_id": "00000008",
    "audit_id": 628,
    "preview": [
      {
        "clb_id": "0000000d",
        "clb_name": "测试用-请勿删除1",
        "listeners": [
          {
            "action": "create_listener_and_append_rs",
            "ca_cloud_id": "",
            "cert_cloud_ids": [],
            "domain": "",
            "health_check": false,
            "name": "listener1",
            "port": [
              8000
            ],
            "protocol": "TCP",
            "rs_infos": [
              {
                "inst_type": "CVM",
                "rsip": "127.0.0.2",
                "rsport": 100,
                "weight": 50
              }
            ],
            "scheduler": "WRR",
            "session_expired": 0,
            "url": ""
          }
        ],
        "new_rs_count": 4,
        "vip": "127.0.0.1"
      }
    ],
    "async_flows": [
      {
        "audit_id": "00000001",
        "flow_id": "00000001"
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
| data    | string | 响应结果 |

#### data字段说明

| 参数名称               | 参数类型   | 描述     |
|--------------------|--------|--------|
| batch_operation_id | string | 批量操作ID |
| audit_id           | int32  | 审计ID   |
| preview            | array  | 预览结果   |
| async_flows        | array  | 异步流程   |

preview字段解析详见 batch_bind_rs_preview.md 和 batch_modify_weight_preview.md

#### async_flows字段解析

| 参数名称     | 参数类型   | 描述     |
|----------|--------|--------|
| audit_id | string | 审计ID   |
| flow_id  | string | 异步流程ID |