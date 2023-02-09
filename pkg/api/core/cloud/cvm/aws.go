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

// AwsCvmExtension cvm extension.
type AwsCvmExtension struct {
	BlockDeviceMapping []BlockDeviceMapping `json:"block_device_mapping,omitempty"`
	// ClientToken The idempotency token you provided when you launched the instance, if applicable.
	ClientToken *string        `json:"client_token,omitempty"`
	CpuOptions  *AwsCpuOptions `json:"cpu_options,omitempty"`
	// DnsName (IPv4 only) The public DNS name assigned to the instance. This name is not available until the
	// instance enters the running state. For EC2-VPC, this name is only available if you've enabled DNS hostnames
	// for your VPC.
	DnsName *string `json:"dns_name,omitempty"`
	// EbsOptimized Indicates whether the instance is optimized for Amazon EBS I/O. This optimization
	// provides dedicated throughput to Amazon EBS and an optimized configuration stack to provide
	// optimal I/O performance. This optimization isn't available with all instance types. Additional
	// usage charges apply when using an EBS Optimized instance.
	EbsOptimized          *bool                  `json:"ebs_optimized,omitempty"`
	CloudSecurityGroupIDs []string               `json:"cloud_security_group_ids,omitempty"`
	HibernationOptions    *AwsHibernationOptions `json:"hibernation_options,omitempty"`
	// Platform The value is Windows for Windows instances; otherwise blank.
	Platform *string `json:"platform,omitempty"`
	// PrivateDnsName (IPv4 only) The private DNS hostname name assigned to the instance.
	// This DNS hostname can only be used inside the Amazon EC2 network. This name is not
	// available until the instance enters the running state.
	// [EC2-VPC] The Amazon-provided DNS server resolves Amazon-provided private DNS hostnames
	// if you've enabled DNS resolution and DNS hostnames in your VPC. If you are not using
	// the Amazon-provided DNS server in your VPC, your custom domain name servers must resolve
	// the hostname as appropriate.
	PrivateDnsName        *string                   `json:"private_dns_name,omitempty"`
	PrivateDnsNameOptions *AwsPrivateDnsNameOptions `json:"private_dns_name_options,omitempty"`
	CloudRamDiskID        *string                   `json:"cloud_ram_disk_id,omitempty"`
	// RootDeviceName The device name of the root device volume (for example, /dev/sda1).
	RootDeviceName *string `json:"root_device_name,omitempty"`
	// RootDeviceType The root device type used by the AMI. The AMI can use an EBS volume or an instance store volume.
	// Valid Values: ebs | instance-store
	RootDeviceType *string `json:"root_device_type,omitempty"`
	// SourceDestCheck Indicates whether source/destination checking is enabled.
	SourceDestCheck *bool `json:"source_dest_check,omitempty"`
	// SriovNetSupport Specifies whether enhanced networking with the Intel 82599 Virtual Function interface is enabled.
	SriovNetSupport *string `json:"sriov_net_support,omitempty"`
	// VirtualizationType The virtualization type of the instance.
	VirtualizationType *string `json:"virtualization_type,omitempty"`
}

// BlockDeviceMapping Describes a block device mapping.
type BlockDeviceMapping struct {
	// DeviceName The device name (for example, /dev/sdh or xvdh).
	DeviceName *string `json:"device_name,omitempty"`
	// AttachTime The time stamp when the attachment initiated.
	AttachTime *string `json:"attach_time,omitempty"`
	// DeleteOnTermination Indicates whether the volume is deleted on instance termination.
	DeleteOnTermination *bool `json:"delete_on_termination,omitempty"`
	// Status The attachment state. (attaching | attached | detaching | detached)
	Status *string `json:"status,omitempty"`
	// CloudVolumeID The ID of the EBS volume.
	CloudVolumeID *string `json:"cloud_volume_id,omitempty"`
}

// AwsCpuOptions The CPU options for the instance.
type AwsCpuOptions struct {
	// CoreCount The number of CPU cores for the instance.
	CoreCount *int64 `json:"core_count,omitempty"`
	// ThreadsPerCore The number of threads per CPU core.
	ThreadsPerCore *int64 `json:"threads_per_core,omitempty"`
}

// AwsElasticGpuAssociation The Elastic GPU associated with the instance.
type AwsElasticGpuAssociation struct {
	// CloudElasticGpuAssociationID The ID of the association.
	CloudElasticGpuAssociationID *string `json:"cloud_elastic_gpu_association_id,omitempty"`
	// ElasticGpuAssociationState The state of the association between the instance and the Elastic
	// Graphics accelerator.
	ElasticGpuAssociationState *string `json:"elastic_gpu_association_state,omitempty"`
	// ElasticGpuAssociationTime The time the Elastic Graphics accelerator was associated with the instance.
	ElasticGpuAssociationTime *string `json:"elastic_gpu_association_time,omitempty"`
	// CloudElasticGpuID The ID of the Elastic Graphics accelerator.
	CloudElasticGpuID *string `json:"cloud_elastic_gpu_id,omitempty"`
}

// AwsElasticInferenceAcceleratorAssociationSet The elastic inference accelerator associated with the instance.
type AwsElasticInferenceAcceleratorAssociationSet struct {
	Arn             *string `json:"arn,omitempty"`
	CloudID         *string `json:"cloud_id,omitempty"`
	State           *string `json:"state,omitempty"`
	AssociationTime *string `json:"association_time,omitempty"`
}

