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

package gcp

import (
	"encoding/json"
	"fmt"

	"hcm/cmd/task-server/logics/action/bill/dailypull/registry"
	actcli "hcm/cmd/task-server/logics/action/cli"
	typesBill "hcm/pkg/adaptor/types/bill"
	billcore "hcm/pkg/api/core/bill"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/api/hc-service/bill"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"

	"github.com/shopspring/decimal"
)

const (
	gcpMaxBill = int32(1000)
)

func init() {
	registry.PullerRegistry[enumor.Gcp] = &GcpPuller{}
}

// GcpPuller gcp puller
type GcpPuller struct {
}

// Pull pull gcp bill data
func (gcp *GcpPuller) Pull(kt run.ExecuteKit, opt *registry.PullDailyBillOption) (*registry.PullerResult, error) {
	offset := uint64(0)
	count := uint64(0)
	cost := decimal.NewFromInt(0)
	currency := enumor.CurrencyCode("")
	for {
		limit := uint64(gcpMaxBill)
		itemLen, tmpResult, err := gcp.doPull(kt, opt, offset, limit)
		if err != nil {
			return nil, err
		}
		currency = tmpResult.Currency
		cost = cost.Add(tmpResult.Cost)
		count += uint64(itemLen)
		logs.Infof("get raw bill item %d / total %d of puller %+v", itemLen, tmpResult.Count, opt)
		if uint64(itemLen) < limit {
			break
		}
		offset = offset + uint64(gcpMaxBill)
	}
	return &registry.PullerResult{
		Count:    int64(count),
		Currency: currency,
		Cost:     cost,
	}, nil
}

func getRawBillCost(rawBills []dsbill.RawBillItem) decimal.Decimal {
	cost := decimal.NewFromInt(0)
	for _, bill := range rawBills {
		cost = cost.Add(bill.BillCost)
	}
	return cost
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

func (gcp *GcpPuller) createRawBill(
	kt run.ExecuteKit, opt *registry.PullDailyBillOption,
	filename string, billItems []dsbill.RawBillItem) error {

	storeReq := &dsbill.RawBillCreateReq{
		Vendor:        enumor.Gcp,
		RootAccountID: opt.RootAccountID,
		AccountID:     opt.MainAccountID,
		BillYear:      fmt.Sprintf("%d", opt.BillYear),
		BillMonth:     fmt.Sprintf("%02d", opt.BillMonth),
		BillDate:      fmt.Sprintf("%02d", opt.BillDay),
		Version:       fmt.Sprintf("%d", opt.VersionID),
		FileName:      filename,
	}
	storeReq.Items = billItems
	databillCli := actcli.GetDataService().Global.Bill
	_, err := databillCli.CreateRawBill(kt.Kit(), storeReq)
	if err != nil {
		return fmt.Errorf("create raw bill to dataservice failed, err %s", err.Error())
	}
	return nil
}

func (gcp *GcpPuller) doPull(
	kt run.ExecuteKit, opt *registry.PullDailyBillOption, offset, limit uint64) (
	int, *registry.PullerResult, error) {

	hcCli := actcli.GetHCService()
	resp, err := hcCli.Gcp.Bill.RootAccountBillList(kt.Kit().Ctx, kt.Kit().Header(), &bill.GcpRootAccountBillListReq{
		RootAccountID: opt.RootAccountID,
		MainAccountID: opt.MainAccountID,
		BeginDate:     fmt.Sprintf("%d-%02d-%02dT00:00:00Z", opt.BillYear, opt.BillMonth, opt.BillDay),
		EndDate:       fmt.Sprintf("%d-%02d-%02dT23:59:59Z", opt.BillYear, opt.BillMonth, opt.BillDay),
		Page: &typesBill.GcpBillPage{
			Offset: offset,
			Limit:  limit,
		},
	})
	if err != nil {
		return 0, nil, fmt.Errorf("list gcp root account bill list for %+v, offset %d, limit %d, err %s",
			opt, offset, limit, err.Error())
	}
	var itemList []interface{}
	itemLen := 0
	if resp.Details != nil {
		ok := false
		itemList, ok = resp.Details.([]interface{})
		if !ok {
			logs.Warnf("response %v is not []billcore.GcpRawBillItem", resp.Details)
			return 0, nil, fmt.Errorf("response %v is not []billcore.GcpRawBillItem", resp.Details)
		}
		itemLen = len(itemList)
		if itemLen == 0 {
			return 0, &registry.PullerResult{
				Count:    int64(0),
				Currency: "",
				Cost:     decimal.NewFromFloat(0),
			}, nil
		}
	} else {
		return 0, &registry.PullerResult{
			Count:    int64(0),
			Currency: "",
			Cost:     decimal.NewFromFloat(0),
		}, nil
	}

	currency := enumor.CurrencyCode("")
	var recordList []billcore.GcpRawBillItem
	for _, item := range itemList {
		respData, err := json.Marshal(item)
		if err != nil {
			return 0, nil, fmt.Errorf("marshal gcp response failed, err %s", err.Error())
		}
		record := billcore.GcpRawBillItem{}
		if err := json.Unmarshal(respData, &record); err != nil {
			return 0, nil, fmt.Errorf("decode gcp response failed, err %s", err.Error())
		}
		if record.Currency != nil {
			currency = enumor.CurrencyCode(*record.Currency)
		}
		recordList = append(recordList, record)
	}
	filename := fmt.Sprintf("%d-%d.csv", offset, itemLen)
	billItems, err := convertToRawBill(recordList)
	if err != nil {
		return 0, nil, err
	}
	cost := getRawBillCost(billItems)
	if err := gcp.createRawBill(kt, opt, filename, billItems); err != nil {
		return 0, nil, err
	}
	return itemLen, &registry.PullerResult{
		Count:    int64(len(itemList)),
		Currency: currency,
		Cost:     cost,
	}, nil
}
