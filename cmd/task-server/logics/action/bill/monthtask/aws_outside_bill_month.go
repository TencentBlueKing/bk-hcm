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
	"hcm/pkg/api/data-service/bill"
	hcbill "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/tidwall/gjson"
)

// AwsOutsideBillMonthTask ...
type AwsOutsideBillMonthTask struct {
	awsMonthTaskBaseRunner
}

// Pull outside bill month item
func (a AwsOutsideBillMonthTask) Pull(kt *kit.Kit, opt *MonthTaskActionOption, index uint64) (
	itemList []bill.RawBillItem, isFinished bool, err error) {

	rootBillReq := &hcbill.AwsRootOutsideMonthBillListReq{
		RootAccountID: opt.RootAccountID,
		BillYear:      uint(opt.BillYear),
		BillMonth:     uint(opt.BillMonth),
		Page:          &hcbill.AwsBillListPage{Offset: index, Limit: a.GetBatchSize(kt)},
	}
	billResp, err := actcli.GetHCService().Aws.Bill.ListRootOutsideMonthBill(kt, rootBillReq)
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
func (a AwsOutsideBillMonthTask) Split(kt *kit.Kit, opt *MonthTaskActionOption,
	rawItemList []*bill.RawBillItem) ([]bill.BillItemCreateReq[json.RawMessage], error) {

	if len(rawItemList) == 0 {
		return nil, nil
	}
	a.initExtension(opt)

	cloudIdSummaryMainMap, err := a.listSummaryMains(kt, opt)
	if err != nil {
		logs.Errorf("fail to list summary main for spilt outside bill month, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 按实际使用账号分摊到对应账号下即可
	billItems := make([]bill.BillItemCreateReq[json.RawMessage], 0, len(rawItemList))
	for i := range rawItemList {
		item := rawItemList[i]
		accountCloudId := gjson.Get(string(item.Extension), "line_item_usage_account_id").String()
		if len(accountCloudId) == 0 {
			return nil, fmt.Errorf("empty line item usage account id for idx: %d", i)
		}
		summaryMain := cloudIdSummaryMainMap[accountCloudId]
		if summaryMain == nil {
			logs.Errorf("can not found main account(%s) for aws outside month bill split, rid: %s",
				accountCloudId, kt.Rid)
			return nil, fmt.Errorf("can not found main account(%s) for aws outside month bill split", accountCloudId)
		}
		usageBillItem := bill.BillItemCreateReq[json.RawMessage]{
			RootAccountID: opt.RootAccountID,
			MainAccountID: summaryMain.MainAccountID,
			Vendor:        opt.Vendor,
			ProductID:     summaryMain.ProductID,
			BkBizID:       summaryMain.BkBizID,
			BillYear:      opt.BillYear,
			BillMonth:     opt.BillMonth,
			BillDay:       enumor.MonthTaskSpecialBillDay,
			VersionID:     summaryMain.CurrentVersion,
			Currency:      item.BillCurrency,
			Cost:          item.BillCost,
			HcProductCode: constant.BillOutsideMonthBillName,
			HcProductName: item.HcProductName,
			ResAmount:     item.ResAmount,
			ResAmountUnit: item.ResAmountUnit,
			Extension:     cvt.ValToPtr(json.RawMessage(item.Extension)),
		}
		billItems = append(billItems, usageBillItem)
	}

	return billItems, nil
}

// 获取summary信息
func (a AwsOutsideBillMonthTask) listSummaryMains(kt *kit.Kit, opt *MonthTaskActionOption) (
	map[string]*bill.BillSummaryMain, error) {

	summaryListReq := &bill.BillSummaryMainListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("root_account_id", opt.RootAccountID),
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
	summaryMap := make(map[string]*bill.BillSummaryMain, len(summaryMainResp.Details))
	for i := range summaryMainResp.Details {
		s := summaryMainResp.Details[i]
		summaryMap[s.MainAccountCloudID] = s
	}
	return summaryMap, nil
}

// GetHcProductCodes hc product code ranges
func (a *AwsOutsideBillMonthTask) GetHcProductCodes() []string {
	return []string{constant.BillOutsideMonthBillName}
}
