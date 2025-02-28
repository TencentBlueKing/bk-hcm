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

	actbill "hcm/cmd/task-server/logics/action/bill/common"
	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/data-service/bill"
	hcbill "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/times"

	"github.com/tidwall/gjson"
)

// AwsDeductMonthTask ...
type AwsDeductMonthTask struct {
	awsMonthTaskBaseRunner
}

// Pull deduct bill item
func (a AwsDeductMonthTask) Pull(kt *kit.Kit, opt *MonthTaskActionOption, index uint64) (
	itemList []bill.RawBillItem, isFinished bool, err error) {

	a.initExtension(kt, opt)
	// 检查当前root账号是否有抵扣项配置
	if len(a.deductItemTypes) == 0 || a.deductItemTypes[opt.RootAccountID] == nil {
		logs.Infof("skip aws deduct month task, root account: %s, deductItemTypes: %+v, reason: not need deduct, "+
			"opt: %+v, rid: %s", opt.RootAccountID, a.deductItemTypes, cvt.PtrToVal(opt), kt.Rid)
		return nil, true, nil
	}

	logs.V(3).Infof("[%s] deduct month task pull start, opt: %+v, index: %d, deductItemTypes: %+v, rid: %s", enumor.Aws,
		cvt.PtrToVal(opt), index, a.deductItemTypes, kt.Rid)

	// 解析当前root账号需要查询的字段及值
	fieldsMap := make(map[string][]string)
	for fieldKey, fieldValues := range a.deductItemTypes[opt.RootAccountID] {
		fieldsMap[fieldKey] = fieldValues
	}

	// 获取指定月份最后一天
	lastDay, err := times.GetLastDayOfMonth(opt.BillYear, opt.BillMonth)
	if err != nil {
		logs.Errorf("fail get last day of month for aws month task, year: %d, month: %d, err: %v, rid: %s",
			opt.BillYear, opt.BillMonth, err, kt.Rid)
		return nil, false, err
	}

	rootInfo, err := actcli.GetDataService().Global.RootAccount.GetBasicInfo(kt, opt.RootAccountID)
	if err != nil {
		logs.Errorf("fail to get deduct root account(%s), err: %v, rid: %s", opt.RootAccountID, err, kt.Rid)
		return nil, false, fmt.Errorf("fail to get deduct root account, err: %w", err)
	}

	hcCli := actbill.GetHCServiceByAwsSite(rootInfo.Site)
	billReq := &hcbill.AwsRootBillItemsListReq{
		RootAccountID: opt.RootAccountID,
		Year:          uint(opt.BillYear),
		Month:         uint(opt.BillMonth),
		BeginDate:     fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, 1),
		EndDate:       fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, lastDay),
		FieldsMap:     fieldsMap,
		Page:          &hcbill.AwsBillListPage{Offset: index, Limit: a.GetBatchSize(kt)},
	}
	billResp, err := hcCli.Aws.Bill.ListRootBillItems(kt, billReq)
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
			return nil, false, fmt.Errorf("marshal aws cloud bill item failed, record: %v, err: %w", record, err)
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
	deductDone := uint64(len(billResp.Details)) < a.GetBatchSize(kt)
	return itemList, deductDone, nil
}

// Split aws deduct to main account
func (a AwsDeductMonthTask) Split(kt *kit.Kit, opt *MonthTaskActionOption,
	rawItemList []*bill.RawBillItem) ([]bill.BillItemCreateReq[json.RawMessage], error) {

	if len(rawItemList) == 0 {
		return nil, nil
	}
	a.initExtension(kt, opt)

	cloudIdSummaryMainMap, err := listSummaryMains(kt, opt)
	if err != nil {
		logs.Errorf("fail to list summary main for spilt deduct bill month, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 按实际使用账号抵消到对应账号下即可
	billItems := make([]bill.BillItemCreateReq[json.RawMessage], 0, len(rawItemList))
	for i := range rawItemList {
		item := rawItemList[i]
		// 抵消的账单金额为0，跳过，不再写入account_bill_item表
		if item.BillCost.IsZero() {
			continue
		}

		accountCloudId := gjson.Get(string(item.Extension), "line_item_usage_account_id").String()
		if len(accountCloudId) == 0 {
			logs.Errorf("empty line item usage account id for aws deduct, idx: %d, item: %+v, rid: %s", i, item, kt.Rid)
			return nil, fmt.Errorf("empty line item usage account id for idx: %d", i)
		}
		summaryMain := cloudIdSummaryMainMap[accountCloudId]
		if summaryMain == nil {
			logs.Errorf("can not found main account(%s) for aws deduct bill split, rid: %s", accountCloudId, kt.Rid)
			return nil, fmt.Errorf("can not found main account(%s) for aws deduct bill split", accountCloudId)
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
			Cost:          item.BillCost.Neg(),
			HcProductCode: constant.AwsDeductCostCodeReverse,
			HcProductName: item.HcProductName,
			ResAmount:     item.ResAmount,
			ResAmountUnit: item.ResAmountUnit,
			Extension:     cvt.ValToPtr(json.RawMessage(item.Extension)),
		}
		billItems = append(billItems, usageBillItem)
	}

	return billItems, nil
}

// GetHcProductCodes hc product code ranges
func (a *AwsDeductMonthTask) GetHcProductCodes() []string {
	return []string{constant.AwsDeductCostCodeReverse}
}
