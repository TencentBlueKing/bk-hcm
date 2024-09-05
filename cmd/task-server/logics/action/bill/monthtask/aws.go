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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"hcm/cmd/task-server/logics/action/bill/dailysplit"
	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/account-set"
	billcore "hcm/pkg/api/core/bill"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/api/data-service/bill"
	hcbill "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/times"

	"github.com/shopspring/decimal"
)

const (
	// AwsCommonExpenseExcludeCloudIDKey ...
	AwsCommonExpenseExcludeCloudIDKey = "aws_common_expense_exclude_account_cloud_id"
	// AwsCommonExpenseReverseName common expense reverse
	AwsCommonExpenseReverseName = "CommonExpenseReverse"
	// AwsCommonExpenseName common expense
	AwsCommonExpenseName = "CommonExpense"
)

func newAwsRunner() MonthTaskRunner {
	return &AwsMonthTask{}
}

// AwsMonthTask ...
type AwsMonthTask struct {
	excludeAccountCloudIds []string
	spArnPrefix            string
	spMainAccountCloudID   string
}

// GetBatchSize for aws is always 999
func (a AwsMonthTask) GetBatchSize(kt *kit.Kit) uint64 {
	return 999
}

// Pull aws root account bill
func (a AwsMonthTask) Pull(kt *kit.Kit, opt *MonthTaskActionOption, index uint64) (
	itemList []bill.RawBillItem, isFinished bool, err error) {

	a.initExtension(opt)

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
	if !supportDone {
		return itemList, false, nil
	}
	// 结束前增加分账金额 TODO 支持多类型month task
	spUsageReverseItem, err := a.getSpUsage(kt, opt, uint(lastDay), err)
	if err != nil {
		logs.Errorf("get sp usage failed, err: %v, rid: %s", err, kt.Rid)
		return nil, false, err
	}

	itemList = append(itemList, cvt.PtrToVal(spUsageReverseItem))
	return itemList, supportDone, nil
}

