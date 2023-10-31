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

package coredisk

import "hcm/pkg/api/core"

// BaseDisk define base disk.
type BaseDisk struct {
	ID            string  `json:"id"`
	Vendor        string  `json:"vendor"`
	AccountID     string  `json:"account_id"`
	Name          string  `json:"name"`
	BkBizID       int64   `json:"bk_biz_id"`
	CloudID       string  `json:"cloud_id"`
	Region        string  `json:"region"`
	Zone          string  `json:"zone"`
	DiskSize      uint64  `json:"disk_size"`
	DiskType      string  `json:"disk_type"`
	Status        string  `json:"status"`
	RecycleStatus string  `json:"recycle_status"`
	IsSystemDisk  bool    `json:"is_system_disk"`
	Memo          *string `json:"memo"`
	core.Revision `json:",inline"`
}

// Disk define disk.
type Disk[Ext Extension] struct {
	BaseDisk  `json:",inline"`
	Extension *Ext `json:"extension"`
}

// GetID ...
func (disk Disk[T]) GetID() string {
	return disk.ID
}

// GetCloudID ...
func (disk Disk[T]) GetCloudID() string {
	return disk.CloudID
}

// Extension ...
type Extension interface {
	TCloudExtension | AwsExtension | AzureExtension | GcpExtension | HuaWeiExtension
}

// TCloudExtension ...
type TCloudExtension struct {
	DiskChargeType     string                   `json:"disk_charge_type" validate:"required"`
	DiskChargePrepaid  *TCloudDiskChargePrepaid `json:"disk_charge_prepaid,omitempty"`
	Encrypted          *bool                    `json:"encrypted,omitempty"`
	Attached           *bool                    `json:"attached,omitempty"`
	DiskUsage          *string                  `json:"disk_usage,omitempty"`
	InstanceId         *string                  `json:"instance_id,omitempty"`
	InstanceType       *string                  `json:"instance_type,omitempty"`
	DeleteWithInstance *bool                    `json:"delete_with_instance,omitempty"`
	DeadlineTime       *string                  `json:"deadline_time,omitempty"`
	BackupDisk         *bool                    `json:"backup_disk,omitempty"`
}

// TCloudDiskChargePrepaid ...
type TCloudDiskChargePrepaid struct {
	Period    *int64  `json:"period"`
	RenewFlag *string `json:"renew_flag"`
}

// AwsExtension ...
type AwsExtension struct {
	Attachment []*AwsDiskAttachment `json:"attachment,omitempty"`
	Encrypted  *bool                `json:"encrypted,omitempty"`
}

// AwsDiskAttachment ...
type AwsDiskAttachment struct {
	// The time stamp when the attachment initiated.
	AttachTime string `json:"attach_time"`
	// Indicates whether the EBS volume is deleted on instance termination.
	DeleteOnTermination *bool `json:"delete_on_termination,omitempty"`
	// The device name.
	DeviceName *string `json:"device_name"`
	// The ID of the instance.
	InstanceId *string `json:"instance_id"`
	// The attachment state of the volume.
	Status *string `json:"status"`
	// The ID of the volume.
	DiskId *string `json:"disk_id"`
}

// AzureExtension ...
type AzureExtension struct {
	ResourceGroupName string    `json:"resource_group_name" validate:"required"`
	Encrypted         *bool     `json:"encrypted,omitempty"`
	OSType            string    `json:"os_type"`
	SKUName           *string   `json:"sku_name,omitempty"`
	SKUTier           *string   `json:"sku_tier,omitempty"`
	Zones             []*string `json:"zones,omitempty"`
}

// GcpExtension ...
type GcpExtension struct {
	SelfLink    string `json:"self_link" validate:"required"`
	SourceImage string `json:"source_image,omitempty"`
	Description string `json:"description,omitempty"`
	Encrypted   *bool  `json:"encrypted,omitempty"`
}

// HuaWeiExtension ...
type HuaWeiExtension struct {
	ChargeType    string                   `json:"charge_type" validate:"omitempty"`
	ExpireTime    string                   `json:"expire_time" validate:"omitempty"`
	ChargePrepaid *HuaWeiDiskChargePrepaid `json:"charge_prepaid" validate:"omitempty"`
	// 服务类型，结果为EVS、DSS、DESS.
	ServiceType string `json:"service_type"`
	// 当前云硬盘服务不支持该字段
	Encrypted  *bool                   `json:"encrypted,omitempty"`
	Attachment []*HuaWeiDiskAttachment `json:"attachment,omitempty"`
	Bootable   string                  `json:"bootable"`
}

// HuaWeiDiskChargePrepaid ...
type HuaWeiDiskChargePrepaid struct {
	PeriodNum   *int32  `json:"period_num"`
	PeriodType  *string `json:"period_type"`
	IsAutoRenew *string `json:"is_auto_renew"`
}

// HuaWeiDiskAttachment ...
type HuaWeiDiskAttachment struct {
	// 挂载的时间信息。  时间格式：UTC YYYY-MM-DDTHH:MM:SS.XXXXXX
	AttachedAt string `json:"attached_at"`
	// 挂载信息对应的ID
	AttachmentId string `json:"attachment_id"`
	// 挂载点
	DeviceName string `json:"device_name"`
	// 云硬盘挂载到的云服务器对应的物理主机的名称
	HostName string `json:"host_name"`
	// 挂载的资源ID
	Id string `json:"id"`
	// 云硬盘挂载到的云服务器的 ID
	InstanceId string `json:"instance_id"`
	// 云硬盘ID
	DiskId string `json:"disk_id"`
}
