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

package rootsummary

import (
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/shopspring/decimal"
)

// RootAccountSummaryActionOption option for root account summary action
type RootAccountSummaryActionOption struct {
	RootAccountID string        `json:"root_account_id" validate:"required"`
	BillYear      int           `json:"bill_year" validate:"required"`
	BillMonth     int           `json:"bill_month" validate:"required"`
	Vendor        enumor.Vendor `json:"vendor" validate:"required"`
}

// RootAccountSummaryAction define root account summary
type RootAccountSummaryAction struct{}

// ParameterNew return request params.
func (act RootAccountSummaryAction) ParameterNew() interface{} {
	return new(RootAccountSummaryActionOption)
}

// Name return action name
func (act RootAccountSummaryAction) Name() enumor.ActionName {
	return enumor.ActionRootAccountSummary
}

// Run run task
func (act RootAccountSummaryAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*RootAccountSummaryActionOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}
	summary, err := act.getBillSummary(kt.Kit(), opt)
	if err != nil {
		logs.Warnf("get bill summary for opt %v failed, err %s, rid: %s", opt, err.Error(), kt.Kit().Rid)
		return nil, fmt.Errorf("get bill summary for opt %v failed, err %s, rid: %s", opt, err.Error(), kt.Kit().Rid)
	}
	// 获取一级账号信息
	rAccountResult, err := actcli.GetDataService().Global.RootAccount.GetBasicInfo(kt.Kit(), summary.RootAccountID)
	if err != nil {
		logs.Warnf("get root account with id %s failed, err %s, rid: %s",
			summary.RootAccountID, err.Error(), kt.Kit().Rid)
		return nil, fmt.Errorf("get root account with id %s failed, err %s, rid: %s",
			summary.RootAccountID, err.Error(), kt.Kit().Rid)
	}

	mainSummaryList, err := act.listAllMainSummary(kt.Kit(), opt)
	if err != nil {
		logs.Warnf("list main account summary of opt %v, err %s, rid: %s", opt, err.Error(), kt.Kit().Rid)
		return nil, fmt.Errorf("list main account summary of opt %v, err %s, rid: %s", opt, err.Error(), kt.Kit().Rid)
	}

	rate := float64(0)
	currency := enumor.CurrencyUSD
	lastMonthSyncedCost := decimal.NewFromFloat(0)
	lastMonthSyncedRMBCost := decimal.NewFromFloat(0)
	currentCostSynced := decimal.NewFromFloat(0)
	currentCostRMBSynced := decimal.NewFromFloat(0)
	currentCost := decimal.NewFromFloat(0)
	currentRMBCost := decimal.NewFromFloat(0)
	isAccounted := true
	bkBizNum := uint64(0)
	productNum := uint64(0)
	for _, mainSummary := range mainSummaryList {
		if mainSummary.State != constant.MainAccountBillSummaryStateAccounted {
			isAccounted = false
		}
		currency = mainSummary.Currency
		rate = mainSummary.Rate
		lastMonthSyncedCost = lastMonthSyncedCost.Add(mainSummary.LastMonthCostSynced)
		lastMonthSyncedRMBCost = lastMonthSyncedRMBCost.Add(mainSummary.LastMonthRMBCostSynced)
		currentCostSynced = currentCostSynced.Add(mainSummary.CurrentMonthCostSynced)
		currentCostRMBSynced = currentCostRMBSynced.Add(mainSummary.CurrentMonthCostSynced)
		currentCost = currentCost.Add(mainSummary.CurrentMonthCost)
		currentRMBCost = currentRMBCost.Add(mainSummary.CurrentMonthRMBCost)
		if mainSummary.BkBizID > 0 {
			bkBizNum = bkBizNum + 1
		}
		if mainSummary.ProductID > 0 {
			productNum = productNum + 1
		}
	}
	if isAccounted {
		// 防止主账号账单汇总还没有创建的，判断都已经核算完成了
		mainAccountCount, err := act.countMainAccount(kt.Kit(), opt)
		if err != nil {
			return nil, err
		}
		if len(mainSummaryList) != int(mainAccountCount) {
			isAccounted = false
		}
	}
	req := &bill.BillSummaryRootUpdateReq{
		ID:                        summary.ID,
		Currency:                  currency,
		RootAccountName:           rAccountResult.Name,
		LastMonthCostSynced:       &lastMonthSyncedCost,
		LastMonthRMBCostSynced:    &lastMonthSyncedRMBCost,
		CurrentMonthCostSynced:    &currentCostSynced,
		CurrentMonthRMBCostSynced: &currentCostRMBSynced,
		CurrentMonthCost:          &currentCost,
		CurrentMonthRMBCost:       &currentRMBCost,
		Rate:                      rate,
		BkBizNum:                  bkBizNum,
		ProductNum:                productNum,
	}
	if !lastMonthSyncedCost.IsZero() {
		req.MonthOnMonthValue = currentCostSynced.DivRound(lastMonthSyncedCost, 5).InexactFloat64()
	}
	if isAccounted {
		req.State = constant.RootAccountBillSummaryStateAccounted
	} else {
		req.State = constant.RootAccountBillSummaryStateAccounting
	}
	if err := actcli.GetDataService().Global.Bill.UpdateBillSummaryRoot(kt.Kit(), req); err != nil {
		logs.Warnf("failed to update root account bill summary %v, err %s, rid: %s", opt, err.Error(), kt.Kit().Rid)
		return nil, fmt.Errorf("failed to update root account bill summary %v, err %s", opt, err.Error())
	}
	logs.Infof("sucessfully update root account bill summary %v", req)
	return nil, nil
}

