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

// Package aws daily puller
package aws

import (
	"encoding/json"
	"errors"
	"fmt"

	"hcm/cmd/task-server/logics/action/bill/dailypull/registry"
	actcli "hcm/cmd/task-server/logics/action/cli"
	dsbill "hcm/pkg/api/data-service/bill"
	hcbill "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/shopspring/decimal"
)

const (
	awsMaxBill = int32(999)
)

func init() {
	registry.PullerRegistry[enumor.Aws] = &AwsPuller{}
}

// AwsPuller aws puller
type AwsPuller struct{}

// Pull pull aws data
func (hp *AwsPuller) Pull(kt run.ExecuteKit, opt *registry.PullDailyBillOption) (*registry.PullerResult, error) {
	offset := int32(0)
	count := int64(0)
	cost := decimal.NewFromInt(0)
	var currency enumor.CurrencyCode
	for {
		limit := awsMaxBill
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
		offset = offset + awsMaxBill
	}
	return &registry.PullerResult{
		Count:    count,
		Currency: currency,
		Cost:     cost,
	}, nil
}

func (hp *AwsPuller) doPull(kt run.ExecuteKit, opt *registry.PullDailyBillOption, offset *int32, limit *int32) (
	int, *registry.PullerResult, error) {

	hcCli := actcli.GetHCService()
	resp, err := hcCli.Aws.Bill.GetRootAccountBillList(kt.Kit(),
		&hcbill.AwsRootBillListReq{
			RootAccountID:      opt.RootAccountID,
			MainAccountCloudID: opt.BillAccountID,
			BeginDate:          fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, opt.BillDay),
			EndDate:            fmt.Sprintf("%d-%02d-%02d", opt.BillYear, opt.BillMonth, opt.BillDay),
			Page: &hcbill.AwsBillListPage{
				Offset: uint64(cvt.PtrToVal(offset)),
				Limit:  uint64(cvt.PtrToVal(limit)),
			},
		})
	if err != nil {
		return 0, nil, fmt.Errorf("list aws bill failed, err %w", err)
	}

	recordList := resp.Details
	itemLen := len(recordList)
	if itemLen == 0 {
		return 0, &registry.PullerResult{
			Count: int64(0),
			Cost:  decimal.NewFromFloat(0),
		}, nil
	}
	billItems := make([]dsbill.RawBillItem, 0, itemLen)
	cost := decimal.NewFromInt(0)
	currency := enumor.CurrencyCode("")

	for _, record := range recordList {
		bill, err := convertToRawBill(record)
		if err != nil {
			logs.Errorf("fail to convert aws raw bill, err: %v, rid: %s", err, kt.Kit().Rid)
			return 0, nil, err
		}
		if bill.BillCurrency != "" {
			currency = bill.BillCurrency
		}
		cost = cost.Add(bill.BillCost)
		billItems = append(billItems, *bill)
	}
	filename := fmt.Sprintf("%d-%d.csv", *offset, itemLen)

	if err := hp.createRawBill(kt, opt, filename, billItems); err != nil {
		return 0, nil, err
	}
	return itemLen, &registry.PullerResult{
		Count:    int64(resp.Count),
		Currency: currency,
		Cost:     cost,
	}, nil
}
func convertToRawBill(record map[string]string) (*dsbill.RawBillItem, error) {

	cost, err := getDecimal(record, "line_item_net_unblended_cost")
	if err != nil {
		return nil, err
	}
	amount, err := getDecimal(record, "line_item_usage_amount")
	if err != nil {
		return nil, err
	}

	extensionBytes, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("marshal aws bill item %v failed, err: %w", record, err)
	}

	return &dsbill.RawBillItem{
		Region:        record["product_region"],
		HcProductCode: record["line_item_product_code"],
		HcProductName: record["product_product_name"],
		BillCurrency:  enumor.CurrencyCode(record["line_item_currency_code"]),
		BillCost:      *cost,
		ResAmount:     *amount,
		ResAmountUnit: record["pricing_unit"],
		Extension:     types.JsonField(extensionBytes),
	}, nil
}

func (hp *AwsPuller) createRawBill(
	kt run.ExecuteKit, opt *registry.PullDailyBillOption,
	filename string, billItems []dsbill.RawBillItem) error {

	storeReq := &dsbill.RawBillCreateReq{
		Vendor:        enumor.Aws,
		RootAccountID: opt.RootAccountID,
		AccountID:     opt.MainAccountID,
		BillYear:      fmt.Sprintf("%d", opt.BillYear),
		BillMonth:     fmt.Sprintf("%02d", opt.BillMonth),
		BillDate:      fmt.Sprintf("%02d", opt.BillDay),
		Version:       fmt.Sprintf("%d", opt.VersionID),
		FileName:      filename,
	}
	storeReq.Items = billItems
	dataBill := actcli.GetDataService().Global.Bill
	_, err := dataBill.CreateRawBill(kt.Kit(), storeReq)
	if err != nil {
		return fmt.Errorf("call dataservice to create raw bill failed, err %s", err.Error())
	}
	return nil
}

func getDecimal(dict map[string]string, key string) (*decimal.Decimal, error) {
	val, ok := dict[key]
	if !ok {
		return nil, errors.New("key not found: " + key)
	}
	d, err := decimal.NewFromString(val)
	if err != nil {
		return nil, fmt.Errorf("fail to convert to decimal, key: %s, value: %s, err: %v", key, val, err)
	}
	return &d, nil
}
