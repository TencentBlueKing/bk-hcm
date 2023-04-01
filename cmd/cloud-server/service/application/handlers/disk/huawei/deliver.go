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

package huawei

import (
	"hcm/cmd/cloud-server/service/application/handlers/disk/logics"
	hcproto "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/enumor"
)

// Deliver ...
func (a *ApplicationOfCreateHuaWeiDisk) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	result, err := a.Client.HCService().HuaWei.Disk.CreateDisk(a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(), a.toHcProtoCreateReq())
	if err != nil {
		return enumor.DeliverError, map[string]interface{}{"error": err}, err
	}

	return logics.CheckResultAndAssign(a.Cts.Kit, a.Client.DataService(), result, uint32(a.req.DiskCount),
		a.req.BkBizID)
}

// toHcProtoCreateReq ...
func (a *ApplicationOfCreateHuaWeiDisk) toHcProtoCreateReq() *hcproto.HuaWeiDiskCreateReq {
	req := a.req
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
