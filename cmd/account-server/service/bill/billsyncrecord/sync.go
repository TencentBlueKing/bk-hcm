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

package billsyncrecord

import (
	"fmt"

	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"

	"github.com/shopspring/decimal"
)

// CreateSyncRecord 创建同步记录
func (b *service) CreateSyncRecord(cts *rest.Contexts) (any, error) {

	req := new(bill.BillSyncRecordCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := b.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Create}})
	if err != nil {
		return nil, err
	}
	currency, cost, rmbCost, err := b.collectAllBillSummaryRoot(cts, req)
	if err != nil {
		return nil, err
	}

	dataReq := &dsbill.BatchBillSyncRecordCreateReq{
		Items: []dsbill.BillSyncRecordCreateReq{
			{
				Vendor:    req.Vendor,
				BillYear:  req.BillYear,
				BillMonth: req.BillMonth,
				State:     enumor.BillSyncRecordStateSyncing,
				Currency:  currency,
				Cost:      *cost,
				RMBCost:   *rmbCost,
				Operator:  cts.Kit.User,
			},
		},
	}
	// TODO: 增加全局锁
	result, err := b.client.DataService().Global.Bill.BatchCreateBillSyncRecord(cts.Kit, dataReq)
	if err != nil {
		logs.Errorf("fail create bill sync record, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := b.client.DataService().Global.Bill.BatchSyncBillSummaryRoot(cts.Kit, &dsbill.BillSummaryBatchSyncReq{
		Vendor:    req.Vendor,
		BillYear:  req.BillYear,
		BillMonth: req.BillMonth,
	}); err != nil {
		logs.Errorf("batch sync bill summary root of req %v failed, err %s, rid: %s", req, err.Error(), cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

func (b *service) collectAllBillSummaryRoot(cts *rest.Contexts, req *bill.BillSyncRecordCreateReq) (
	enumor.CurrencyCode, *decimal.Decimal, *decimal.Decimal, error) {

	cost := decimal.NewFromFloat(0)
	rmbCost := decimal.NewFromFloat(0)
	var currency enumor.CurrencyCode
	expressions := []*filter.AtomRule{
		tools.RuleEqual("vendor", req.Vendor),
		tools.RuleEqual("bill_year", req.BillYear),
		tools.RuleEqual("bill_month", req.BillMonth),
	}
	result, err := b.client.DataService().Global.Bill.ListBillSummaryRoot(cts.Kit, &dsbill.BillSummaryRootListReq{
		Filter: tools.ExpressionAnd(expressions...),
		Page: &core.BasePage{
			Count: true,
		},
	})
	if err != nil {
		logs.Errorf("count bill summary root failed, err %s, rid: %s", err.Error(), cts.Kit.Rid)
		return "", nil, nil, fmt.Errorf("count bill summary root failed, err %s", err.Error())
	}
	if result.Count == nil {
		logs.Errorf("result count is empty, result %v, rid: %s", result, cts.Kit.Rid)
		return "", nil, nil, fmt.Errorf("result count is empty, result %v", result)
	}
	for offset := uint64(0); offset < *result.Count; offset = offset + uint64(core.DefaultMaxPageLimit) {
		result, err := b.client.DataService().Global.Bill.ListBillSummaryRoot(cts.Kit, &dsbill.BillSummaryRootListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Start: uint32(offset),
				Limit: core.DefaultMaxPageLimit,
			},
		})
		if err != nil {
			logs.Errorf("list bill summary root failed, err %s, rid: %s", err.Error(), cts.Kit.Rid)
			return "", nil, nil, fmt.Errorf("list bill summary root failed, err %s", err.Error())
		}
		for _, item := range result.Details {
			if item.State != enumor.RootAccountBillSummaryStateConfirmed {
				logs.Errorf("bill summary root %s is state %s, cannot do sync, rid: %s",
					item.ID, item.State, cts.Kit.Rid)
				return "", nil, nil, fmt.Errorf("bill summary root %s is state %s, cannot do sync",
					item.ID, item.State)
			}
			cost = cost.Add(item.CurrentMonthCost)
			cost = cost.Add(item.AdjustmentCost)
			rmbCost = rmbCost.Add(item.CurrentMonthRMBCost)
			rmbCost = rmbCost.Add(item.AdjustmentRMBCost)
			if len(item.Currency) != 0 {
				currency = item.Currency
			}
		}
	}
	return currency, &cost, &rmbCost, nil
}
