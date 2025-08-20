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

// Package cvm ...
package cvm

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// SummaryCVM define summary cvm.
type SummaryCVM struct {
	ID      string        `json:"id"`
	CloudID string        `json:"cloud_id"`
	Name    string        `json:"name"`
	Vendor  enumor.Vendor `json:"vendor"`
	BkBizID int64         `json:"bk_biz_id"`
	Region  string        `json:"region"`
	Zone    string        `json:"zone"`

	CloudVpcIDs    []string `json:"cloud_vpc_ids"`
	CloudSubnetIDs []string `json:"cloud_subnet_ids"`

	Status string `json:"status"`

	// PrivateIPv4Addresses 内网IP
	PrivateIPv4Addresses []string `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []string `json:"private_ipv6_addresses"`
	// PublicIPv6Addresses 公网IP
	PublicIPv4Addresses []string `json:"public_ipv4_addresses"`
	PublicIPv6Addresses []string `json:"public_ipv6_addresses"`
}

// BaseCvm define base cvm.
type BaseCvm struct {
	ID        string        `json:"id"`
	CloudID   string        `json:"cloud_id"`
	Name      string        `json:"name"`
	Vendor    enumor.Vendor `json:"vendor"`
	BkBizID   int64         `json:"bk_biz_id"`
	BkHostID  int64         `json:"bk_host_id"`
	BkCloudID int64         `json:"bk_cloud_id"`
	AccountID string        `json:"account_id"`
	Region    string        `json:"region"`
	Zone      string        `json:"zone"`

	CloudVpcIDs    []string `json:"cloud_vpc_ids"`
	VpcIDs         []string `json:"vpc_ids"`
	CloudSubnetIDs []string `json:"cloud_subnet_ids"`
	SubnetIDs      []string `json:"subnet_ids"`

	CloudImageID string `json:"cloud_image_id"`
	// ImageID 预留字段，因为目前 hcm 还没有支持镜像资源。
	ImageID string  `json:"image_id,omitempty"`
	OsName  string  `json:"os_name"`
	Memo    *string `json:"memo"`
	/*
		tcloud: PENDING：表示创建中、LAUNCH_FAILED：表示创建失败、RUNNING：表示运行中、STOPPED：表示关机、
			STARTING：表示开机中、STOPPING：表示关机中、REBOOTING：表示重启中、SHUTDOWN：表示停止待销毁、TERMINATING：表示销毁中。

		huawei: ACTIVE、BUILD、ERROR、HARD_REBOOT、MIGRATING、REBOOT、REBUILD、RESIZE、
			REVERT_RESIZE、SHUTOFF、VERIFY_RESIZE、DELETED
		huawei_link: https://support.huaweicloud.com/api-ecs/ecs_08_0002.html

		gcp: PROVISIONING, STAGING, RUNNING, STOPPING, SUSPENDING, SUSPENDED, REPAIRING, and TERMINATED
		aws: pending | running | shutting-down | terminated | stopping | stopped
		azure：PowerState/running｜PowerState/stopped｜PowerState/deallocating｜PowerState/deallocated
	*/
	Status        string `json:"status"`
	RecycleStatus string `json:"recycle_status,omitempty"`

	// PrivateIPv4Addresses 内网IP
	PrivateIPv4Addresses []string `json:"private_ipv4_addresses"`
	PrivateIPv6Addresses []string `json:"private_ipv6_addresses"`
	// PublicIPv6Addresses 公网IP
	PublicIPv4Addresses []string `json:"public_ipv4_addresses"`
	PublicIPv6Addresses []string `json:"public_ipv6_addresses"`

	// MachineType 设备类型
	MachineType string `json:"machine_type"`

	CloudCreatedTime  string `json:"cloud_created_time"`
	CloudLaunchedTime string `json:"cloud_launched_time"`
	CloudExpiredTime  string `json:"cloud_expired_time"`
	*core.Revision    `json:",inline"`
}

// Cvm define cvm.
type Cvm[Ext Extension] struct {
	BaseCvm   `json:",inline"`
	Extension *Ext `json:"extension"`
}

// GetID ...
func (cvm Cvm[T]) GetID() string {
	return cvm.BaseCvm.ID
}

// GetCloudID ...
func (cvm Cvm[T]) GetCloudID() string {
	return cvm.BaseCvm.CloudID
}

// GetCloudID ...
func (b BaseCvm) GetCloudID() string {
	return b.CloudID
}

// Extension cvm extension.
type Extension interface {
	TCloudCvmExtension | AwsCvmExtension | HuaWeiCvmExtension | AzureCvmExtension | GcpCvmExtension | OtherCvmExtension
}
