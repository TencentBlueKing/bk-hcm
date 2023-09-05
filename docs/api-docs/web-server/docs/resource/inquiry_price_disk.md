### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：IaaS资源创建。
- 该接口功能描述：查询硬盘售价。

### URL

POST /api/v1/cloud/disks/prices/inquiry

### 输入参数

**tcloud**
| 参数名称 | 参数类型 | 必选 | 描述 |
|-----------------------------|--------|----|-------|
| account_id | string | 是 | 账号ID |
| disk_name | string | 是 | 云盘名称 |
| region | string | 是 | 地域 |
| zone | string | 是 | 可用区 |
| disk_size | uint64 | 是 | 云盘大小 |
| disk_type | string | 是 | 云盘类型 |
| disk_count | uint32 | 是 | 云盘数量 |
| disk_charge_type | string | 是 | 计费类型 |
| disk_charge_prepaid | object | 否 | 预付费配置 |
| memo | string | 否 | 备注 |

**TCloudDiskChargePrepaid**
| 参数名称 | 参数类型 | 必选 | 描述 |
|-------------|-----------|----|--------------------------------------------------------------------------------------------------------------------------------------------------------------|
| period | uint64 | 否 | 购买云盘的时长，默认单位为月 |
| renew_flag | string | 否 |
自动续费标识（NOTIFY_AND_AUTO_RENEW：通知过期且自动续费，NOTIFY_AND_MANUAL_RENEW：通知过期不自动续费，DISABLE_NOTIFY_AND_MANUAL_RENEW：不通知过期不自动续费。默认取值：NOTIFY_AND_MANUAL_RENEW：通知过期不自动续费。） |

### 调用示例

**tcloud**

```json
{
  "account_id": "00000003",
  "disk_name": "20230704-test",
  "region": "ap-guangzhou",
  "zone": "ap-guangzhou-3",
  "disk_size": 50,
  "disk_type": "CLOUD_BSSD",
  "disk_count": 1,
  "disk_charge_type": "PREPAID",
  "disk_charge_prepaid": {
    "period": 6,
    "renew_flag": "NOTIFY_AND_AUTO_RENEW"
  },
  "memo": "test"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "discount_price": 0.03,
    "original_price": 0.13
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

| 参数名称           | 参数类型  | 描述     |
|----------------|-------|--------|
| discount_price | float | 折扣后的价格 |
| original_price | float | 原价     |
