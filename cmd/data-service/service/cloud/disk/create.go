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
	"hcm/pkg/api/core"
	coredisk "hcm/pkg/api/core/cloud/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablecloud "hcm/pkg/dal/table/cloud/disk"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// BatchCreateDiskExt 批量创建云盘(支持 extension 字段)
func (dSvc *diskSvc) BatchCreateDiskExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	switch vendor {
	case enumor.TCloud:
		return batchCreateDiskExt[coredisk.TCloudExtension](cts, dSvc, vendor)
	case enumor.Aws:
		return batchCreateDiskExt[coredisk.AwsExtension](cts, dSvc, vendor)
	case enumor.Gcp:
		return batchCreateDiskExt[coredisk.GcpExtension](cts, dSvc, vendor)
	case enumor.Azure:
		return batchCreateDiskExt[coredisk.AzureExtension](cts, dSvc, vendor)
	case enumor.HuaWei:
		return batchCreateDiskExt[coredisk.HuaWeiExtension](cts, dSvc, vendor)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func batchCreateDiskExt[T coredisk.Extension](
	cts *rest.Contexts,
	dSvc *diskSvc,
	vendor enumor.Vendor,
) (interface{}, error) {
	req := new(dataproto.DiskExtBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	diskIDs, err := dSvc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		disks := make([]*tablecloud.DiskModel, len(*req))
		for indx, diskReq := range *req {
			extensionJson, err := json.MarshalToString(diskReq.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			bkBizID := diskReq.BkBizID
			if bkBizID == 0 {
				bkBizID = constant.UnassignedBiz
			}

			disks[indx] = &tablecloud.DiskModel{
				Vendor:       string(vendor),
				AccountID:    diskReq.AccountID,
				CloudID:      diskReq.CloudID,
				BkBizID:      bkBizID,
				Name:         diskReq.Name,
				Region:       diskReq.Region,
				Zone:         diskReq.Zone,
				DiskSize:     diskReq.DiskSize,
				DiskType:     diskReq.DiskType,
				Status:       diskReq.Status,
				IsSystemDisk: converter.ValToPtr(diskReq.IsSystemDisk),
				Memo:         diskReq.Memo,
				Extension:    tabletype.JsonField(extensionJson),
				Creator:      cts.Kit.User,
				Reviser:      cts.Kit.User,
			}
		}
		return dSvc.dao.Disk().BatchCreateWithTx(cts.Kit, txn, disks)
	})
	if err != nil {
		return nil, err
	}

	return &core.BatchCreateResult{IDs: diskIDs.([]string)}, nil
}
