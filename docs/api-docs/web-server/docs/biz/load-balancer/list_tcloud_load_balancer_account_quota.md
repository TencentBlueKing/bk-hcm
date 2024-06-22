### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：获取腾讯云账号负载均衡的配额。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/tcloud/load_balancers/accounts/{account_id}/quotas

### 请求参数

| 参数名称    | 参数类型 | 必选  | 描述  |
|------------|--------|------|-------|
| bk_biz_id  | int    | 是   | 业务ID |
| account_id | string | 是   | 账号ID |
| region     | string | 是   | 地域   |

### 调用示例

#### 请求参数示例

查询腾讯云账号的负载均衡配额。
```json
{
  "region": "ap-guangzhou"
}
```

#### 返回参数示例

查询腾讯云账号配额响应。
```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "quota_id": "TOTAL_OPEN_CLB_QUOTA",
      "quota_current": null,
      "quota_limit": 10
    }
  ]
}
```

### 响应参数说明

| 参数名称 | 参数类型       | 描述    |
|---------|--------------|---------|
| code    | int          | 状态码   |
| message | string       | 请求信息 |
| data    | array object | 响应数据 |

#### data[tcloud]

| 参数名称        | 参数类型  | 描述                           |
|----------------|---------|--------------------------------|
| quota_id       | string  | 配额名称                        |
| quota_current  | int     | 当前使用数量，为 null 时表示无意义 |
| quota_limit    | int     | 配额数量                        |

#### 配额名称，取值范围：
 - TOTAL_OPEN_CLB_QUOTA：用户当前地域下的公网CLB配额
 - TOTAL_INTERNAL_CLB_QUOTA：用户当前地域下的内网CLB配额
 - TOTAL_LISTENER_QUOTA：一个CLB下的监听器配额
 - TOTAL_LISTENER_RULE_QUOTA：一个监听器下的转发规则配额
 - TOTAL_TARGET_BIND_QUOTA：一条转发规则下可绑定设备的配额
 - TOTAL_SNAT_IP_QUOTA： 一个CLB实例下跨地域2.0的SNAT IP配额
 - TOTAL_ISP_CLB_QUOTA：用户当前地域下的三网CLB配额
