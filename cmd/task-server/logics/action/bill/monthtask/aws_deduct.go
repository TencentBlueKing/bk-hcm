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
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/shopspring/decimal"
)

// AwsDeductMonthTask ...
type AwsDeductMonthTask struct {
	awsMonthTaskBaseRunner
}

// Pull deduct tax bill item
func (a AwsDeductMonthTask) Pull(kt *kit.Kit, opt *MonthTaskActionOption, _ uint64) (itemList []bill.RawBillItem,
	isFinished bool, err error) {

	// 查询指定月份的账单明细记录
	commonOpt := &bill.ItemCommonOpt{
		Vendor: opt.Vendor,
		Year:   opt.BillYear,
		Month:  opt.BillMonth,
	}
	billItemReq := &bill.BillItemListReq{
		ItemCommonOpt: commonOpt,
		ListReq: &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("root_account_id", opt.RootAccountID),
				tools.RuleIn("extension.line_item_line_item_type", a.deductItemTypes),
			),
			Page: core.NewDefaultBasePage(),
		},
	}
	billItemResp, err := actcli.GetDataService().Global.Bill.ListBillItem(kt, billItemReq)
	if err != nil {
		return nil, false, err
	}
	if len(billItemResp.Details) == 0 {
		return nil, true, nil
	}
	for _, item := range billItemResp.Details {
		extBytes, err := json.Marshal(item)
		if err != nil {
			logs.Errorf("marshal aws bill item failed, err: %v, billItem: %+v, rid: %s", err, item, kt.Rid)
			return nil, false, fmt.Errorf("marshal aws bill item failed, billItem: %+v, err: %w", item, err)
		}
		itemList = append(itemList, bill.RawBillItem{
			Region:        "any",
			HcProductCode: constant.AwsDeductPlanCostCodeReverse,
			HcProductName: item.HcProductName,
			BillCurrency:  item.Currency,
			BillCost:      item.Cost,
			ResAmount:     item.ResAmount,
			ResAmountUnit: item.ResAmountUnit,
			Extension:     types.JsonField(extBytes),
		})
	}
	deductDone := uint64(len(billItemResp.Details)) < a.GetBatchSize(kt)
	return itemList, deductDone, nil
}

// Split aws deduct tax fee to main account
func (a AwsDeductMonthTask) Split(kt *kit.Kit, opt *MonthTaskActionOption,
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

	commonItems, err := a.splitProductCodeReverse(kt, opt, mainAccountMap, rootAsMainAccount, rawItemList,
		constant.AwsDeductPlanCostCodeReverse)
	if err != nil {
		logs.Errorf("fail to split deduct for aws month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}
	return commonItems, nil
}

func (a AwsDeductMonthTask) splitProductCodeReverse(kt *kit.Kit, opt *MonthTaskActionOption,
	mainAccountMap map[string]*protocore.BaseMainAccount, rootAsMainAccount *protocore.BaseMainAccount,
	rawItemList []*bill.RawBillItem, hcProductCode string) ([]bill.BillItemCreateReq[json.RawMessage], error) {

	if len(rawItemList) == 0 {
		return nil, nil
	}

	// 聚合本批次 账单总额，并分摊给每个主账号
	batchSum := decimal.Zero
	for _, item := range rawItemList {
		batchSum = batchSum.Add(item.BillCost)
	}

	var summaryList []*bill.BillSummaryMain
	summaryList, err := a.listSummaryMainByRootCloudID(kt, opt, mainAccountMap, rootAsMainAccount.CloudID)
	if err != nil {
		logs.Errorf("fail to get summary main list for aws month task split step, err: %v, opt: %#v, rid: %s",
			err, opt, kt.Rid)
		return nil, err
	}
	if len(summaryList) == 0 {
		logs.Warnf("no main account for aws month task reverse expense, hcProductCode: %s, opt: %#v, rid: %s",
			hcProductCode, opt, kt.Rid)
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
		// 此处冲平根账号支出
		reverseCost := cost.Neg()
		reverseExtJson, err := convAwsBillItemExtension(hcProductCode, opt, summary.RootAccountCloudID,
			mainAccount.CloudID, summary.Currency, reverseCost)
		if err != nil {
			logs.Errorf("fail to marshal aws deduct expense reverse extension to json, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		reverseBillItem := convSummaryToProductCodeReverse(hcProductCode, rootAsMainAccount, summary, reverseCost,
			reverseExtJson)
		billItems = append(billItems, reverseBillItem)
	}
	return billItems, nil
}

// 不包含根账号自身以及用户设定的排除账号的 二级账号汇总信息
func (a AwsDeductMonthTask) listSummaryMainByRootCloudID(kt *kit.Kit, opt *MonthTaskActionOption,
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
		logs.Errorf("fail to list main account bill summary for %s month task, err: %v, type: %s, step: %s, rid: %s",
			enumor.Aws, err, opt.Type, opt.Step, kt.Rid)
		return nil, err
	}
	return summaryMainResp.Details, nil
}

func convSummaryToProductCodeReverse(hcProductCode string, mainAccount *protocore.BaseMainAccount,
	summary *bill.BillSummaryMain, cost decimal.Decimal, extension []byte) bill.BillItemCreateReq[json.RawMessage] {

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
		HcProductCode: hcProductCode,
		HcProductName: hcProductCode,
		Extension:     cvt.ValToPtr[json.RawMessage](extension),
	}
	return reverseBillItem
}

// GetHcProductCodes hc product code ranges
func (a *AwsDeductMonthTask) GetHcProductCodes() []string {
	return []string{constant.AwsDeductPlanCostCodeReverse}
}
