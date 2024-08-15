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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/times"

	"github.com/shopspring/decimal"
)

func newGcpRunner() *Gcp {
	return &Gcp{}
}

type Gcp struct {
}

// GetBatchSize get batch size
func (gcp *Gcp) GetBatchSize(kt *kit.Kit) uint64 {
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
			newBillItem.BillCost = *record.Cost
		}
		if record.ReturnCost != nil {
			newBillItem.BillCost = newBillItem.BillCost.Add(*record.ReturnCost)
		}
		newBillItem.Extension = types.JsonField(string(extensionBytes))
		retList = append(retList, newBillItem)
	}
	return retList, nil
}

// Pull pull gcp bill
func (gcp *Gcp) Pull(kt *kit.Kit, opt *MonthTaskActionOption, index uint64) (
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
		return nil, false, fmt.Errorf("times.GetFirstDayOfMonth failed, err: %s", err.Error())
	}
	lastDay, err := times.GetLastDayOfMonth(opt.BillYear, opt.BillMonth)
	if err != nil {
		return nil, false, fmt.Errorf("times.GetLastDayOfMonth failed, err: %s", err.Error())
	}
	beginDate := fmt.Sprintf("%d-%02d-%02dT00:00:00Z", opt.BillYear, opt.BillMonth, firstDay)
	endDate := fmt.Sprintf("%d-%02d-%02dT23:59:59Z", opt.BillYear, opt.BillMonth, lastDay)
	var recordList []billcore.GcpRawBillItem
	for _, item := range itemList {
		respData, err := json.Marshal(item)
		if err != nil {
			return nil, false, fmt.Errorf("marshal gcp response failed, err %s", err.Error())
		}
		record := billcore.GcpRawBillItem{}
		if err := json.Unmarshal(respData, &record); err != nil {
			return nil, false, fmt.Errorf("decode gcp response failed, err %s", err.Error())
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
func (gcp *Gcp) Split(kt *kit.Kit, opt *MonthTaskActionOption,
	rawItemList []*dsbill.RawBillItem) ([]dsbill.BillItemCreateReq[json.RawMessage], error) {

	if len(rawItemList) <= 0 {
		return nil, nil
	}
	summaryMainList, err := gcp.getSummaryMainList(kt, opt.RootAccountID, opt.BillYear, opt.BillMonth)
	if err != nil {
		return nil, err
	}
	// 聚合本批次 账单总额，并分摊给每个主账号
	batchCost := decimal.Zero
	for _, item := range rawItemList {
		batchCost = batchCost.Add(item.BillCost)
	}
	// 按比例分摊给各个二级账号
	summaryTotal := decimal.Zero
	for _, summaryMain := range summaryMainList {
		summaryTotal = summaryTotal.Add(summaryMain.CurrentMonthCost)
	}
	billItems := make([]dsbill.BillItemCreateReq[json.RawMessage], 0, len(summaryMainList))
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
			BillDay:       0,
			VersionID:     summaryMain.CurrentVersion,
			Currency:      summaryMain.Currency,
			Cost:          cost,
			HcProductCode: "CommonExpense",
			HcProductName: "CommonExpense",
			Extension:     cvt.ValToPtr(json.RawMessage("{}")),
		}
		billItems = append(billItems, costBillItem)
	}

	return billItems, nil
}

func (gcp *Gcp) getSummaryMainList(
	kt *kit.Kit, rootAccountID string, billYear, billMonth int) ([]*dsbill.BillSummaryMain, error) {

	filter := tools.ExpressionAnd(
		tools.RuleEqual("root_account_id", rootAccountID),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	)
	summaryMainCountResp, err := actcli.GetDataService().Global.Bill.ListBillSummaryMain(kt,
		&dsbill.BillSummaryMainListReq{
			Filter: filter,
			Page:   core.NewCountPage(),
		})
	if err != nil {
		logs.Errorf(
			"count gcp summary main list failed, err: %v, rootAccountID: %s, billYear: %d, billMonth: %d, rid: %s",
			err, rootAccountID, billYear, billMonth, kt.Rid)
		return nil, err
	}
	var summaryMainResultList []*dsbill.BillSummaryMain
	for offset := uint64(0); offset < summaryMainCountResp.Count; offset = offset + uint64(core.DefaultMaxPageLimit) {
		summaryMainCountResp, err := actcli.GetDataService().Global.Bill.ListBillSummaryMain(kt,
			&dsbill.BillSummaryMainListReq{
				Filter: filter,
				Page: &core.BasePage{
					Start: uint32(offset),
					Limit: core.DefaultMaxPageLimit,
				},
			})
		if err != nil {
			logs.Errorf(
				"get gcp summary main list failed, err: %v, rootAccountID: %s, billYear: %d, billMonth: %d, rid: %s",
				err, rootAccountID, billYear, billMonth, kt.Rid)
			return nil, err
		}
		summaryMainResultList = append(summaryMainResultList, summaryMainCountResp.Details...)
	}

	return summaryMainResultList, nil
}
