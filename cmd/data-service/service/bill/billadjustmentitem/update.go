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

package billadjustmentitem

import (
	"fmt"

	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// UpdateBillAdjustmentItem account with options
func (svc *service) UpdateBillAdjustmentItem(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.BillAdjustmentItemUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	adjustment := &tablebill.AccountBillAdjustmentItem{
		ID:            req.ID,
		RootAccountID: req.RootAccountID,
		MainAccountID: req.MainAccountID,
		ProductID:     req.ProductID,
		BkBizID:       req.BkBizID,
		BillYear:      req.BillYear,
		BillMonth:     req.BillMonth,
		BillDay:       req.BillDay,
		Type:          string(req.Type),
		Memo:          req.Memo,
		Operator:      req.Operator,
		Currency:      req.Currency,
		State:         req.State,
	}
	if req.Cost != nil {
		adjustment.Cost = &types.Decimal{Decimal: *req.Cost}
	}
	if req.RMBCost != nil {
		adjustment.RMBCost = &types.Decimal{Decimal: *req.RMBCost}
	}
	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if err := svc.dao.AccountBillAdjustmentItem().UpdateByIDWithTx(
			cts.Kit, txn, adjustment.ID, adjustment); err != nil {
			return nil, fmt.Errorf("update bill adjustment item failed, err: %v", err)
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchConfirmBillAdjustmentItem batch confirm
func (svc *service) BatchConfirmBillAdjustmentItem(cts *rest.Contexts) (any, error) {
	req := new(core.BatchDeleteReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {

		for _, id := range req.IDs {
			updateReq := &tablebill.AccountBillAdjustmentItem{
				ID:    id,
				State: enumor.BillAdjustmentStateConfirmed,
			}
			if err := svc.dao.AccountBillAdjustmentItem().UpdateByIDWithTx(cts.Kit, txn, id, updateReq); err != nil {
				logs.Errorf("fail to set bill adjustment item state to confirm, err: %v,rid: %v", err, cts.Kit.Rid)
				return nil, fmt.Errorf("confirm bill adjustment item failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