// AwsEnclaveOptions Indicates whether the instance is enabled for AWS Nitro Enclaves.
type AwsEnclaveOptions struct {
	// Enabled If this parameter is set to true, the instance is enabled for AWS Nitro Enclaves;
	// otherwise, it is not enabled for AWS Nitro Enclaves.
	Enabled *bool `json:"enabled,omitempty"`
}

// AwsHibernationOptions Indicates whether the instance is enabled for hibernation.
type AwsHibernationOptions struct {
	Configured *bool `json:"configured,omitempty"`
}

// AwsIamInstanceProfile The IAM instance profile associated with the instance, if applicable.
type AwsIamInstanceProfile struct {
	// Arn The Amazon Resource Name (ARN) of the instance profile.
	Arn *string `json:"arn,omitempty"`
	// ID The ID of the instance profile.
	ID *string `json:"id,omitempty"`
}

// AwsLicenseConfiguration Describes a license configuration.
type AwsLicenseConfiguration struct {
	Arn *string `json:"arn,omitempty"`
}

// AwsInstanceNetworkInterfaceAssociation The association information for an Elastic IPv4 associated with
// the network interface.
type AwsInstanceNetworkInterfaceAssociation struct {
	// CarrierIP The carrier IP address associated with the network interface.
	CarrierIP *string `json:"carrier_ip,omitempty"`
	// CustomerOwnedIP The customer-owned IP address associated with the network interface.
	CustomerOwnedIP *string `json:"customer_owned_ip,omitempty"`
	// CloudIPOwnerID The ID of the owner of the Elastic IP address.
	CloudIPOwnerID *string `json:"cloud_ip_owner_id,omitempty"`
	// PublicDnsName The public DNS name.
	PublicDnsName *string `json:"public_dns_name,omitempty"`
	// PublicIP The public IP address or Elastic IP address bound to the network interface.
	PublicIP *string `json:"public_ip,omitempty"`
}

// AwsInstanceNetworkInterfaceAttachment The network interface attachment.
type AwsInstanceNetworkInterfaceAttachment struct {
	// CloudID The ID of the network interface attachment.
	CloudID *string `json:"cloud_id,omitempty"`
	// AttachTime The time stamp when the attachment initiated.
	AttachTime *string `json:"attach_time,omitempty"`
	// DeleteOnTermination Indicates whether the network interface is deleted when the instance is terminated.
	DeleteOnTermination *bool `json:"delete_on_termination,omitempty"`
	// DeviceIndex The index of the device on the instance for the network interface attachment.
	DeviceIndex *int64 `json:"device_index,omitempty"`
	// NetworkCardIndex The index of the network card.
	NetworkCardIndex *string `json:"network_card_index,omitempty"`
	// Status The attachment state.
	// Valid Values: attaching | attached | detaching | detached
	Status *string `json:"status,omitempty"`
}

// AwsInstanceIPv4Prefix Information about an IPv4 prefix.
type AwsInstanceIPv4Prefix struct {
	// IPv4Prefix One or more IPv4 prefixes assigned to the network interface.
	IPv4Prefix *string `json:"ipv4_prefix,omitempty"`
}

// AwsInstanceIPv6Addresses The IPv6 addresses associated with the network interface.
type AwsInstanceIPv6Addresses struct {
	IPv6Address *string `json:"ipv6_address,omitempty"`
}

// AwsInstanceIPv6Prefix The IPv6 delegated prefixes that are assigned to the network interface.
type AwsInstanceIPv6Prefix struct {
	// IPv6Prefix One or more IPv6 prefixes assigned to the network interface.
	IPv6Prefix *string `json:"ipv6_prefix,omitempty"`
}

// AwsInstancePrivateIpAddress Describes a private IPv4 address.
type AwsInstancePrivateIpAddress struct {
	Association *AwsInstanceNetworkInterfaceAssociation `json:"association,omitempty"`
	// Primary Indicates whether this IPv4 address is the primary private IP address of the network interface.
	Primary *bool `json:"primary,omitempty"`
	// PrivateDnsName The private IPv4 DNS name.
	PrivateDnsName *string `json:"private_dns_name,omitempty"`
	// PrivateIPAddress The private IPv4 address of the network interface.
	PrivateIPAddress *string `json:"private_ip_address,omitempty"`
}

// AwsPrivateDnsNameOptions Describes the options for instance hostnames.
type AwsPrivateDnsNameOptions struct {
	// EnableResourceNameDnsAAAARecord Indicates whether to respond to DNS queries for instance hostnames with
	// DNS AAAA records.
	EnableResourceNameDnsAAAARecord *string `json:"enable_resource_name_dns_aaaa_record,omitempty"`
	// EnableResourceNameDnsARecord Indicates whether to respond to DNS queries for instance hostnames
	// with DNS A records.
	EnableResourceNameDnsARecord *string `json:"enable_resource_name_dns_a_record,omitempty"`
	// HostnameType The type of hostname to assign to an instance.
	// Valid Values: ip-name | resource-name
	HostnameType *string `json:"hostname_type,omitempty"`
}

// AwsProductCode Describes a product code.
type AwsProductCode struct {
	ProductCode *string `json:"product_code,omitempty"`
	// Type Valid Values: devpay | marketplace
	Type *string `json:"type,omitempty"`
}
