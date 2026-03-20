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
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/util"

	"github.com/jmoiron/sqlx"
)

// OverwriteResPlanDemandGpuSubOrders atomically deletes all existing sub orders of the given order
// and creates new ones within a single database transaction.
func (svc *service) OverwriteResPlanDemandGpuSubOrders(cts *rest.Contexts) (interface{}, error) {
	req := new(rpproto.ResPlanDemandGpuSubOrderOverwriteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	delIDs, err := svc.listSubOrderIDsByOrderID(cts.Kit, req.OrderID)
	if err != nil {
		return nil, err
	}

	createIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, batchIDs := range slice.Split(delIDs, constant.BatchOperationMaxLimit) {
			delFilter := tools.ContainersExpression("id", batchIDs)
			if err := svc.dao.ResPlanDemandGpuSubOrder().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
				return nil, err
			}
		}

		recordIDs, err := svc.batchCreateResPlanDemandGpuSubOrderWithTx(cts.Kit, txn, req.SubOrders)
		if err != nil {
			logs.Errorf("failed to create sub orders in overwrite tx, orderID: %s, err: %v, rid: %s",
				req.OrderID, err, cts.Kit.Rid)
			return nil, err
		}

		return recordIDs, nil
	})
	if err != nil {
		logs.Errorf("overwrite resource plan demand gpu sub orders failed, orderID: %s, err: %v, rid: %s",
			req.OrderID, err, cts.Kit.Rid)
		return nil, err
	}

	ids, err := util.GetStrSliceByInterface(createIDs)
	if err != nil {
		logs.Errorf("overwrite resource plan demand gpu sub orders returned non-[]string ids, "+
			"orderID: %s, err: %v, rid: %s", req.OrderID, err, cts.Kit.Rid)
		return nil, fmt.Errorf("overwrite resource plan demand gpu sub orders returned non-[]string ids, err: %v", err)
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// listSubOrderIDsByOrderID collects all sub order IDs belonging to the given order (paginated).
func (svc *service) listSubOrderIDsByOrderID(kt *kit.Kit, orderID string) ([]string, error) {
	listOpt := &types.ListOption{
		Filter: tools.EqualExpression("order_id", orderID),
		Page:   core.NewDefaultBasePage(),
	}

	ids := make([]string, 0)
	for {
		listResp, err := svc.dao.ResPlanDemandGpuSubOrder().List(kt, listOpt)
		if err != nil {
			logs.Errorf("list resource plan demand gpu sub orders failed, orderID: %s, err: %v, rid: %s",
				orderID, err, kt.Rid)
			return nil, fmt.Errorf("list resource plan demand gpu sub orders failed, err: %v", err)
		}

		for _, one := range listResp.Details {
			ids = append(ids, one.ID)
		}

		if len(listResp.Details) < int(listOpt.Page.Limit) {
			break
		}
		listOpt.Page.Start += uint32(listOpt.Page.Limit)
	}

	return ids, nil
}
