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

package disk

import (
	"hcm/pkg/criteria/validator"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// AzureCachingTypes ...
var AzureCachingTypes = map[string]armcompute.CachingTypes{
	"None":      armcompute.CachingTypesNone,
	"ReadOnly":  armcompute.CachingTypesReadOnly,
	"ReadWrite": armcompute.CachingTypesReadWrite,
}

// AzureDiskCreateOption ...
type AzureDiskCreateOption struct {
	DiskName          string  `json:"disk_name" validate:"required"`
	ResourceGroupName string  `json:"resource_group_name" validate:"required"`
	Region            string  `json:"region" validate:"required"`
	Zone              string  `json:"zone" validate:"required"`
	DiskType          string  `json:"disk_type" validate:"required"`
	DiskSize          int32   `json:"disk_size" validate:"required"`
	DiskCount         *uint64 `json:"disk_count" validate:"required"`
}

// Validate ...
func (opt *AzureDiskCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToCreateDiskRequest 转换成接口需要的 *armcompute.Disk 结构
func (opt *AzureDiskCreateOption) ToCreateDiskRequest() (*armcompute.Disk, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	skuName := armcompute.DiskStorageAccountTypes(opt.DiskType)
	sku := &armcompute.DiskSKU{Name: to.Ptr(skuName)}
	prop := &armcompute.DiskProperties{DiskSizeGB: to.Ptr(opt.DiskSize)}

	return &armcompute.Disk{
		Zones:      to.SliceOfPtrs[string](opt.Zone),
		Location:   to.Ptr(opt.Region),
		SKU:        sku,
		Properties: prop,
	}, nil
}

// AzureDiskListOption define azure disk list option.
type AzureDiskListOption struct {
	ResourceGroupName string   `json:"resource_group_name" validate:"required"`
	CloudIDs          []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate azure disk list option.
func (opt AzureDiskListOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AzureDiskGetOption ...
type AzureDiskGetOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	DiskName          string `json:"disk_name" validate:"required"`
}

// Validate ...
func (opt *AzureDiskGetOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AzureDiskDeleteOption ...
type AzureDiskDeleteOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	DiskName          string `json:"disk_name" validate:"required"`
}

// Validate ...
func (opt *AzureDiskDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AzureDiskAttachOption ...
type AzureDiskAttachOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	CvmName           string `json:"cvm_name" validate:"required"`
	DiskName          string `json:"disk_name" validate:"required"`
	CachingType       string `json:"caching_type" validate:"required,eq=None|eq=ReadOnly|eq=ReadWrite"`
}

// Validate ...
func (opt *AzureDiskAttachOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AzureDiskDetachOption ...
type AzureDiskDetachOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	CvmName           string `json:"cvm_name" validate:"required"`
	DiskName          string `json:"disk_name" validate:"required"`
}

// Validate ...
func (opt *AzureDiskDetachOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AzureDisk define azure disk.
type AzureDisk struct {
	ID       *string   `json:"id"`
	Name     *string   `json:"name"`
	Location *string   `json:"location"`
	Type     *string   `json:"type"`
	Status   *string   `json:"status"`
	DiskSize *int64    `json:"disk_size"`
	Zones    []*string `json:"zone"`
}
