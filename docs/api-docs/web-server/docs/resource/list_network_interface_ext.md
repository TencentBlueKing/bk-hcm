### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询网络接口列表(带扩展信息)。

### URL

POST /api/v1/cloud/vendors/{vendor}/network_interfaces/list

### 输入参数

| 参数名称     | 参数类型         | 必选  | 描述           |
|----------|--------------|-----|--------------|
| vendor   | string       | 是   | 云厂商          |
| filter   | object       | 是   | 查询过滤条件       |
| page     | object       | 是   | 分页设置         |

#### filter

| 参数名称  | 参数类型        | 必选  | 描述                                                      |
|-------|-------------|-----|-----------------------------------------------------------------|
| op    | enum string | 是   | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是   | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### filter.rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选  | 描述                                 |
|-------|-------------|-----|--------------------------------------------|
| field | string      | 是   | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍 |
| op    | enum string | 是   | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis）           |
| value | 可变类型     | 是   | 查询条件Value值                              |

##### rules 表达式说明：

##### 1. 操作符

| 操作符 | 描述                                        | 操作符的value支持的数据类型                             |
|-----|-------------------------------------------|----------------------------------------------|
| eq  | 等于。不能为空字符串                                | boolean, numeric, string                     |
| neq | 不等。不能为空字符串                                | boolean, numeric, string                     |
| gt  | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte | 大于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lt  | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lte | 小于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in  | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string                     |
| nin | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string                     |
| cs  | 模糊查询，区分大小写                                | string                                       |
| cis | 模糊查询，不区分大小写                               | string                                       |

##### 2. 协议示例

查询 name 是 "Jim" 且 age 大于18小于30 且 servers 类型是 "api" 或者是 "web" 的数据。

```json
{
  "op": "and",
  "rules": [
    {
      "field": "name",
      "op": "eq",
      "value": "Jim"
    },
    {
      "field": "age",
      "op": "gt",
      "value": 18
    },
    {
      "field": "age",
      "op": "lt",
      "value": 30
    },
    {
      "field": "servers",
      "op": "in",
      "value": [
        "api",
        "web"
      ]
    }
  ]
}
```

#### page

| 参数名称   | 参数类型    | 必选  | 描述                                                                                                                                               |
|--------|---------|-----|--------------------------------------------------------------------------------------------------------------------------------------------------|
| count	 | bool	   | 是	 | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据 detail，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count |
| start	 | uint32  | 否	 | 记录开始位置，start 起始值为0                                                                                                                               |
| limit	 | uint32  | 否	 | 每页限制条数，最大500，不能为0                                                                                                                                |
| sort	 | string  | 否	 | 排序字段，返回数据将按该字段进行排序                                                                                                                               |
| order	 | string  | 否	 | 排序顺序（枚举值：ASC、DESC）                                                                                                                               |

#### 查询参数介绍：

| 参数名称        | 参数类型   | 描述                             |
|-------------|--------|-------------------------------------- |
| id          | string | 主键ID                                |
| vendor      | string | 云厂商（枚举值：azure、huawei、gcp）      |
| name        | string | 网络接口名称                            |
| account_id  | string | 云资源的账号ID                          |
| region      | string | 地区ID                                 |
| zone        | string | 可用区                                 |
| vpc_id      | string | VPC的ID                               |
| cloud_vpc_id | string | 云VPC的ID                             |
| subnet_id   | string | 子网ID                                 |
| cloud_subnet_id | string | 云子网ID                           |
| private_ipv4  | string array | 内网IPv4                       |
| private_ipv6  | string array | 内网IPv6                       |
| public_ipv4   | string array | 公网IPv4                       |
| public_ipv6   | string array | 公网IPv6                       |
| bk_biz_id   | int     | 业务ID                                |
| instance_id | string  | 关联的实例ID                           |
| creator     | string | 创建者                                  |
| reviser     | string | 更新者                                  |
| created_at  | string | 创建时间，标准格式：2006-01-02T15:04:05Z    |
| updated_at  | string | 更新时间，标准格式：2006-01-02T15:04:05Z    |

接口调用者可以根据以上参数自行根据查询场景设置查询规则。

### 调用示例

#### 获取详细信息请求参数示例

如查询账号ID为"00000024"的Azure网络接口列表。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "account_id",
        "op": "eq",
        "value": "00000024"
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

如查询账号ID为"00000024"的Azure网络接口数量。

```json
{
  "filter": {
    "op": "and",
    "rules": [
      {
        "field": "account_id",
        "op": "eq",
        "value": "00000024"
      }
    ]
  },
  "page": {
    "count": true
  }
}
```

