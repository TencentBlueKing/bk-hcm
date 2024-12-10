### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：主机查看权限。
- 该接口功能描述：查询主机下关联的安全组对应的安全组规则。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/cvms/{cvm_id}/security_groups/{security_group_id}/rules/list

### 输入参数

| 参数名称              | 参数类型   | 必选 | 描述     |
|-------------------|--------|----|--------|
| bk_biz_id         | int64  | 是  | 业务ID   |
| cvm_id            | string | 是  | 云主机ID  |
| security_group_id | string | 是  | 安全组ID  |
| filter            | object | 是  | 查询过滤条件 |
| page              | object | 是  | 分页设置   |

#### filter

| 参数名称  | 参数类型        | 必选  | 描述                                                              |
|-------|-------------|-----|-----------------------------------------------------------------|
| op    | enum string | 是   | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是   | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选 | 描述                                          |
|-------|-------------|----|---------------------------------------------|
| field | string      | 是  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| op    | enum string | 是  | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis） |
| value | 可变类型        | 是  | 查询条件Value值                                  |

#### page

| 参数名称  | 参数类型   | 必选  | 描述                                                                                                                                                  |
|-------|--------|-----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是   | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否   | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否   | 每页限制条数，最大500，不能为0                                                                                                                                   |

#### 查询参数介绍：

| 参数名称 | 参数类型   | 描述                        |
|------|--------|---------------------------|
| type | string | 规则类型。（枚举值：egress、ingress） |

### 调用示例

#### 获取详细信息请求参数示例

如查询主机为1、安全组ID为1的安全组出站规则列表。

