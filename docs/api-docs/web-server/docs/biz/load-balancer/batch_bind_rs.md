### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：目标组更新。
- 该接口功能描述：excel导入，批量绑定RS。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/batch_operations/bind_rs


### 输入参数

| 参数名称       | 参数类型     | 必选 | 描述     |
|------------|----------|----|--------|
| account_id | string   | 是  | 账号ID   |
| data       | []object | 是  | 修改权重数据 |

#### data 字段说明

| 参数名称         | 参数类型   | 必选     | 描述     |
|--------------|--------|--------|--------|
| clb_id       | string | 是      | 负载均衡ID |
| clb_name     | string | 是      | 负载均衡名称 |
| vip          | string | 是      | 负载均衡IP |
| new_rs_count | int    | 新增rs数量 |
| listeners    | object | 是      | 监听器列表  |

#### listeners 字段说明
| 参数名称            | 参数类型     | 描述                            |
|-----------------|----------|-------------------------------|
| action          | string   | 操作类型                          | 
| name            | string   | 监听器名称                         | 
| protocol        | string   | 协议                            | 
| ports           | []int    | 监听器端口，正常情况下只有一个端口，如果是端口段则有两个值 | 
| domain          | string   | 域名                            | 
| url             | string   | url                           | 
| rs_infos        | []RSInfo | RS的相关信息                       | 
| scheduler       | string   | 轮询规则                          | 
| session_expired | int      | 会话保持，0则不开启，单位：秒               |
| health_check    | boolean  | 健康检查                          |
| ca_cloud_id     | string   | 客户端证书id                       |
| cert_cloud_ids  | []string | 服务端证书id                       |

##### RSInfo 字段说明
| 参数名称      | 参数类型   | 描述   |
|-----------|--------|------|
| inst_type | string | 实例类型 |
| rsip      | string | RSIP |
| rsport    | int    | RS端口 |
| weight    | int    | 权重   |


### 调用示例
```json
{
  "account_id": "0000001",
  "data": [
    {
      "clb_id": "0000000d",
      "clb_name": "测试用-请勿删除1",
      "vip": "1.13.126.116",
      "new_rs_count": 1,
      "listeners": [
        {
          "action": "create_listener_and_append_rs",
          "name": "listener1",
          "protocol": "TCP",
          "ports": [
            7994,
            7995
          ],
          "domain": "",
          "url": "",
          "cert_cloud_ids": [],
          "ca_cloud_id": "",
          "rs_infos": [
            {
              "inst_type": "CVM",
              "rsip": "10.206.16.13",
              "rsport": 101,
              "weight": 50
            }
          ],
          "scheduler": "WRR",
          "session_expired": 0,
          "health_check": false
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
  "data": "0000000d"
}
```

### 响应参数说明
| 参数名称    | 参数类型   | 描述     |
|---------|--------|--------|
| code    | int32  | 状态码    |
| message | string | 请求信息   |
| data    | string | 批量操作ID |
