### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：负载均衡操作。
- 该接口功能描述：更新腾讯云负载均衡云上属性。

### URL

PATCH /api/v1/cloud/bizs/{bk_biz_id}/vendors/tcloud/load_balancers/{id}

### 输入参数

| 参数名称                         | 参数类型    | 必选 | 描述                                                                                                      |
|------------------------------|---------|----|---------------------------------------------------------------------------------------------------------|
| bk_biz_id                    | int64   | 是  | 业务ID                                                                                                    |
| id                           | string  | 是  | 负载均衡ID                                                                                                  |
| name                         | string  | 否  | 名字                                                                                                      |
| internet_charge_type         | string  | 否  | 计费模式 TRAFFIC_POSTPAID_BY_HOUR 按流量按小时后计费 ; BANDWIDTH_POSTPAID_BY_HOUR 按带宽按小时后计费; BANDWIDTH_PACKAGE 带宽包计费 |
| internet_max_bandwidth_out   | int64   | 否  | 最大出带宽，单位Mbps                                                                                            |
| delete_protect               | boolean | 否  | 删除保护                                                                                                    |
| load_balancer_pass_to_target | boolean | 否  | Target是否放通来自CLB的流量。开启放通（true）：只验证CLB上的安全组；不开启放通（false）：需同时验证CLB和后端实例上的安全组。                              |
| memo                         | string  | 否  | 备注                                                                                                      |

接口调用者可以根据以上参数自行根据更新场景设置更新的字段，除了ID之外的更新字段至少需要填写一个。

### 调用示例

```json
{
  "memo": "default clb"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |