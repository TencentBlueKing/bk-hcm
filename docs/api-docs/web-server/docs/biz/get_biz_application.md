### 描述

- 该接口提供版本：v1.8.11+。
- 该接口所需权限：业务访问。
- 该接口功能描述：业务视角下查看申请单详情。

### URL

GET /api/v1/cloud/bizs/{bk_biz_id}/applications/{application_id}

### 输入参数

| 参数名称           | 参数类型   | 必选 | 描述    |
|----------------|--------|----|-------|
| bk_biz_id      | int64  | 是  | 业务ID  |
| application_id | string | 是  | 申请单ID |

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "000000e4",
    "source": "itsm",
    "sn": "REQ20260401000001",
    "type": "create_load_balancer",
    "operation": "create_load_balancer",
    "status": "pending",
    "applicant": "admin",
    "content": "{...}",
    "delivery_detail": "{...}",
    "memo": "申请负载均衡",
    "creator": "admin",
    "reviser": "admin",
    "created_at": "2026-04-01T10:00:00Z",
    "updated_at": "2026-04-01T10:30:00Z",
    "ticket_url": "https://itsm.example.com/ticket/xxx"
  }
}
```

#### 创建负载均衡独占集群响应示例

`content` 为 JSON 字符串，需先解析再读字段。当 `type` 为 `create_load_balancer`（腾讯云购买负载均衡）时，解析后的结构在现有购买字段基础上新增独占集群相关字段 `exclusive`、`cluster_tag`、`cloud_cluster_ids`、`clusters`（`clusters` 字段形态与负载均衡详情 `extension.clusters` 一致）。示例如下：

```json
{
  "vendor": "tcloud",
  "bk_biz_id": 2005000002,
  "account_id": "0000002a",
  "region": "ap-nanjing",
  "name": "xxxx",
  "load_balancer_type": "OPEN",
  "address_ip_version": "IPV4",
  "zones": [],
  "backup_zones": [],
  "cloud_vpc_id": "vpc-9xd34ofn",
  "cloud_subnet_id": null,
  "vip": "1.1.1.1",
  "cloud_eip_id": null,
  "vip_isp": "BGP",
  "internet_charge_type": null,
  "internet_max_bandwidth_out": null,
  "bandwidth_package_id": null,
  "bandwidthpkg_sub_type": null,
  "egress": null,
  "sla_type": "",
  "auto_renew": null,
  "require_count": 2,
  "load_balancer_pass_to_target": false,
  "memo": "",
  "exclusive": 1,
  "cluster_tag": "ziyan_chiji",
  "cloud_cluster_ids": [
    "tgw-38feq8c6"
  ],
  "clusters": [
    {
      "cloud_cluster_id": "tgw-38feq8c6",
      "cluster_id": "00000001",
      "cluster_name": "ziyan-l4-1",
      "cluster_tag": "ziyan_chiji",
      "cluster_type": "TGW"
    },
    {
      "cloud_cluster_id": "",
      "cluster_id": "",
      "cluster_name": "",
      "cluster_tag": "ziyan_chiji",
      "cluster_type": "STGW"
    }
  ]
}
```

##### content[create_load_balancer] 独占集群相关字段

其余字段与创建负载均衡申请接口入参一致，此处只说明新增字段。单据详情展示建议读 `exclusive` + `clusters`，与负载均衡详情页同一套渲染。

| 参数名称             | 参数类型         | 描述                                                                 |
|------------------|--------------|--------------------------------------------------------------------|
| exclusive        | int          | 是否独占型：1是、0否。与 `sla_type` 一起合成「实例规格」：`exclusive` 为 1 展示独占型；否则 `sla_type` 非空展示对应档位；否则展示共享型 |
| cluster_tag      | string       | 七层独占集群标签，购买入参                                                      |
| cloud_cluster_ids | string array | 四层（TGW）独占集群云上ID列表，购买入参。填一个表示指定集群；填多个表示在列表中随机挑选                 |
| clusters         | array        | 独占集群信息列表，字段与负载均衡详情 `extension.clusters` 一致。非独占型为空数组             |

##### content.clusters[n]

| 参数名称             | 参数类型   | 描述                          |
|------------------|--------|-----------------------------|
| cloud_cluster_id | string | 集群云上ID，如tgw-38feq8c6        |
| cluster_id       | string | 集群本地ID，未指定或本地表未同步到该集群时为空字符串 |
| cluster_name     | string | 集群名称，未指定或本地表未同步到该集群时为空字符串   |
| cluster_tag      | string | 集群标签                        |
| cluster_type     | string | 集群类型（枚举值：TGW、STGW），前端映射为四层/七层 |

说明：

- `exclusive` 为 0 或不传时 `clusters` 为空数组，`cluster_tag`、`cloud_cluster_ids` 为空。
- 七层（STGW）由云侧调度，提单时通常没有具体集群ID，该元素仅 `cluster_tag`、`cluster_type` 有值，其余字段为空字符串。
- 四层选择「随机分配」时，对应 TGW 元素的 `cloud_cluster_id`、`cluster_id`、`cluster_name` 为空字符串，`vip` 也为空；实际落地集群与 IP 需从负载均衡详情接口读取。

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称            | 参数类型   | 描述                                                                                           |
|-----------------|--------|----------------------------------------------------------------------------------------------|
| id              | string | 申请ID                                                                                         |
| source          | string | 来源（枚举值：itsm、bpaas）                                                                          |
| sn              | string | 序列号                                                                                          |
| type            | string | 申请类型（枚举值：add_account、create_cvm、create_vpc、create_disk、create_load_balancer）                 |
| operation       | string | 操作类型（枚举值：add_account、create_cvm、create_vpc、create_disk、create_load_balancer）                 |
| status          | string | 申请状态（枚举值：pending、pass、rejected、cancelled、delivering、completed、deliver_partial、deliver_error） |
| applicant       | string | 申请人                                                                                          |
| content         | string | 申请内容（已脱敏）                                                                                    |
| delivery_detail | string | 交付详情                                                                                         |
| memo            | string | 备注                                                                                           |
| creator         | string | 创建者                                                                                          |
| reviser         | string | 更新者                                                                                          |
| created_at      | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                                               |
| updated_at      | string | 更新时间，标准格式：2006-01-02T15:04:05Z                                                               |
| ticket_url      | string | ITSM审批链接                                                                                     |

### 错误码

| 错误码            | 描述                                   |
|----------------|--------------------------------------|
| RecordNotFound | 申请单不存在、用户无业务访问权限、或申请单不归属当前业务时均返回此错误 |
