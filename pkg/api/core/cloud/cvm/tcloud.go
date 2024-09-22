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

// TCloudCvmExtension cvm extension.
type TCloudCvmExtension struct {
	Placement *TCloudPlacement `json:"placement,omitempty"`

	// InstanceChargeType 实例计费模式。(PREPAID：表示预付费，即包年包月、POSTPAID_BY_HOUR：表示后付费，即按量计费、
	// CDHPAID：专用宿主机付费，即只对专用宿主机计费，不对专用宿主机上的实例计费。、SPOTPAID：表示竞价实例付费。)
	InstanceChargeType  *string                    `json:"instance_charge_type,omitempty"`
	Cpu                 *int64                     `json:"cpu,omitempty"`
	Memory              *int64                     `json:"memory,omitempty"`
	CloudSystemDiskID   *string                    `json:"cloud_system_disk_id,omitempty"`
	CloudDataDiskIDs    []string                   `json:"cloud_data_disk_ids,omitempty"`
	InternetAccessible  *TCloudInternetAccessible  `json:"internet_accessible,omitempty"`
	VirtualPrivateCloud *TCloudVirtualPrivateCloud `json:"virtual_private_cloud,omitempty"`
	/*
		自动续费标识。注意：后付费模式本项为null。
		取值范围：
		- NOTIFY_AND_MANUAL_RENEW：表示通知即将过期，但不自动续费
		- NOTIFY_AND_AUTO_RENEW：表示通知即将过期，而且自动续费
		- DISABLE_NOTIFY_AND_MANUAL_RENEW：表示不通知即将过期，也不自动续费。
	*/
	RenewFlag             *string  `json:"renew_flag,omitempty"`
	CloudSecurityGroupIDs []string `json:"cloud_security_group_ids,omitempty"`

	/*
		实例的关机计费模式。
		取值范围：
		- KEEP_CHARGING：关机继续收费
		- STOP_CHARGING：关机停止收费
		- NOT_APPLICABLE：实例处于非关机状态或者不适用关机停止计费的条件
	*/
	StopChargingMode *string `json:"stop_charging_mode,omitempty"`

	UUID *string `json:"uuid,omitempty"`

	/*
		实例隔离类型。取值范围：
		- ARREAR：表示欠费隔离
		- EXPIRE：表示到期隔离
		- MANMADE：表示主动退还隔离
		- NOTISOLATED：表示未隔离
	*/
	IsolatedSource *string `json:"isolated_source,omitempty"`

	/*
		实例销毁保护标志，表示是否允许通过api接口删除实例。默认取值：FALSE。
		取值范围：
		- TRUE：表示开启实例保护，不允许通过api接口删除实例
		- FALSE：表示关闭实例保护，允许通过api接口删除实例
	*/
	DisableApiTermination *bool   `json:"disable_api_termination,omitempty"`
	BandwidthPackageID    *string `json:"bandwidth_package_id,omitempty"`
}

// TCloudPlacement 描述了实例的抽象位置，包括其所在的可用区，所属的项目，宿主机（仅专用宿主机产品可用），母机IP等。
type TCloudPlacement struct {
	// ProjectID 实例所属项目ID。
	CloudProjectID *int64 `json:"cloud_project_id,omitempty"`
}

// TCloudDataDisk 数据盘的信息
type TCloudDataDisk struct {
	/*
		数据盘类型。数据盘类型限制详见存储概述。默认取值：LOCAL_BASIC。
		取值范围：
		- LOCAL_BASIC：本地硬盘
		- LOCAL_SSD：本地SSD硬盘
		- LOCAL_NVME：本地NVME硬盘，与InstanceType强相关，不支持指定
		- LOCAL_PRO：本地HDD硬盘，与InstanceType强相关，不支持指定
		- CLOUD_BASIC：普通云硬盘
		- CLOUD_PREMIUM：高性能云硬盘
		- CLOUD_SSD：SSD云硬盘
		- CLOUD_HSSD：增强型SSD云硬盘
		- CLOUD_TSSD：极速型SSD云硬盘
		- CLOUD_BSSD：通用型SSD云硬盘
	*/
	DiskType *string `json:"disk_type,omitempty"`

	// DiskSizeGB 数据盘大小，单位：GB。
	DiskSizeGB *int64 `json:"disk_size_gb,omitempty"`

	// CloudDiskID 数据盘ID。LOCAL_BASIC 和 LOCAL_SSD 类型没有ID。
	CloudDiskID *string `json:"cloud_disk_id,omitempty"`

	// DeleteWithInstance 数据盘是否随子机销毁。取值范围：
	// TRUE：子机销毁时，销毁数据盘，只支持按小时后付费云盘
	// FALSE：子机销毁时，保留数据盘
	DeleteWithInstance *bool `json:"delete_with_instance,omitempty"`

	CloudSnapshotID *string `json:"cloud_snapshot_id,omitempty"`
	Encrypt         *bool   `json:"encrypt,omitempty"`
	CloudKmsKeyID   *string `json:"cloud_kms_key_id,omitempty"`
	// ThroughputPerformance 云硬盘性能，单位：MB/s
	ThroughputPerformance *int64 `json:"throughput_performance,omitempty"`

	// CloudCdcID 所属的独享集群ID。
	CloudCdcID *string `json:"cloud_cdc_id,omitempty"`
}

// TCloudInternetAccessible 描述了实例的公网可访问性，声明了实例的公网使用计费模式，最大带宽等
type TCloudInternetAccessible struct {
	/*
		网络计费类型。取值范围：
		- BANDWIDTH_PREPAID：预付费按带宽结算
		- TRAFFIC_POSTPAID_BY_HOUR：流量按小时后付费
		- BANDWIDTH_POSTPAID_BY_HOUR：带宽按小时后付费
		- BANDWIDTH_PACKAGE：带宽包用户
		默认取值：非带宽包用户默认与子机付费类型保持一致。
	*/
	InternetChargeType *string `json:"internet_charge_type,omitempty"`

	// InternetMaxBandwidthOut 公网出带宽上限，单位：Mbps。默认值：0Mbps。
	InternetMaxBandwidthOut *int64 `json:"internet_max_bandwidth_out,omitempty"`

	/*
		是否分配公网IP。取值范围：
		- TRUE：表示分配公网IP
		- FALSE：表示不分配公网IP
		当公网带宽大于0Mbps时，可自由选择开通与否，默认开通公网IP；当公网带宽为0，则不允许分配公网IP。该参数仅在RunInstances接口中作为入参使用。
	*/
	PublicIPAssigned *bool `json:"public_ip_assigned,omitempty"`
	// CloudBandwidthPackageID 带宽包ID。
	CloudBandwidthPackageID *string `json:"cloud_bandwidth_package_id,omitempty"`
}

// TCloudVirtualPrivateCloud 描述了网络信息等
type TCloudVirtualPrivateCloud struct {
	/*
		是否用作公网网关。公网网关只有在实例拥有公网IP以及处于私有网络下时才能正常使用。默认取值：FALSE。
		取值范围：
		- TRUE：表示用作公网网关
		- FALSE：表示不作为公网网关
	*/
	AsVpcGateway *bool `json:"as_vpc_gateway,omitempty"`
}
