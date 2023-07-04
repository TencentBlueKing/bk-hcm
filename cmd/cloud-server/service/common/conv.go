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

package common

import (
	cloudserver "hcm/pkg/api/cloud-server/csdisk"
	hcproto "hcm/pkg/api/hc-service/disk"
)

// ConvTCloudDiskCreateReq conv disk create req.
func ConvTCloudDiskCreateReq(req *cloudserver.TCloudDiskCreateReq) *hcproto.TCloudDiskCreateReq {
	return &hcproto.TCloudDiskCreateReq{
		DiskBaseCreateReq: &hcproto.DiskBaseCreateReq{
			AccountID: req.AccountID,
			DiskName:  &req.DiskName,
			Region:    req.Region,
			Zone:      req.Zone,
			DiskSize:  req.DiskSize,
			DiskType:  req.DiskType,
			DiskCount: req.DiskCount,
			Memo:      req.Memo,
		},
		Extension: &hcproto.TCloudDiskExtensionCreateReq{
			DiskChargeType:    req.DiskChargeType,
			DiskChargePrepaid: req.DiskChargePrepaid,
		},
	}
}

// ConvHuaWeiDiskCreateReq conv disk create req.
func ConvHuaWeiDiskCreateReq(req *cloudserver.HuaWeiDiskCreateReq) *hcproto.HuaWeiDiskCreateReq {
	return &hcproto.HuaWeiDiskCreateReq{
		DiskBaseCreateReq: &hcproto.DiskBaseCreateReq{
			AccountID: req.AccountID,
			DiskName:  req.DiskName,
			Region:    req.Region,
			Zone:      req.Zone,
			DiskSize:  uint64(req.DiskSize),
			DiskType:  req.DiskType,
			DiskCount: uint32(req.DiskCount),
			Memo:      req.Memo,
		},
		Extension: &hcproto.HuaWeiDiskExtensionCreateReq{
			DiskChargeType:    *req.DiskChargeType,
			DiskChargePrepaid: req.DiskChargePrepaid,
		},
	}
}

// ConvAwsDiskCreateReq conv disk create req.
func ConvAwsDiskCreateReq(req *cloudserver.AwsDiskCreateReq) *hcproto.AwsDiskCreateReq {
	return &hcproto.AwsDiskCreateReq{
		DiskBaseCreateReq: &hcproto.DiskBaseCreateReq{
			AccountID: req.AccountID,
			Region:    req.Region,
			Zone:      req.Zone,
			DiskSize:  uint64(req.DiskSize),
			DiskType:  req.DiskType,
			DiskCount: uint32(req.DiskCount),
			Memo:      req.Memo,
		},
	}
}

// ConvGcpDiskCreateReq conv disk create req.
func ConvGcpDiskCreateReq(req *cloudserver.GcpDiskCreateReq) *hcproto.GcpDiskCreateReq {
	return &hcproto.GcpDiskCreateReq{
		DiskBaseCreateReq: &hcproto.DiskBaseCreateReq{
			AccountID: req.AccountID,
			DiskName:  &req.DiskName,
			Region:    req.Region,
			Zone:      req.Zone,
			DiskSize:  uint64(req.DiskSize),
			DiskType:  req.DiskType,
			DiskCount: uint32(req.DiskCount),
			Memo:      req.Memo,
		},
	}
}

// ConvAzureDiskCreateReq conv disk create req.
func ConvAzureDiskCreateReq(req *cloudserver.AzureDiskCreateReq) *hcproto.AzureDiskCreateReq {
	return &hcproto.AzureDiskCreateReq{
		DiskBaseCreateReq: &hcproto.DiskBaseCreateReq{
			AccountID: req.AccountID,
			DiskName:  &req.DiskName,
			Region:    req.Region,
			Zone:      req.Zone,
			DiskSize:  uint64(req.DiskSize),
			DiskType:  req.DiskType,
			DiskCount: uint32(req.DiskCount),
			Memo:      req.Memo,
		},
		Extension: &hcproto.AzureDiskExtensionCreateReq{
			ResourceGroupName: req.ResourceGroupName,
		},
	}
}
