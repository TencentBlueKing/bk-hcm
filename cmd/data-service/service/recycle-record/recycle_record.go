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

package recyclerecord

import (
	"fmt"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/recycle-record"
	protodata "hcm/pkg/api/data-service/recycle-record"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	protodao "hcm/pkg/dal/dao/types/recycle-record"
	prototable "hcm/pkg/dal/table/recycle-record"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// InitRecycleRecordService initialize the recycle record service.
func InitRecycleRecordService(cap *capability.Capability) {
	svc := &recycleRecordSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchRecycleCloudResource", "POST", "/cloud/resources/batch/recycle", svc.BatchRecycleCloudResource)
	h.Add("BatchRecoverCloudResource", "POST", "/cloud/resources/batch/recover", svc.BatchRecoverCloudResource)
	h.Add("ListRecycleRecord", "POST", "/recycle_records/list", svc.ListRecycleRecord)
	h.Add("BatchUpdateRecycleRecord", "PATCH", "/recycle_records/batch", svc.BatchUpdateRecycleRecord)

	h.Load(cap.WebService)
}

type recycleRecordSvc struct {
	dao dao.Set
}

// BatchRecycleCloudResource batch recycle cloud resource.
func (svc *recycleRecordSvc) BatchRecycleCloudResource(cts *rest.Contexts) (interface{}, error) {
	req := new(protodata.BatchRecycleReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	resIDs := make([]string, 0, len(req.Infos))
	for _, info := range req.Infos {
		resIDs = append(resIDs, info.ID)
	}

	resourceInfo, err := svc.dao.RecycleRecord().ListResourceInfo(cts.Kit, req.ResType, resIDs)
	if err != nil {
		return nil, err
	}

	if len(resourceInfo) != len(req.Infos) {
		return nil, errf.Newf(errf.InvalidParameter, "recycle resource count is invalid")
	}

	taskID, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		recycleRecords := make([]prototable.RecycleRecordTable, 0, len(resourceInfo))
		for idx, info := range resourceInfo {
			detail, err := tabletype.NewJsonField(req.Infos[idx].Detail)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			recycleRecords = append(recycleRecords, prototable.RecycleRecordTable{
				Vendor:     info.Vendor,
				ResType:    req.ResType,
				ResID:      info.ID,
				CloudResID: info.CloudID,
				ResName:    info.Name,
				BkBizID:    info.BkBizID,
				AccountID:  info.AccountID,
				Region:     info.Region,
				Detail:     detail,
				Status:     enumor.WaitingRecycleRecordStatus,
				Creator:    cts.Kit.User,
				Reviser:    cts.Kit.User})
		}

		// recycle resource
		updateResOpt := &protodao.ResourceUpdateOptions{ResType: req.ResType, IDs: resIDs, Status: enumor.RecycleStatus,
			BkBizID: constant.UnassignedBiz}

		if err := svc.dao.RecycleRecord().UpdateResource(cts.Kit, txn, updateResOpt); err != nil {
			return nil, fmt.Errorf("update recycled resource info failed, err: %v", err)
		}

		// create recycle record
		taskID, err := svc.dao.RecycleRecord().BatchCreateWithTx(cts.Kit, txn, recycleRecords)
		if err != nil {
			return nil, fmt.Errorf("create recycle record failed, err: %v", err)
		}

		return taskID, nil
	})

	if err != nil {
		return nil, err
	}

	return taskID, nil
}

