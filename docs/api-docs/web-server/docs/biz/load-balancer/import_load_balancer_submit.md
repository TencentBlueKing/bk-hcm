### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：负载均衡excel导入，提交接口。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/load_balancers/operations/{operation_type}/submit

### 输入参数

| 参数名称           | 参数类型   | 必选 | 描述      |
|----------------|--------|----|---------|
| bk_biz_id      | int    | 是  | 业务ID    |
| operation_type | string | 是  | 操作类型    |
| data           | object | 是  | 批量导入的数据 |

#### operation_type 说明

| 操作类型                   | 说明        |
|------------------------|-----------|
| create_layer4_listener | 创建四层监听器   |
| create_layer7_listener | 创建七层监听器   |
| create_layer7_rule     | 创建URL规则   |
| binding_layer4_rs      | 四层监听器绑定RS |
| binding_layer7_rs      | 七层监听器绑定RS |


#### data参数解析


| 参数名称       | 参数类型   | 描述                                |
|------------|--------|-----------------------------------|
| account_id | string | 账户id                              |
| region_ids | array  | 云地域id列表                           |
| source     | string | 任务来源, 标准运维插件(sops)、excel导入(excel) |
| details    | object | excel导入详情                         |

不同的operation_type对应的details不同, 具体查看 [上传excel文件接口](import_load_balancer_preview)
将上传excel文件接口返回的data作为该接口的入参即可

### 调用示例
```json
{
  "account_id": "",
  "region_ids": ["",""],
  "source": "excel",
  "details":[
    {
      "clb_vip_domain": "127.0.0.1",
      "cloud_clb_id": "lb-xxxxxxx1",
      "protocol": "https",
      "listener_port": [8888],
      "domain": "www.tencent.com",
      "url_path": "/",
      "target_type": "ENI",
      "rs_ip": "127.0.0.1",
      "rs_port": 80,
      "weight": 50,
      "user_remark": "this is a create listener item",
      "status": "executable",
      "validate_result": "-"
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "task_management_id": "xxxxxx"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |


#### data参数说明

| 参数名称    | 参数类型   | 描述     |
|---------|--------|--------|
| task_id | string | 任务管理id |
