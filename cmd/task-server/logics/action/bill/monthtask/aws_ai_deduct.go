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
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"

	"github.com/tidwall/gjson"
)

// AwsAIDeductMonthTask ...
type AwsAIDeductMonthTask struct {
	awsMonthTaskBaseRunner
}

// Pull ...
func (a *AwsAIDeductMonthTask) Pull(kt *kit.Kit, opt *MonthTaskActionOption, index uint64) (itemList []bill.RawBillItem,
	isFinished bool, err error) {

	// ai 账单抵扣
	rules := []filter.RuleFactory{
		tools.RuleEqual("root_account_id", opt.RootAccountID),
	}
	aiRules, err := GenAIFilterRules(kt, opt.Vendor)
	if err != nil {
		logs.Errorf("fail to gen ai filter rules, err: %v, rid: %s", err, kt.Rid)
		return nil, false, fmt.Errorf("fail to gen ai filter rules, err: %w", err)
	}
	rules = append(rules, aiRules...)

	flt := &filter.Expression{
		Op:    filter.And,
		Rules: rules,
	}
	page := &core.BasePage{
		Start: uint32(index),
		Limit: uint(a.GetBatchSize(kt)),
		Sort:  "id",
	}
	req := &bill.BillItemListReq{
		ItemCommonOpt: &bill.ItemCommonOpt{
			Vendor: opt.Vendor,
			Year:   opt.BillYear,
			Month:  opt.BillMonth,
		},
		ListReq: &core.ListReq{
			Filter: flt,
			Page:   page,
			Fields: nil,
		},
	}
	itemResult, err := actcli.GetDataService().Global.Bill.ListBillItemRaw(kt, req)
	if err != nil {
		logs.Errorf("fail to get bill item list, err: %v, rid: %s", err, kt.Rid)
		return nil, false, fmt.Errorf("fail to get bill item list, err: %w", err)
	}
	if len(itemResult.Details) == 0 {
		return nil, true, nil
	}
	itemList = make([]bill.RawBillItem, len(itemResult.Details))
	for i := range itemResult.Details {
		item := itemResult.Details[i]
		region := gjson.Get(string(item.Extension), "extension.product_region").String()
		itemList[i] = bill.RawBillItem{
			Region:        region,
			HcProductCode: item.HcProductCode,
			HcProductName: item.HcProductName,
			BillCurrency:  item.Currency,
			BillCost:      item.Cost,
			ResAmount:     item.ResAmount,
			ResAmountUnit: item.ResAmountUnit,
			Extension:     types.JsonField(item.Extension),
		}
	}
	finished := len(itemList) < int(page.Limit)
	return itemList, finished, nil
}

// Split ...
func (a *AwsAIDeductMonthTask) Split(kt *kit.Kit, opt *MonthTaskActionOption,
	rawItemList []*bill.RawBillItem) ([]bill.BillItemCreateReq[json.RawMessage], error) {

	if len(rawItemList) == 0 {
		return nil, nil
	}
	a.initExtension(kt, opt)

	cloudIdSummaryMainMap, err := listSummaryMains(kt, opt)
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
			// 取反抵消账单
			Cost: item.BillCost.Neg(),
			// 覆盖HcProductCode 和HcProductName，防止被被当成原始账单
			HcProductCode: constant.AwsAIDeductProductCode,
			HcProductName: constant.AwsAIDeductProductCode,
			ResAmount:     item.ResAmount,
			ResAmountUnit: item.ResAmountUnit,
			Extension:     cvt.ValToPtr(json.RawMessage(item.Extension)),
		}
		billItems = append(billItems, usageBillItem)
	}
	return billItems, nil
}

// GetHcProductCodes ...
func (a *AwsAIDeductMonthTask) GetHcProductCodes() []string {
	return []string{constant.AwsAIDeductProductCode}
}

// GetBatchSize for aws is always 999
func (a AwsAIDeductMonthTask) GetBatchSize(kt *kit.Kit) uint64 {
	return uint64(core.DefaultMaxPageLimit)
}
