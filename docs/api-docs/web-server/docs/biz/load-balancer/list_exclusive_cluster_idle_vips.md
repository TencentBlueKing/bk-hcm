### 描述

- 该接口提供版本：v9.9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询指定四层独占集群内的闲置 VIP 列表，用于负载均衡购买页「独占型」指定 IP。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/load_balancers/exclusive_clusters/idle_vips/list

### 输入参数

| 参数名称             | 参数类型   | 必选 | 描述                                       |
|------------------|--------|----|------------------------------------------|
| bk_biz_id        | int64  | 是  | 业务ID，路径参数                                |
| account_id       | string | 是  | 账号ID                                     |
| region           | string | 是  | 地域                                       |
| cloud_cluster_id | string | 是  | 集群云上ID，取自标签聚合查询接口返回的 `cloud_cluster_id`   |

说明：

- 仅支持四层（TGW）集群；目标集群必须已分配给路径中的业务，否则返回 `PermissionDenied`。
- 本接口实时查询云上数据、不落库，返回结果可能随时被其它实例占用，创建负载均衡时后端会再复核一次。
- 集群选择「随机分配」（购买时不传 `cloud_cluster_ids`）时不需要调用本接口，IP 只能为「随机分配」。
- 集群改选后需清空已选 IP 并重新调用本接口。

### 调用示例

```json
{
  "account_id": "00000001",
  "region": "ap-guangzhou",
  "cloud_cluster_id": "tgw-xxxxxxxx"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 2,
    "details": [
      "1.1.1.1",
      "1.1.1.2"
    ]
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

| 参数名称    | 参数类型         | 描述                                  |
|---------|--------------|-------------------------------------|
| count   | uint64       | 闲置 VIP 总数                           |
| details | string array | 闲置的 IP 地址数组，元素在购买时作为 `vip` 参数传入     |
