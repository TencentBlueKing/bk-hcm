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

package azure

import (
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	hcproto "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/enumor"
)

// Deliver ...
func (a *ApplicationOfCreateAzureDisk) Deliver() (
	status enumor.ApplicationStatus,
	deliverDetail map[string]interface{},
	err error,
) {
	resp, err := a.Client.HCService().Azure.Disk.CreateDisk(a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(), a.toHcProtoCreateReq())
	if err != nil {
		status = enumor.DeliverError
		deliverDetail["error"] = err
		return
	}

	deliverDetail["disk_ids"] = resp.IDs

	status = enumor.Completed
	_, err = a.Client.DataService().Global.BatchUpdateDisk(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataproto.DiskBatchUpdateReq{IDs: resp.IDs, BkBizID: uint64(a.req.BkBizID)},
	)
	if err != nil {
		status = enumor.DeliverError
		deliverDetail["error"] = err
	}
	return
}

// toHcProtoCreateReq ...
func (a *ApplicationOfCreateAzureDisk) toHcProtoCreateReq() *hcproto.AzureDiskCreateReq {
	req := a.req
	return &hcproto.AzureDiskCreateReq{
		DiskBaseCreateReq: &hcproto.DiskBaseCreateReq{
			AccountID: req.AccountID,
			DiskName:  &req.DiskName,
			Region:    req.Region,
			Zone:      req.Zone,
			DiskSize:  uint64(req.DiskSize),
			DiskType:  req.DiskType,
			DiskCount: uint32(*req.DiskCount),
			Memo:      req.Memo,
		},
		Extension: &hcproto.AzureDiskExtensionCreateReq{ResourceGroupName: req.ResourceGroupName},
	}
}
