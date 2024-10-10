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

package monthtask

import (
	"encoding/json"
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	typesbill "hcm/pkg/adaptor/types/bill"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/api/hc-service/bill"
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

// GcpSupportMonthTask gcp support month task
type GcpSupportMonthTask struct {
	gcpMonthTaskBaseRunner
}

// GetBatchSize get batch size
func (gcp *GcpSupportMonthTask) GetBatchSize(kt *kit.Kit) uint64 {
	return 1000
}

func convertToRawBill(recordList []billcore.GcpRawBillItem) ([]dsbill.RawBillItem, error) {
	var retList []dsbill.RawBillItem
	for _, record := range recordList {
		extensionBytes, err := json.Marshal(record)
		if err != nil {
			return nil, fmt.Errorf("marshal gcp bill item %v failed", record)
		}
		newBillItem := dsbill.RawBillItem{}
		if record.Region != nil {
			newBillItem.Region = *record.Region
		}
		if record.ServiceID != nil {
			newBillItem.HcProductCode = *record.ServiceID
		}
		if record.ServiceDescription != nil {
			newBillItem.HcProductName = *record.ServiceDescription
		}
		if record.UsageUnit != nil {
			newBillItem.ResAmountUnit = *record.UsageUnit
		}
		if record.UsageAmount != nil {
			newBillItem.ResAmount = *record.UsageAmount
		}
		if record.Currency != nil {
			newBillItem.BillCurrency = enumor.CurrencyCode(*record.Currency)
		}
		if record.Cost != nil {
			// use original cost with non promotion cost
			newBillItem.BillCost = *record.TotalCost
		}
		newBillItem.Extension = types.JsonField(string(extensionBytes))
		retList = append(retList, newBillItem)
	}
	return retList, nil
}

// Pull pull gcp bill
func (gcp *GcpSupportMonthTask) Pull(kt *kit.Kit, opt *MonthTaskActionOption, index uint64) (
	[]dsbill.RawBillItem, bool, error) {

	limit := gcp.GetBatchSize(kt)
	hcCli := actcli.GetHCService()
	req := &bill.GcpRootAccountBillListReq{
		RootAccountID: opt.RootAccountID,
		Month:         fmt.Sprintf("%d%02d", opt.BillYear, opt.BillMonth),
		Page: &typesbill.GcpBillPage{
			Offset: index,
			Limit:  limit,
		},
	}
	resp, err := hcCli.Gcp.Bill.RootAccountBillList(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Warnf("list gcp root account bill list failed, req: %+v, err: %s, rid: %s", req, err.Error(), kt.Rid)
		return nil, false, fmt.Errorf(
			"list gcp root account bill list failed, req: %+v, offset: %d, limit: %d, err: %s",
			req, index, limit, err.Error())
	}
	var itemList []interface{}
	itemLen := 0
	if resp.Details != nil {
		ok := false
		itemList, ok = resp.Details.([]interface{})
		if !ok {
			logs.Warnf("response %v is not []billcore.GcpRawBillItem", resp.Details)
			return nil, false, fmt.Errorf("response %v is not []billcore.GcpRawBillItem", resp.Details)
		}
		itemLen = len(itemList)
		if itemLen == 0 {
			return nil, true, nil
		}
	} else {
		return nil, true, nil
	}

	firstDay, err := times.GetFirstDayOfMonth(opt.BillYear, opt.BillMonth)
	if err != nil {
		return nil, false, fmt.Errorf("times.GetFirstDayOfMonth failed, err: %v", err)
	}
	lastDay, err := times.GetLastDayOfMonth(opt.BillYear, opt.BillMonth)
	if err != nil {
		return nil, false, fmt.Errorf("times.GetLastDayOfMonth failed, err: %v", err)
	}
	beginDate := fmt.Sprintf("%d-%02d-%02dT00:00:00Z", opt.BillYear, opt.BillMonth, firstDay)
	endDate := fmt.Sprintf("%d-%02d-%02dT23:59:59Z", opt.BillYear, opt.BillMonth, lastDay)
	var recordList []billcore.GcpRawBillItem
	for _, item := range itemList {
		respData, err := json.Marshal(item)
		if err != nil {
			return nil, false, fmt.Errorf("marshal gcp response failed, err: %v", err)
		}
		record := billcore.GcpRawBillItem{}
		if err := json.Unmarshal(respData, &record); err != nil {
			return nil, false, fmt.Errorf("decode gcp response failed, err: %v", err)
		}
		record.UsageStartTime = &beginDate
		record.UsageEndTime = &endDate
		recordList = append(recordList, record)
	}
	billItems, err := convertToRawBill(recordList)
	if err != nil {
		return nil, false, err
	}
	return billItems, uint64(itemLen) < limit, nil
}

// Split split gcp bill
func (gcp *GcpSupportMonthTask) Split(kt *kit.Kit, opt *MonthTaskActionOption,
	rawItemList []*dsbill.RawBillItem) ([]dsbill.BillItemCreateReq[json.RawMessage], error) {

	err := gcp.initExtension(opt)
	if err != nil {
		logs.Errorf("fail to init gcp extension for gcp support month task, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(rawItemList) <= 0 {
		return nil, nil
	}
	summaryMainList, err := gcp.getSummaryMainListExcluded(kt, opt.RootAccountID, opt.BillYear, opt.BillMonth)
	if err != nil {
		return nil, err
	}

	billItems := make([]dsbill.BillItemCreateReq[json.RawMessage], 0, len(summaryMainList))
	// 聚合本批次 账单总额，并分摊给每个主账号
	batchCost := decimal.Zero
	for _, item := range rawItemList {
		// 不计算赠金的支出
		cost := item.BillCost
		gcpRaw := billcore.GcpRawBillItem{}
		err := json.Unmarshal([]byte(item.Extension), &gcpRaw)
		if err != nil {
			logs.Errorf("unmarshal gcp raw bill item failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		for _, credit := range gcpRaw.CreditInfos {
			if gcp.creditReturnMap[credit.ID] == "" {
				// 未返还的，正常抵扣即可, credit 金额为负数，直接相加
				cost = cost.Add(cvt.PtrToVal(credit.Amount))
			}
		}
		batchCost = batchCost.Add(cost)
	}
	// 按比例分摊给各个二级账号
	summaryTotal := decimal.Zero
	for _, summaryMain := range summaryMainList {
		summaryTotal = summaryTotal.Add(summaryMain.CurrentMonthCost)
	}
	for _, summaryMain := range summaryMainList {
		cost := batchCost.Mul(summaryMain.CurrentMonthCost).Div(summaryTotal)
		costBillItem := dsbill.BillItemCreateReq[json.RawMessage]{
			RootAccountID: opt.RootAccountID,
			MainAccountID: summaryMain.MainAccountID,
			Vendor:        summaryMain.Vendor,
			ProductID:     summaryMain.ProductID,
			BkBizID:       summaryMain.BkBizID,
			BillYear:      summaryMain.BillYear,
			BillMonth:     summaryMain.BillMonth,
			BillDay:       enumor.MonthTaskSpecialBillDay,
			VersionID:     summaryMain.CurrentVersion,
			Currency:      summaryMain.Currency,
			Cost:          cost,
			HcProductCode: constant.BillCommonExpenseName,
			HcProductName: constant.BillCommonExpenseName,
			Extension:     cvt.ValToPtr(json.RawMessage("{}")),
		}
		billItems = append(billItems, costBillItem)
	}

	return billItems, nil
}

func (gcp *GcpSupportMonthTask) getSummaryMainListExcluded(kt *kit.Kit, rootAccountID string, billYear, billMonth int) (
	[]*dsbill.BillSummaryMain, error) {

	filter := tools.ExpressionAnd(
		tools.RuleEqual("root_account_id", rootAccountID),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
		// 去掉用户需要排除的数据
		tools.RuleNotIn("main_account_cloud_id", gcp.excludeAccountCloudIds),
	)
	req := &dsbill.BillSummaryMainListReq{
		Filter: filter,
		Page:   core.NewCountPage(),
	}
	summaryMainCountResp, err := actcli.GetDataService().Global.Bill.ListBillSummaryMain(kt, req)
	if err != nil {
		logs.Errorf("get gcp %s %d-%02d summary main list failed, err: %v, rid: %s",
			rootAccountID, billYear, billMonth, err, kt.Rid)
		return nil, err
	}
	var summaryMainResultList []*dsbill.BillSummaryMain
	for offset := uint64(0); offset < summaryMainCountResp.Count; offset = offset + uint64(core.DefaultMaxPageLimit) {
		req.Page = &core.BasePage{Start: uint32(offset), Limit: core.DefaultMaxPageLimit}
		summaryMainResp, err := actcli.GetDataService().Global.Bill.ListBillSummaryMain(kt, req)
		if err != nil {
			logs.Errorf("get gcp %s %d-%02d summary main list failed, err: %v, rid: %s",
				rootAccountID, billYear, billMonth, err, kt.Rid)
			return nil, err
		}
		summaryMainResultList = append(summaryMainResultList, summaryMainResp.Details...)
	}

	return summaryMainResultList, nil
}

// GetHcProductCodes hc product code ranges
func (gcp *GcpSupportMonthTask) GetHcProductCodes() []string {
	return []string{constant.BillCommonExpenseName, constant.BillCommonExpenseReverseName}
}
