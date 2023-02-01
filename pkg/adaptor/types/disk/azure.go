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

import "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"

// AzureDiskCreateOption ...
type AzureDiskCreateOption struct {
	Name              string
	ResourceGroupName string
	Region            *string
	Zone              *string
	DiskType          string
	DiskSize          *int32
}

// ToCreateDiskRequest 转换成接口需要的 *armcompute.Disk 结构
func (opt *AzureDiskCreateOption) ToCreateDiskRequest() (*armcompute.Disk, error) {
	skuName := armcompute.DiskStorageAccountTypes(opt.DiskType)
	sku := &armcompute.DiskSKU{Name: &skuName}
	prop := &armcompute.DiskProperties{DiskSizeGB: opt.DiskSize}

	return &armcompute.Disk{
		Zones:      []*string{opt.Zone},
		Location:   opt.Region,
		SKU:        sku,
		Properties: prop,
	}, nil
}
