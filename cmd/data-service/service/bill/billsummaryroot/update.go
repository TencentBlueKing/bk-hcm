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

package billsummaryroot

import (
	"fmt"

	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	daotypes "hcm/pkg/dal/dao/types"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// UpdateBillSummaryRoot update bill summary of root account with options
func (svc *service) UpdateBillSummaryRoot(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.BillSummaryRootUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	billSummaryRoot := &tablebill.AccountBillSummaryRoot{
		ID:                req.ID,
		RootAccountName:   req.RootAccountName,
		LastSyncedVersion: req.LastSyncedVersion,
		CurrentVersion:    req.CurrentVersion,
		Currency:          req.Currency,
		MonthOnMonthValue: req.MonthOnMonthValue,
		Rate:              req.Rate,
		State:             req.State,
		BkBizNum:          req.BkBizNum,
		ProductNum:        req.ProductNum,
	}
	if req.LastMonthCostSynced != nil {
		billSummaryRoot.LastMonthCostSynced = &types.Decimal{Decimal: *req.LastMonthCostSynced}
	}
	if req.LastMonthRMBCostSynced != nil {
		billSummaryRoot.LastMonthRMBCostSynced = &types.Decimal{Decimal: *req.LastMonthRMBCostSynced}
	}
	if req.CurrentMonthCostSynced != nil {
		billSummaryRoot.CurrentMonthCostSynced = &types.Decimal{Decimal: *req.CurrentMonthCostSynced}
	}
	if req.CurrentMonthRMBCostSynced != nil {
		billSummaryRoot.CurrentMonthRMBCostSynced = &types.Decimal{Decimal: *req.CurrentMonthRMBCostSynced}
	}
	if req.CurrentMonthCost != nil {
		billSummaryRoot.CurrentMonthCost = &types.Decimal{Decimal: *req.CurrentMonthCost}
	}
	if req.CurrentMonthRMBCost != nil {
		billSummaryRoot.CurrentMonthRMBCost = &types.Decimal{Decimal: *req.CurrentMonthRMBCost}
	}
	if req.AjustmentCost != nil {
		billSummaryRoot.AjustmentCost = &types.Decimal{Decimal: *req.AjustmentCost}
	}
	if req.AjustmentRMBCost != nil {
		billSummaryRoot.AjustmentRMBCost = &types.Decimal{Decimal: *req.AjustmentRMBCost}
	}
	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if err := svc.dao.AccountBillSummaryRoot().UpdateByIDWithTx(
			cts.Kit, txn, billSummaryRoot.ID, billSummaryRoot); err != nil {
			return nil, fmt.Errorf("update bill summary of root account failed, err: %v", err)
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchSyncBillSummaryRoot batch update bill summary of root account to syncing state
func (svc *service) BatchSyncBillSummaryRoot(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.BillSummaryBatchSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	expressions := []*filter.AtomRule{
		tools.RuleEqual("vendor", req.Vendor),
		tools.RuleEqual("bill_year", req.BillYear),
		tools.RuleEqual("bill_month", req.BillMonth),
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		countOpt := &daotypes.ListOption{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Count: true,
			},
		}
		result, err := svc.dao.AccountBillSummaryRoot().ListWithTx(cts.Kit, txn, countOpt)
		if err != nil {
			logs.Errorf(
				"list bill summary root txn by option failed, err %s, rid: %s", err.Error(), cts.Kit.Rid)
			return nil, fmt.Errorf(
				"list bill summary root txn by option failed, err %s", err.Error())
		}
		if len(result.Details) == 0 {
			return nil, nil
		}
		var ids []string
		for offset := uint64(0); offset < result.Count; offset = offset + uint64(core.DefaultMaxPageLimit) {
			listOpt := &daotypes.ListOption{
				Filter: tools.ExpressionAnd(expressions...),
				Page: &core.BasePage{
					Start: uint32(offset),
					Limit: core.DefaultMaxPageLimit,
				},
			}
			tmpResult, err := svc.dao.AccountBillSummaryRoot().ListWithTx(cts.Kit, txn, listOpt)
			if err != nil {
				logs.Errorf(
					"list bill summary root txn by option failed, err %s, rid: %s", err.Error(), cts.Kit.Rid)
				return nil, fmt.Errorf(
					"list bill summary root txn by option failed, err %s", err.Error())
			}
			for _, item := range tmpResult.Details {
				if item.State != constant.RootAccountBillSummaryStateConfirmed {
					logs.Errorf("bill root summary of %s %s %d-%d is not confirmed, cannot do sync, rid: %s",
						item.RootAccountID, item.Vendor, item.BillYear, item.BillMonth, cts.Kit.Rid)
					return nil, fmt.Errorf("bill root summary of %s %s %d-%d is not confirmed, cannot do sync",
						item.RootAccountID, item.Vendor, item.BillYear, item.BillMonth)
				}
				ids = append(ids, item.ID)
			}
		}

		for _, id := range ids {
			updateReq := &tablebill.AccountBillSummaryRoot{
				ID:    id,
				State: constant.RootAccountBillSummaryStateSyncing,
			}
			if err := svc.dao.AccountBillSummaryRoot().UpdateByIDWithTx(cts.Kit, txn, id, updateReq); err != nil {
				logs.Errorf("fail to set bill summary root state to syncing, err: %v,rid: %v", err, cts.Kit.Rid)
				return nil, fmt.Errorf("fail to set bill summary root state to syncing, err: %v", err)
			}
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}
