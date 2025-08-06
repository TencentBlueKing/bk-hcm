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

// Package mainsummary ...
package mainsummary

import (
	"fmt"
	"time"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/times"

	"github.com/shopspring/decimal"
)

// MainAccountSummaryActionOption option for main account summary action
type MainAccountSummaryActionOption struct {
	RootAccountID string        `json:"root_account_id" validate:"required"`
	MainAccountID string        `json:"main_account_id" validate:"required"`
	BillYear      int           `json:"bill_year" validate:"required"`
	BillMonth     int           `json:"bill_month" validate:"required"`
	Vendor        enumor.Vendor `json:"vendor" validate:"required"`

	MonthTaskTypes []enumor.MonthTaskType `json:"month_task_types"`
}

// String ...
func (opt MainAccountSummaryActionOption) String() string {

	return fmt.Sprintf("{%s/%s/%s %d-%02d}",
		opt.Vendor, opt.RootAccountID, opt.MainAccountID, opt.BillYear, opt.BillMonth)
}

// Validate ...
func (opt *MainAccountSummaryActionOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// MainAccountSummaryAction define main account summary action
type MainAccountSummaryAction struct{}

// ParameterNew return request params.
func (act MainAccountSummaryAction) ParameterNew() interface{} {
	return new(MainAccountSummaryActionOption)
}

// Name return action name
func (act MainAccountSummaryAction) Name() enumor.ActionName {
	return enumor.ActionMainAccountSummary
}

// Run task
func (act MainAccountSummaryAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*MainAccountSummaryActionOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}
	if err := opt.Validate(); err != nil {
		return nil, err
	}
	rootSummary, summary, err := act.getRootSummary(kt.Kit(), opt)
	if err != nil {
		logs.Errorf("fail to get root summary for main summary, err: %v, rid: %s", err, kt.Kit().Rid)
		return nil, err
	}

	// 获取主账号信息
	mAccountResult, err := actcli.GetDataService().Global.MainAccount.GetBasicInfo(kt.Kit(), summary.MainAccountID)
	if err != nil {
		return nil, err
	}

	// 计算上月同步成本
	lastMonthCostSynced, lastMonthRMBCostSynced, err := act.getLastMonthSyncedCost(kt.Kit(), opt)
	if err != nil {
		return nil, fmt.Errorf("get last month synced cost failed, err %s", err.Error())
	}

	// 计算当月已同步成本
	var curMonthCostSynced *decimal.Decimal
	// 主账号账单已处于确认或者同步状态，则计算已同步成本
	if rootSummary.State == enumor.RootAccountBillSummaryStateConfirmed ||
		rootSummary.State == enumor.RootAccountBillSummaryStateSyncing ||
		rootSummary.State == enumor.RootAccountBillSummaryStateSynced {
		curMonthCostSynced, _, _, err = act.getDailyVersionCost(kt.Kit(), opt, summary.CurrentVersion)
		if err != nil {
			logs.Errorf("fail get cur month synced cost faild, err: %v, rid: %s", err, kt.Kit().Rid)
			return nil, fmt.Errorf("get current month synced cost failed, err %s", err.Error())
		}
	}

	// 计算当月实时成本
	curMonthCost, isCurMonthAccounted, currency, err := act.getDailyVersionCost(kt.Kit(), opt, summary.CurrentVersion)
	if err != nil {
		logs.Errorf("fail get current month cost failed, err: %v, rid: %s", err, kt.Kit().Rid)
		return nil, fmt.Errorf("get current month cost failed, err %s", err.Error())
	}

	// 获取当月平均汇率
	var exchangeRate *decimal.Decimal
	if len(currency) != 0 {
		exchangeRate, err = act.getExchangeRate(kt.Kit(), currency, enumor.CurrencyRMB, opt.BillYear, opt.BillMonth)
		if err != nil {
			return nil, err
		}
	}
	// 计算调账成本
	adjCost, err := act.getAdjustmentSummary(kt.Kit(), opt, currency)
	if err != nil {
		return nil, err
	}
	req := &bill.BillSummaryMainUpdateReq{
		ID:                     summary.ID,
		ProductID:              mAccountResult.OpProductID,
		BkBizID:                mAccountResult.BkBizID,
		Currency:               currency,
		CurrentMonthCost:       curMonthCost,
		CurrentMonthCostSynced: curMonthCostSynced,
		LastMonthRMBCostSynced: lastMonthRMBCostSynced,
		AdjustmentCost:         adjCost,
	}

	if curMonthCostSynced != nil && lastMonthCostSynced != nil && !lastMonthCostSynced.IsZero() {
		req.LastMonthCostSynced = lastMonthCostSynced
		req.MonthOnMonthValue = curMonthCostSynced.DivRound(*lastMonthCostSynced, 5).InexactFloat64()
	}

	if isCurMonthAccounted {
		// 如果当月所有日账单都已经分账，那么就获取月度账单状态
		extraCost, isFinished, err := act.calculateMonthTaskStatus(kt.Kit(), rootSummary, summary, opt.MonthTaskTypes)
		if err != nil {
			logs.Errorf("failed to check if month pull task finished, err: %v, rid: %s", err, kt.Kit().Rid)
			return nil, err
		}
		if isFinished {
			req.CurrentMonthCost = cvt.ValToPtr(extraCost.Add(cvt.PtrToVal(req.CurrentMonthCost)))
			req.State = enumor.MainAccountBillSummaryStateAccounted
		} else {
			req.State = enumor.MainAccountBillSummaryStateWaitMonthTask
		}
	}
	req = calRMBCost(req, exchangeRate)
	if err := actcli.GetDataService().Global.Bill.UpdateBillSummaryMain(kt.Kit(), req); err != nil {
		logs.Errorf("failed to update main account bill summary %+v, err: %v, rid: %s", opt, err, kt.Kit().Rid)
		return nil, fmt.Errorf("failed to update main account bill summary %+v, err %v", opt, err)
	}
	logs.Infof("sucessfully update main account bill summary %+v, rid: %s", req, kt.Kit().Rid)
	return nil, nil
}

