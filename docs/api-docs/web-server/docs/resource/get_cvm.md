### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询虚拟机详情。

### URL

GET /api/v1/cloud/cvms/{id}

### 输入参数

| 参数名称 | 参数类型     | 必选 | 描述    |
|------|----------|----|-------|
| id   | string   | 是  | 虚拟机ID |

### 调用示例

#### 获取详细信息请求参数示例

```json
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "id": "00000001",
    "cloud_id": "cvm-123",
    "name": "cvm-test",
    "vendor": "tcloud",
    "bk_biz_id": -1,
    "bk_cloud_id": 100,
    "account_id": "0000001",
    "region": "ap-hk",
    "zone": "ap-hk-1",
    "cloud_vpc_ids": [
      "vpc-123"
    ],
    "cloud_subnet_ids": [
      "subnet-123"
    ],
    "cloud_image_id": "image-123",
    "os_name": "linux",
    "memo": "cvm test",
    "status": "init",
    "private_ipv4_addresses": [
      "127.0.0.1"
    ],
    "private_ipv6_addresses": [],
    "public_ipv4_addresses": [
      "127.0.0.2"
    ],
    "public_ipv6_addresses": [],
    "machine_type": "s5",
    "cloud_created_time": "2022-01-20",
    "cloud_launched_time": "2022-01-21",
    "cloud_expired_time": "2022-02-22",
    "creator": "Jim",
    "reviser": "Jim",
    "created_at": "2023-02-12T14:47:39Z",
    "updated_at": "2023-02-12T14:55:40Z"
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

| 参数名称                   | 参数类型           | 描述                                   |
|------------------------|----------------|--------------------------------------|
| id                     | uint64         | 资源ID                                 |
| cloud_id               | string         | 云资源ID                                |
| name                   | string         | 名称                                   |
| vendor                 | string         | 供应商（枚举值：tcloud、aws、azure、gcp、huawei） |
| bk_biz_id              | int64          | 业务ID                                 |
| bk_cloud_id            | int64          | 云区域ID                                |
| account_id             | string         | 账号ID                                 |
| region                 | string         | 地域                                   |
| zone                   | string         | 可用区                                  |
| cloud_vpc_ids          | string array   | 云VpcID列表                             |
| cloud_subnet_ids       | string array   | 云子网ID列表                              |
| cloud_image_id         | string         | 云镜像ID                                |
| os_name                | string         | 操作系统名称                               |
| memo                   | string         | 备注                                   |
| status                 | string         | 状态                                   |
| private_ipv4_addresses | string array   | 内网IPv4地址                             |
| private_ipv6_addresses | string array   | 内网IPv6地址                             |
| public_ipv4_addresses  | string array   | 公网IPv4地址                             |
| public_ipv6_addresses  | string array   | 公网IPv6地址                             |
| machine_type           | string         | 设备类型                                 |
| cloud_created_time     | string         | Cvm在云上创建时间，标准格式：2006-01-02T15:04:05Z                           |
| cloud_launched_time    | string         | Cvm启动时间，标准格式：2006-01-02T15:04:05Z                              |
| cloud_expired_time     | string         | Cvm过期时间，标准格式：2006-01-02T15:04:05Z                              |
| extension              | object[vendor] | 混合云差异字段                       |
| creator                | string         | 创建者                                  |
| reviser                | string         | 修改者                                  |
| created_at             | string         | 创建时间，标准格式：2006-01-02T15:04:05Z                                 |
| updated_at             | string         | 修改时间，标准格式：2006-01-02T15:04:05Z                                 |

#### extension[tcloud]

| 参数名称                       | 参数类型                        | 描述                                                                                                                                                                |
|----------------------------|-----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| placement                  | TCloudPlacement             | 位置信息。                                                                                                                                                             |
| instance_charge_type       | string                      | 实例计费模式。(PREPAID：表示预付费，即包年包月、POSTPAID_BY_HOUR：表示后付费，即按量计费、CDHPAID：专用宿主机付费，即只对专用宿主机计费，不对专用宿主机上的实例计费。、SPOTPAID：表示竞价实例付费。)。                                           |
| cpu                        | int64                       | Cpu。                                                                                                                                                              |
| memory                     | int64                       | 内存。                                                                                                                                                               |
| cloud_system_disk_id       | string                      | 云系统硬盘ID。                                                                                                                                                          |
| cloud_data_disk_ids        | string array                | 云数据盘ID。                                                                                                                                                           |
| internet_accessible        | TCloudInternetAccessible    | 描述了实例的公网可访问性，声明了实例的公网使用计费模式，最大带宽等。                                                                                                                                |
| virtual_private_cloud      | TCloudVirtualPrivateCloud   | 描述了网络信息等。                                                                                                                                                         |
| renew_flag                 | string                      | 自动续费标识。注意：后付费模式本项为null。取值范围：- NOTIFY_AND_MANUAL_RENEW：表示通知即将过期，但不自动续费 - NOTIFY_AND_AUTO_RENEW：表示通知即将过期，而且自动续费 - DISABLE_NOTIFY_AND_MANUAL_RENEW：表示不通知即将过期，也不自动续费。 |
| cloud_security_group_ids   | string array                | 云安全组ID。                                                                                                                                                           |
| stop_charging_mode         | string                      | 实例的关机计费模式。取值范围：- KEEP_CHARGING：关机继续收费- STOP_CHARGING：关机停止收费- NOT_APPLICABLE：实例处于非关机状态或者不适用关机停止计费的条件。                                                              |
| uuid                       | string                      | 云UUID。                                                                                                                                                            |
| isolated_source            | string                      | 实例隔离类型。取值范围：- ARREAR：表示欠费隔离XPIRE：表示到期隔离ANMADE：表示主动退还隔离OTISOLATED：表示未隔离。                                                                                                                                                            |
| disable_api_termination    | bool                        | 实例销毁保护标志，表示是否允许通过api接口删除实例。默认取值：FALSE。取值范围：- TRUE：表示开启实例保护，不允许通过api接口删除实例ALSE：表示关闭实例保护，允许通过api接口删除实例                                                                                                                                                            |

#### TCloudPlacement

| 参数名称             | 参数类型   | 描述     |
|------------------|--------|--------|
| cloud_project_id | string | 云项目id。 |

#### TCloudInternetAccessible

| 参数名称                       | 参数类型   | 描述        |
|----------------------------|--------|-----------|
| internet_charge_type       | string | 网络计费类型。取值范围：- BANDWIDTH_PREPAID：预付费按带宽结算RAFFIC_POSTPAID_BY_HOUR：流量按小时后付费ANDWIDTH_POSTPAID_BY_HOUR：带宽按小时后付费ANDWIDTH_PACKAGE：带宽包用户值：非带宽包用户默认与子机付费类型保持一致。 |
| internet_max_bandwidth_out | int64  | 公网出带宽上限，单位：Mbps。默认值：0Mbps。 |
| public_ip_assigned         | bool   | 是否分配公网IP。取值范围：- TRUE：表示分配公网IPALSE：表示不分配公网IP带宽大于0Mbps时，可自由选择开通与否，默认开通公网IP；当公网带宽为0，则不允许分配公网IP。该参数仅在RunInstances接口中作为入参使用。 |
| cloud_bandwidth_package_id | string | 带宽包ID。 |

#### TCloudInternetAccessible

| 参数名称             | 参数类型 | 描述        |
|------------------|------|-----------|
| as_vpc_gateway   | bool | 是否用作公网网关。公网网关只有在实例拥有公网IP以及处于私有网络下时才能正常使用。默认取值：FALSE。取值范围：- TRUE：表示用作公网网关ALSE：表示不作为公网网关。 |

#### extension[aws]

| 参数名称                     | 参数类型                        | 描述                      |
|--------------------------|-----------------------------|-------------------------|
| block_device_mapping     | AwsBlockDeviceMapping array | 硬盘相关信息。                 |
| cpu_options              | AwsCpuOptions               | Cpu选项。                  |
| ebs_optimized            | bool                        | 是否开启了 ebs 优化。           |
| cloud_security_group_ids | string array                | 云安全组ID列表。               |
| hibernation_options      | AwsHibernationOptions       | 实例休眠信息。                 |
| platform                 | string                      | 平台。                     |
| private_dns_name         | string                      | 私有DNS名称。                |
| private_dns_name_options | AwsPrivateDnsNameOptions    | 私有DNS名称选项。              |
| cloud_ram_disk_id        | string                      | 云内存硬盘ID。                |
| root_device_name         | string                      | 根设备硬盘的设备名称（例如 /dev/sda1）。 |
| root_device_type         | string                      | 根设备类型。                |
| source_dest_check        | bool                        | 指示是否启用源目标检查。                |
| sriov_net_support        | string                      | 指定是否启用与英特尔 82599 虚拟功能接口的增强网络。                |
| virtualization_type      | string                      | 实例的虚拟化类型。                |

#### AwsBlockDeviceMapping

| 参数名称            | 参数类型 | 描述     |
|-----------------|------|--------|
| status          | string | 状态。    |
| cloud_volume_id | string | 云硬盘ID。 |

#### AwsCpuOptions

| 参数名称             | 参数类型 | 描述     |
|------------------|------|--------|
| core_count       | int64 | CPU数量。 |
| threads_per_core | int64 | 线程数量。  |

#### AwsHibernationOptions

| 参数名称             | 参数类型 | 描述        |
|------------------|------|-----------|
| configured | bool | 是否开启休眠功能。 |

#### AwsPrivateDnsNameOptions

| 参数名称              | 参数类型 | 描述        |
|-------------------|------|-----------|
| carrier_ip        | string | 与网络接口关联的运营商 IP 地址。 |
| customer_owned_ip | string | 与网络接口关联的客户拥有的 IP 地址。 |
| cloud_ip_owner_id | string | 弹性 IP 地址所有者的 ID。 |
| public_dns_name   | string | 公共 DNS 名称。 |
| public_ip         | string | 绑定到网络接口的公有 IP 地址或弹性 IP 地址。 |

#### extension[huawei]

| 参数名称                        | 参数类型                  | 描述                                                                 |
|-----------------------------|-----------------------|--------------------------------------------------------------------|
| alias_name                  | string                | 弹性云服务器别名。                                                          |
| hypervisor_hostname         | string                | 弹性云服务器所在虚拟化主机名。                                                    |
| flavor                      | HuaWeiFlavor          | 弹性云服务器规格信息。                                                          |
| cloud_security_group_ids    | string array          | 云安全组ID。                                                            |
| cloud_tenant_id             | string                | 云租户ID。                                                             |
| disk_config                 | string                | 扩展属性， diskConfig的类型。MANUAL，镜像空间不会扩展。AUTO，系统盘镜像空间会自动扩展为与flavor大小一致。 |
| power_state                 | string                | 弹性云服务器电源状态。0：NOSTATE 1：RUNNING 4：SHUTDOWN                          |
| config_drive                | string                | config drive信息。                                                    |
| metadata                    | HuaWeiMetadata        | 弹性云服务器元数据。                                                          |
| volumes_attached            | HuaWeiVolumesAttached | 挂载到弹性云服务器上的磁盘。                                                          |
| root_device_name            | string                | 弹性云服务器系统盘的设备名称，例如当系统盘的磁盘模式是VDB，为/dev/vda，磁盘模式是SCSI，为/dev/sda。。     |
| cloud_enterprise_project_id | string                | 弹性云服务器所属的企业项目ID。                                                   |
| cpu_options                 | HuaWeiCpuOptions      | Cpu选项。                                                             |

#### HuaWeiFlavor

| 参数名称          | 参数类型 | 描述        |
|---------------|------|-----------|
| cloud_id      | string | 云服务器规格ID。 |
| name          | string | 云服务器规格名称。 |
| disk          | string | 该云服务器规格对应要求系统盘大小，0为不限制。 |
| vcpus         | string | 该云服务器规格对应的CPU核数。 |
| ram           | string | 该云服务器规格对应的内存大小，单位为MB。 |

#### HuaWeiMetadata

| 参数名称                | 参数类型 | 描述        |
|---------------------|------|-----------|
| charging_mode       | string | ChargingMode 云服务器的计费类型。“0”：按需计费（即postPaid-后付费方式）。“1”：按包年包月计费（即prePaid-预付费方式）。"2"：竞价实例计费 |
| cloud_order_id      | string | 按“包年/包月”计费的云服务器对应的订单ID。 |
| cloud_product_id    | string | 按“包年/包月”计费的云服务器对应的产品ID。 |
| ecm_res_status      | string | 云服务器的冻结状态。normal：云服务器正常状态（未被冻结）。freeze：云服务器被冻结。 |
| image_type          | string | 镜像类型，目前支持： 公共镜像（gold） 私有镜像（private） 共享镜像（shared） |
| resource_spec_code  | string | 云服务器对应的资源规格。 |
| resource_type       | string | 云服务器对应的资源类型。取值为“1”，代表资源类型为云服务器。 |
| instance_extra_info | string | 系统内部虚拟机扩展信息。 |
| image_name          | string | 云服务器操作系统对应的镜像名称。 |
| agency_name         | string | 委托的名称。委托是由租户管理员在统一身份认证服务（Identity and Access Management，IAM）上创建的，可以为弹性云服务器提供访问云服务器的临时凭证。 |
| os_bit              | string | 操作系统位数，一般取值为“32”或者“64”。 |
| os_type             | string | 操作系统类型，取值为：Linux、Windows。 |
| support_agent_list  | string | 云服务器是否支持企业主机安全、主机监控。“hss”：企业主机安全“ces”：主机监控 |

#### HuaWeiVolumesAttached

| 参数名称                  | 参数类型 | 描述                 |
|-----------------------|------|--------------------|
| cloud_id              | string | 云硬盘ID。             |
| delete_on_termination | string | 删除云服务器时是否一并删除该磁盘。- true：是- false。 |
| boot_index            | string | 云硬盘启动顺序。 0为系统盘。非0为数据盘。 |

#### HuaWeiCpuOptions

| 参数名称              | 参数类型 | 描述        |
|-------------------|------|-----------|
| cpu_threads        | int64 | CPU超线程数， 决定CPU是否开启超线程。取值范围：1，2。1: 关闭超线程。2: 打开超线程。 |

#### extension[azure]

| 参数名称                        | 参数类型                        | 描述                   |
|-----------------------------|-----------------------------|----------------------|
| resource_group_name         | string                      | 资源组名称。               |
| additional_capabilities     | AzureAdditionalCapabilities | 启用或禁用的其他功能。。         |
| billing_profile             | AzureBillingProfile         | 计费相关详细信息。            |
| eviction_policy             | string                      | 规模集的逐出策略。(Deallocate | Delete)    |
| hardware_profile            | AzureHardwareProfile        | 硬件设置。                |
| license_type                | string                      | 许可证类型。               |
| cloud_network_interface_ids | string                      | 云网络接口ID。             |
| priority                    | string                      | 优先级。                 |
| storage_profile             | AzureStorageProfile         | 存储配置。                |
| zones                       | string array                | 可用区。                 |

#### AzureAdditionalCapabilities

| 参数名称                        | 参数类型   | 描述        |
|-----------------------------|--------|-----------|
| hibernation_enabled         | bool   | 是否开启休眠。   |
| ultra_ssd_enabled           | bool   | 启用超级固态硬盘。 |

#### AzureBillingProfile

| 参数名称                        | 参数类型                        | 描述    |
|-----------------------------|-----------------------------|-------|
| max_price         | int64                      | 最高价格。 |

#### AzureHardwareProfile

| 参数名称               | 参数类型                  | 描述       |
|--------------------|-----------------------|----------|
| vm_size            | string                |  虚拟机大小。  |
| vm_size_properties | AzureVmSizeProperties | 虚拟机大小属性。 |

#### AzureVmSizeProperties

| 参数名称             | 参数类型   | 描述       |
|------------------|--------|----------|
| vcpus_available  | int64 | 可用于 VM 的 vCPU 数。   |
| vcpus_per_core   | int64 | vCPU 与物理核心的比率。 |

#### AzureStorageProfile

| 参数名称                       | 参数类型         | 描述        |
|----------------------------|--------------|-----------|
| cloud_data_disk_ids        | string array | 云数据盘ID列表。 |
| cloud_os_disk_id           | string       | 云操作系统盘ID。 |

#### extension[gcp]

| 参数名称                        | 参数类型                       | 描述                       |
|-----------------------------|----------------------------|--------------------------|
| vpc_self_links              | string array               | Vpc资源URL。                |
| subnet_self_links           | string array               | Subnet资源URL。             |
| deletion_protection         | bool                       | 是否应保护从此机器映像创建的实例不被删除。    |
| can_ip_forward              | bool                       | 是否开启IP转发。                |
| cloud_network_interface_ids | string                     | 云网络接口ID。                 |
| disks                       | GcpAttachedDisk            | 项目id，默认0。                |
| self_link                   | string                     | gcp self_link。           |
| cpu_platform                | string                     | Cpu平台。                   |
| labels                      | map[string]string          | 实例标签。                    |
| min_cpu_platform            | string                     | 实例的最低 CPU 平台。。           |
| start_restricted            | bool                       | VM 是否因计算引擎检测到可疑活动而限制启动。。 |
| resource_policies           | string array               | 应用于此实例的资源策略。             |
| reservation_affinity        | GcpReservationAffinity     | 实例可以使用的预留。               |
| fingerprint                 | string                     | 指纹。                      |
| advanced_machine_features   | GcpAdvancedMachineFeatures | 高级计算机功能的选项。              |

#### GcpAttachedDisk

| 参数名称      | 参数类型   | 描述                                                          |
|-----------|--------|-------------------------------------------------------------|
| boot      | bool   | 是否是启动盘。                                                     |
| index     | int64  | 此磁盘的从零开始的索引，其中 0 保留用于启动磁盘。如果您将多个磁盘附加到一个实例，则每个磁盘都有一个唯一的索引号。。 |
| cloud_id  | string | 云硬盘ID。                                                      |
| self_link | string | 云硬盘URL。                                                     |

#### GcpReservationAffinity

| 参数名称                     | 参数类型         | 描述        |
|--------------------------|--------------|-----------|
| consume_reservation_type | string       | 可以使用资源的预留类型。 |
| key                      | string       | 对应于预留资源的标签键。 |
| values                   | string array | 对应于预留资源的标签值。 |

#### GcpAdvancedMachineFeatures

| 参数名称                         | 参数类型 | 描述        |
|------------------------------|------|-----------|
| enable_nested_virtualization | bool | 是否启用嵌套虚拟化（默认值为 false）。 |
| enable_uefi_networking       | bool | 是否为实例创建启用 UEFI 网络。 |
| threads_per_core             | bool | 每个物理内核的线程数。要禁用同时多线程 （SMT），请将此项设置为 1。如果未设置，则假定基础处理器每个内核支持的最大线程数。 |
| visible_core_count           | bool | 要向实例公开的物理内核数。乘以每个内核的线程数，计算要向实例公开的虚拟 CPU 总数。如果未设置，则根据实例的标称 CPU 计数和底层平台的 SMT 宽度推断内核数。 |
