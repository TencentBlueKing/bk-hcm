### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：获取腾讯云账号配额。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/tcloud/accounts/{account_id}/zones/quotas

### 请求参数

| 参数名称       | 参数类型   | 必选  | 描述   |
|------------|--------|-----|------|
| bk_biz_id  | int64  | 是   | 业务ID |
| account_id | string | 是   | 账号ID |
| vendor     | string | 是   | 供应商  |
| region     | string | 是   | 地域   |
| zone       | string | 是   | 可用区  |

### 调用示例

#### 请求参数示例

查询腾讯云账号配额。
```json
{
  "region": "ap-guangzhou",
  "zone": "ap-guangzhou-6"
}
```

#### 返回参数示例

查询腾讯云账号配额响应。
```json
{
  "code": 0,
  "message": "",
  "data": {
    "post_paid_quota_set": {
      "used_quota": 3,
      "remaining_quota": 497,
      "total_quota": 500
    },
    "pre_paid_quota": {
      "used_quota": 1,
      "once_quota": 500,
      "remaining_quota": 499,
      "total_quota": 500
    },
    "spot_paid_quota": {
      "used_quota": 0,
      "remaining_quota": 100,
      "total_quota": 100
    },
    "image_quota": {
      "used_quota": 8,
      "total_quota": 10
    },
    "disaster_recover_group_quota": {
      "group_quota": 1000,
      "current_num": 0,
      "cvm_in_host_group_quota": 30,
      "cvm_in_switch_group_quota": 50,
      "cvm_in_rack_group_quota": 20
    }
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data[tcloud]

| 参数名称                         | 参数类型                            | 描述                                         |
|------------------------------|---------------------------------|--------------------------------------------|
| post_paid_quota_set          | TCloudPostPaidQuota             | 后付费配额列表。     |
| pre_paid_quota               | TCloudPrePaidQuota              | 预付费配额列表。                                      |
| spot_paid_quota              | TCloudSpotPaidQuota             | spot配额列表。                                    |
| image_quota                  | TCloudImageQuota                | 镜像配额列表。                                   |
| disaster_recover_group_quota | TCloudDisasterRecoverGroupQuota | 置放群组配额列表。                                   |

#### TCloudPostPaidQuota

| 参数名称                  | 参数类型                 | 描述                                         |
|-----------------------|----------------------|--------------------------------------------|
| used_quota            | uint64               | 累计已使用配额。     |
| remaining_quota       | uint64               | 剩余配额。                                      |
| total_quota           | uint64               | 总配额。                                    |

#### TCloudPrePaidQuota

| 参数名称                  | 参数类型                 | 描述                                         |
|-----------------------|----------------------|--------------------------------------------|
| used_quota            | uint64               | 累计已使用配额。     |
| once_quota            | uint64               | 单次购买最大数量。     |
| remaining_quota       | uint64               | 剩余配额。                                      |
| total_quota           | uint64               | 总配额。                                    |

#### TCloudSpotPaidQuota

| 参数名称                  | 参数类型                 | 描述                                         |
|-----------------------|----------------------|--------------------------------------------|
| used_quota            | uint64               | 累计已使用配额。     |
| remaining_quota       | uint64               | 剩余配额。                                      |
| total_quota           | uint64               | 总配额。                                    |

#### TCloudImageQuota

| 参数名称                  | 参数类型                 | 描述                                         |
|-----------------------|----------------------|--------------------------------------------|
| used_quota            | uint64               | 累计已使用配额。     |
| total_quota           | uint64               | 总配额。                                    |

#### TCloudDisasterRecoverGroupQuota

| 参数名称                      | 参数类型                 | 描述                                        |
|---------------------------|----------------------|-------------------------------------------|
| group_quota               | int64               | 可创建置放群组数量的上限。     |
| current_num               | int64               | 当前用户已经创建的置放群组数量。                                     |
| cvm_in_host_group_quota   | int64               | 物理机类型容灾组内实例的配额数。                                   |
| cvm_in_switch_group_quota | int64               | 交换机类型容灾组内实例的配额数。                                   |
| cvm_in_rack_group_quota   | int64               | 机架类型容灾组内实例的配额数。                                   |
