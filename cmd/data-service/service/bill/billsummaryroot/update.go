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

	dataservice "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/rest"

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
		ID:                        req.ID,
		RootAccountID:             req.RootAccountID,
		RootAccountName:           req.RootAccountName,
		Vendor:                    req.Vendor,
		BillYear:                  req.BillYear,
		BillMonth:                 req.BillMonth,
		LastSyncedVersion:         req.LastSyncedVersion,
		CurrentVersion:            req.CurrentVersion,
		Currency:                  req.Currency,
		LastMonthCostSynced:       &types.Decimal{Decimal: req.LastMonthCostSynced},
		LastMonthRMBCostSynced:    &types.Decimal{Decimal: req.LastMonthRMBCostSynced},
		CurrentMonthCostSynced:    &types.Decimal{Decimal: req.CurrentMonthCostSynced},
		CurrentMonthRMBCostSynced: &types.Decimal{Decimal: req.CurrentMonthRMBCostSynced},
		MonthOnMonthValue:         req.MonthOnMonthValue,
		CurrentMonthCost:          &types.Decimal{Decimal: req.CurrentMonthCost},
		CurrentMonthRMBCost:       &types.Decimal{Decimal: req.CurrentMonthRMBCost},
		Rate:                      req.Rate,
		AjustmentCost:             &types.Decimal{Decimal: req.AjustmentCost},
		AjustmentRMBCost:          &types.Decimal{Decimal: req.AjustmentRMBCost},
		State:                     req.State,
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