### Azure响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "detail": [
      {
        "id": "00000001",
        "vendor": "azure",
        "name": "resource_group_test",
        "account_id": "00000024",
        "region": "eastasia",
        "zone": "us-central1-a",
        "vpc_id": "000001",
        "cloud_vpc_id": "/subscriptions/1001-0000-0000-0000-111111111111/resourceGroups/rsg1001/providers/Microsoft.Network/virtualNetworks/test-vnet",
        "subnet_id": "000002",
        "cloud_subnet_id": "/subscriptions/1001-0000-0000-0000-111111111111/resourceGroups/rsg1001/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/default",
        "private_ipv4": ["127.0.0.1"],
        "private_ipv6": ["xx:xx:xx:xx:xx"],
        "public_ipv4": ["127.0.0.2"],
        "public_ipv6": ["xx:xx:xx:xx:xx"],
        "bk_biz_id": 10010,
        "instance_id": "1001-0000-0000-0000-xxxxxxxxx",
        "creator": "tom",
        "reviser": "tom",
        "created_at": "2019-07-29 11:57:20",
        "updated_at": "2019-07-29 11:57:20",
        "extension": {
            "resource_group_name": "test",
            "mac_address": "xx-xx-xx-xx-xx-xx",
            "enable_ip_forwarding": true,
            "enable_accelerated_networking": true,
            "dns_settings": {
              "dns_servers": [
                  "127.0.0.1",
                  "127.0.0.2"
              ],
              "applied_dns_servers": [
                  "127.0.0.1",
                  "127.0.0.2"
              ]
            },
            "cloud_gateway_load_balancer_id": "clb_001",
            "cloud_security_group_id": "/subscriptions/1001-0000-0000-0000-xxxxxxxxxxx/resourceGroups/rsg1001/providers/Microsoft.Network/networkSecurityGroups/test-nsg",
            "security_group_id": "000000xx",
            "ip_configurations": [
            {
                "cloud_id": "ip-001",
                "name": "ipconfig-001",
                "type": "",
                "properties": {
                    "primary": true,
                    "private_ip_address_version": "IPv4",
                    "private_ip_address": "127.0.0.x",
                    "private_ip_allocation_method": "Dynamic",
                    "public_ip_address": {
                        "cloud_id": "ip-id-1001",
                        "location": "eastasia",
                        "name": "ip-name-001",
                        "properties": {
                            "ip_address": "127.0.0.x",
                            "public_ip_address_version": "IPv4",
                            "public_ip_allocation_method": "Static"
                        }
                    },
                    "cloud_subnet_id": "/subscriptions/1001-0000-0000-0000-xxxxxxxxx/resourceGroups/rsg1001/providers/Microsoft.Network/subnet/test-sub"
                }
            }],
            "cloud_virtual_machine_id": "/subscriptions/1001-0000-0000-0000-xxxxxxxxx/resourceGroups/test_group/providers/Microsoft.Compute/virtualMachines/test001"
          }
      }
    ]
  }
}
```

### Gcp响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "detail": [
    {
        "id": "3",
        "vendor": "gcp",
        "name": "nic0",
        "account_id": "0000002d",
        "region": "eastasia",
        "zone": "us-central1-a",
        "vpc_id": "",
        "cloud_vpc_id": "https://www.googleapis.com/compute/v1/projects/tencentgcpieg6/global/networks/test001",
        "subnet_id": "",
        "cloud_subnet_id": "https://www.googleapis.com/compute/v1/projects/tencentgcpieg6/regions/us-west1/subnetworks/test-sub1",
        "private_ipv4": ["127.0.0.1"],
        "private_ipv6": ["xx:xx:xx:xx:xx"],
        "public_ipv4": ["127.0.0.2"],
        "public_ipv6": ["xx:xx:xx:xx:xx"],
        "bk_biz_id": 10010,
        "instance_id": "1000000001",
        "creator": "tom",
        "reviser": "tom",
        "created_at": "2019-07-29 12:57:20",
        "updated_at": "2019-07-29 12:57:20",
        "extension": {
            "can_ip_forward": false,
            "status": "RUNNING",
            "stack_type": "IPV4_ONLY",
            "access_configs":[
            {
              "type":"ONE_TO_ONE_NAT",
              "name":"External NAT",
              "nat_ip":"127.0.0.1",
              "network_tier":"PREMIUM"
            }
          ]
        }
    }]
  }
}
```

### HuaWei响应示例

#### 获取详细信息返回结果示例

