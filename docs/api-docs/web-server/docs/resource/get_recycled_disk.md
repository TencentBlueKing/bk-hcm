### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：回收站查看。
- 该接口功能描述：查询回收站中的硬盘详情。

### URL

GET /api/v1/cloud/recycled/disks/{id}

#### 路径参数说明

| 参数名称 | 参数类型   | 必选  | 描述    |
|------|--------|-----|-------|
| id   | string | 是   | 云盘 ID |

### 调用示例

如查询云厂商是 tcloud , ID 是 00000002 的云盘信息

#### 返回参数示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000002",
    "vendor": "tcloud",
    "account_id": "abc",
    "name": "ab222c2222221",
    "cloud_id": "disk-123",
    "region": "abc",
    "zone": "abc",
    "disk_size": 500,
    "disk_type": "ssd",
    "memo": "abc",
    "creator": "james",
    "reviser": "james",
    "created_at": "2023-01-16T03:30:41Z",
    "updated_at": "2023-01-16T08:39:28Z",
    "extension": {
      "disk_charge_type": "PREPAID",
      "disk_charge_prepaid": {
        "period": 6,
        "renew_flag": "true"
      }
    }
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | Data   | 响应数据 |

#### Data

| 参数名称       | 参数类型                  | 描述                     |
|------------|-----------------------|------------------------|
| id         | string                | 云盘 ID                  |
| vendor     | string                | 云厂商                    |
| account_id | string                | 云账号 ID                 |
| name       | string                | 云盘名                    |
| bk_biz_id  | int                   | 分配给的cc 业务 ID， -1 表示未分配 |
| cloud_id   | string                | 云盘在云厂商上的 ID            |
| region     | string                | 地域                     |
| zone       | string                | 可用区                    |
| disk_size  | uint                  | 云盘大小                   |
| disk_type  | string                | 云盘类型                   |
| memo       | string                | 云盘备注                   | 
| creator    | string                | 创建者                    |
| reviser    | string                | 更新者                    |
| created_at | string                | 创建时间，标准格式：2006-01-02T15:04:05Z                   |
| updated_at | string                | 更新时间                   | 
| extension  | DiskExtension[vendor] | 各云厂商的差异化字段             | 

#### DiskExtension[tcloud]

| 参数名称                | 参数类型                    | 描述                                 |
|---------------------|-------------------------|------------------------------------|
| disk_charge_type    | string                  | 计费类型。范围[PREPAID, POSTPAID_BY_HOUR] |
| disk_charge_prepaid | TCloudDiskChargePrepaid | 预付费配置                              |

#### TCloudDiskChargePrepaid

| 参数名称       | 参数类型   | 描述                                                                                                                                                          |
|------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------------------------|
| period     | uint   | 购买云盘的时长，默认单位为月                                                                                                                                              |
| renew_flag | string | 自动续费标识。NOTIFY_AND_AUTO_RENEW：通知过期且自动续费，NOTIFY_AND_MANUAL_RENEW：通知过期不自动续费，DISABLE_NOTIFY_AND_MANUAL_RENEW：不通知过期不自动续费。默认取值：NOTIFY_AND_MANUAL_RENEW：通知过期不自动续费。 |

#### DiskExtension[azure]

| 参数名称                | 参数类型   | 描述   |
|---------------------|--------|------|
| resource_group_name | string | 资源组名 |

#### DiskExtension[huawei]

| 参数名称                | 参数类型                    | 描述                          |
|---------------------|-------------------------|-----------------------------|
| disk_charge_type    | string                  | 计费类型。可选值[prePaid, postPaid] |
| disk_charge_prepaid | HuaWeiDiskChargePrepaid | 预付费参数。                      |

#### HuaWeiDiskChargePrepaid

| 参数名称          | 参数类型   | 描述                                                                |
|---------------|--------|-------------------------------------------------------------------|
| period_num    | int    | 订购周期数，取值范围：period_type 为 month时，为[1-9]。period_type 为 year时，为[1-1] |
| period_type   | string | 订购周期单位                                                            |
| is_auto_renew | string | 是否自动续订                                                            |

#### DiskExtension[gcp]

暂时是空字典，后续会补充

#### DiskExtension [aws]

暂时是空字典，后续会补充