// BatchRecoverCloudResource batch recover cloud resource.
func (svc *recycleRecordSvc) BatchRecoverCloudResource(cts *rest.Contexts) (interface{}, error) {
	req := new(protodata.BatchRecoverReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", req.RecordIDs),
		Page: &core.BasePage{
			Limit: core.DefaultMaxPageLimit,
		},
		Fields: []string{"id", "bk_biz_id", "res_id"},
	}
	listResp, err := svc.dao.RecycleRecord().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list recycle record failed, err: %v, ids: %+v, rid: %s", err, req.RecordIDs, cts.Kit.Rid)
		return nil, fmt.Errorf("list recycle record failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	recordIDs := make([]string, len(listResp.Details))
	bizCvmIDMap := make(map[int64][]string)
	for index, one := range listResp.Details {
		recordIDs[index] = one.ID
		bizCvmIDMap[one.BkBizID] = append(bizCvmIDMap[one.BkBizID], one.ResID)
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		// recover resource
		for bizID, ids := range bizCvmIDMap {
			updateResOpt := &protodao.ResourceUpdateOptions{ResType: req.ResType, IDs: ids,
				Status: enumor.RecoverStatus, BkBizID: bizID}
			err := svc.dao.RecycleRecord().UpdateResource(cts.Kit, txn, updateResOpt)
			if err != nil {
				return nil, fmt.Errorf("update recycled resource status failed, err: %v", err)
			}
		}

		// delete recycle records
		updateFilter := tools.ContainersExpression("id", recordIDs)
		updateData := &prototable.RecycleRecordTable{
			Status:  enumor.RecoverRecycleRecordStatus,
			Reviser: cts.Kit.User,
		}
		if err := svc.dao.RecycleRecord().Update(cts.Kit, txn, updateFilter, updateData); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete recycle record failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListRecycleRecord list recycle records.
func (svc *recycleRecordSvc) ListRecycleRecord(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
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
	res, err := svc.dao.RecycleRecord().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list recycle record failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list recycle record failed, err: %v", err)
	}
	if req.Page.Count {
		return &protodata.ListResult{Count: res.Count}, nil
	}

	records := make([]protocore.RecycleRecord, 0, len(res.Details))
	for _, recycleRecord := range res.Details {
		records = append(records, protocore.RecycleRecord{
			BaseRecycleRecord: protocore.BaseRecycleRecord{
				ID:         recycleRecord.ID,
				TaskID:     recycleRecord.TaskID,
				Vendor:     recycleRecord.Vendor,
				ResType:    recycleRecord.ResType,
				ResID:      recycleRecord.ResID,
				CloudResID: recycleRecord.CloudResID,
				ResName:    recycleRecord.ResName,
				BkBizID:    recycleRecord.BkBizID,
				AccountID:  recycleRecord.AccountID,
				Region:     recycleRecord.Region,
				Status:     enumor.RecycleRecordStatus(recycleRecord.Status),
				Revision: core.Revision{
					Creator:   recycleRecord.Creator,
					Reviser:   recycleRecord.Reviser,
					CreatedAt: recycleRecord.CreatedAt.String(),
					UpdatedAt: recycleRecord.UpdatedAt.String(),
				},
			},
			Detail: recycleRecord.Detail,
		})
	}

	return &protodata.ListResult{Details: records}, nil
}

// BatchUpdateRecycleRecord batch update recycle records.
func (svc *recycleRecordSvc) BatchUpdateRecycleRecord(cts *rest.Contexts) (interface{}, error) {
	req := new(protodata.BatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.Data))
	for _, data := range req.Data {
		ids = append(ids, data.ID)
	}

	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id", "detail"},
	}
	res, err := svc.dao.RecycleRecord().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list recycle record failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list recycle record failed, err: %v", err)
	}

	if len(req.Data) != len(res.Details) {
		return nil, fmt.Errorf("list recycle record failed, some recycle record(ids=%+v) doesn't exist", ids)
	}

	detailMap := make(map[string]tabletype.JsonField)
	for _, recycleRecord := range res.Details {
		detailMap[recycleRecord.ID] = recycleRecord.Detail
	}

	record := &prototable.RecycleRecordTable{
		Reviser: cts.Kit.User,
	}
	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, updateReq := range req.Data {
			record.Status = string(updateReq.Status)

			if updateReq.Detail != nil {
				updatedDetail, err := json.UpdateMerge(updateReq.Detail, string(detailMap[updateReq.ID]))
				if err != nil {
					return nil, fmt.Errorf("extension update merge failed, err: %v", err)
				}

				record.Detail = tabletype.JsonField(updatedDetail)
			}

			err = svc.dao.RecycleRecord().Update(cts.Kit, txn, tools.EqualExpression("id", updateReq.ID), record)
			if err != nil {
				logs.Errorf("update recycle record failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, fmt.Errorf("update subnet failed, err: %v", err)
			}
		}

		return nil, nil
	})

	return nil, nil
}
