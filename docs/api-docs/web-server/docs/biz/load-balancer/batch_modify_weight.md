### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：目标组更新。
- 该接口功能描述：excel导入，批量更新权重。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/batch_operations/modify_weight

### 输入参数

| 参数名称       | 参数类型     | 必选 | 描述     |
|------------|----------|----|--------|
| account_id | string   | 是  | 账号ID   |
| data       | []object | 是  | 修改权重数据 |

#### data 字段说明

| 参数名称      | 参数类型   | 必选 | 描述     |
|-----------|--------|----|--------|
| clb_id    | string | 是  | 负载均衡ID |
| clb_name  | string | 是  | 负载均衡名称 |
| vip       | string | 是  | 负载均衡IP |
| listeners | object | 是  | 监听器列表  |

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


##### RSInfo 字段说明
| 参数名称       | 参数类型   | 描述   |
|------------|--------|------|
| rsip       | string | RSIP |
| rsport     | int    | RS端口 |
| old_weight | int    | 原权重  |
| new_weight | int    | 新权重  |


### 调用示例
```json
{
  "account_id": "0000001",
  "data": [
    {
      "clb_id": "0000000d",
      "clb_name": "测试用-请勿删除1",
      "vip": "127.0.0.1",
      "update_weight_count": 1,
      "listeners": [
        {
          "Action": "modify_rs_weight",
          "name": "listener2",
          "protocol": "TCP",
          "ports": [
            8000
          ],
          "domain": "",
          "url": "",
          "rs_infos": [
            {
              "rsip": "127.0.0.2",
              "rsport": 100,
              "old_weight": 66,
              "new_weight": 30
            }
          ]
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
