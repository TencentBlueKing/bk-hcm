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

package billsummarymain

import (
	"fmt"

	dataservice "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// UpdateBillSummary update account bill summary main with options
func (svc *service) UpdateBillSummaryMain(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.BillSummaryMainUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	billSummaryMain := &tablebill.AccountBillSummaryMain{
		ID:                req.ID,
		RootAccountName:   req.RootAccountName,
		MainAccountName:   req.MainAccountName,
		ProductID:         req.ProductID,
		ProductName:       req.ProductName,
		BkBizID:           req.BkBizID,
		BkBizName:         req.BkBizName,
		LastSyncedVersion: req.LastSyncedVersion,
		CurrentVersion:    req.CurrentVersion,
		Currency:          req.Currency,
		MonthOnMonthValue: req.MonthOnMonthValue,
		Rate:              req.Rate,
		State:             req.State,
	}
	if req.LastMonthCostSynced != nil {
		billSummaryMain.LastMonthCostSynced = &types.Decimal{Decimal: *req.LastMonthCostSynced}
	}
	if req.LastMonthRMBCostSynced != nil {
		billSummaryMain.LastMonthRMBCostSynced = &types.Decimal{Decimal: *req.LastMonthRMBCostSynced}
	}
	if req.CurrentMonthCostSynced != nil {
		billSummaryMain.CurrentMonthCostSynced = &types.Decimal{Decimal: *req.CurrentMonthCostSynced}
	}
	if req.CurrentMonthRMBCostSynced != nil {
		billSummaryMain.CurrentMonthRMBCostSynced = &types.Decimal{Decimal: *req.CurrentMonthRMBCostSynced}
	}
	if req.CurrentMonthCost != nil {
		billSummaryMain.CurrentMonthCost = &types.Decimal{Decimal: *req.CurrentMonthCost}
	}
	if req.CurrentMonthRMBCost != nil {
		billSummaryMain.CurrentMonthRMBCost = &types.Decimal{Decimal: *req.CurrentMonthRMBCost}
	}
	if req.AdjustmentCost != nil {
		billSummaryMain.AdjustmentCost = &types.Decimal{Decimal: *req.AdjustmentCost}
	}
	if req.AdjustmentRMBCost != nil {
		billSummaryMain.AdjustmentRMBCost = &types.Decimal{Decimal: *req.AdjustmentRMBCost}
	}
	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if err := svc.dao.AccountBillSummaryMain().UpdateByIDWithTx(
			cts.Kit, txn, billSummaryMain.ID, billSummaryMain); err != nil {
			return nil, fmt.Errorf("update bill summary main failed, err: %v", err)
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
