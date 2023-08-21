### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务-配额管理。
- 该接口功能描述：同步业务配额。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/quotas/sync

### 输入参数

| 参数名称        | 参数类型         | 必选 | 描述    |
|-------------|--------------|----|-------|
| bk_biz_id   | int          | 是  | 业务ID  |
| account_id  | string       | 是  | 云账号ID |
| quota_infos | array object | 是  | 配额信息  |

#### quota_infos

| 参数名称       | 参数类型         | 必选 | 描述     |
|------------|--------------|----|--------|
| region     | string       | 是  | 地域     |
| zone       | string       | 是  | 可用区    |
| levels     | array object | 是  | 配额管理层级 |
| dimensions | array object | 是  | 配额管理维度 |
| memo       | string       | 否  | 备注     |

#### levels

| 参数名称  | 参数类型   | 描述                                             |
|-------|--------|------------------------------------------------|
| name  | string | 配额管理层级名称（枚举值：res_type：资源类型、instance_type：实例类型） |
| value | string | 配额管理层级值                                        |

#### dimension

| 参数名称        | 参数类型   | 描述                                              |
|-------------|--------|-------------------------------------------------|
| type        | string | 维度类型（枚举值：cvm_num：主机数、core_num：核数、mem_size：内存容量） |
| total_quota | int    | 总额度                                             |
| used_quota  | int    | 已用额度                                            |

### 调用示例

```json
{
  "bk_biz_id": 310,
  "account_id": "00000001",
  "quota_infos": [
    {
      "region": "ap-guangzhou",
      "zone": "ap-guangzhou-1",
      "levels": [
        {
          "name": "res_type",
          "value": "cvm"
        },
        {
          "name": "instance_type",
          "value": "s5.xxxx"
        }
      ],
      "dimensions": [
        {
          "type": "cvm_num",
          "total_quota": 100,
          "used_quota": 10
        }
      ],
      "memo": "cvm quota"
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "ids": [
      "00000003"
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 调用数据 |

#### data

| 参数名称 | 参数类型         | 描述       |
|------|--------------|----------|
| ids  | string array | 配额条目ID列表 |
