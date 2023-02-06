### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询TCloud安全组规则列表。

### 输入参数

| 参数名称              | 参数类型   | 必选  | 描述    |
|-------------------|--------|-----|-------|
| security_group_id | string | 是   | 安全组ID |
| page              | object | 是   | 分页设置  |

#### page

| 参数名称  | 参数类型   | 必选  | 描述                                                                                                                                                  |
|-------|--------|-----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是   | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否   | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否   | 每页限制条数，最大500，不能为0                                                                                                                                   |

### 调用示例

#### 获取详细信息请求参数示例

如查询腾讯云安全组ID为1的安全组规则列表。

```json
{
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
  }
}
```

#### 获取数量请求参数示例

如查询腾讯云安全组ID为1的安全组规则数量。

```json
{
  "page": {
    "count": true
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

如查询腾讯云安全组ID为1的安全组规则列表响应示例。

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "details": [
      {
        "id": 1,
        "cloud_policy_index": 1,
        "version": "27",
        "protocol": "TCP",
        "port": "8080",
        "cloud_service_id": "",
        "cloud_service_group_id": "",
        "ipv4_cidr": "0.0.0.0/0",
        "ipv6_cidr": "",
        "cloud_target_security_group_id": "",
        "cloud_security_group_id": "sg-xxxxxx",
        "cloud_address_id": "",
        "cloud_address_group_id": "",
        "action": "ACCEPT",
        "memo": "security_group_rule",
        "type": "egress",
        "account_id": "1",
        "security_group_id": "1",
        "creator": "tom",
        "reviser": "tom",
        "created_at": "2019-07-29 11:57:20",
        "updated_at": "2019-07-29 11:57:20"
      }
    ]
  }
}
```

#### 获取数量返回结果示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "count": 1
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
| count   | uint64 | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据        |

#### data.details[n]

| 参数名称                           | 参数类型   | 描述                                                          |
|--------------------------------|--------|-------------------------------------------------------------|
| id                             | string | 安全组规则ID                                                     |
| cloud_policy_index             | uint64 | 安全组规则云索引号，值会随着安全组规则的变更动态变化                                  |
| version                        | string | 安全组规则当前版本。用户每次更新安全规则版本会自动加1。                                |
| protocol                       | string | 协议, 取值: TCP,UDP,ICMP,ICMPv6,ALL                             |
| port                           | string | 端口(all, 离散port, range)。 说明：如果Protocol设置为ALL，则Port也需要设置为all。 |
| cloud_service_id               | string | 协议端口云ID，例如：ppm-f5n1f8da。                                    |
| cloud_service_group_id         | string | 协议端口组云ID，例如：ppmg-f5n1f8da。                                  |
| ipv4_cidr                      | string | IPv4网段或IP(互斥)。                                              |
| ipv6_cidr                      | string | IPv4网段或IPv6(互斥)。                                            |
| cloud_target_security_group_id | string | 下一跳安全组实例云ID，例如：sg-ohuuioma。                                 |
| cloud_address_id               | string | IP地址云ID，例如：ipm-2uw6ujo6。                                    |
| cloud_address_group_id         | string | IP地址组云ID，例如：ipmg-2uw6ujo6。                                  |
| action                         | string | ACCEPT 或 DROP。                                              |
| memo                           | string | 备注。                                                         |
| type                           | string | 规则类型。（枚举值：egress、ingress）                                   |
| cloud_security_group_id        | string | 规则所属安全组云ID。                                                 |
| account_id                     | string | 账号ID                                                        |
| security_group_id              | string | 规则所属安全组ID                                                   |
| creator                        | string | 创建者                                                         |
| reviser                        | string | 最后一次修改的修改者                                                  |
| created_at                     | string | 创建时间                                                        |
| updated_at                     | string | 最后一次修改时间                                                    |
