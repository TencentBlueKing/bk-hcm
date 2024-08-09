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

// Package azure daily puller
package azure

import (
	"encoding/json"
	"errors"
	"fmt"

	"hcm/cmd/task-server/logics/action/bill/dailypull/registry"
	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/adaptor/types/bill"
	dsbill "hcm/pkg/api/data-service/bill"
	hcbill "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/consumption/armconsumption"
	"github.com/shopspring/decimal"
)

const (
	azureMaxBill = int32(50)
)

func init() {
	registry.PullerRegistry[enumor.Azure] = &AzurePuller{}
}

// AzurePuller azure puller
type AzurePuller struct{}

// Pull pull azure data
func (hp *AzurePuller) Pull(kt run.ExecuteKit, opt *registry.PullDailyBillOption) (*registry.PullerResult, error) {
	nextLink := ""
	count := int64(0)
	// to build raw bill filename, format: [offset]-[length].csv
	offset := 0
	limit := azureMaxBill
	cost := decimal.NewFromInt(0)
	var currency enumor.CurrencyCode
	hcCli := actcli.GetHCService()
	for {
		billReq := &hcbill.AzureRootBillListReq{
			RootAccountID:  opt.RootAccountID,
			SubscriptionID: opt.MainAccountCloudID,
			BeginDate:      buildDateArg(opt),
			EndDate:        buildDateArg(opt),
			Page:           &bill.AzureBillPage{Limit: limit, NextLink: nextLink},
		}
		billResp, err := hcCli.Azure.Bill.GetRootAccountBillList(kt.Kit(), billReq)
		if err != nil {
			logs.Errorf("fail to call Azure to get root account bills, err: %v, rid: %s", err, kt.Kit().Rid)
			return nil, fmt.Errorf("list azure bill failed, err %w", err)
		}

		ret, err := hp.batchCreateRawBill(kt.Kit(), opt, billResp, offset)
		if err != nil {
			logs.Errorf("fail to batch create azure raw bill, err: %v, rid: %s", err, kt.Kit().Rid)
			return nil, err
		}

		nextLink = billResp.NextLink

		currency = ret.Currency
		cost = cost.Add(ret.Cost)
		count += ret.Count
		offset += int(count)
		logs.Infof("get raw azure bill item, count: %d, offset: %d, main account: %s, date: %d-%02d-%02d, rid: %s",
			ret.Count, offset, opt.MainAccountCloudID, opt.BillYear, opt.BillMonth, opt.BillDay, kt.Kit().Rid)
		if int32(ret.Count) < limit {
			break
		}

	}
	pullResult := &registry.PullerResult{
		Count:    count,
		Currency: currency,
		Cost:     cost,
	}
	return pullResult, nil
}

func (hp *AzurePuller) batchCreateRawBill(kt *kit.Kit, opt *registry.PullDailyBillOption,
	billResp *hcbill.AzureLegacyBillListResult, offset int) (*registry.PullerResult, error) {

	ret := new(registry.PullerResult)
	var itemList []dsbill.RawBillItem
	for _, record := range billResp.Details {
		item, err := convertToRawBill(record)
		if err != nil {
			logs.Errorf("fail to convert azure raw bill, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if item.BillCurrency != "" {
			ret.Currency = item.BillCurrency
		}
		ret.Cost = ret.Cost.Add(item.BillCost)
		itemList = append(itemList, cvt.PtrToVal(item))
	}
	filename := buildRawBillFilename(offset, len(itemList))
	if err := hp.createRawBill(kt, opt, filename, itemList); err != nil {
		logs.Errorf("fail to call data servcie to create raw bills, err: %v, filename: %s, rid: %s",
			err, filename, kt.Rid)
		return nil, err
	}
	// current count
	ret.Count = int64(len(itemList))
	return ret, nil
}

// format: [offset]-[length].csv
func buildRawBillFilename(offset, length int) string {
	return fmt.Sprintf("%d-%d.csv", offset, length)
}

// format: yyyy-mm-dd
func buildDateArg(opt *registry.PullDailyBillOption) string {
	return fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, opt.BillDay)
}

func convertToRawBill(record armconsumption.LegacyUsageDetail) (*dsbill.RawBillItem, error) {
	if record.Properties == nil {
		return nil, errors.New("nil azure bill properties")
	}
	extensionBytes, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("marshal azure bill item %v failed, err: %w", record, err)
	}

	item := &dsbill.RawBillItem{
		Region:        cvt.PtrToVal(record.Properties.ResourceLocation),
		HcProductCode: cvt.PtrToVal(record.Properties.ConsumedService),
		HcProductName: cvt.PtrToVal(record.Properties.Product),
		BillCurrency:  enumor.CurrencyCode(cvt.PtrToVal(record.Properties.BillingCurrency)),
		BillCost:      decimal.NewFromFloat(cvt.PtrToVal(record.Properties.Cost)),
		ResAmount:     decimal.NewFromFloat(cvt.PtrToVal(record.Properties.Quantity)),
		ResAmountUnit: cvt.PtrToVal(record.Properties.MeterDetails.UnitOfMeasure),
		Extension:     types.JsonField(extensionBytes),
	}
	return item, nil
}

func (hp *AzurePuller) createRawBill(kt *kit.Kit, opt *registry.PullDailyBillOption, filename string,
	billItems []dsbill.RawBillItem) error {

	storeReq := &dsbill.RawBillCreateReq{
		RawBillPathParam: dsbill.RawBillPathParam{
			Vendor:        enumor.Azure,
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
	dataBill := actcli.GetDataService().Global.Bill
	_, err := dataBill.CreateRawBill(kt, storeReq)
	if err != nil {
		logs.Errorf("fail to create raw bill, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("call dataservice to create raw bill failed, err: %v", err)
	}
	return nil
}