// calRMBCost calculate RMB cost based on exchange rate
func calRMBCost(req *bill.BillSummaryMainUpdateReq, exchangeRate *decimal.Decimal) *bill.BillSummaryMainUpdateReq {
	if exchangeRate == nil {
		return req
	}
	req.AdjustmentRMBCost = cvt.ValToPtr(req.AdjustmentCost.Mul(*exchangeRate))
	if req.CurrentMonthCost != nil {
		req.CurrentMonthRMBCost = cvt.ValToPtr(req.CurrentMonthCost.Mul(*exchangeRate))
	}
	if req.CurrentMonthCostSynced != nil {
		req.CurrentMonthRMBCostSynced = cvt.ValToPtr(req.CurrentMonthCostSynced.Mul(*exchangeRate))
	}
	return req
}

// getExchangeRate get exchange rate from data service
func (act *MainAccountSummaryAction) getExchangeRate(kt *kit.Kit, fromCurrency, toCurrency enumor.CurrencyCode,
	billYear, billMonth int) (*decimal.Decimal, error) {

	if fromCurrency == toCurrency {
		one := decimal.NewFromInt(1)
		return &one, nil
	}

	expressions := []*filter.AtomRule{
		tools.RuleEqual("from_currency", fromCurrency),
		tools.RuleEqual("to_currency", toCurrency),
		tools.RuleEqual("year", billYear),
		tools.RuleEqual("month", billMonth),
	}
	result, err := actcli.GetDataService().Global.Bill.ListExchangeRate(kt, &core.ListReq{
		Filter: tools.ExpressionAnd(expressions...),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get exchange rate from %s to %s in %d-%d failed, err %s",
			fromCurrency, toCurrency, billYear, billMonth, err.Error())
	}
	if len(result.Details) == 0 {
		logs.Infof("get no exchange rate from %s to %s in %d-%d, rid %s",
			fromCurrency, toCurrency, billYear, billMonth, kt.Rid)
		return nil, nil
	}
	if len(result.Details) != 1 {
		logs.Infof("get invalid resp length from exchange rate from %s to %s in %d-%d, resp %v, rid %s",
			fromCurrency, toCurrency, billYear, billMonth, result.Details, kt.Rid)
		return nil, fmt.Errorf("get invalid resp length from exchange rate from %s to %s in %d-%d, resp %v",
			fromCurrency, toCurrency, billYear, billMonth, result.Details)
	}
	return result.Details[0].ExchangeRate, nil
}

