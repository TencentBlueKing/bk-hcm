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
	"strconv"

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

// AwsSavingsPlanMonthTask ...
type AwsSavingsPlanMonthTask struct {
	awsMonthTaskBaseRunner
}

// GetBatchSize for aws is always 999
func (a awsMonthTaskBaseRunner) GetBatchSize(kt *kit.Kit) uint64 {
	return 999
}

// Pull aws root account bill
func (a AwsSavingsPlanMonthTask) Pull(kt *kit.Kit, opt *MonthTaskActionOption, index uint64) (
	itemList []bill.RawBillItem, isFinished bool, err error) {

	a.initExtension(opt)
	// 获取指定月份最后一天
	lastDay, err := times.GetLastDayOfMonth(opt.BillYear, opt.BillMonth)
	if err != nil {
		logs.Errorf("fail get last day of month for aws month task, year: %d, month: %d, err: %v, rid: %s",
			opt.BillYear, opt.BillMonth, err, kt.Rid)
		return nil, false, err
	}
	// 拉取 sp 分账金额
	spReq := &hcbill.AwsRootSpUsageTotalReq{
		RootAccountID: opt.RootAccountID,
		SpArnPrefix:   a.spArnPrefix,
		Year:          uint(opt.BillYear),
		Month:         uint(opt.BillMonth),
		StartDay:      1,
		EndDay:        uint(lastDay),
	}

	spUsage, err := actcli.GetHCService().Aws.Bill.GetRootAccountSpTotalUsage(kt, spReq)
	if err != nil {
		logs.Errorf("get root account sp usage failed, err: %v, rid: %s", err, kt.Rid)
		return nil, false, err
	}
	extension := map[string]string{
		"line_item_product_code":       constant.AwsSavingsPlansCostCodeReverse,
		"product_product_name":         constant.AwsSavingsPlansCostCodeReverse,
		"pricing_unit":                 "Account",
		"line_item_currency_code":      string(enumor.CurrencyUSD),
		"line_item_net_unblended_cost": spUsage.SPNetCost.Neg().String(),
		"line_item_usage_amount":       strconv.FormatUint(spUsage.AccountCount, 10),
	}
	extBytes, err := json.Marshal(extension)
	if err != nil {
		logs.Errorf("marshal sp usage failed, err: %v, rid: %s", err, kt.Rid)
		return nil, false, err
	}
	spUsageReverseItem := bill.RawBillItem{
		Region:        "any",
		HcProductCode: constant.AwsSavingsPlansCostCodeReverse,
		HcProductName: constant.AwsSavingsPlansCostCodeReverse,
		BillCurrency:  enumor.CurrencyUSD,
		BillCost:      spUsage.SPNetCost.Neg(),
		ResAmount:     decimal.NewFromUint64(spUsage.AccountCount),
		ResAmountUnit: "Account",
		Extension:     types.JsonField(extBytes),
	}

	itemList = []bill.RawBillItem{spUsageReverseItem}
	return itemList, true, nil
}

// Split root account bill into main accounts
func (a AwsSavingsPlanMonthTask) Split(kt *kit.Kit, opt *MonthTaskActionOption, rawItemList []*bill.RawBillItem) (
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
	mainAccountMap, _, err := a.listMainAccount(kt, rootAccount)
	if err != nil {
		logs.Errorf("fail to list main account for aws month task split step, err: %v, opt: %#v, rid: %s",
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

	spItems, err := a.splitSpReverseExpense(kt, opt, spMainAccount, rawItemList)
	if err != nil {
		logs.Errorf("fail to split sp reverse expense for aws month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}
	return spItems, nil
}

func (a AwsSavingsPlanMonthTask) splitSpReverseExpense(kt *kit.Kit, opt *MonthTaskActionOption,
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

	extJson, err := convAwsBillItemExtension(constant.AwsSavingsPlansCostCodeReverse, opt,
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
		HcProductCode: constant.AwsSavingsPlansCostCodeReverse,
		HcProductName: constant.AwsSavingsPlansCostCodeReverse,
		Extension:     cvt.ValToPtr[json.RawMessage](extJson),
	}
	return []bill.BillItemCreateReq[json.RawMessage]{item}, nil
}

// GetHcProductCodes hc product code ranges
func (a AwsSavingsPlanMonthTask) GetHcProductCodes() []string {
	return []string{constant.AwsSavingsPlansCostCode, constant.AwsSavingsPlansCostCodeReverse}
}
