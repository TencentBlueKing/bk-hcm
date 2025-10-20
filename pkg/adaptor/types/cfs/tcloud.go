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
// 腾讯云Cfs-SDK-Models
package cfs

import (
	cfs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cfs/v20190719"

	apicore "hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// TCloudCreateStorageOption defines tencent cloud create storage options.
type TCloudCreateStorageOption struct {
	// -----------必填参数-----------
	// BkBizID 业务id
	BkBizID int64 `json:"bk_biz_id" validate:"required"`
	//// AccountID 账号id
	//AccountID string `json:"account_id" validate:"required"`
	// Name 资源名称
	Name string `json:"name" validate:"required"`
	// Region 所属地域
	Region string `json:"region" validate:"required"`
	// Zone 所属可用区
	Zone string `json:"zone" validate:"required"`
	// NetInterface 网络类型,可选值为VPC,CCN;其中VPC为私有网络,CCN为云联网.通用标准型/性能型请选择VPC, Turbo标准型/性能型请选择CCN.
	NetInterface string `json:"net_interface" validate:"required"`
	// PGroupId 权限组 ID. pgroupbasic 是默认权限组.
	// 通过控制查询权限组列表接口获取[DescribeCfsPGroups](https://cloud.tencent.com/document/product/582/38157)
	PGroupId string `json:"p_group_id" validate:"required"`

	// -----------选填参数-----------
	// CloudVpcID 用户指定的VPC ID. 腾讯云侧非必选.
	CloudVpcID *string `json:"cloud_vpc_id" validate:"omitempty"`
	// CloudSubnetID 用户指定的子网的网络ID. 腾讯云侧非必选.
	CloudSubnetID *string `json:"cloud_subnet_id" validate:"omitempty"`
	// Protocol 文件系统协议类型, 值为 NFS、CIFS、TURBO ; 若留空则默认为 NFS协议,turbo系列必须选择TURBO,不支持NFS、CIFS
	Protocol *string `json:"protocol" validate:"omitempty"`
	// StorageType 文件系统存储类型,默认值为 SD ;其中 SD 为通用标准型存储, HP为通用性能型存储, TB为Turbo标准型, TP 为Turbo性能型.
	StorageType *string `json:"storage_type" validate:"omitempty"`
	// Capacity 文件系统容量,turbo系列必填,单位为GiB. turbo标准型单位GB,起售20TiB,即20480 GiB;
	// 扩容步长10TiB,即10240 GiB.turbo性能型起售10TiB,即10240 GiB;扩容步长10TiB,10240 GiB.
	Capacity *uint64 `json:"capacity" validate:"omitempty"`
	// EnableAutoScaleUp 是否开启,默认扩容,仅turbo类型文件存储支持
	EnableAutoScaleUp *bool `json:"enable_auto_scale_up" validate:"omitempty"`
	// CfsVersion 文件系统版本; v1.5:创建普通版的通用文件系统; v3.1:创建增强版的通用文件系统
	// 说明:增强版的通用系统需要开通白名单才能使用,如有需要请提交工单与我们联系.
	CfsVersion *string `json:"cfs_version" validate:"omitempty"`
	//// MetaType 属性; turbo文件系统元数据属性; basic:创建标准型的元数据;enhanced:创建增强型的元数据
	//MetaType *string `json:"meta_type" validate:"omitempty"`
	// Memo 备注
	Memo *string `json:"memo" validate:"omitempty"`
	// Tags tag计费使用
	Tags []apicore.TagPair `json:"tags,omitempty"`
}

// Validate TCloudCreateStorageOption.
func (c TCloudCreateStorageOption) Validate() error {

	return validator.Validate.Struct(c)
}

// -------------------------- Delete --------------------------

// TCloudDeleteStorageOption defines tencent cloud delete storage options.
type TCloudDeleteStorageOption struct {
	// -----------必填参数-----------
	// Name 资源名称
	Name string `json:"name" validate:"required"`
	// CloudID 云上资源id
	CloudID string `json:"cloud_id" validate:"required"`
	// Region 所属地域
	Region string `json:"region" validate:"required"`
}

// Validate TCloudDeleteStorageOption.
func (c TCloudDeleteStorageOption) Validate() error {
	return validator.Validate.Struct(c)
}

// TCloudDeleteStorageResult defines tencent cloud delete storage result.
// response
type TCloudDeleteStorageResult struct {
	// CloudID 云上资源id(文件系统ID)
	CloudID string `json:"cloud_id"`
	// Name 资源名称
	Name string `json:"name"`
	// Status 状态
	Status bool `json:"status"`
	// RequestId 腾讯云request id
	RequestId *string `json:"request_id"`
}

// -------------------------- List --------------------------

// TCloudListStorageOption defines tencent cloud list storage options.
type TCloudListStorageOption struct {
	// -----------必填参数-----------
	// Region 所属地域
	Region string `json:"region" validate:"required"`

	// -----------选填参数-----------
	// CloudID 云上资源id(文件系统ID)
	CloudID *string `json:"cloud_id" validate:"omitempty"`
	// VpcId 私有网络(VPC) ID
	VpcId *string `json:"vpc_id" validate:"omitempty"`
	// SubnetId 子网 ID
	SubnetId *string `json:"subnet_id" validate:"omitempty"`
	// Offset 分页码,默认0
	Offset *uint64 `json:"offset" validate:"omitempty"`
	// Limit 页面大小,默认10
	Limit *uint64 `json:"limit" validate:"omitempty"`
	// CreationToken 用户自定义名称
	CreationToken *string `json:"creation_token" validate:"omitempty"`
}

// Validate TCloudListStorageOption.
func (c TCloudListStorageOption) Validate() error {
	return validator.Validate.Struct(c)
}

// TCloudGetStorageOption defines tencent cloud get storage options.
type TCloudGetStorageOption struct {
	// -----------必填参数-----------
	// Region 所属地域
	Region string `json:"region" validate:"required"`
	// CloudID 云上资源id(文件系统ID)
	CloudID string `json:"cloud_id" validate:"required"`

	// -----------选填参数-----------
	// VpcId 私有网络(VPC) ID
	VpcId *string `json:"vpc_id" validate:"omitempty"`
	// SubnetId 子网 ID
	SubnetId *string `json:"subnet_id" validate:"omitempty"`
}

// Validate TCloudListStorageOption.
func (c TCloudGetStorageOption) Validate() error {
	return validator.Validate.Struct(c)
}

// TCloudListStorageResult defines tencent cloud list storage result.
// response
type TCloudListStorageResult struct {
	// Storages 文件系统信息
	Storages []*StorageInfo `json:"data"`
}

// StorageInfo cfs storage
type StorageInfo struct {
	// Cfs 文件系统信息
	Cfs *cfs.FileSystemInfo `json:"cfs"`

	// MountInfo 挂载点信息
	MountInfo *cfs.MountInfo `json:"mount_info"`
}