// getSpUsage 获取SP使用总额冲平支出账号
func (a AwsMonthTask) getSpUsage(kt *kit.Kit, opt *MonthTaskActionOption, lastDay uint, err error) (
	*bill.RawBillItem, error) {

	// 拉取 sp 分账金额
	spReq := &hcbill.AwsRootSpUsageTotalReq{
		RootAccountID: opt.RootAccountID,
		SpArnPrefix:   a.spArnPrefix,
		Year:          uint(opt.BillYear),
		Month:         uint(opt.BillMonth),
		StartDay:      1,
		EndDay:        lastDay,
	}
	spUsage, err := actcli.GetHCService().Aws.Bill.GetRootAccountSpTotalUsage(kt, spReq)
	if err != nil {
		logs.Errorf("get root account sp usage failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	extension := map[string]string{
		"line_item_product_code":       dailysplit.AwsSavingsPlansCostCodeReverse,
		"product_product_name":         dailysplit.AwsSavingsPlansCostCodeReverse,
		"pricing_unit":                 "Account",
		"line_item_currency_code":      string(enumor.CurrencyUSD),
		"line_item_net_unblended_cost": spUsage.SPNetCost.Neg().String(),
		"line_item_usage_amount":       strconv.FormatUint(spUsage.AccountCount, 10),
	}
	extBytes, err := json.Marshal(extension)
	if err != nil {
		logs.Errorf("marshal sp usage failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	spUsageReverseItem := &bill.RawBillItem{
		Region:        "any",
		HcProductCode: dailysplit.AwsSavingsPlansCostCodeReverse,
		HcProductName: dailysplit.AwsSavingsPlansCostCodeReverse,
		BillCurrency:  enumor.CurrencyUSD,
		BillCost:      spUsage.SPNetCost.Neg(),
		ResAmount:     decimal.NewFromUint64(spUsage.AccountCount),
		ResAmountUnit: "Account",
		Extension:     types.JsonField(extBytes),
	}
	return spUsageReverseItem, nil
}

func (a *AwsMonthTask) initExtension(opt *MonthTaskActionOption) {
	if opt.Extension == nil {
		return
	}
	a.spArnPrefix = opt.Extension[dailysplit.AwsSavingsPlanARNPrefixKey]
	a.spMainAccountCloudID = opt.Extension[dailysplit.AwsSavingsPlanAccountCloudIDKey]
	if opt.Extension[AwsCommonExpenseExcludeCloudIDKey] != "" {
		excludeCloudIDStr := opt.Extension[AwsCommonExpenseExcludeCloudIDKey]
		excluded := strings.Split(excludeCloudIDStr, ",")
		a.excludeAccountCloudIds = excluded
	}
}

// Split root account bill into main accounts
func (a AwsMonthTask) Split(kt *kit.Kit, opt *MonthTaskActionOption, rawItemList []*bill.RawBillItem) (
	[]bill.BillItemCreateReq[json.RawMessage], error) {

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

	var commonRawItems []*bill.RawBillItem
	var spRawItems []*bill.RawBillItem
	for _, item := range rawItemList {
		if item.HcProductCode == dailysplit.AwsSavingsPlansCostCodeReverse {
			spRawItems = append(spRawItems, item)
			continue
		}
		commonRawItems = append(commonRawItems, item)

	}

	commonItems, err := a.splitCommonExpense(kt, opt, mainAccountMap, rootAsMainAccount,
		commonRawItems)
	if err != nil {
		logs.Errorf("fail to split common expense for aws month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}

	var spMainAccount *protocore.BaseMainAccount
	for id := range mainAccountMap {
		if mainAccountMap[id].CloudID == a.spMainAccountCloudID {
			spMainAccount = mainAccountMap[id]
		}
	}
	if spMainAccount == nil {
		return nil, errors.New("sp main account not found")
	}

	spItems, err := a.splitSpReverseExpense(kt, opt, spMainAccount, spRawItems)
	if err != nil {
		logs.Errorf("fail to split sp reverse expense for aws month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}
	billItems := append(commonItems, spItems...)
	return billItems, nil
}
func (a AwsMonthTask) splitSpReverseExpense(kt *kit.Kit, opt *MonthTaskActionOption,
	spAccount *protocore.BaseMainAccount, rawItemList []*bill.RawBillItem) (
	[]bill.BillItemCreateReq[json.RawMessage], error) {

	// 查找summary
	summaryReq := &bill.BillSummaryMainListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bill_year", opt.BillYear),
			tools.RuleEqual("bill_month", opt.BillMonth),
			tools.RuleEqual("vendor", opt.Vendor),
			tools.RuleEqual("main_account_id", spAccount.ID)),
		Page: core.NewDefaultBasePage(),
	}
	summaryResp, err := actcli.GetDataService().Global.Bill.ListBillSummaryMain(kt, summaryReq)
	if err != nil {
		logs.Errorf("fail to get sp summary for month task split, err: %v, opt: %#v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}
	if len(summaryResp.Details) != 1 {
		return nil, errors.New("sp summary not found")
	}
	summary := summaryResp.Details[0]

	batchSum := decimal.Zero
	for _, item := range rawItemList {
		batchSum = batchSum.Add(item.BillCost)
	}

	extJson, err := a.convCommonExpenseExt(dailysplit.AwsSavingsPlansCostCodeReverse, opt,
		summary.RootAccountCloudID, summary.RootAccountCloudID, summary.Currency, batchSum)
	if err != nil {
		logs.Errorf("fail to convert common expense extension for aws month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}

	item := bill.BillItemCreateReq[json.RawMessage]{
		RootAccountID: opt.RootAccountID,
		MainAccountID: summary.MainAccountID,
		Vendor:        enumor.Aws,
		ProductID:     summary.ProductID,
		BkBizID:       summary.BkBizID,
		BillYear:      opt.BillYear,
		BillMonth:     opt.BillMonth,
		BillDay:       enumor.MonthTaskSpecialBillDay,
		VersionID:     summary.CurrentVersion,
		Currency:      summary.Currency,
		Cost:          batchSum,
		HcProductCode: dailysplit.AwsSavingsPlansCostCodeReverse,
		HcProductName: dailysplit.AwsSavingsPlansCostCodeReverse,
		Extension:     cvt.ValToPtr[json.RawMessage](extJson),
	}
	return []bill.BillItemCreateReq[json.RawMessage]{item}, nil
}

func (a AwsMonthTask) splitCommonExpense(kt *kit.Kit, opt *MonthTaskActionOption,
	mainAccountMap map[string]*protocore.BaseMainAccount, rootAsMainAccount *protocore.BaseMainAccount,
	rawItemList []*bill.RawBillItem) ([]bill.BillItemCreateReq[json.RawMessage], error) {

	if len(rawItemList) == 0 {
		return nil, nil
	}
	var summaryList []*bill.BillSummaryMain
	summaryList, err := a.getSummaryMainListExcludeRootAndOwn(kt, opt, mainAccountMap, rootAsMainAccount.CloudID)
	if err != nil {
		logs.Errorf("fail to get summary main list for aws month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}
	if len(summaryList) == 0 {
		logs.Warnf("no main account for aws month task common expense, opt: %#v, rid: %s", opt, kt.Rid)
		return nil, nil
	}

	// 聚合本批次 账单总额，并分摊给每个主账号
	batchSum := decimal.Zero
	for _, item := range rawItemList {
		batchSum = batchSum.Add(item.BillCost)
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
		extJson, err := a.convCommonExpenseExt(
			AwsCommonExpenseName, opt, summary.RootAccountCloudID, mainAccount.CloudID, summary.Currency, cost)
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
		reverseExtJson, err := a.convCommonExpenseExt(AwsCommonExpenseReverseName,
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

func (a AwsMonthTask) getSummaryMainListExcludeRootAndOwn(kt *kit.Kit, opt *MonthTaskActionOption,
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

func (a AwsMonthTask) listMainAccount(kt *kit.Kit, rootAccount *dataproto.AwsRootAccount) (
	mainAccountMap map[string]*protocore.BaseMainAccount, rootAsMainAccount *protocore.BaseMainAccount, err error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("parent_account_id", rootAccount.ID)),
		Page:   core.NewDefaultBasePage(),
	}
	mainAccountsResp, err := actcli.GetDataService().Global.MainAccount.List(kt, listReq)
	if err != nil {
		logs.Errorf("failt to list main account for %s month task, err: %v, rid: %s",
			enumor.Aws, err, kt.Rid)
		return nil, nil, err
	}
	mainAccountMap = make(map[string]*protocore.BaseMainAccount, len(mainAccountsResp.Details))
	for _, account := range mainAccountsResp.Details {
		mainAccountMap[account.ID] = account
		if account.CloudID == rootAccount.CloudID {
			rootAsMainAccount = account
		}
	}

	return mainAccountMap, rootAsMainAccount, nil
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
		HcProductCode: AwsCommonExpenseReverseName,
		HcProductName: AwsCommonExpenseReverseName,
		Extension:     cvt.ValToPtr[json.RawMessage](extension),
	}
	return reverseBillItem
}

func (a AwsMonthTask) convCommonExpenseExt(productName string, opt *MonthTaskActionOption, rootAccountCloudID string,
	mainAccountCloudID string, currencyCode enumor.CurrencyCode, cost decimal.Decimal) ([]byte, error) {

	ext := billcore.AwsRawBillItem{
		Month:                    fmt.Sprintf("%02d", opt.BillMonth),
		Year:                     fmt.Sprintf("%4d", opt.BillYear),
		BillPayerAccountId:       rootAccountCloudID,
		LineItemUsageAccountId:   mainAccountCloudID,
		LineItemCurrencyCode:     string(currencyCode),
		LineItemNetUnblendedCost: cost.String(),
		LineItemProductCode:      productName,
		ProductProductName:       productName,
		PricingCurrency:          string(currencyCode),
	}
	return json.Marshal(ext)
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
		HcProductCode: AwsCommonExpenseName,
		HcProductName: AwsCommonExpenseName,
		Extension:     cvt.ValToPtr[json.RawMessage](extJson),
	}
}

func getDecimal(dict map[string]string, key string) (*decimal.Decimal, error) {
	val, ok := dict[key]
	if !ok {
		return nil, errors.New("key not found: " + key)
	}
	d, err := decimal.NewFromString(val)
	if err != nil {
		return nil, fmt.Errorf("fail to convert to decimal, key: %s, value: %s, err: %v", key, val, err)
	}
	return &d, nil
}

// BuildAwsMonthTaskOptionExt build aws month task option extension
func BuildAwsMonthTaskOptionExt(arnPrefix, spMainCloudID string, excludeAccountCloudIds []string) map[string]string {

	return map[string]string{
		AwsCommonExpenseExcludeCloudIDKey:          strings.Join(excludeAccountCloudIds, ","),
		dailysplit.AwsSavingsPlanARNPrefixKey:      arnPrefix,
		dailysplit.AwsSavingsPlanAccountCloudIDKey: spMainCloudID,
	}
}
