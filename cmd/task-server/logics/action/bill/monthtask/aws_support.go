/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package monthtask

import (
	"encoding/json"
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/account-set"
	"hcm/pkg/api/data-service/bill"
	hcbill "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/times"

	"github.com/shopspring/decimal"
)

// AwsSupportMonthTask ...
type AwsSupportMonthTask struct {
	awsMonthTaskBaseRunner
}

// Pull support bill item
func (a AwsSupportMonthTask) Pull(kt *kit.Kit, opt *MonthTaskActionOption, index uint64) (itemList []bill.RawBillItem,
	isFinished bool, err error) {

	// 查询根账号信息
	rootAccount, err := actcli.GetDataService().Aws.RootAccount.Get(kt, opt.RootAccountID)
	if err != nil {
		return nil, false, err
	}

	// 获取指定月份最后一天
	lastDay, err := times.GetLastDayOfMonth(opt.BillYear, opt.BillMonth)
	if err != nil {
		logs.Errorf("fail get last day of month for aws month task, year: %d, month: %d, err: %v, rid: %s",
			opt.BillYear, opt.BillMonth, err, kt.Rid)
		return nil, false, err
	}
	rootBillReq := &hcbill.AwsRootBillListReq{
		RootAccountID:      opt.RootAccountID,
		MainAccountCloudID: rootAccount.CloudID,
		BeginDate:          fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, 1),
		EndDate:            fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, lastDay),
		Page:               &hcbill.AwsBillListPage{Offset: index, Limit: a.GetBatchSize(kt)},
	}
	billResp, err := actcli.GetHCService().Aws.Bill.GetRootAccountBillList(kt, rootBillReq)
	if err != nil {
		return nil, false, err
	}
	if len(billResp.Details) == 0 {
		return nil, true, nil
	}
	for _, record := range billResp.Details {
		cost, err := getDecimal(record, "line_item_net_unblended_cost")
		if err != nil {
			return nil, false, err
		}
		amount, err := getDecimal(record, "line_item_usage_amount")
		if err != nil {
			return nil, false, err
		}

		extensionBytes, err := json.Marshal(record)
		if err != nil {
			return nil, false, fmt.Errorf("marshal aws bill item %v failed, err: %w", record, err)
		}
		itemList = append(itemList, bill.RawBillItem{
			Region:        record["product_region"],
			HcProductCode: record["line_item_product_code"],
			HcProductName: record["product_product_name"],
			BillCurrency:  enumor.CurrencyCode(record["line_item_currency_code"]),
			BillCost:      cvt.PtrToVal(cost),
			ResAmount:     cvt.PtrToVal(amount),
			ResAmountUnit: record["pricing_unit"],
			Extension:     types.JsonField(extensionBytes),
		})
	}
	supportDone := uint64(len(billResp.Details)) < a.GetBatchSize(kt)
	return itemList, supportDone, nil
}