```json
{
    "code": 0,
    "message": "ok",
    "data": {
      "detail": [
      {
          "id": "3",
          "vendor": "huawei",
          "name": "resource_group_test",
          "account_id": "0000002d",
          "region": "eastasia",
          "zone": "us-central1-a",
          "vpc_id": "000001",
          "cloud_vpc_id": "/subscriptions/1001-0000-0000-0000-xxxxxxxx/resourceGroups/rsgtest/providers/Microsoft.Network/virtualNetworks/test-vnet",
          "subnet_id": "000002",
          "cloud_subnet_id": "/subscriptions/1001-0000-0000-0000-xxxxxxxx/resourceGroups/rsgtest/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/default",
          "private_ipv4": ["127.0.0.1"],
          "private_ipv6": ["xx:xx:xx:xx:xx"],
          "public_ipv4": ["127.0.0.2"],
          "public_ipv6": ["xx:xx:xx:xx:xx"],
          "bk_biz_id": 10010,
          "instance_id": "1001-0000-0000-0000-xxxxxxxxx",
          "creator": "tom",
          "reviser": "tom",
          "created_at": "2019-07-29 12:57:20",
          "updated_at": "2019-07-29 12:57:20",
          "extension": {
              "port_state": "ACTIVE",
              "fixed_ips": [
              {
                "subnet_id": "1001-0000-0000-0000-xxxxxxx",
                "ip_address": "127.0.0.1"
              }
              ],
              "mac_addr": "xx:xx:xx:xx:xx:xx",
              "delete_on_termination": false,
              "driver_mode": "virtio",
              "min_rate": 100,
              "multiqueue_num": 8,
              "pci_address": "xxxxxxx",
              "ipv6": "1001:xxx:xxxx:xxxx:00",
              "virtual_ip_list": [{
                  "ip": "127.0.0.1",
                  "elasticity_ip": "127.0.0.2"
              }],
              "addresses": {
                  "bandwidth_id": "00000-0000-0000-0000-xxxxxxxxx",
                  "bandwidth_size": 1,
                  "bandwidth_type": "5_bgp"
              },
              "cloud_security_group_ids": ["1001","1002"]
          }
      }]
  }
}
```

#### 获取数量返回结果示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "count": 0
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

| 参数名称   | 参数类型   | 描述                                       |
|--------|--------|------------------------------------------|
| count  | uint64 | 当前规则能匹配到的总记录条数，仅在 count 查询参数设置为 true 时返回 |
| detail | array  | 查询返回的数据，仅在 count 查询参数设置为 false 时返回       |

#### data.detail[n]

| 参数名称        | 参数类型   | 描述                            |
|-------------|--------|--------------------------------------|
| id          | string | 主键ID                                |
| vendor      | string | 云厂商（枚举值：azure、huawei、gcp）      |
| name        | string | 网络接口名称                            |
| account_id  | string | 云资源的账号ID                          |
| region      | string | 地区ID                                 |
| zone        | string | 可用区                                 |
| vpc_id      | string | VPC的ID                               |
| cloud_vpc_id | string | 云VPC的ID                             |
| subnet_id   | string | 子网ID                                 |
| cloud_subnet_id | string | 云子网ID，格式：半角逗号分割           |
| private_ipv4  | string array | 内网IPv4                       |
| private_ipv6  | string array | 内网IPv6                       |
| public_ipv4   | string array | 公网IPv4                       |
| public_ipv6   | string array | 公网IPv6                       |
| bk_biz_id   | int     | 业务ID                                |
| instance_id | string  | 关联的实例ID                           |
| creator     | string | 创建者                                  |
| reviser     | string | 更新者                                  |
| created_at  | string | 创建时间，标准格式：2006-01-02T15:04:05Z    |
| updated_at  | string | 更新时间，标准格式：2006-01-02T15:04:05Z    |
| extension   | object | 云厂商私有结构                            |

#### data.extension(azure)

| 参数名称           | 参数类型         | 描述  |
|----------------|--------------|------------|
| resource_group_name | string  | 资源组名称   |
| mac_address    | string       | MAC地址     |
| enable_ip_forwarding | bool  | 是否允许IP转发(0:否1:是)  |
| enable_accelerated_networking | bool | 是否启用加速网络(0:否1:是) |
| dns_settings   | object array | DNS设置                    |
| cloud_gateway_load_balancer_id    | string | 网关负载均衡ID |
| cloud_security_group_id   | string | 云厂商网络安全组ID        |
| security_group_id   | string | 网络安全组ID                 |
| ip_configurations   | list array    | IP配置                 |
| cloud_virtual_machine_id | string | 虚拟机ID               |


