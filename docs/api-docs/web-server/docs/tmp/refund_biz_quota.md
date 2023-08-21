### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务-配额管理。
- 该接口功能描述：退还业务额度。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/quotas/refund

### 输入参数

| 参数名称       | 参数类型         | 必选 | 描述       |
|------------|--------------|----|----------|
| bk_biz_id  | int          | 是  | 业务ID     |
| id         | string       | 是  | 业务配额条目ID |
| dimensions | array object | 是  | 配额维度     |

#### dimensions

| 参数名称         | 参数类型   | 描述                                              |
|--------------|--------|-------------------------------------------------|
| type         | string | 维度类型（枚举值：cvm_num：主机数、core_num：核数、mem_size：内存容量） |
| refund_quota | int    | 申请额度数量                                          |

### 调用示例

```json
{
  "id": "00000001",
  "dimensions": [
    {
      "type": "cvm_num",
      "refund_quota": 100
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
    "sn": "NO2019090519542603"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称 | 参数类型   | 描述 |
|------|--------|----|
| sn   | string | 单号 |
