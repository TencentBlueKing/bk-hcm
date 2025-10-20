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

// Package cfs 如下:
// api core
package cfs

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// BaseCfs define base cfs.
type BaseCfs struct {
	// ID 主键
	ID string `json:"id"`
	// BkBizID 业务id
	BkBizID int64 `json:"bk_biz_id"`
	// AccountID 账号id
	AccountID string `json:"account_id"`
	// Vendor 云标识
	Vendor enumor.Vendor `json:"vendor"`

	// CloudID  云上资源id
	CloudID string `json:"cloud_id"`
	// Name 资源名称
	Name string `json:"name"`
	// Region 所属地域
	Region string `json:"region"`
	// Zone 所属可用区
	Zone string `json:"zone"`
	// SizeLimit 文件系统最大空间限制(单位:GiB, 示例值：50)
	SizeLimit uint64 `json:"size_limit"`
	// SizeByte 文件系统已使用容量.单位：Byte; 示例值：10
	SizeByte uint64 `json:"size_byte"`
	// AvailCapacity 文件系统剩余容量. 单位：Byte. 示例值：10
	AvailCapacity uint64 `json:"avail_capacity"`
	// BandwidthLimit 文件系统吞吐上限，吞吐上限是根据文件系统当前已使用存储量、绑定的存储资源包以及吞吐资源包一同确定. 单位MiB/s
	BandwidthLimit float64 `json:"bandwidth_limit" `
	// Protocol 文件系统协议类型, 支持 NFS,CIFS,TURBO; 示例值：NFS
	Protocol string `json:"protocol" `
	// StorageType 文件系统存储类型. HP：通用性能型;SD：通用标准型;TP:turbo性能型;TB：turbo标准型;THP：吞吐型; 示例值：HP
	StorageType string `json:"storage_type" `
	// Encrypted 文件系统是否加密
	Encrypted bool `json:"encrypted"`
	// CryptKeyId 加密所使用的密钥，可以为密钥的 ID 或者 ARN
	CryptKeyId string `json:"crypt_key_id"`

	// CloudVpcIDs 云上vpc
	CloudVpcIDs []string `json:"cloud_vpc_ids"`
	// CloudSubnetIDs 云上子网
	CloudSubnetIDs []string `json:"cloud_subnet_ids"`
	// VpcIDs vpc
	VpcIDs []string `json:"vpc_ids"`
	// SubnetIDs 子网
	SubnetIDs []string `json:"subnet_ids"`

	// Status 状态
	/*
		tcloud: creating:创建中; mounting:挂载中;create_failed:创建失败;available:可使用;unserviced:停服中;upgrading:升级中;

		huawei: '121'表示扩容中;'132'表示修改安全组中;'137'表示添加VPC中;'138'表示删除VPC中;'150'表示配置联动后端中;
		 '151'表示删除联动后端配置中. '221'表示扩容成功;'232'表示修改安全组成功;'237'表示添加VPC成功;'238'表示删除VPC成功;
		 '250'表示配置联动后端成功;'251'表示删除联动后端配置成功.'321'表示扩容失败;'332'表示修改安全组失败;'337'表示添加VPC失败;
		 '338'表示删除VPC失败;'350'表示配置联动后端失败;'351'表示删除联动后端配置失败.
			(https://support.huaweicloud.com/api-sfs/ShowShare.html)
	*/
	Status string `json:"status"`
	// CloudCreatedTime 云上资源创建时间
	CloudCreatedTime string `json:"cloud_created_time"`
	// Memo 备注
	Memo *string `json:"memo"`
	// Revision 创建人
	*core.Revision `json:",inline"`
}

// Cfs define cfs.
type Cfs[Ext Extension] struct {
	BaseCfs   `json:",inline"`
	Extension *Ext `json:"extension"`
}

// GetID ...
func (c Cfs[T]) GetID() string {
	return c.BaseCfs.ID
}

// GetCloudID ...
func (c Cfs[T]) GetCloudID() string {
	return c.BaseCfs.CloudID
}

// GetCloudID ...
func (b BaseCfs) GetCloudID() string {
	return b.CloudID
}

// Extension cfs extension.
type Extension interface {
	TCloudCfsExtension
}
