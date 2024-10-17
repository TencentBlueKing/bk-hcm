# 创建URL规则 参数说明

**operation-type=create_layer7_rule**

## 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details":[
      {
        "clb_vip_domain": "127.0.0.1",
        "cloud_clb_id": "lb-xxxxxxx1",
        "protocol": "https",
        "listener_port": [8888],
        "domain": "www.tencent.com",
        "default_domain": true,
        "url_path": "/",
        "scheduler": "LEAST_CONN",
        "session": 60,
        "health_check": false,
        "user_remark": "this is a create listener item",
        "status": "executable",
        "validate_result": []
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
| data    | object | 响应结果 |


### data参数说明

| 参数名称       | 参数类型   | 描述        |
|------------|--------|-----------|
| details    | array  | excel导入详情 |


#### details[n] 参数说明

| 参数名称            | 参数类型     | 描述                           |
|-----------------|----------|------------------------------|
| clb_vip_domain  | string   | 监听器绑定的vip或域名                 |
| cloud_clb_id    | string   | 监听器绑定的clb id                 |
| protocol        | string   | 监听器协议                        |
| listener_port   | []int    | 监听器端口, 通常长度为1, 如果为端口段则长度为2   |
| domain          | string   | 监听器绑定的域名                     |
| default_domain  | bool     | 是否为默认域名                      |
| url_path        | string   | 监听器绑定的url路径                  |
| scheduler       | string   | 负载均衡算法                       |
| session         | int      | 会话保持时间                       |
| health_check    | bool     | 是否开启健康检查                     |
| user_remark     | string   | 用户备注                         |
| validate_result | []string | 参数校验详情, 当状态为不可执行时, 会有具体的报错原因 |
| status          | string   | 校验结果状态                       |

##### status枚举

| 枚举值            | 描述   |
|----------------|------|
| executable     | 可执行  |
| not_executable | 不可执行 |
| existing       | 已存在  |


