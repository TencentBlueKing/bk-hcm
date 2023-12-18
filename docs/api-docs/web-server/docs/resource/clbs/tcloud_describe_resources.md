### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：x。
- 该接口功能描述：查询用户在当前地域支持可用区列表和资源列表。腾讯云代理接口 DescribeResources

### URL

POST /api/v1/cloud/vendors/tcloud/clbs/resources/describe

### 输入参数

| 参数名称   | 参数类型   | 必选 | 描述                                   |
|--------|--------|----|--------------------------------------|
| region | string | 是  | 地域                                   |
| zone   | string | 否  | 指定可用区                                |
| isp    | string | 否  | 指定运营商类型，如："BGP","CMCC","CUCC","CTCC" |

### 响应参数说明

| 参数名称    | 参数类型                  | 描述   |
|---------|-----------------------|------|
| code    | int32                 | 状态码  |
| message | string                | 请求信息 |
| data    | array of ZoneResource | 响应数据 |

#### ZoneResource

可用区资源

| 参数名称               | 参数类型              | 描述                                    |
|--------------------|-------------------|---------------------------------------|
| master_zone        | string            | 主可用区                                  |
| slave_zone         | string            | 备可用区                                  |
| resource_set	      | array of Resource | 资源列表                                  |
| ip_version         | string            | ip版本（枚举值：IPv4，IPv6，IPv6_Nat）          |
| local_zone         | bool              | 是否本地可用区                               |
| zone_resource_type | string            | 可用区资源的类型，SHARED表示共享资源，EXCLUSIVE表示独占资源 |
| zone_region        | string            | 所属地域                                  |
| edge_zone          | bool              | 可用区是否是EdgeZone可用区                     |
| egress             | string            | 网络出口                                  |

#### Resource

资源

| 参数名称               | 参数类型                          | 描述                                                    |
|--------------------|-------------------------------|-------------------------------------------------------|
| ip_version         | string                        | ip版本（枚举值：IPv4，IPv6，IPv6_Nat）                          |
| isp                | string                        | 运营商信息，如"CMCC", "CUCC", "CTCC", "BGP", "INTERNAL"      |
| type               | array of  string              | 运营商内具体资源信息，如"CMCC", "CUCC", "CTCC", "BGP", "INTERNAL" |
| local_zone         | bool                          | 是否本地可用区                                               |
| zone_resource_type | string                        | 可用区资源的类型，SHARED表示共享资源，EXCLUSIVE表示独占资源                 |
| zone_region        | string                        | 所属地域                                                  |
| edge_zone          | bool                          | 可用区是否是EdgeZone可用区                                     |
| egress             | string                        | 网络出口                                                  |
| availability_set	  | array of ResourceAvailability | 可用资源。                                                 |
| type_info	         | array of TypeInfo             | 运营商类型信息。                                              |

#### ResourceAvailability

资源可用性

| 参数名称         | 参数类型   | 描述                                                    |
|--------------|--------|-------------------------------------------------------|
| availability | string | 资源可用性，"Available"：可用，"Unavailable"：不可用                |
| type         | string | 运营商内具体资源信息，如"CMCC", "CUCC", "CTCC", "BGP", "INTERNAL" |

#### TypeInfo

运营商类型信息

| 参数名称                  | 参数类型                      | 描述           |
|-----------------------|---------------------------|--------------|
| spec_availability_set | array of SpecAvailability | 规格可用性        |
| type                  | string                    | 运营商类型，如"BGP" |

#### SpecAvailability

规格可用性

| 参数名称         | 参数类型   | 描述                                     |
|--------------|--------|----------------------------------------|
| availability | string | 规格可用性，"Available"：可用，"Unavailable"：不可用 |
| spec_type    | string | 规格类型, 如 "shared"                       |


