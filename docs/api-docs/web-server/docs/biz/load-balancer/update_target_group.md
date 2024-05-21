### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：目标组更新。
- 该接口功能描述：业务下更新目标组基本信息。

### URL

PATCH /api/v1/cloud/bizs/{bk_biz_id}/target_groups/{id}

### 输入参数

#### tcloud

| 参数名称         | 参数类型         | 必选 | 描述         |
|--------------|--------------|----|------------|
| id           | string       | 是  | 目标组ID      |
| name         | string       | 是  | 名称         |
| protocol     | string       | 是  | 协议         |
| port         | int          | 是  | 端口         |
| region       | string       | 是  | 地域         |
| cloud_vpc_id | string array | 是  | 云端vpc的ID数组 |
| memo         | string       | 否  | 备注         |

### 调用示例

#### tcloud

```json
{
  "account_id": "0000001",
  "name": "xxx",
  "protocol": "TCP",
  "port": 22,
  "region": "ap-hk",
  "cloud_vpc_id": [
    "xxxx",
    "xxxx"
  ],
  "memo": ""
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
| code    | int    | 状态码  |
| message | string | 请求信息 |

