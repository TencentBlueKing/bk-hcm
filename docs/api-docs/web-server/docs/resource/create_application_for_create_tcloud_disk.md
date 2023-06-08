### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：IaaS资源创建。
- 该接口功能描述：创建用于创建腾讯云云盘的申请。

### URL

POST /api/v1/cloud/vendors/tcloud/applications/types/create_disk

### 输入参数

| 参数名称                        | 参数类型   | 必选 | 描述    |
|-----------------------------|--------|----|-------|
| bk_biz_id                   | int64  | 是  | 业务ID  |
| account_id                  | string | 是  | 账号ID  |
| disk_name                   | string | 是  | 云盘名称  |
| region                      | string | 是  | 地域    |
| zone                        | string | 是  | 可用区   |
| disk_size                   | uint64 | 是  | 云盘大小  |
| disk_type                   | string | 是  | 云盘类型  |
| disk_count                  | uint32 | 是  | 云盘数量  |
| disk_charge_type            | string | 是  | 计费类型  |
| disk_charge_prepaid         | object | 否  | 预付费配置 |
| memo                        | string | 否  | 备注    |

#### TCloudDiskChargePrepaid
| 参数名称        | 参数类型      | 必选 | 描述                                                                                                                                                           |
|-------------|-----------|----|--------------------------------------------------------------------------------------------------------------------------------------------------------------|
| period      | uint64    | 否  | 购买云盘的时长，默认单位为月                                                                                                                                               |
| renew_flag  | string    | 否  | 自动续费标识（NOTIFY_AND_AUTO_RENEW：通知过期且自动续费，NOTIFY_AND_MANUAL_RENEW：通知过期不自动续费，DISABLE_NOTIFY_AND_MANUAL_RENEW：不通知过期不自动续费。默认取值：NOTIFY_AND_MANUAL_RENEW：通知过期不自动续费。） |

### 调用示例
```json
{
  "bk_biz_id": 100,
  "account_id": "0000001",
  "disk_name": "test",
  "region": "ap-hk",
  "zone": "ap-hk-1",
  "disk_size": 1,
  "disk_type": "ssd",
  "disk_count": 1,
  "disk_charge_type": "PREPAID",
  "disk_charge_prepaid": {
    "period": 6,
    "renew_flag": "NOTIFY_AND_AUTO_RENEW"
  },
  "memo": ""
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000001"
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

| 参数名称 | 参数类型   | 描述   |
|------|--------|------|
| id   | string | 单据ID |
