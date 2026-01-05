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

// Package resplandemand ...
package resplandemand

import (
	"fmt"

	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-demand"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/util"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateResPlanDemand update resource plan demand
func (svc *service) BatchUpdateResPlanDemand(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandBatchUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		_, err := svc.batchUpdateResPlanDemandWithTx(cts.Kit, txn, req.Demands)
		if err != nil {
			logs.Errorf("failed to batch update res plan demand with tx, err: %v, rid: %v", err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("update res plan demand failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *service) batchUpdateResPlanDemandWithTx(kt *kit.Kit, txn *sqlx.Tx,
	updateReqs []rpproto.ResPlanDemandUpdateReq) (
	[]string, error) {

	for _, updateReq := range updateReqs {
		coreType := enumor.CoreType(updateReq.CoreType)
		if coreType != "" {
			if err := coreType.Validate(); err != nil {
				logs.Errorf("invalid core type: %s, rid: %s", coreType, kt.Rid)
				return nil, err
			}
		}

		record := &tablers.ResPlanDemandTable{
			ID:              updateReq.ID,
			OpProductID:     updateReq.OpProductID,
			OpProductName:   updateReq.OpProductName,
			PlanProductID:   updateReq.PlanProductID,
			PlanProductName: updateReq.PlanProductName,
			VirtualDeptID:   updateReq.VirtualDeptID,
			VirtualDeptName: updateReq.VirtualDeptName,
			CoreType:        coreType,
			Reviser:         kt.User,
		}
		if updateReq.OS != nil {
			record.OS = &types.Decimal{Decimal: cvt.PtrToVal(updateReq.OS)}
		}
		if updateReq.CpuCore != nil {
			record.CpuCore = updateReq.CpuCore
		}
		if updateReq.Memory != nil {
			record.Memory = updateReq.Memory
		}
		if updateReq.DiskSize != nil {
			record.DiskSize = updateReq.DiskSize
		}
		// 系统创建的情况下，reviser需要从参数中取
		if updateReq.Reviser != "" {
			record.Reviser = updateReq.Reviser
		}
		// 技术分类
		if updateReq.TechnicalClass != "" {
			record.TechnicalClass = updateReq.TechnicalClass
		}
		// is_confirmed 是否确认
		record.IsConfirmed = updateReq.IsConfirmed

		if err := svc.dao.ResPlanDemand().UpdateWithTx(kt, txn,
			tools.EqualExpression("id", updateReq.ID), record); err != nil {
			logs.Errorf("update res plan demand failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}
	return nil, nil
}

// LockResPlanDemand lock resource plan demand
func (svc *service) LockResPlanDemand(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandLockOpReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(true); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.dao.ResPlanDemand().ExamineAndLockAllRPDemand(cts.Kit, req.LockedItems); err != nil {
		logs.Errorf("lock res plan demand failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// UnlockResPlanDemand unlock resource plan demand
func (svc *service) UnlockResPlanDemand(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandLockOpReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := slice.Map(req.LockedItems, func(item rpproto.ResPlanDemandLockOpItem) string {
		return item.ID
	})

	if err := svc.dao.ResPlanDemand().UnlockAllResPlanDemand(cts.Kit, ids); err != nil {
		logs.Errorf("unlock res plan demand failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchUpsertResPlanDemand batch upsert res plan demand, contains create or update
func (svc *service) BatchUpsertResPlanDemand(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandBatchUpsertReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	createIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		createIDs := make([]string, 0)

		if len(req.CreateDemands) > 0 {
			for _, batch := range slice.Split(req.CreateDemands, constant.BatchOperationMaxLimit) {
				newIDs, err := svc.batchCreateResPlanDemandWithTx(cts.Kit, txn, batch)
				if err != nil {
					logs.Errorf("batch create res plan demand failed, err: %v, rid: %v", err, cts.Kit.Rid)
					return nil, err
				}
				createIDs = append(createIDs, newIDs...)
			}
		}

		if len(req.UpdateDemands) > 0 {
			for _, batch := range slice.Split(req.UpdateDemands, constant.BatchOperationMaxLimit) {
				_, err := svc.batchUpdateResPlanDemandWithTx(cts.Kit, txn, batch)
				if err != nil {
					logs.Errorf("batch update res plan demand failed, err: %v, rid: %v", err, cts.Kit.Rid)
					return nil, err
				}
			}
		}

		return createIDs, nil
	})
	if err != nil {
		logs.Errorf("batch upsert resource plan demand failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ids, err := util.GetStrSliceByInterface(createIDs)
	if err != nil {
		logs.Errorf("upsert resource plan demand but return ids type not []string, err: %v, rid: %s", err,
			cts.Kit.Rid)
		return nil, fmt.Errorf("upsert resource plan demand but return ids type not []string, err: %v", err)
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
