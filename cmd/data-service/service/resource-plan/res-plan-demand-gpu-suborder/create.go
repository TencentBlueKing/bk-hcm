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
	"fmt"

	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/util"

	"github.com/jmoiron/sqlx"
)

// BatchCreateResPlanDemandGpuSubOrder create resource plan demand gpu sub order.
func (svc *service) BatchCreateResPlanDemandGpuSubOrder(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandGpuSubOrderBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	createIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		recordIDs, err := svc.batchCreateResPlanDemandGpuSubOrderWithTx(cts.Kit, txn, req.SubOrders)
		if err != nil {
			logs.Errorf("failed to batch create resource plan demand gpu sub order with tx, err: %v, rid: %s",
				err, cts.Kit.Rid)
			return nil, err
		}
		return recordIDs, nil
	})
	if err != nil {
		logs.Errorf("create resource plan demand gpu sub order failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ids, err := util.GetStrSliceByInterface(createIDs)
	if err != nil {
		logs.Errorf("create resource plan demand gpu sub order but return ids type not []string, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, fmt.Errorf("create resource plan demand gpu sub order but return ids type not []string, err: %v", err)
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func (svc *service) batchCreateResPlanDemandGpuSubOrderWithTx(kt *kit.Kit, txn *sqlx.Tx,
	createReqs []rpproto.ResPlanDemandGpuSubOrderCreateReq) ([]string, error) {

	models := make([]tablers.ResPlanDemandGpuSubOrderTable, len(createReqs))
	for idx, item := range createReqs {
		createT := tablers.ResPlanDemandGpuSubOrderTable{
			OrderID:       item.OrderID,
			BkBizID:       item.BkBizID,
			OpProductID:   item.OpProductID,
			OpProductName: item.OpProductName,
			DemandType:    item.DemandType,
			DemandYear:    item.DemandYear,
			DemandMonth:   item.DemandMonth,
			GPUNum:        item.GPUNum,
			QpmMax:        item.QpmMax,
			Status:        item.Status,
			Comment:       item.Comment,
			Extension:     item.Extension,
			Remark:        item.Remark,
			Creator:       kt.User,
			Reviser:       kt.User,
		}

		if item.Creator != "" {
			createT.Creator = item.Creator
			createT.Reviser = item.Creator
		}

		models[idx] = createT
	}

	recordIDs, err := svc.dao.ResPlanDemandGpuSubOrder().CreateWithTx(kt, txn, models)
	if err != nil {
		logs.Errorf("create resource plan demand gpu sub order failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("create resource plan demand gpu sub order failed, err: %v", err)
	}
	return recordIDs, nil
}
