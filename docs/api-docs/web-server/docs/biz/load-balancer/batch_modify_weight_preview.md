### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：。
- 该接口功能描述：批量绑定RS,excel导入预览。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/batch_operations/modify_weight/preview

### 输入参数

| 参数名称              | 参数类型   | 必选 | 描述              |
|-------------------|--------|----|-----------------|
| excel_file_base64 | string | 是  | excel文件base64编码 |

### 调用示例
```json
{
  "excel_file_base64": "XXXXXXXX"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "clb_id": "0000000d",
      "clb_name": "测试用-请勿删除1",
      "vip": "127.0.0.1",
      "ip_domain_type": "IPv4",
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
              "inst_type" : "CVM",
              "rsip": "127.0.0.1",
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

### 响应参数说明
| 参数名称       | 参数类型     | 描述   |
|------------|----------|------|
| code       | int32    | 状态码  |
| message    | string   | 请求信息 |
| data       | []object | 返回数据 |

#### data 字段说明
| 参数名称                | 参数类型   | 描述     |
|---------------------|--------|--------|
| clb_id              | string | 负载均衡ID |
| clb_name            | string | 负载均衡名称 |
| vip                 | string | 负载均衡IP |
| ip_domain_type      | string | IP类型   |
| update_weight_count | int    | 更新权重数量 |
| listeners           | object | 监听器列表  |

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
| inst_type  | string | 实例类型 |
| rsip       | string | RSIP |
| rsport     | int    | RS端口 |
| old_weight | int    | 原权重  |
| new_weight | int    | 新权重  |


### 错误响应示例
```json
{
    "code": 0,
    "message": "",
    "data": [
        {
            "reason": "127.0.0.1 TCP:8000 target 127.0.0.2:100 oldweight not match, input 50, actual 66",
            "ext": ""
        }
    ]
}
```
