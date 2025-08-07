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

// Package gcp contains gcp bill pull function
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
	"hcm/pkg/tools/times"

	"github.com/shopspring/decimal"
)

const (
	gcpMaxBill            = int32(1000)
	gcpTimestampExtraDays = 2
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
		logs.Infof("get raw bill item %d / total %d, cost: %s, of puller %+v, rid: %s", itemLen, tmpResult.Count,
			tmpResult.Cost.String(), opt, kt.Kit().Rid)
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
		if record.TotalCost != nil {
			// use original cost with non promotion cost
			newBillItem.BillCost = *record.TotalCost
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
		RawBillPathParam: dsbill.RawBillPathParam{
			Vendor:        enumor.Gcp,
			RootAccountID: opt.RootAccountID,
			MainAccountID: opt.MainAccountID,
			BillYear:      fmt.Sprintf("%d", opt.BillYear),
			BillMonth:     fmt.Sprintf("%02d", opt.BillMonth),
			BillDate:      fmt.Sprintf("%02d", opt.BillDay),
			Version:       fmt.Sprintf("%d", opt.VersionID),
			FileName:      filename,
		},
	}
	storeReq.Items = billItems
	databillCli := actcli.GetDataService().Global.Bill
	_, err := databillCli.CreateRawBill(kt.Kit(), storeReq)
	if err != nil {
		return fmt.Errorf("create raw bill to dataservice failed, err %s", err.Error())
	}
	return nil
}

func (gcp *GcpPuller) getPullDate(kt run.ExecuteKit, opt *registry.PullDailyBillOption) (string, string, error) {
	beginDate := fmt.Sprintf("%d-%02d-%02dT00:00:00Z", opt.BillYear, opt.BillMonth, opt.BillDay)
	endDate := fmt.Sprintf("%d-%02d-%02dT23:59:59Z", opt.BillYear, opt.BillMonth, opt.BillDay)

	// 由于GCP账单TIMESTAMP与invoice.month的月份可能不一致，
	// 比如：TIMESTAMP处于7月1日的账单，实际其invoice.month可能是6。
	// 所以如果是该月最后一天，那么则将TIMESTAMP放大gcpTimestampExtraDays天，保证能够拉取到延迟出帐的那部分账单
	isLastDay, err := times.IsLastDayOfMonth(opt.BillMonth, opt.BillDay)
	if err != nil {
		logs.Warnf("is last day of month failed, err: %v, rid %s", err, kt.Kit().Rid)
		return "", "", err
	}
	if isLastDay {
		tmpYear, tmpMonth, tmpDay, err := times.AddDaysToDate(
			opt.BillYear, opt.BillMonth, opt.BillDay, gcpTimestampExtraDays)
		if err != nil {
			logs.Warnf("add days to date failed, err: %v, rid %s", err, kt.Kit().Rid)
			return "", "", err
		}
		endDate = fmt.Sprintf("%d-%02d-%02dT23:59:59Z", tmpYear, tmpMonth, tmpDay)
	}
	return beginDate, endDate, nil
}

func (gcp *GcpPuller) doPull(
	kt run.ExecuteKit, opt *registry.PullDailyBillOption, offset, limit uint64) (
	int, *registry.PullerResult, error) {

	hcCli := actcli.GetHCService()
	beginDate, endDate, err := gcp.getPullDate(kt, opt)
	if err != nil {
		return 0, nil, err
	}
	resp, err := hcCli.Gcp.Bill.RootAccountBillList(kt.Kit().Ctx, kt.Kit().Header(), &bill.GcpRootAccountBillListReq{
		RootAccountID: opt.RootAccountID,
		MainAccountID: opt.MainAccountID,
		Month:         fmt.Sprintf("%d%02d", opt.BillYear, opt.BillMonth),
		BeginDate:     beginDate,
		EndDate:       endDate,
		Page: &typesBill.GcpBillPage{
			Offset: offset,
			Limit:  limit,
		},
	})
	if err != nil {
		return 0, nil, fmt.Errorf("fail to list gcp root account bill list for %+v, offset %d, limit %d, err %s",
			opt, offset, limit, err.Error())
	}
	var itemList []interface{}
	itemLen := 0
	zeroResult := &registry.PullerResult{
		Count:    int64(0),
		Currency: "",
		Cost:     decimal.NewFromFloat(0),
	}
	if resp.Details != nil {
		ok := false
		itemList, ok = resp.Details.([]interface{})
		if !ok {
			logs.Warnf("response %v is not []billcore.GcpRawBillItem", resp.Details)
			return 0, nil, fmt.Errorf("response %v is not []billcore.GcpRawBillItem", resp.Details)
		}
		itemLen = len(itemList)
		if itemLen == 0 {
			return 0, zeroResult, nil
		}
	} else {
		return 0, zeroResult, nil
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
		record.UsageStartTime = &beginDate
		record.UsageEndTime = &endDate
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
