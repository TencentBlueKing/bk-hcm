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
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	diskdao "hcm/pkg/dal/dao/cloud/disk"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud/disk"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

type diskSvc struct {
	dao.Set
	objectDao *diskdao.DiskDao
}

// Init 注册 diskdao.DiskDao 到 Capability.Dao, 并设置 objectDao
func (dSvc *diskSvc) Init() {
	d := &diskdao.DiskDao{}
	registeredDao := dSvc.GetObjectDao(d.Name())
	if registeredDao == nil {
		d.ObjectDaoManager = new(dao.ObjectDaoManager)
		dSvc.RegisterObjectDao(d)
	}

	dSvc.objectDao = dSvc.GetObjectDao(d.Name()).(*diskdao.DiskDao)
	dSvc.objectDao.Audit = dSvc.Audit()
}

// BatchCreateDiskExt 批量创建云盘(支持 extension 字段)
func (dSvc *diskSvc) BatchCreateDiskExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	switch vendor {
	case enumor.TCloud:
		return batchCreateDiskExt[dataproto.TCloudDiskExtensionCreateReq](cts, dSvc, vendor)
	case enumor.Aws:
		return batchCreateDiskExt[dataproto.AwsDiskExtensionCreateReq](cts, dSvc, vendor)
	case enumor.Gcp:
		return batchCreateDiskExt[dataproto.GcpDiskExtensionCreateReq](cts, dSvc, vendor)
	case enumor.Azure:
		return batchCreateDiskExt[dataproto.AzureDiskExtensionCreateReq](cts, dSvc, vendor)
	case enumor.HuaWei:
		return batchCreateDiskExt[dataproto.HuaWeiDiskExtensionCreateReq](cts, dSvc, vendor)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// RetrieveDiskExt 获取云盘详情
func (dSvc *diskSvc) RetrieveDiskExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	diskID := cts.PathParameter("id").String()
	opt := &types.ListOption{
		Filter: tools.EqualWithOpExpression(
			filter.And,
			map[string]interface{}{"id": diskID, "vendor": string(vendor)},
		),
		Page: &core.BasePage{Count: false, Start: 0, Limit: 1},
	}

	data, err := dSvc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	if count := len(data.Details); count != 1 {
		return nil, fmt.Errorf("retrieve disk failed: query id(%s) return total %d", diskID, count)
	}

	diskData := data.Details[0]
	switch vendor {
	case enumor.TCloud:
		return toProtoDiskExtResult[dataproto.TCloudDiskExtensionResult](diskData)
	case enumor.Aws:
		return toProtoDiskExtResult[dataproto.AwsDiskExtensionResult](diskData)
	case enumor.Gcp:
		return toProtoDiskExtResult[dataproto.GcpDiskExtensionResult](diskData)
	case enumor.Azure:
		return toProtoDiskExtResult[dataproto.AzureDiskExtensionResult](diskData)
	case enumor.HuaWei:
		return toProtoDiskExtResult[dataproto.HuaWeiDiskExtensionResult](diskData)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

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
	if err := dSvc.objectDao.Update(cts.Kit, tools.ContainersExpression("id", req.IDs), updateData); err != nil {
		return nil, err
	}
	return nil, nil
}

// ListDisk 查询云盘列表
func (dSvc *diskSvc) ListDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.DiskListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	data, err := dSvc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	details := make([]*dataproto.DiskResult, len(data.Details))
	for indx, d := range data.Details {
		details[indx] = toProtoDiskResult(d)
	}

	return &dataproto.DiskListResult{Details: details, Count: data.Count}, nil
}

// ListDiskExt 获取云盘列表(带 extension 字段)
func (dSvc *diskSvc) ListDiskExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(dataproto.DiskListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	data, err := dSvc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return toProtoDiskExtListResult[dataproto.TCloudDiskExtensionResult](data)
	case enumor.Aws:
		return toProtoDiskExtListResult[dataproto.AwsDiskExtensionResult](data)
	case enumor.Gcp:
		return toProtoDiskExtListResult[dataproto.GcpDiskExtensionResult](data)
	case enumor.Azure:
		return toProtoDiskExtListResult[dataproto.AzureDiskExtensionResult](data)
	case enumor.HuaWei:
		return toProtoDiskExtListResult[dataproto.HuaWeiDiskExtensionResult](data)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// BatchDeleteDisk 删除云盘
func (dSvc *diskSvc) BatchDeleteDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.DiskDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id"},
		Filter: req.Filter,
		Page:   core.DefaultBasePage,
	}
	listResp, err := dSvc.objectDao.List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list disk failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}

	_, err = dSvc.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delFilter := tools.ContainersExpression("id", delIDs)
		if err := dSvc.objectDao.DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			return nil, err
		}

		// delete disk related recycle records
		delRecycleRecordFilter, err := tools.And(tools.ContainersExpression("res_id", delIDs),
			tools.EqualExpression("res_type", enumor.DiskCloudResType))
		if err != nil {
			return nil, err
		}
		if err := dSvc.RecycleRecord().BatchDeleteWithTx(cts.Kit, txn, delRecycleRecordFilter); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// CountDisk 统计云盘数量
func (dSvc *diskSvc) CountDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.DiskCountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.CountOption{
		Filter: req.Filter,
	}

	data, err := dSvc.objectDao.Count(cts.Kit, opt)
	if err != nil {
		return nil, err
	}
	return &dataproto.DiskCountResult{Count: data.Count}, nil
}

// rawExtensions 根据条件查询原始的 extension 字段, 返回字典结构 {"云盘 ID": "原始的 extension 字段"}
func (dSvc *diskSvc) rawExtensions(
	cts *rest.Contexts,
	filterExp *filter.Expression,
) (map[string]tabletype.JsonField, error) {
	opt := &types.ListOption{
		Filter: filterExp,
		Page:   core.DefaultBasePage,
		Fields: []string{"id", "extension"},
	}
	data, err := dSvc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	extensions := make(map[string]tabletype.JsonField)
	for _, d := range data.Details {
		extensions[d.ID] = d.Extension
	}

	return extensions, nil
}

func batchCreateDiskExt[T dataproto.DiskExtensionCreateReq](
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

	diskIDs, err := dSvc.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
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
				Vendor:    string(vendor),
				AccountID: diskReq.AccountID,
				CloudID:   diskReq.CloudID,
				BkBizID:   bkBizID,
				Name:      diskReq.Name,
				Region:    diskReq.Region,
				Zone:      diskReq.Zone,
				DiskSize:  diskReq.DiskSize,
				DiskType:  diskReq.DiskType,
				Status:    diskReq.Status,
				Memo:      diskReq.Memo,
				Extension: tabletype.JsonField(extensionJson),
				Creator:   cts.Kit.User,
				Reviser:   cts.Kit.User,
			}
		}
		return dSvc.objectDao.BatchCreateWithTx(cts.Kit, txn, disks)
	})
	if err != nil {
		return nil, err
	}

	return &core.BatchCreateResult{IDs: diskIDs.([]string)}, nil
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

	_, err = dSvc.Set.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, diskReq := range *req {
			updateData := &tablecloud.DiskModel{
				Name:    diskReq.Name,
				Region:  diskReq.Region,
				BkBizID: int64(diskReq.BkBizID),
				Status:  diskReq.Status,
				IsSystemDisk: diskReq.IsSystemDisk,
				Memo:    diskReq.Memo,
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

			if err := dSvc.objectDao.UpdateByIDWithTx(cts.Kit, txn, diskReq.ID, updateData); err != nil {
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
