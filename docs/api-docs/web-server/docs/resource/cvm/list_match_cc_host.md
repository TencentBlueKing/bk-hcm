### 描述

- 该接口提供版本：v1.8.0+。
- 该接口所需权限：资源分配。
- 该接口功能描述：查询分配cvm时匹配的cc主机。

### URL

POST /api/v1/cloud/cvms/assign/hosts/match/list

### 输入参数

| 参数名称       | 参数类型         | 必选 | 描述       |
|------------|--------------|----|----------|
| account_id | string       | 是  | 账号ID     |
| private_ipv4_addresses  | array string | 是  | 主机内网IP列表 |

### 调用示例

```json
{
  "account_id": "00000001",
  "private_ipv4_addresses": ["127.0.0.1"]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "bk_host_id": 1,
        "private_ipv4_addresses": ["127.0.0.1"],
        "public_ipv4_addresses": ["127.0.0.1"],
        "bk_cloud_id": 1,
        "bk_biz_id": 1,
        "region": "ap-guangzhou",
        "bk_host_name": "test",
        "bk_os_name": "linux",
        "create_time": "2019-04-28 17:55:11"
      }
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

| 参数名称    | 参数类型   | 描述             |
|---------|--------|----------------|
| details | array  | 查询返回的数据        |

#### data.details[n]

| 参数名称                   | 参数类型         | 描述       |
|------------------------|--------------|----------|
| bk_host_id             | int64        | cc主机唯一id |
| private_ipv4_addresses | string array | 内网IP     |
| public_ipv4_addresses  | string array | 公网IP     |
| bk_cloud_id            | int64        | 管控区域id   |
| bk_biz_id              | int64        | 业务ID     |
| region                 | string       | 地域       |
| bk_host_name           | string       | OS主机名称   |
| bk_os_name             | string       | 操作系统名称   |
| create_time            | string       | 配置平台录入时间 |