func (act *RootAccountSummaryAction) getBillSummary(
	kt *kit.Kit, opt *RootAccountSummaryActionOption) (*bill.BillSummaryRootResult, error) {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("vendor", opt.Vendor),
		tools.RuleEqual("bill_year", opt.BillYear),
		tools.RuleEqual("bill_month", opt.BillMonth),
	}
	result, err := actcli.GetDataService().Global.Bill.ListBillSummaryRoot(
		kt, &bill.BillSummaryRootListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("get root account bill summary failed, err %s", err.Error())
	}
	if len(result.Details) != 1 {
		return nil, fmt.Errorf("get invalid length root account bill summary resp %+v", result)
	}
	return result.Details[0], nil
}

func (act *RootAccountSummaryAction) countMainAccount(
	kt *kit.Kit, opt *RootAccountSummaryActionOption) (uint64, error) {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("parent_account_id", opt.RootAccountID),
		tools.RuleEqual("vendor", opt.Vendor),
	}
	result, err := actcli.GetDataService().Global.MainAccount.List(kt, &core.ListReq{
		Filter: tools.ExpressionAnd(expressions...),
		Page: &core.BasePage{
			Count: true,
		},
	})
	if err != nil {
		return 0, err
	}
	return result.Count, nil
}

func (act *RootAccountSummaryAction) listAllMainSummary(
	kt *kit.Kit, opt *RootAccountSummaryActionOption) ([]*bill.BillSummaryMainResult, error) {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("vendor", opt.Vendor),
		tools.RuleEqual("bill_year", opt.BillYear),
		tools.RuleEqual("bill_month", opt.BillMonth),
	}
	result, err := actcli.GetDataService().Global.Bill.ListBillSummaryMain(
		kt, &bill.BillSummaryMainListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Count: true,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("list main account bill summary of %v failed, err %s", opt, err.Error())
	}
	if result.Count == 0 {
		return nil, fmt.Errorf("empty count in result %+v", result)
	}
	logs.Infof("found %d main account summary for opt: %v, rid: %s", result.Count, opt, kt.Rid)
	var mainSummaryList []*bill.BillSummaryMainResult
	for offset := uint64(0); offset < result.Count; offset = offset + uint64(core.DefaultMaxPageLimit) {
		result, err = actcli.GetDataService().Global.Bill.ListBillSummaryMain(
			kt, &bill.BillSummaryMainListReq{
				Filter: tools.ExpressionAnd(expressions...),
				Page: &core.BasePage{
					Start: 0,
					Limit: core.DefaultMaxPageLimit,
				},
			})
		if err != nil {
			return nil, fmt.Errorf("list main account bill summary of %v failed, err %s", opt, err.Error())
		}
		mainSummaryList = append(mainSummaryList, result.Details...)
	}
	return mainSummaryList, nil
}