#### data.extension(azure).dns_settings

| 参数名称     | 参数类型   | 描述                                     |
|----------|--------|-----------------------------------------------|
| dns_servers | string array | DNS服务器列表                          |
| applied_dns_servers | string array | 应用的DNS服务器列表             |


#### data.extension(azure).ip_configurations

| 参数名称     | 参数类型   | 描述                                     |
|----------|--------|-----------------------------------------------|
| cloud_id | string | IP配置ID                                      |
| name     | string | IP配置名称                                     |
| type     | string | 资源类型                                       |
| properties | object array | IP配置属性                             |

#### data.extension(azure).ip_configurations.properties（需要把primary==true的IP数据，显示到详情页面）

| 参数名称     | 参数类型   | 描述                                     |
|----------|--------|-----------------------------------------------|
| primary  | bool   | 类型(主要、辅助)                                 |
| private_ip_address_version | string | IP版本(IPv4、IPv6)             |
| private_ip_address | string | 专用IP地址                            |
| private_ip_allocation_method | string | IP分配(Dynamic、Static)     |
| public_ip_address  | object array | 公共IP地址                      |
| cloud_subnet_id    | string | 子网ID                               |


#### data.extension(azure).ip_configurations.public_ip_address

| 参数名称     | 参数类型   | 描述                                   |
|----------|--------|---------------------------------------------|
| cloud_id | string | 公共IP的ID                                   |
| location | string | 公共IP的地区                                  |
| name     | string | 公共IP地址名称                                |
| zone     | string array | 可用区列表                              |
| properties | object array | 公共IP配置属性                         |

#### data.extension(azure).ip_configurations.public_ip_address.properties

| 参数名称     | 参数类型   | 描述                                    |
|----------|--------|----------------------------------------------|
| ip_address | string | 公共IP地址                                   |
| public_ip_allocation_method | string | 公共IP分配(Dynamic、Static) |
| public_ip_address_version   | string | 公共IP版本(IPv4、IPv6)      |


#### data.extension(gcp)

| 参数名称  | 参数类型   | 描述                      |
|----------|-----------|--------------------------|
| can_ip_forward | bool       | 是否允许IP转发      |
| status         | string     | 状态(RUNNING)      |
| stack_type     | string     | 堆栈类型(IPV4_ONLY) |
| access_configs | list array | 公网IP列表          |


#### data.extension(gcp).access_configs

| 参数名称     | 参数类型   | 描述                        |
|----------|--------|----------------------------------|
| type         | string | 外网IP类型                    |
| name         | string | 外网IP名称                    |
| nat_ip       | string | 外网IP                       |
| network_tier | string | 网络层级(PREMIUM、STANDARD)   |

#### data.extension(huawei)

| 参数名称  | 参数类型   | 描述                      |
|----------|-----------|--------------------------|
| port_state  | string  | 网卡端口状态                  |
| fixed_ips   | list array | 网卡私网IP信息列表          |
| mac_addr    | string  | 网卡Mac地址信息                |
| delete_on_termination | bool | 卸载网卡时，是否删除网卡(true: 删除； false: 不删除)  |
| driver_mode | string  | 从guest os中，网卡的驱动类型。可选值为virtio和hinic，默认为virtio |
| min_rate    | int     | 网卡带宽下限                   |
| multiqueue_num | int  | 队列个数(取值范围为 1, 2, 4, 8, 16，28) |
| pci_address | string  | 弹性网卡在Linux GuestOS里的BDF号,网卡不支持时，返回为空 |
| ipv6        | string         | IpV6地址                 |
| virtual_ip_list | list array   | 虚拟IP地址数组           |
| addresses       | object array | 云服务器对应的弹性网卡信息 |
| cloud_security_group_ids | string array | 云服务器所属安全组列表 |


#### data.extension(huawei).fixed_ips

| 参数名称 | 参数类型 |         描述           |
|----------|-----------|--------------------|
| subnet_id  | string | 网卡私网IP对应子网信息 |
| ip_address | string | 网卡私网IP信息        |


#### data.extension(huawei).virtual_ip_list

| 参数名称 | 参数类型 |         描述           |
|----------|-----------|--------------------|
| ip            | string |  虚拟IP地址       |
| elasticity_ip | string |  弹性公网IP地址    |


#### data.extension(huawei).addresses

| 参数名称  | 参数类型 |         描述           |
|----------|-----------|--------------------|
| bandwidth_id   | string | 带宽ID        |
| bandwidth_size | string | 带宽大小       |
| bandwidth_type | string | 带宽类型，示例：5_bgp（全动态BGP） |

