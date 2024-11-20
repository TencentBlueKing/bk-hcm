# 四层监听器绑定RS 参数说明

**operation-type=binding_layer4_rs**

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
        "inst_type": "ENI",
        "rs_ip": "127.0.0.1",
        "rs_port": [80],
        "weight": 50,
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
| inst_type       | string   | RS类型，CVM/ENI                 |
| rs_ip           | string   | rs ip                        |
| rs_port         | []int    | rs 端口                        |
| weight          | int      | 监听器绑定的目标权重, 权重范围：0-100       |
| user_remark     | string   | 用户备注                         |
| validate_result | []string | 参数校验详情, 当状态为不可执行时, 会有具体的报错原因 |
| status          | string   | 校验结果状态                       |

##### status枚举

| 枚举值            | 描述   |
|----------------|------|
| executable     | 可执行  |
| not_executable | 不可执行 |
| existing       | 已存在  |