/api/v1/cloud/bizs/{bk_biz_id}/cvms/1/security_groups/1/rules/list

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "type",
        "op": "eq",
        "value": "egress"
      }
    ]
  },
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
  }
}
```

#### 获取数量请求参数示例

如查询主机为1、安全组ID为1的安全组出站规则数量。

/api/v1/cloud/bizs/{bk_biz_id}/cvms/1/security_groups/1/rules/list

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "type",
        "op": "eq",
        "value": "egress"
      }
    ]
  },
  "page": {
    "count": true
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

| 参数名称    | 参数类型          | 描述             |
|---------|---------------|----------------|
| count   | uint64        | 当前规则能匹配到的总记录条数 |
| details | array  object | 查询返回的数据        |


#### data.details[n] & vendor=huawei

| 参数名称                    | 参数类型   | 描述                                                                                                                       |
|-------------------------|--------|--------------------------------------------------------------------------------------------------------------------------|
| id                      | string | 安全组规则ID                                                                                                                  |
| cloud_id                | string | 安全组规则云ID。                                                                                                                |
| protocol                | string | 协议类型取值范围：icmp、tcp、udp、icmpv6或IP协议号约束：为空表示支持所有协议协议为icmpv6时，网络类型应该为IPv6协议为icmp时，网络类型应该为IPv4。                               |
| ethertype               | string | IP地址协议类型取值范围：IPv4，IPv6约束：不填默认值为IPv4                                                                                      |
| cloud_remote_group_id   | string | 远端安全组ID，表示该安全组内的流量允许或拒绝取值范围：与remote_ip_prefix，remote_address_group_id功能互斥。                                               |
| remote_ip_prefix        | string | 远端IP地址，当type是egress时，为虚拟机访问端的地址。当type是ingress时，为访问虚拟机的地址取值范围：IP地址，或者cidr格式约束：与remote_group_id、remote_address_group_id互斥。 |
| remote_address_group_id | string | 远端地址组ID取值范围：租户下存在的地址组ID约束：与remote_ip_prefix，remote_group_id功能互斥。                                                         |
| port                    | string | 端口取值范围取值范围：支持单端口(80)，连续端口(1-30)以及不连续端口(22,3389,80)。                                                                      |
| priority                | uint32 | 功能说明：优先级取值范围：1~100，1代表最高优先级。                                                                                             |
| memo                    | string | 备注。                                                                                                                      |
| action                  | string | 安全组规则生效策略。取值范围：allow表示允许，deny表示拒绝。                                                                                       |
| type                    | string | 规则类型。（枚举值：egress、ingress）                                                                                                |
| cloud_security_group_id | string | 规则所属安全组云ID。                                                                                                              |
| cloud_project_id        | string | 安全组规则所属项目云ID。                                                                                                            |
| account_id              | string | 账号ID                                                                                                                     |
| security_group_id       | string | 规则所属安全组ID                                                                                                                |
| creator                 | string | 创建者                                                                                                                      |
| reviser                 | string | 最后一次修改的修改者                                                                                                               |
| created_at              | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                                                                           |
| updated_at              | string | 最后一次修改时间，标准格式：2006-01-02T15:04:05Z                                                                                       |


#### data.details[n] & vendor=azure

| 参数名称                                     | 参数类型   | 描述                                                                                                           |
|------------------------------------------|--------|--------------------------------------------------------------------------------------------------------------|
| id                                       | string | 安全组规则ID                                                                                                      |
| cloud_id                                 | string | 安全组规则云ID。                                                                                                    |
| etag                                     | string | 在更新资源时更改的唯一只读字符串。                                                                                            |
| name                                     | string | 资源组中唯一的资源名称。此名称可用于访问资源。                                                                                      |
| memo                                     | string | 备注。                                                                                                          |
| destination_address_prefix               | string | 目的地址前缀。CIDR或目标IP范围。星号‘*’也可用于匹配所有源IP。也可以使用‘VirtualNetwork’、‘AzureLoadBalancer’和‘Internet’等默认标签。               |
| destination_address_prefixes             | string | 目的地址带有前缀。CIDR或目标IP范围。                                                                                        |
| cloud_destination_app_security_group_ids | string | 目标应用安全组云ID列表。                                                                                                |
| destination_port_range                   | string | 目标端口或范围。介于0和65535之间的整数或范围。星号‘*’也可用于匹配所有端口。                                                                   |
| destination_port_ranges                  | string | 目的端口范围。                                                                                                      |
| protocol                                 | string | 网络协议。（枚举值：*、Ah、Esp、Icmp、Tcp、Udp）                                                                             |
| provisioning_state                       | string | 调度状态。（枚举值：Deleting、Failed、Succeeded、Updating）                                                                |
| source_address_prefix                    | string | CIDR或来源IP范围。星号‘*’也可用于匹配所有源IP。也可以使用‘VirtualNetwork’、‘AzureLoadBalancer’和‘Internet’等默认标签。如果这是入口规则，则指定网络流量源自何处。 |
| source_address_prefixes                  | string | CIDR或来源IP范围。                                                                                                 |
| cloud_source_app_security_group_ids      | string | 源安全组云ID列表。                                                                                                   |
| source_port_range                        | string | 源端口或范围。介于0和65535之间的整数或范围。星号‘*’也可用于匹配所有端口。                                                                    |
| source_port_ranges                       | string | 源端口范围。                                                                                                       |
| priority                                 | uint32 | 规则的优先级。该值可以介于100和4096之间。对于集合中的每个规则，优先级编号必须是唯一的。优先级数字越小，规则的优先级越高。                                             |
| type                                     | string | 规则类型。（枚举值：egress、ingress）                                                                                    |
| access                                   | string | 允许或拒绝网络流量。（枚举值：Allow、Deny）                                                                                   |
| cloud_security_group_id                  | string | 安全组规则所属安全组云ID。                                                                                               |
| account_id                               | string | 账号ID                                                                                                         |
| security_group_id                        | string | 规则所属安全组ID                                                                                                    |
| creator                                  | string | 创建者                                                                                                          |
| reviser                                  | string | 最后一次修改的修改者                                                                                                   |
| created_at                               | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                                                               |
| updated_at                               | string | 最后一次修改时间，标准格式：2006-01-02T15:04:05Z                                                                           |


#### data.details[n] & vendor=aws

| 参数名称                           | 参数类型   | 描述                                                                                                                               |
|--------------------------------|--------|----------------------------------------------------------------------------------------------------------------------------------|
| id                             | string | 安全组规则ID                                                                                                                          |
| cloud_id                       | string | 安全组规则云ID。                                                                                                                        |
| protocol                       | string | 协议, 取值: `tcp`, `udp`, `icmp`, `icmpv6`,用数字 `-1` 代表所有协议 。                                                                         |
| from_port                      | uint32 | 起始端口，与 to_port 配合使用。<br />port: 8080 (from_port: 8080, to_port: 8080) <br />port_range: 8080-9000(from_port: 8080, to_port:9000) |
| to_port                        | uint32 | 结束端口，与from_port配合使用。                                                                                                             |
| cloud_prefix_list_id           | string | 前缀列表的云ID。                                                                                                                        |
| ipv4_cidr                      | string | IPv4网段。                                                                                                                          |
| ipv6_cidr                      | string | IPv4网段。                                                                                                                          |
| cloud_target_security_group_id | string | 下一跳安全组实例云ID，例如：sg-ohuuioma。                                                                                                      |
| memo                           | string | 备注。                                                                                                                              |
| type                           | string | 规则类型。（枚举值：egress、ingress）                                                                                                        |
| cloud_security_group_id        | string | 规则所属安全组云ID。                                                                                                                      |
| cloud_group_owner_id           | string | 规则所属账号云ID。                                                                                                                       |
| account_id                     | string | 账号ID                                                                                                                             |
| security_group_id              | string | 规则所属安全组ID                                                                                                                        |
| creator                        | string | 创建者                                                                                                                              |
| reviser                        | string | 最后一次修改的修改者                                                                                                                       |
| created_at                     | string | 创建时间，标准格式：2006-01-02T15:04:05Z                                                                                                   |
| updated_at                     | string | 最后一次修改时间，标准格式：2006-01-02T15:04:05Z                                                                                               |

#### data.details[n] & vendor=tcloud

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
| created_at                     | string | 创建时间，标准格式：2006-01-02T15:04:05Z                              |
| updated_at                     | string | 最后一次修改时间，标准格式：2006-01-02T15:04:05Z                          |