// Split aws support fee to main account
func (a AwsSupportMonthTask) Split(kt *kit.Kit, opt *MonthTaskActionOption,
	rawItemList []*bill.RawBillItem) ([]bill.BillItemCreateReq[json.RawMessage], error) {

	if len(rawItemList) == 0 {
		return nil, nil
	}
	a.initExtension(opt)

	// 查询根账号信息
	rootAccount, err := actcli.GetDataService().Aws.RootAccount.Get(kt, opt.RootAccountID)
	if err != nil {
		logs.Errorf("failt to get root account info, err: %v, accountID: %s, rid: %s", err, opt.RootAccountID, kt.Rid)
		return nil, err
	}

	// rootAsMainAccount 作为二级账号存在的根账号，将分摊后的账单抵冲该账号支出
	mainAccountMap, rootAsMainAccount, err := a.listMainAccount(kt, rootAccount)
	if err != nil {
		logs.Errorf("fail to list main account for aws month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}

	commonItems, err := a.splitCommonExpense(kt, opt, mainAccountMap, rootAsMainAccount, rawItemList)
	if err != nil {
		logs.Errorf("fail to split common expense for aws month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}
	return commonItems, nil
}

func (a AwsSupportMonthTask) splitCommonExpense(kt *kit.Kit, opt *MonthTaskActionOption,
	mainAccountMap map[string]*protocore.BaseMainAccount, rootAsMainAccount *protocore.BaseMainAccount,
	rawItemList []*bill.RawBillItem) ([]bill.BillItemCreateReq[json.RawMessage], error) {

	if len(rawItemList) == 0 {
		return nil, nil
	}

	// 聚合本批次 账单总额，并分摊给每个主账号
	batchSum := decimal.Zero
	for _, item := range rawItemList {
		batchSum = batchSum.Add(item.BillCost)
	}

	var summaryList []*bill.BillSummaryMain
	summaryList, err := a.listSummaryMainForSupport(kt, opt, mainAccountMap, rootAsMainAccount.CloudID)
	if err != nil {
		logs.Errorf("fail to get summary main list for aws month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}
	if len(summaryList) == 0 {
		logs.Warnf("no main account for aws month task common expense, opt: %#v, rid: %s", opt, kt.Rid)
		return nil, nil
	}

	// 计算总额，再按比例分摊给各个二级账号
	summaryTotal := decimal.Zero
	for _, summaryMain := range summaryList {
		summaryTotal = summaryTotal.Add(summaryMain.CurrentMonthCost)
	}

	billItems := make([]bill.BillItemCreateReq[json.RawMessage], 0, len(summaryList))
	for _, summary := range summaryList {
		mainAccount := mainAccountMap[summary.MainAccountID]
		cost := batchSum.Mul(summary.CurrentMonthCost).Div(summaryTotal)
		extJson, err := convAwsBillItemExtension(
			constant.BillCommonExpenseName, opt, summary.RootAccountCloudID, mainAccount.CloudID, summary.Currency,
			cost)
		if err != nil {
			logs.Errorf("fail to marshal aws common expense extension to json, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		costBillItem := convSummaryToCommonExpense(summary, cost, extJson)
		billItems = append(billItems, costBillItem)

		if rootAsMainAccount == nil {
			// 未将根账号作为主账号录入，跳过
			continue
		}
		// 此处冲平根账号支出
		reverseCost := cost.Neg()
		reverseExtJson, err := convAwsBillItemExtension(constant.BillCommonExpenseReverseName,
			opt, summary.RootAccountCloudID, mainAccount.CloudID, summary.Currency, reverseCost)
		if err != nil {
			logs.Errorf("fail to marshal aws common expense reverse extension to json, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		reverseBillItem := convSummaryToCommonReverse(rootAsMainAccount, summary, reverseCost, reverseExtJson)
		billItems = append(billItems, reverseBillItem)
	}
	return billItems, nil
}

// 不包含根账号自身以及用户设定的排除账号的 二级账号汇总信息
func (a AwsSupportMonthTask) listSummaryMainForSupport(kt *kit.Kit, opt *MonthTaskActionOption,
	mainAccountMap map[string]*protocore.BaseMainAccount, rootCloudID string) ([]*bill.BillSummaryMain, error) {

	mainAccountIDs := make([]string, 0, len(mainAccountMap))

	// 排除根账号自身以及用户设定的账号
	exCloudIdMap := cvt.StringSliceToMap(a.excludeAccountCloudIds)
	exCloudIdMap[rootCloudID] = struct{}{}
	for _, account := range mainAccountMap {
		if _, exist := exCloudIdMap[account.CloudID]; exist {
			continue
		}
		mainAccountIDs = append(mainAccountIDs, account.ID)
	}
	summaryListReq := &bill.BillSummaryMainListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("main_account_id", mainAccountIDs),
			tools.RuleEqual("bill_year", opt.BillYear),
			tools.RuleEqual("bill_month", opt.BillMonth),
		),
		Page: core.NewDefaultBasePage(),
	}
	summaryMainResp, err := actcli.GetDataService().Global.Bill.ListBillSummaryMain(kt, summaryListReq)
	if err != nil {
		logs.Errorf("failt to list main account bill summary for %s month task, err: %v, rid: %s",
			enumor.Aws, err, kt.Rid)
		return nil, err
	}
	return summaryMainResp.Details, nil
}

func convSummaryToCommonReverse(mainAccount *protocore.BaseMainAccount, summary *bill.BillSummaryMain,
	cost decimal.Decimal, extension []byte) bill.BillItemCreateReq[json.RawMessage] {

	reverseBillItem := bill.BillItemCreateReq[json.RawMessage]{
		RootAccountID: mainAccount.ParentAccountID,
		MainAccountID: mainAccount.ID,
		Vendor:        mainAccount.Vendor,
		ProductID:     mainAccount.OpProductID,
		BkBizID:       mainAccount.BkBizID,
		BillYear:      summary.BillYear,
		BillMonth:     summary.BillMonth,
		BillDay:       enumor.MonthTaskSpecialBillDay,
		VersionID:     summary.CurrentVersion,
		Currency:      summary.Currency,
		Cost:          cost,
		HcProductCode: constant.BillCommonExpenseReverseName,
		HcProductName: constant.BillCommonExpenseReverseName,
		Extension:     cvt.ValToPtr[json.RawMessage](extension),
	}
	return reverseBillItem
}

func convSummaryToCommonExpense(summary *bill.BillSummaryMain,
	cost decimal.Decimal, extJson []byte) bill.BillItemCreateReq[json.RawMessage] {

	return bill.BillItemCreateReq[json.RawMessage]{
		RootAccountID: summary.RootAccountID,
		MainAccountID: summary.MainAccountID,
		Vendor:        summary.Vendor,
		ProductID:     summary.ProductID,
		BkBizID:       summary.BkBizID,
		BillYear:      summary.BillYear,
		BillMonth:     summary.BillMonth,
		BillDay:       enumor.MonthTaskSpecialBillDay,
		VersionID:     summary.CurrentVersion,
		Currency:      summary.Currency,
		Cost:          cost,
		HcProductCode: constant.BillCommonExpenseName,
		HcProductName: constant.BillCommonExpenseName,
		Extension:     cvt.ValToPtr[json.RawMessage](extJson),
	}
}

// GetHcProductCodes hc product code ranges
func (a *AwsSupportMonthTask) GetHcProductCodes() []string {
	return []string{constant.BillCommonExpenseName, constant.BillCommonExpenseReverseName}
}
