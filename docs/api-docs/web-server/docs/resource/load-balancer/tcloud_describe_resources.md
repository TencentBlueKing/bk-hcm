### 描述

- 该接口提供版本：v1.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询用户在当前地域支持可用区列表和资源列表。腾讯云代理接口 DescribeResources

### URL

POST /api/v1/cloud/vendors/tcloud/load_balancers/resources/describe

### 输入参数

| 参数名称        | 参数类型         | 必选 | 描述                                   |
|-------------|--------------|----|--------------------------------------|
| account_id  | string       | 是  | 云账户id                                |
| region      | string       | 是  | 地域                                   |
| master_zone | string array | 否  | 指定可用区                                |
| ip_version  | string array | 否  | 指定IP版本，如"IPv4"、"IPv6"、"IPv6_Nat"     |
| isp         | string array | 否  | 指定运营商类型，如："BGP","CMCC","CUCC","CTCC" |
| limit       | int          | 否  | 返回可用区资源列表数目，默认20，最大值100。             |
| offset      | int          | 否  | 返回可用区资源列表起始偏移量，默认0。                  |

### 响应参数说明

| 参数名称    | 参数类型                      | 描述   |
|---------|---------------------------|------|
| code    | int32                     | 状态码  |
| message | string                    | 请求信息 |
| data    | DescribeResourcesResponse | 响应数据 |

#### DescribeResourcesResponse

| 参数名称            | 参数类型                  | 描述         |
|-----------------|-----------------------|------------|
| ZoneResourceSet | array of ZoneResource | 响应数据       |
| TotalCount      | int                   | 符合条件的总记录条数 |

#### ZoneResource

可用区资源

| 参数名称               | 参数类型              | 描述                                    |
|--------------------|-------------------|---------------------------------------|
| MasterZone         | string            | 主可用区                                  |
| SlaveZone          | string            | 备可用区                                  |
| ResourceSet	       | array of Resource | 资源列表                                  |
| IPVersion          | string            | ip版本（枚举值：IPv4，IPv6，IPv6_Nat）          |
| LocalZone          | bool              | 是否本地可用区                               |
| zone_resource_type | string            | 可用区资源的类型，SHARED表示共享资源，EXCLUSIVE表示独占资源 |
| ZoneRegion         | string            | 所属地域                                  |
| EdgeZone           | bool              | 可用区是否是EdgeZone可用区                     |
| Egress             | string            | 网络出口                                  |

#### Resource

资源

| 参数名称             | 参数类型                          | 描述                                                    |
|------------------|-------------------------------|-------------------------------------------------------|
| Isp              | string                        | 运营商信息，如"CMCC", "CUCC", "CTCC", "BGP", "INTERNAL"      |
| Type             | array of  string              | 运营商内具体资源信息，如"CMCC", "CUCC", "CTCC", "BGP", "INTERNAL" |
| AvailabilitySet	 | array of ResourceAvailability | 可用资源。                                                 |
| TypeSet	         | array of TypeInfo             | 运营商类型信息。                                              |

#### ResourceAvailability

资源可用性

| 参数名称         | 参数类型   | 描述                                                    |
|--------------|--------|-------------------------------------------------------|
| Availability | string | 资源可用性，"Available"：可用，"Unavailable"：不可用                |
| Type         | string | 运营商内具体资源信息，如"CMCC", "CUCC", "CTCC", "BGP", "INTERNAL" |

#### TypeInfo

运营商类型信息

| 参数名称                | 参数类型                      | 描述           |
|---------------------|---------------------------|--------------|
| SpecAvailabilitySet | array of SpecAvailability | 规格可用性        |
| Type                | string                    | 运营商类型，如"BGP" |

#### SpecAvailability

规格可用性

| 参数名称         | 参数类型   | 描述                                     |
|--------------|--------|----------------------------------------|
| Availability | string | 规格可用性，"Available"：可用，"Unavailable"：不可用 |
| SpecType     | string | 规格类型, 如 "shared"                       |


