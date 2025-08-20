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

// AzureCvmExtension cvm extension.
type AzureCvmExtension struct {
	ResourceGroupName      string                       `json:"resource_group_name,omitempty"`
	AdditionalCapabilities *AzureAdditionalCapabilities `json:"additional_capabilities,omitempty"`
	BillingProfile         *AzureBillingProfile         `json:"billing_profile,omitempty"`
	// EvictionPolicy Specifies the eviction policy for the Azure Spot virtual machine and Azure Spot scale set.
	// (Deallocate | Delete)
	EvictionPolicy  *string               `json:"eviction_policy,omitempty"`
	HardwareProfile *AzureHardwareProfile `json:"hardware_profile,omitempty"`

	// Possible values for Windows Server operating system are:
	// - Windows_Client
	// - Windows_Server
	// Possible values for Linux Server operating system are:
	// - RHEL_BYOS (for RHEL)
	// - SLES_BYOS (for SUSE)
	LicenseType              *string              `json:"license_type,omitempty"`
	CloudNetworkInterfaceIDs []string             `json:"cloud_network_interface_ids,omitempty"`
	Priority                 *string              `json:"priority,omitempty"`
	StorageProfile           *AzureStorageProfile `json:"storage_profile,omitempty"`
	Zones                    []string             `json:"zones,omitempty"`
}

// AzureStorageProfile ...
// https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/list?tabs=HTTP#storageprofile
type AzureStorageProfile struct {
	CloudDataDiskIDs []string `json:"cloud_data_disk_ids,omitempty"`
	CloudOsDiskID    string   `json:"cloud_os_disk_id,omitempty"`
}

// AzureKeyVaultKeyReference ...
// https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/list?tabs=HTTP#keyvaultkeyreference
type AzureKeyVaultKeyReference struct {
	KeyUrl      *string           `json:"key_url,omitempty"`
	SourceVault *AzureSubResource `json:"source_vault,omitempty"`
}

// AzureKeyVaultSecretReference ...
// https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/list?tabs=HTTP#keyvaultsecretreference
type AzureKeyVaultSecretReference struct {
	SecretUrl   *string           `json:"secret_url,omitempty"`
	SourceVault *AzureSubResource `json:"source_vault,omitempty"`
}

// AzureDiffDiskSettings ...
// https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/list?tabs=HTTP#diffdisksettings
type AzureDiffDiskSettings struct {
	Option    *string `json:"option,omitempty"`
	Placement *string `json:"placement,omitempty"`
}

// AzureAdditionalCapabilities Specifies additional capabilities enabled or disabled on the virtual machine.
type AzureAdditionalCapabilities struct {
	HibernationEnabled *bool `json:"hibernation_enabled,omitempty"`
	UltraSSDEnabled    *bool `json:"ultra_ssd_enabled,omitempty"`
}

// AzureSubResource Specifies information about the availability set that the virtual machine should be
// assigned to. Virtual machines specified in the same availability set are allocated to different nodes
// to maximize availability.
type AzureSubResource struct {
	CloudID *string `json:"cloud_id,omitempty"`
}

// AzureBillingProfile Specifies the billing related details of a Azure Spot virtual machine.
type AzureBillingProfile struct {
	// https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/list?tabs=HTTP#billingprofile
	MaxPrice *float64 `json:"max_price,omitempty"`
}

// AzureHardwareProfile Specifies the hardware settings for the virtual machine.
type AzureHardwareProfile struct {
	// VmSize Specifies the size of the virtual machine.
	// https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/list?tabs=HTTP#virtualmachinesizetypes
	VmSize           *string                `json:"vm_size,omitempty"`
	VmSizeProperties *AzureVmSizeProperties `json:"vm_size_properties,omitempty"`
}

// AzureVmSizeProperties Specifies the properties for customizing the size of the virtual machine.
type AzureVmSizeProperties struct {
	// VCPUsAvailable Specifies the number of vCPUs available for the VM.
	VCPUsAvailable *int32 `json:"vcpus_available,omitempty"`
	// VCPUsPerCore Specifies the vCPU to physical core ratio.
	VCPUsPerCore *int32 `json:"vcpus_per_core,omitempty"`
}
