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
	adbilltypes "hcm/pkg/adaptor/types/bill"
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

// HuaweiSupportMonthTask
// 1. 拉根账号下的support plan 账单 直接用根账号cloud_id 接口拉取，再分摊到个各个账号下
// 2. 分摊到子账号下
type HuaweiSupportMonthTask struct {
	huaweiMonthTaskBaseRunner
}

// Pull support bill item
func (a HuaweiSupportMonthTask) Pull(kt *kit.Kit, opt *MonthTaskActionOption, index uint64) (
	itemList []bill.RawBillItem, isFinished bool, err error) {

	// 查询根账号信息
	rootAccount, err := actcli.GetDataService().HuaWei.RootAccount.Get(kt, opt.RootAccountID)
	if err != nil {
		return nil, false, err
	}

	// 获取指定月份最后一天
	lastDay, err := times.GetLastDayOfMonth(opt.BillYear, opt.BillMonth)
	if err != nil {
		logs.Errorf("fail get last day of month for huawei month task, year: %d, month: %d, err: %v, rid: %s",
			opt.BillYear, opt.BillMonth, err, kt.Rid)
		return nil, false, err
	}

	rootBillReq := &hcbill.HuaWeiFeeRecordListReq{
		RootAccountID:      opt.RootAccountID,
		MainAccountCloudID: rootAccount.CloudID,
		Month:              fmt.Sprintf("%d-%02d", opt.BillYear, opt.BillMonth),
		BillDateBegin:      fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, 1),
		BillDateEnd:        fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, lastDay),
		Page: &adbilltypes.HuaWeiBillPage{
			Offset: cvt.ValToPtr(int32(index)),
			Limit:  cvt.ValToPtr(int32(a.GetBatchSize(kt))),
		},
	}
	billResp, err := actcli.GetHCService().HuaWei.Bill.ListFeeRecord(kt, rootBillReq)
	if err != nil {
		return nil, false, err
	}
	if len(billResp.Details) == 0 {
		return nil, true, nil
	}
	for _, record := range billResp.Details {

		creditCost := decimal.NewFromFloat(0)
		if record.CreditAmount != nil {
			creditCost = decimal.NewFromFloat(*record.CreditAmount)
		}
		debtCost := decimal.NewFromFloat(0)
		if record.DebtAmount != nil {
			debtCost = decimal.NewFromFloat(*record.DebtAmount)
		}
		extensionBytes, err := json.Marshal(record)
		if err != nil {
			return nil, false, fmt.Errorf("marshal huawei support bill item %v failed", record)
		}
		newBillItem := bill.RawBillItem{}
		if record.Region != nil {
			newBillItem.Region = *record.Region
		}
		if record.CloudServiceType != nil {
			newBillItem.HcProductCode = *record.CloudServiceType
		}
		if record.CloudServiceTypeName != nil {
			newBillItem.HcProductName = *record.CloudServiceTypeName
		}
		newBillItem.BillCurrency = billResp.Currency
		newBillItem.BillCost = creditCost.Add(debtCost)
		if record.Usage != nil {
			newBillItem.ResAmount = decimal.NewFromFloat(*record.Usage)
		}
		if record.UsageMeasureId != nil {
			newBillItem.ResAmountUnit = fmt.Sprintf("%d", *record.UsageMeasureId)
		}
		newBillItem.Extension = types.JsonField(extensionBytes)
		itemList = append(itemList, newBillItem)
	}
	supportDone := uint64(len(billResp.Details)) < a.GetBatchSize(kt)
	return itemList, supportDone, nil
}

// Split huawei support fee to main account
func (a HuaweiSupportMonthTask) Split(kt *kit.Kit, opt *MonthTaskActionOption,
	rawItemList []*bill.RawBillItem) ([]bill.BillItemCreateReq[json.RawMessage], error) {

	if len(rawItemList) == 0 {
		return nil, nil
	}
	a.initExtension(opt)

	// 查询根账号信息
	rootAccount, err := actcli.GetDataService().HuaWei.RootAccount.Get(kt, opt.RootAccountID)
	if err != nil {
		logs.Errorf("failt to get root account info, err: %v, accountID: %s, rid: %s", err, opt.RootAccountID, kt.Rid)
		return nil, err
	}

	// rootAsMainAccount 作为二级账号存在的根账号，将分摊后的账单抵冲该账号支出
	mainAccountMap, rootAsMainAccount, err := a.listMainAccount(kt, rootAccount)
	if err != nil {
		logs.Errorf("fail to list main account for huawei month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}

	commonItems, err := a.splitCommonExpense(kt, opt, mainAccountMap, rootAsMainAccount, rawItemList)
	if err != nil {
		logs.Errorf("fail to split common expense for huawei month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}
	return commonItems, nil
}

func (a HuaweiSupportMonthTask) splitCommonExpense(kt *kit.Kit, opt *MonthTaskActionOption,
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
		logs.Errorf("fail to get summary main list for huawei month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}
	if len(summaryList) == 0 {
		logs.Warnf("no main account for huawei month task common expense, opt: %#v, rid: %s", opt, kt.Rid)
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
		extJson, err := convHuaweiBillItemExtension(constant.BillCommonExpenseName, opt, mainAccount.CloudID, cost)
		if err != nil {
			logs.Errorf("fail to marshal huawei common expense extension to json, err: %v, rid: %s", err, kt.Rid)
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
		reverseExtJson, err := convHuaweiBillItemExtension(constant.BillCommonExpenseReverseName, opt,
			mainAccount.CloudID, reverseCost)
		if err != nil {
			logs.Errorf("fail to marshal huawei common expense reverse extension to json, err: %v, rid: %s",
				err, kt.Rid)
			return nil, err
		}

		reverseBillItem := convSummaryToCommonReverse(rootAsMainAccount, summary, reverseCost, reverseExtJson)
		billItems = append(billItems, reverseBillItem)
	}
	return billItems, nil
}

// 不包含根账号自身以及用户设定的排除账号的 二级账号汇总信息
func (a HuaweiSupportMonthTask) listSummaryMainForSupport(kt *kit.Kit, opt *MonthTaskActionOption,
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
			enumor.HuaWei, err, kt.Rid)
		return nil, err
	}
	return summaryMainResp.Details, nil
}

// GetHcProductCodes hc product code ranges
func (a *HuaweiSupportMonthTask) GetHcProductCodes() []string {
	return []string{constant.BillCommonExpenseName, constant.BillCommonExpenseReverseName}
}

// GetBatchSize ...
func (a HuaweiSupportMonthTask) GetBatchSize(kt *kit.Kit) uint64 {
	return 1000
}