// calculateMonthTaskStatus calculate month task status and cost
func (act *MainAccountSummaryAction) calculateMonthTaskStatus(kt *kit.Kit, summaryRoot *billcore.SummaryRoot,
	summary *bill.BillSummaryMain, monthTaskTypes []enumor.MonthTaskType) (extraCost decimal.Decimal, isFinished bool,
	err error) {

	if len(monthTaskTypes) == 0 {
		// unsupported, skip
		logs.Infof("[%s] %s(%s)/%s(%s) do not support month task, skip, rid: %s ", summaryRoot.Vendor,
			summary.RootAccountCloudID, summary.RootAccountID, summary.MainAccountCloudID, summary.MainAccountCloudID,
			kt.Rid)
		return decimal.Zero, true, nil
	}
	monthTasks, err := getMonthTask(kt, summaryRoot, monthTaskTypes)
	if err != nil {
		return decimal.Zero, false, err
	}
	if len(monthTasks) != len(monthTaskTypes) {
		logs.Infof("[%s] %s(%s) %d-%02d month task length not match, got: %d, want: %d, rid: %s",
			summary.Vendor, summary.RootAccountCloudID, summary.RootAccountID,
			summary.BillYear, summary.BillMonth, len(monthTasks), len(monthTaskTypes), kt.Rid)
		return decimal.Zero, false, nil
	}
	mtNameMap := make(map[enumor.MonthTaskType]struct{}, len(monthTaskTypes))
	for _, mt := range monthTasks {
		mtNameMap[mt.Type] = struct{}{}
	}
	cost := decimal.Zero
	for _, monthTask := range monthTasks {
		if _, ok := mtNameMap[monthTask.Type]; !ok {
			return decimal.Zero, false, fmt.Errorf("get invalid month task type: %s ", monthTask.Type)
		}
		if monthTask.State != enumor.RootAccountMonthBillTaskStateAccounted {
			logs.Infof("[%s] %s(%s) %d-%02d month task %s not accounted, rid: %s",
				summary.Vendor, summary.RootAccountCloudID, summary.RootAccountID,
				summary.BillYear, summary.BillMonth, monthTask.Type, kt.Rid)
			return decimal.Zero, false, err
		}

		for _, item := range monthTask.SummaryDetail {
			if item.MainAccountID == summary.MainAccountID {
				cost = cost.Add(item.Cost)
			}
		}
	}

	return cost, true, nil
}

// getMonthTask retrieves month tasks for the given summary and task types
func getMonthTask(kt *kit.Kit, summary *billcore.SummaryRoot, taskTypes []enumor.MonthTaskType) (
	[]*billcore.MonthTask, error) {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", summary.RootAccountID),
		tools.RuleEqual("bill_year", summary.BillYear),
		tools.RuleEqual("bill_month", summary.BillMonth),
		tools.RuleIn("type", taskTypes),
	}
	req := &bill.BillMonthTaskListReq{
		Filter: tools.ExpressionAnd(expressions...),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := actcli.GetDataService().Global.Bill.ListBillMonthTask(kt, req)
	if err != nil {
		logs.Errorf("get month pull task for %s(%s) %d-%02d failed, err: %v, rid: %s",
			summary.RootAccountCloudID, summary.RootAccountID, summary.BillYear, summary.BillMonth, err, kt.Rid)
		return nil, fmt.Errorf("get month pull task for %s(%s) %d-%02d failed, err: %v",
			summary.RootAccountCloudID, summary.RootAccountID, summary.BillYear, summary.BillMonth, err)
	}
	return result.Details, nil
}

// getDailyVersionCost retrieves the daily version cost for the given options
func (act *MainAccountSummaryAction) getDailyVersionCost(kt *kit.Kit, opt *MainAccountSummaryActionOption,
	versionID int) (total *decimal.Decimal, isAccounted bool, currencyCode enumor.CurrencyCode, err error) {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("main_account_id", opt.MainAccountID),
		tools.RuleEqual("vendor", opt.Vendor),
		tools.RuleEqual("bill_year", opt.BillYear),
		tools.RuleEqual("bill_month", opt.BillMonth),
		tools.RuleEqual("version_id", versionID),
	}
	result, err := actcli.GetDataService().Global.Bill.ListBillSummaryDaily(kt, &bill.BillSummaryDailyListReq{
		Filter: tools.ExpressionAnd(expressions...),
		Page: &core.BasePage{
			Start: 0,
			Limit: 31,
		},
	})
	if err != nil {
		return nil, false, "", fmt.Errorf("get main account summary of %v failed, err %s", opt, err.Error())
	}
	totalCost := decimal.NewFromInt(0)
	currencyCode = enumor.CurrencyUSD
	for _, dailySummary := range result.Details {
		if len(dailySummary.Currency) != 0 {
			currencyCode = dailySummary.Currency
		}
		totalCost = totalCost.Add(dailySummary.Cost)
	}
	isAccounted = true
	if len(result.Details) != times.DaysInMonth(opt.BillYear, time.Month(opt.BillMonth)) {
		isAccounted = false
	}

	return &totalCost, isAccounted, currencyCode, nil
}

