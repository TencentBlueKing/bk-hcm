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
	asbillapi "hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	dsbillapi "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"

	"github.com/shopspring/decimal"
)

// SumRootAccountSummary Summarize the main account bill
func (s *service) SumRootAccountSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(asbillapi.RootAccountSummarySumReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	expressions := []filter.RuleFactory{
		tools.RuleEqual("bill_year", req.BillYear),
		tools.RuleEqual("bill_month", req.BillMonth),
	}
	if req.Filter != nil {
		expressions = append(expressions, req.Filter)
	}
	bizFilter, err := tools.And(
		expressions...)
	if err != nil {
		return nil, err
	}

	result, err := s.client.DataService().Global.Bill.ListBillSummaryRoot(cts.Kit, &dsbillapi.BillSummaryRootListReq{
		Filter: bizFilter,
		Page: &core.BasePage{
			Count: true,
		},
	})
	if err != nil {
		return nil, err
	}
	var rootSummaryList []*billcore.SummaryRoot
	for offset := uint64(0); offset < *result.Count; offset = offset + uint64(core.DefaultMaxPageLimit) {
		tmpResult, err := s.client.DataService().Global.Bill.ListBillSummaryRoot(cts.Kit,
			&dsbillapi.BillSummaryRootListReq{
				Filter: bizFilter,
				Page: &core.BasePage{
					Start: uint32(offset),
					Limit: core.DefaultMaxPageLimit,
				},
			})
		if err != nil {
			return nil, err
		}
		rootSummaryList = append(rootSummaryList, tmpResult.Details...)
	}
	return s.doCalcalcute(rootSummaryList, *result.Count)
}

func (s *service) doCalcalcute(rootSummaryList []*billcore.SummaryRoot, count uint64) (interface{}, error) {
	retMap := make(map[enumor.CurrencyCode]*billcore.CostWithCurrency)
	for _, rootSummary := range rootSummaryList {
		if _, ok := retMap[rootSummary.Currency]; !ok {
			retMap[enumor.CurrencyCode(rootSummary.Currency)] = &billcore.CostWithCurrency{
				Cost:     decimal.NewFromFloat(0),
				RMBCost:  decimal.NewFromFloat(0),
				Currency: rootSummary.Currency,
			}
		}
		retMap[rootSummary.Currency].Cost = retMap[rootSummary.Currency].Cost.Add(rootSummary.CurrentMonthCost)
		retMap[rootSummary.Currency].RMBCost = retMap[rootSummary.Currency].RMBCost.Add(rootSummary.CurrentMonthRMBCost)
	}
	return &asbillapi.RootAccountSummarySumResult{
		Count:   count,
		CostMap: retMap,
	}, nil
}
