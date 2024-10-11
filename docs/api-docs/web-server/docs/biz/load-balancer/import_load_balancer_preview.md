### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：负载均衡excel导入，上传文件接口。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/load_balancers/operations/{operation_type}/preview

### 输入参数

| 参数名称           | 参数类型   | 必选      | 描述      |
|----------------|--------|---------|---------|
| bk_biz_id      | int    | 是       | 业务ID    |
| vendor         | string | 是       | 云厂商     |
| operation_type | string | 是       | 操作类型    |
| account_id     | string | 账户id    |
| region_ids     | array  | 云地域id列表 |
| file           | file   | 是       | excel文件 |

#### operation_type 说明

| 操作类型                   | 说明        |
|------------------------|-----------|
| create_layer4_listener | 创建四层监听器   |
| create_layer7_listener | 创建七层监听器   |
| create_layer7_rule     | 创建URL规则   |
| binding_layer4_rs      | 四层监听器绑定RS |
| binding_layer7_rs      | 七层监听器绑定RS |


### 调用示例
```multipart/form-data
{
  "account_id": "",
  "region_ids": ["",""],
  "file": "file.xlsx"
}
```

### 响应示例

#### operation-type=create_layer4_listener
[参数说明](import-response/create_layer4_listener_resp.md)

#### operation-type=create_layer7_listener
[参数说明](import-response/create_layer7_listener_resp.md)

#### operation-type=create_layer7_rule
[参数说明](import-response/create_layer7_rule_resp)

#### operation-type=layer4_listener_bind_rs
[参数说明](import-response/binding_layer4_rs_resp)

#### operation-type=layer7_listener_bind_rs
[参数说明](import-response/binding_layer7_rs_resp)