// getLastMonthSyncedCost retrieves the last month's synced cost for the given options
func (act *MainAccountSummaryAction) getLastMonthSyncedCost(kt *kit.Kit, opt *MainAccountSummaryActionOption) (
	*decimal.Decimal, *decimal.Decimal, error) {

	billYear, billMonth, err := times.GetLastMonth(opt.BillYear, opt.BillMonth)
	if err != nil {
		return nil, nil, fmt.Errorf("get last month failed, err %s", err.Error())
	}
	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("main_account_id", opt.MainAccountID),
		tools.RuleEqual("vendor", opt.Vendor),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}
	result, err := actcli.GetDataService().Global.Bill.ListBillSummaryMain(
		kt, &bill.BillSummaryMainListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		return nil, nil, fmt.Errorf("get main account bill summary failed, err %s", err.Error())
	}
	if len(result.Details) > 1 {
		return nil, nil, fmt.Errorf("get invalid length main account bill summary resp %v", result)
	}
	if len(result.Details) == 0 {
		return nil, nil, nil
	}
	lastMonthSummary := result.Details[0]
	return &lastMonthSummary.CurrentMonthCostSynced, &lastMonthSummary.CurrentMonthRMBCostSynced, nil
}

// getAdjustmentSummary ...
func (act *MainAccountSummaryAction) getAdjustmentSummary(kt *kit.Kit, opt *MainAccountSummaryActionOption,
	currency enumor.CurrencyCode) (*decimal.Decimal, error) {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("main_account_id", opt.MainAccountID),
		tools.RuleEqual("vendor", opt.Vendor),
		tools.RuleEqual("bill_year", opt.BillYear),
		tools.RuleEqual("bill_month", opt.BillMonth),
		// only confirmed adjustment is counted
		tools.RuleEqual("state", enumor.BillAdjustmentStateConfirmed),
	}
	result, err := actcli.GetDataService().Global.Bill.ListBillAdjustmentItem(kt, &core.ListReq{
		Filter: tools.ExpressionAnd(expressions...),
		Page: &core.BasePage{
			Count: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("list adjustment item of %v failed, err %s", opt, err.Error())
	}
	logs.Infof("found %d adjustment item for opt %v, rid: %s", result.Count, opt, kt.Rid)
	cost := decimal.NewFromFloat(0)
	for offset := uint64(0); offset < result.Count; offset = offset + uint64(core.DefaultMaxPageLimit) {
		result, err = actcli.GetDataService().Global.Bill.ListBillAdjustmentItem(
			kt, &core.ListReq{
				Filter: tools.ExpressionAnd(expressions...),
				Page: &core.BasePage{
					Start: 0,
					Limit: core.DefaultMaxPageLimit,
				},
			})
		if err != nil {
			return nil, fmt.Errorf("list adjustment item of %v failed, err %s", opt, err.Error())
		}
		for _, item := range result.Details {
			cost = cost.Add(item.Cost)
			if len(item.Currency) != 0 && currency != item.Currency {
				return nil, fmt.Errorf("adjustment currency mismatch, want: %s ,got: %s", currency, item.Currency)
			}
		}
	}
	return &cost, nil
}

// getRootSummary ...
func (act *MainAccountSummaryAction) getRootSummary(kt *kit.Kit, opt *MainAccountSummaryActionOption) (
	*billcore.SummaryRoot, *bill.BillSummaryMain, error) {

	rootAccountExpr := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("vendor", opt.Vendor),
		tools.RuleEqual("bill_year", opt.BillYear),
		tools.RuleEqual("bill_month", opt.BillMonth),
	}
	rootResult, err := actcli.GetDataService().Global.Bill.ListBillSummaryRoot(
		kt, &bill.BillSummaryRootListReq{
			Filter: tools.ExpressionAnd(rootAccountExpr...),
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		return nil, nil, fmt.Errorf("get root account bill summary failed, opt: %s, err %s", opt.String(), err.Error())
	}
	if len(rootResult.Details) != 1 {
		return nil, nil, fmt.Errorf("get invalid length root account bill summary resp %v", rootResult)
	}

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("main_account_id", opt.MainAccountID),
		tools.RuleEqual("vendor", opt.Vendor),
		tools.RuleEqual("bill_year", opt.BillYear),
		tools.RuleEqual("bill_month", opt.BillMonth),
	}
	result, err := actcli.GetDataService().Global.Bill.ListBillSummaryMain(
		kt, &bill.BillSummaryMainListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		return nil, nil, fmt.Errorf("get main account bill summary failed, opt: %s, err: %v", opt.String(), err)
	}
	if len(result.Details) != 1 {
		return nil, nil, fmt.Errorf("get invalid length main account bill summary resp: %v", result)
	}

	return rootResult.Details[0], result.Details[0], nil
}
