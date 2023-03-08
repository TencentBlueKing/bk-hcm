/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package cvm

// HuaWeiCvmExtension cvm extension.
type HuaWeiCvmExtension struct {
	// AliasName 弹性云服务器别名。
	AliasName string `json:"alias_name,omitempty"`
	// HypervisorHostname 弹性云服务器所在虚拟化主机名。
	HypervisorHostname    string        `json:"hypervisor_hostname,omitempty"`
	Flavor                *HuaWeiFlavor `json:"flavor,omitempty"`
	CloudSecurityGroupIDs []string      `json:"cloud_security_group_ids,omitempty"`
	CloudTenantID         string        `json:"cloud_tenant_id,omitempty"`

	// 扩展属性， diskConfig的类型。MANUAL，镜像空间不会扩展。AUTO，系统盘镜像空间会自动扩展为与flavor大小一致。
	DiskConfig *string `json:"disk_config,omitempty"`
	// PowerState 弹性云服务器电源状态。0：NOSTATE 1：RUNNING 4：SHUTDOWN
	PowerState int32 `json:"power_state,omitempty"`
	// ConfigDrive config drive信息。
	ConfigDrive        string          `json:"config_drive,omitempty"`
	Metadata           *HuaWeiMetadata `json:"metadata,omitempty"`
	CloudOSVolumeID    string          `json:"cloud_os_volume_id,omitempty"`
	CloudDataVolumeIDs []string        `json:"cloud_data_volume_ids,omitempty"`

	// RootDeviceName 弹性云服务器系统盘的设备名称，例如当系统盘的磁盘模式是VDB，为/dev/vda，磁盘模式是SCSI，为/dev/sda。
	RootDeviceName string `json:"root_device_name,omitempty"`
	// CloudEnterpriseProjectID 弹性云服务器所属的企业项目ID。
	CloudEnterpriseProjectID *string           `json:"cloud_enterprise_project_id,omitempty"`
	CpuOptions               *HuaWeiCpuOptions `json:"cpu_options,omitempty"`
}

// HuaWeiAddress 弹性云服务器的网络属性。
type HuaWeiAddress struct {
	// Version IP地址版本。“4”：代表IPv4。“6”：代表IPv6。
	Version string `json:"version,omitempty"`
	// Addr IP地址
	Addr string `json:"addr,omitempty"`
	// Type IP地址类型。fixed：代表私有IP地址。floating：代表浮动IP地址。
	Type string `json:"type,omitempty"`
	// MacAddr MAC地址。
	MacAddr string `json:"mac_addr,omitempty"`
	// CloudPortID IP地址对应的端口ID。
	CloudPortID string `json:"cloud_port_id,omitempty"`
}

// HuaWeiFlavor 弹性云服务器规格信息。
type HuaWeiFlavor struct {
	// CloudID 云服务器规格ID。
	CloudID string `json:"cloud_id,omitempty"`
	// Name 云服务器规格名称。
	Name string `json:"name,omitempty"`
	// Disk 该云服务器规格对应要求系统盘大小，0为不限制。
	Disk string `json:"disk,omitempty"`
	// VCpus 该云服务器规格对应的CPU核数。
	VCpus string `json:"vcpus,omitempty"`
	// Ram 该云服务器规格对应的内存大小，单位为MB。
	Ram string `json:"ram,omitempty"`
}

// HuaWeiMetadata 弹性云服务器元数据。
type HuaWeiMetadata struct {
	// ChargingMode 云服务器的计费类型。
	// “0”：按需计费（即postPaid-后付费方式）。
	// “1”：按包年包月计费（即prePaid-预付费方式）。
	// "2"：竞价实例计费
	ChargingMode string `json:"charging_mode,omitempty"`
	// CloudOrderID 按“包年/包月”计费的云服务器对应的订单ID。
	CloudOrderID string `json:"cloud_order_id,omitempty"`
	// CloudProductID 按“包年/包月”计费的云服务器对应的产品ID。
	CloudProductID string `json:"cloud_product_id,omitempty"`
	// EcmResStatus 云服务器的冻结状态。normal：云服务器正常状态（未被冻结）。freeze：云服务器被冻结。
	EcmResStatus string `json:"ecm_res_status,omitempty"`
	// ImageType 镜像类型，目前支持： 公共镜像（gold） 私有镜像（private） 共享镜像（shared）
	ImageType string `json:"image_type,omitempty"`
	// ResourceSpecCode 云服务器对应的资源规格。
	ResourceSpecCode string `json:"resource_spec_code,omitempty"`
	// ResourceType 云服务器对应的资源类型。取值为“1”，代表资源类型为云服务器。
	ResourceType string `json:"resource_type,omitempty"`
	// InstanceExtraInfo 系统内部虚拟机扩展信息。
	InstanceExtraInfo string `json:"instance_extra_info,omitempty"`
	// ImageName 云服务器操作系统对应的镜像名称。
	ImageName string `json:"image_name,omitempty"`
	// AgencyName 委托的名称。委托是由租户管理员在统一身份认证服务（Identity and Access Management，IAM）
	// 上创建的，可以为弹性云服务器提供访问云服务器的临时凭证。
	AgencyName string `json:"agency_name,omitempty"`
	// OSBit 操作系统位数，一般取值为“32”或者“64”。
	OSBit string `json:"os_bit,omitempty"`
	// SupportAgentList 云服务器是否支持企业主机安全、主机监控。
	// “hss”：企业主机安全
	// “ces”：主机监控
	SupportAgentList string `json:"support_agent_list,omitempty"`
}

// HuaWeiVolumesAttached 挂载到弹性云服务器上的磁盘。
type HuaWeiVolumesAttached struct {
	CloudID string `json:"cloud_id,omitempty"`
	// DeleteOnTermination 删除云服务器时是否一并删除该磁盘。
	// - true：是
	// - false：否
	// 微版本2.3及以上版本支持。
	DeleteOnTermination string `json:"delete_on_termination,omitempty"`
	// BootIndex 云硬盘启动顺序。 0为系统盘。非0为数据盘。
	BootIndex string `json:"boot_index,omitempty"`
}

// HuaWeiCpuOptions 自定义CPU选项。
type HuaWeiCpuOptions struct {
	// CpuThreads CPU超线程数， 决定CPU是否开启超线程。取值范围：1，2。
	// 1: 关闭超线程。
	// 2: 打开超线程。
	CpuThreads *int32 `json:"cpu_threads,omitempty"`
}
