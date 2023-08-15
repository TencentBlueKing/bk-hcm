### 描述
- 该接口提供版本：v9.9.9+。
- 该接口所需权限：账号查看。
- 该接口功能描述：查询指定账号资源同步信息。

### URL
GET /api/v1/cloud/accounts/sync_details/{account_id}

### 输入参数
| 参数名称        | 参数类型   | 必选   | 描述    |
|-------------|--------|------|-------|
| account_id  | string | 是    | 账号ID  |

### 调用示例
```json
```

### 响应示例
```json
{
  "code": 0,
  "message": "",
  "data": {
    "iass_res": [ 
        {
          "res_name": "cvm",
          "res_status": "success",
          "res_failed_reason": "",
          "res_end_time": "2023-08-10T08:46:59Z"
        },
        {
          "res_name": "disk",
          "res_status": "success",
          "res_failed_reason": "",
          "res_end_time": "2023-08-10T08:46:59Z"
        },
        {
          "res_name": "security_group",
          "res_status": "success",
          "res_failed_reason": "",
          "res_end_time": "2023-08-10T08:46:59Z"
        },
        {
          "res_name": "gcp_firewall_rule",
          "res_status": "success",
          "res_failed_reason": "",
          "res_end_time": "2023-08-10T08:46:59Z"
        },
        {
          "res_name": "vpc",
          "res_status": "success",
          "res_failed_reason": "",
          "res_end_time": "2023-08-10T08:46:59Z"
        },
        {
          "res_name": "subnet",
          "res_status": "success",
          "res_failed_reason": "",
          "res_end_time": "2023-08-10T08:46:59Z"
        },
        {
          "res_name": "eip",
          "res_status": "failed",
          "res_failed_reason": "sync eip failed",
          "res_end_time": "2023-08-10T08:46:59Z"
        },
        {
          "res_name": "network_interface",
          "res_status": "not_sync",
          "res_failed_reason": "",
          "res_end_time": "2023-08-10T08:46:59Z"
        },
        {
          "res_name": "route_table",
          "res_status": "not_sync",
          "res_failed_reason": "",
          "res_end_time": "2023-08-10T08:46:59Z"
        }
    ],
  }
}
```
### 响应参数说明
| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data
| 参数名称       | 参数类型         | 描述                                                               |
|------------|--------------|------------------------------------------------------------------|
| iass_sync  | object       | iass层资源同步情况                                                        |

##### iass_sync[0]
| 参数名称                     | 参数类型     | 描述      |
|--------------------------|----------|-------------|
| res_name                 | string   | 资源标识         |
| res_status               | string   | 同步状态         |
| res_failed_reason        | string   | 同步失败原因      |
| res_end_time             | string   | 同步结束时间，标准格式：2006-01-02T15:04:05Z |