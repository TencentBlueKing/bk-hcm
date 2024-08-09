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

// Package huawei daily puller
package huawei

import (
	"encoding/json"
	"fmt"

	"hcm/cmd/task-server/logics/action/bill/dailypull/registry"
	actcli "hcm/cmd/task-server/logics/action/cli"
	adbilltypes "hcm/pkg/adaptor/types/bill"
	dsbill "hcm/pkg/api/data-service/bill"
	hcbillservice "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2/model"
	"github.com/shopspring/decimal"
)

const (
	huaweiMaxBill = int32(1000)
)

func init() {
	registry.PullerRegistry[enumor.HuaWei] = &HuaweiPuller{}
}

// HuaweiPuller huawei puller
type HuaweiPuller struct{}

// Pull pull huawei data
func (hp *HuaweiPuller) Pull(kt run.ExecuteKit, opt *registry.PullDailyBillOption) (*registry.PullerResult, error) {
	offset := int32(0)
	count := int64(0)
	cost := decimal.NewFromInt(0)
	var currency enumor.CurrencyCode
	for {
		limit := huaweiMaxBill
		itemLen, tmpResult, err := hp.doPull(kt, opt, &offset, &limit)
		if err != nil {
			return nil, err
		}
		currency = tmpResult.Currency
		cost = cost.Add(tmpResult.Cost)
		count += int64(itemLen)
		logs.Infof("get raw bill item %d / total %d of puller %+v", itemLen, tmpResult.Count, opt)
		if int32(itemLen) < limit {
			break
		}
		offset = offset + huaweiMaxBill
	}
	return &registry.PullerResult{
		Count:    count,
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

func convertToRawBill(currency enumor.CurrencyCode, recordList []model.ResFeeRecordV2) ([]dsbill.RawBillItem, error) {
	var retList []dsbill.RawBillItem
	for _, record := range recordList {
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
			return nil, fmt.Errorf("marshal haiwei bill item %v failed", record)
		}
		newBillItem := dsbill.RawBillItem{}
		if record.Region != nil {
			newBillItem.Region = *record.Region
		}
		if record.CloudServiceType != nil {
			newBillItem.HcProductCode = *record.CloudServiceType
		}
		if record.CloudServiceTypeName != nil {
			newBillItem.HcProductName = *record.CloudServiceTypeName
		}
		newBillItem.BillCurrency = currency
		newBillItem.BillCost = creditCost.Add(debtCost)
		if record.Usage != nil {
			newBillItem.ResAmount = decimal.NewFromFloat(*record.Usage)
		}
		if record.UsageMeasureId != nil {
			newBillItem.ResAmountUnit = fmt.Sprintf("%d", *record.UsageMeasureId)
		}
		newBillItem.Extension = types.JsonField(string(extensionBytes))
		retList = append(retList, newBillItem)
	}
	return retList, nil
}

func (hp *HuaweiPuller) createRawBill(
	kt run.ExecuteKit, opt *registry.PullDailyBillOption,
	filename string, billItems []dsbill.RawBillItem) error {

	storeReq := &dsbill.RawBillCreateReq{
		RawBillPathParam: dsbill.RawBillPathParam{
			Vendor:        enumor.HuaWei,
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

func (hp *HuaweiPuller) doPull(
	kt run.ExecuteKit, opt *registry.PullDailyBillOption, offset *int32, limit *int32) (
	int, *registry.PullerResult, error) {

	hcCli := actcli.GetHCService()
	resp, err := hcCli.HuaWei.Bill.ListFeeRecord(kt.Kit().Ctx, kt.Kit().Header(), &hcbillservice.HuaWeiFeeRecordListReq{
		AccountID:     opt.RootAccountID,
		SubAccountID:  opt.MainAccountCloudID,
		Month:         fmt.Sprintf("%d-%02d", opt.BillYear, opt.BillMonth),
		BillDateBegin: fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, opt.BillDay),
		BillDateEnd:   fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, opt.BillDay),
		Page: &adbilltypes.HuaWeiBillPage{
			Offset: offset,
			Limit:  limit,
		},
	})
	if err != nil {
		return 0, nil, fmt.Errorf("list fee record failed, err %s", err.Error())
	}

	if resp.Count == nil {
		return 0, nil, fmt.Errorf("count in response is empty, resp %v", resp)
	}
	if resp.Details == nil {
		return 0, nil, fmt.Errorf("details in response is empty, resp %v", resp)
	}
	currency := enumor.CurrencyCode(cvt.PtrToVal(resp.Currency))

	itemList, ok := resp.Details.([]interface{})
	if !ok {
		logs.Warnf("response %v is not []model.ResFeeRecordV2", resp.Details)
		return 0, nil, fmt.Errorf("response %v is not []model.ResFeeRecordV2", resp.Details)
	}
	itemLen := len(itemList)
	if itemLen == 0 {
		return 0, &registry.PullerResult{
			Count:    int64(0),
			Currency: currency,
			Cost:     decimal.NewFromFloat(0),
		}, nil
	}

	var recordList []model.ResFeeRecordV2
	for _, item := range itemList {
		itemData, err := json.Marshal(item)
		if err != nil {
			return 0, nil, fmt.Errorf("marshal %v failed", itemData)
		}
		record := model.ResFeeRecordV2{}
		if err := json.Unmarshal(itemData, &record); err != nil {
			return 0, nil, fmt.Errorf("unmarshal %s failed", string(itemData))
		}
		recordList = append(recordList, record)
	}
	filename := fmt.Sprintf("%d-%d.csv", *offset, itemLen)
	billItems, err := convertToRawBill(currency, recordList)
	if err != nil {
		return 0, nil, err
	}
	cost := getRawBillCost(billItems)
	if err := hp.createRawBill(kt, opt, filename, billItems); err != nil {
		return 0, nil, err
	}
	return itemLen, &registry.PullerResult{
		Count:    int64(*resp.Count),
		Currency: currency,
		Cost:     cost,
	}, nil
}
