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

// Package resplandemandgpusuborder ...
package resplandemandgpusuborder

import (
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateResPlanDemandGpuSubOrder update resource plan demand gpu sub order.
func (svc *service) BatchUpdateResPlanDemandGpuSubOrder(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandGpuSubOrderBatchUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		_, err := svc.batchUpdateResPlanDemandGpuSubOrderWithTx(cts.Kit, txn, req.SubOrders)
		if err != nil {
			logs.Errorf("failed to batch update res plan demand gpu sub order with tx, err: %v, rid: %v",
				err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("update res plan demand gpu sub order failed, err: %v, rid: %v", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *service) batchUpdateResPlanDemandGpuSubOrderWithTx(kt *kit.Kit, txn *sqlx.Tx,
	updateReqs []rpproto.ResPlanDemandGpuSubOrderUpdateReq) ([]string, error) {

	for _, updateReq := range updateReqs {
		record := &tablers.ResPlanDemandGpuSubOrderTable{
			BkBizID:       updateReq.BkBizID,
			OpProductID:   updateReq.OpProductID,
			OpProductName: updateReq.OpProductName,
			TemplateID:    updateReq.TemplateID,
			DemandType:    updateReq.DemandType,
			DemandYear:    updateReq.DemandYear,
			DemandMonth:   updateReq.DemandMonth,
			GPUNum:        updateReq.GPUNum,
			QpmMax:        updateReq.QpmMax,
			Status:        updateReq.Status,
			Remark:        updateReq.Remark,
			Reviser:       kt.User,
		}
		if updateReq.Comment != nil {
			record.Comment = updateReq.Comment
		}
		if updateReq.Extension != nil {
			record.Extension = cvt.PtrToVal(updateReq.Extension)
		}

		if err := svc.dao.ResPlanDemandGpuSubOrder().UpdateWithTx(kt, txn,
			tools.EqualExpression("id", updateReq.ID), record); err != nil {
			logs.Errorf("update res plan demand gpu sub order failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}
	return nil, nil
}

// BatchUpdateStatusResPlanDemandGpuSubOrder update all sub orders in the id list to the same status.
func (svc *service) BatchUpdateStatusResPlanDemandGpuSubOrder(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandGpuSubOrderBatchUpdateStatusReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	record := &tablers.ResPlanDemandGpuSubOrderTable{
		Status:  req.Status,
		Reviser: cts.Kit.User,
	}
	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if err := svc.dao.ResPlanDemandGpuSubOrder().UpdateWithTx(cts.Kit, txn,
			tools.ContainersExpression("id", req.IDs), record); err != nil {
			logs.Errorf("batch update res plan demand gpu sub order status failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch update res plan demand gpu sub order status failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
