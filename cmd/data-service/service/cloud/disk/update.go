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
	"fmt"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud/disk"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateDiskExt ...
func (dSvc *diskSvc) BatchUpdateDiskExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	switch vendor {
	case enumor.TCloud:
		return batchUpdateDiskExt[dataproto.TCloudDiskExtensionUpdateReq](cts, dSvc)
	case enumor.Aws:
		return batchUpdateDiskExt[dataproto.AwsDiskExtensionUpdateReq](cts, dSvc)
	case enumor.Gcp:
		return batchUpdateDiskExt[dataproto.GcpDiskExtensionUpdateReq](cts, dSvc)
	case enumor.Azure:
		return batchUpdateDiskExt[dataproto.AzureDiskExtensionUpdateReq](cts, dSvc)
	case enumor.HuaWei:
		return batchUpdateDiskExt[dataproto.HuaWeiDiskExtensionUpdateReq](cts, dSvc)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// BatchUpdateDisk ...
func (dSvc *diskSvc) BatchUpdateDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.DiskBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	updateData := &tablecloud.DiskModel{
		BkBizID: int64(req.BkBizID),
		Status:  req.Status,
		Memo:    req.Memo,
	}
	if err := dSvc.dao.Disk().Update(cts.Kit, tools.ContainersExpression("id", req.IDs), updateData); err != nil {
		return nil, err
	}
	return nil, nil
}

func batchUpdateDiskExt[T dataproto.DiskExtensionUpdateReq](cts *rest.Contexts,
	dSvc *diskSvc,
) (interface{}, error) {
	req := new(dataproto.DiskExtBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	queryIDs := make([]string, len(*req))
	for indx, diskReq := range *req {
		queryIDs[indx] = diskReq.ID
	}
	rawExtensions, err := dSvc.rawExtensions(cts, tools.ContainersExpression("id", queryIDs))
	if err != nil {
		return nil, err
	}

	_, err = dSvc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, diskReq := range *req {
			updateData := &tablecloud.DiskModel{
				Name:         diskReq.Name,
				Region:       diskReq.Region,
				BkBizID:      int64(diskReq.BkBizID),
				Status:       diskReq.Status,
				IsSystemDisk: diskReq.IsSystemDisk,
				Memo:         diskReq.Memo,
			}

			if diskReq.Extension != nil {
				rawExtension, exist := rawExtensions[diskReq.ID]
				if !exist {
					return nil, fmt.Errorf("disk id (%s) not exit", diskReq.ID)
				}
				mergedExtension, err := json.UpdateMerge(diskReq.Extension, string(rawExtension))
				if err != nil {
					return nil, fmt.Errorf("disk id (%s) merge extension failed, err: %v", diskReq.ID, err)
				}
				updateData.Extension = tabletype.JsonField(mergedExtension)
			}

			if err := dSvc.dao.Disk().UpdateByIDWithTx(cts.Kit, txn, diskReq.ID, updateData); err != nil {
				return nil, fmt.Errorf("update disk failed, err: %v", err)
			}
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// rawExtensions 根据条件查询原始的 extension 字段, 返回字典结构 {"云盘 ID": "原始的 extension 字段"}
func (dSvc *diskSvc) rawExtensions(
	cts *rest.Contexts,
	filterExp *filter.Expression,
) (map[string]tabletype.JsonField, error) {
	opt := &types.ListOption{
		Filter: filterExp,
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id", "extension"},
	}
	data, err := dSvc.dao.Disk().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	extensions := make(map[string]tabletype.JsonField)
	for _, d := range data.Details {
		extensions[d.ID] = d.Extension
	}

	return extensions, nil
}
