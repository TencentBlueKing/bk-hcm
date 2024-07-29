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
	"strings"

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
}

// GetBatchSize for aws is always 999
func (a AwsMonthTask) GetBatchSize(kt *kit.Kit) uint64 {
	return 999
}

// Pull aws root account bill
func (a AwsMonthTask) Pull(kt *kit.Kit, opt *MonthTaskActionOption,
	index uint64) (itemList []bill.RawBillItem, isFinished bool, err error) {

	// 查询根账号信息
	rootAccount, err := actcli.GetDataService().Aws.RootAccount.Get(kt, opt.RootAccountID)
	if err != nil {
		return nil, false, err
	}

	// 获取指定月份最后一天
	lastDay, err := times.GetLastDayOfMonth(opt.BillYear, opt.BillMonth)
	if err != nil {
		logs.Errorf("fail get last day of month for aws month task, year: %d, month:%d, err: %v, rid: %s",
			opt.BillYear, opt.BillMonth, err, kt.Rid)
		return nil, false, err
	}

	billResp, err := actcli.GetHCService().Aws.Bill.GetRootAccountBillList(kt, &hcbill.AwsRootBillListReq{
		RootAccountID:      opt.RootAccountID,
		MainAccountCloudID: rootAccount.CloudID,
		BeginDate:          fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, 1),
		EndDate:            fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, lastDay),
		Page: &hcbill.AwsBillListPage{
			Offset: index,
			Limit:  a.GetBatchSize(kt),
		},
	})
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
			BillCost:      *cost,
			ResAmount:     *amount,
			ResAmountUnit: record["pricing_unit"],
			Extension:     types.JsonField(extensionBytes),
		})
	}
	finished := uint64(len(billResp.Details)) < a.GetBatchSize(kt)
	return itemList, finished, nil
}

// Split root account bill into main accounts
func (a AwsMonthTask) Split(kt *kit.Kit, opt *MonthTaskActionOption, rawItemList []*bill.RawBillItem) (
	[]bill.BillItemCreateReq[json.RawMessage], error) {

	if len(rawItemList) == 0 {
		return nil, nil
	}
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

	summaryMainList, err := a.getSummaryMainList(kt, opt, mainAccountMap, rootAccount)
	if err != nil {
		logs.Errorf("fail to get summary main list for aws month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}

	// 聚合本批次 账单总额，并分摊给每个主账号
	batchCost := decimal.Zero
	for _, item := range rawItemList {
		batchCost = batchCost.Add(item.BillCost)
	}
	// 计算总额，再按比例分摊给各个二级账号
	summaryTotal := decimal.Zero
	for _, summaryMain := range summaryMainList {
		summaryTotal = summaryTotal.Add(summaryMain.CurrentMonthCost)
	}

	billItems := make([]bill.BillItemCreateReq[json.RawMessage], 0, len(summaryMainList))
	for _, summary := range summaryMainList {

		mainAccount := mainAccountMap[summary.MainAccountID]
		cost := batchCost.Mul(summary.CurrentMonthCost).Div(summaryTotal)
		extJson, err := a.convCommonExpenseExt(
			AwsCommonExpenseName, opt, rootAccount.CloudID, mainAccount.CloudID, summary.Currency, cost)
		if err != nil {
			logs.Errorf("fail to marshal aws common expense extension to json, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		costBillItem := convSummaryToCommonExpense(rootAccount.ID, summary, cost, extJson)
		billItems = append(billItems, costBillItem)

		if rootAsMainAccount == nil {
			continue
		}

		// 此处冲平根账号支出
		reverseCost := cost.Neg()
		reverseExtJson, err := a.convCommonExpenseExt(AwsCommonExpenseReverseName,
			opt, rootAccount.CloudID, mainAccount.CloudID, summary.Currency, reverseCost)
		if err != nil {
			logs.Errorf("fail to marshal aws common expense reverse extension to json, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		reverseBillItem := convSummaryToCommonReverse(
			rootAccount.ID, rootAsMainAccount.ID, summary, reverseCost, reverseExtJson)
		billItems = append(billItems, reverseBillItem)
	}

	return billItems, nil
}

func (a AwsMonthTask) getSummaryMainList(kt *kit.Kit, opt *MonthTaskActionOption,
	mainAccountMap map[string]*protocore.BaseMainAccount, rootAccount *dataproto.AwsRootAccount) (
	[]*bill.BillSummaryMainResult, error) {

	mainAccountIDs := make([]string, 0, len(mainAccountMap))

	// 排除根账号自身以及用户设定的账号
	excludeCloudIds := []string{rootAccount.CloudID}
	if opt.Extension != nil && opt.Extension["AwsCommonExpenseExcludeCloudIDKey"] != "" {
		excludeCloudIDStr := opt.Extension["AwsCommonExpenseExcludeCloudIDKey"]
		excluded := strings.Split(excludeCloudIDStr, ",")
		excludeCloudIds = append(excludeCloudIds, excluded...)
	}
	exCloudIdMap := cvt.StringSliceToMap(excludeCloudIds)
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

func convSummaryToCommonReverse(rootAccountID, mainAccountID string, summary *bill.BillSummaryMainResult,
	cost decimal.Decimal, extension []byte) bill.BillItemCreateReq[json.RawMessage] {

	reverseBillItem := bill.BillItemCreateReq[json.RawMessage]{
		RootAccountID: rootAccountID,
		MainAccountID: mainAccountID,
		Vendor:        summary.Vendor,
		ProductID:     summary.ProductID,
		BkBizID:       summary.BkBizID,
		BillYear:      summary.BillYear,
		BillMonth:     summary.BillMonth,
		BillDay:       0,
		VersionID:     summary.CurrentVersion,
		Currency:      summary.Currency,
		Cost:          cost.Neg(),
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

func convSummaryToCommonExpense(rootID string, summary *bill.BillSummaryMainResult,
	cost decimal.Decimal, extJson []byte) bill.BillItemCreateReq[json.RawMessage] {

	return bill.BillItemCreateReq[json.RawMessage]{
		RootAccountID: rootID,
		MainAccountID: summary.MainAccountID,
		Vendor:        summary.Vendor,
		ProductID:     summary.ProductID,
		BkBizID:       summary.BkBizID,
		BillYear:      summary.BillYear,
		BillMonth:     summary.BillMonth,
		BillDay:       0,
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
func BuildAwsMonthTaskOptionExt(excludeAccountCloudIds []string) map[string]string {

	return map[string]string{
		AwsCommonExpenseExcludeCloudIDKey: strings.Join(excludeAccountCloudIds, ","),
	}
}
