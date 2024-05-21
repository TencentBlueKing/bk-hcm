### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：目标组创建。
- 该接口功能描述：业务下创建目标组。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/target_groups/create

### 输入参数

#### tcloud

| 参数名称          | 参数类型       | 必选 | 描述             |
|------------------|--------------|------|-----------------|
| bk_biz_id        | int          | 是   | 业务ID           |
| account_id       | string       | 是   | 账号ID           |
| name             | string       | 是   | 名称             |
| protocol         | string       | 是   | 协议             |
| port             | int          | 是   | 端口             |
| region           | string       | 是   | 地域             |
| cloud_vpc_id     | string array | 是   | 云端vpc的ID数组   |
| memo             | string       | 否   | 备注             |
| rs_list          | object array | 否   | RS列表           |

#### rs_list

| 参数名称          | 参数类型       | 必选 | 描述                               |
|------------------|--------------|------|-----------------------------------|
| inst_type        | string       | 是   | 实例类型(CVM:云服务器)               |
| cloud_inst_id    | string       | 是   | 云实例ID                           |
| port             | int          | 是   | 端口                               |
| weight           | string       | 是   | 权重                               |

### 调用示例

#### tcloud

```json
{
  "account_id": "0000001",
  "name": "xxx",
  "protocol": "TCP",
  "port": 22,
  "region": "ap-hk",
  "cloud_vpc_id": ["xxxx", "xxxx"]
  "memo": "",
  "rs_list": [
    {
      "inst_type": "CVM",
      "cloud_inst_id": "cvm-xxxxxx",
      "port": 8000,
      "weight": 10
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
    "id": "00000001"
  }
}
```

### 响应参数说明

| 参数名称  | 参数类型  | 描述    |
|---------|----------|---------|
| code    | int      | 状态码   |
| message | string   | 请求信息 |
| data    | object   | 响应数据 |

#### data

| 参数名称  | 参数类型 | 描述    |
|----------|--------|---------|
| id       | string | 目标组id |
