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

package dailysplit

import (
	rawjson "encoding/json"

	protocore "hcm/pkg/api/core/account-set"
	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/shopspring/decimal"
)

// AwsSplitter default account splitter
type AwsSplitter struct {
	currency enumor.CurrencyCode
	spCost   *decimal.Decimal
}

// DoSplit implements RawBillSplitter
func (ds *AwsSplitter) DoSplit(kt *kit.Kit, opt *DailyAccountSplitActionOption, billDay int,
	item *bill.RawBillItem, mainAccount *protocore.BaseMainAccount) ([]bill.BillItemCreateReq[rawjson.RawMessage],
	error) {

	var ext map[string]string
	if err := rawjson.Unmarshal([]byte(item.Extension), &ext); err != nil {
		return nil, err
	}
	if ext["line_item_line_item_type"] == "SavingsPlanCoveredUsage" {
		spStr := ext["savings_plan_net_savings_plan_effective_cost"]
		spNetCost, err := decimal.NewFromString(spStr)
		if err != nil {
			logs.Errorf("fail to parse aws sp net cost, err: %v, raw str: %s  rid: %s", err, spStr, kt.Rid)
			return nil, err
		}
		if ds.spCost == nil {
			ds.spCost = cvt.ValToPtr(spNetCost)
		} else {
			ds.spCost = cvt.ValToPtr(ds.spCost.Add(spNetCost))
		}
		ds.currency = item.BillCurrency

	}

	billItemCreate := bill.BillItemCreateReq[rawjson.RawMessage]{
		RootAccountID: opt.RootAccountID,
		MainAccountID: opt.MainAccountID,
		Vendor:        opt.Vendor,
		ProductID:     mainAccount.OpProductID,
		BkBizID:       mainAccount.BkBizID,
		BillYear:      opt.BillYear,
		BillMonth:     opt.BillMonth,
		BillDay:       billDay,
		VersionID:     opt.VersionID,
		Currency:      item.BillCurrency,
		Cost:          item.BillCost,
		HcProductCode: item.HcProductCode,
		HcProductName: item.HcProductName,
		ResAmount:     item.ResAmount,
		ResAmountUnit: item.ResAmountUnit,
		Extension:     cvt.ValToPtr(rawjson.RawMessage(item.Extension)),
	}
	return []bill.BillItemCreateReq[rawjson.RawMessage]{billItemCreate}, nil
}

// FinishSplit implements RawBillSplitter
func (ds *AwsSplitter) FinishSplit(kt *kit.Kit, opt *DailyAccountSplitActionOption, billDay int,
	mainAccount *protocore.BaseMainAccount) ([]bill.BillItemCreateReq[rawjson.RawMessage], error) {

	if ds.spCost == nil {
		return nil, nil
	}
	billItemCreate := bill.BillItemCreateReq[rawjson.RawMessage]{
		RootAccountID: opt.RootAccountID,
		MainAccountID: opt.MainAccountID,
		Vendor:        opt.Vendor,
		ProductID:     mainAccount.OpProductID,
		BkBizID:       mainAccount.BkBizID,
		BillYear:      opt.BillYear,
		BillMonth:     opt.BillMonth,
		BillDay:       billDay,
		VersionID:     opt.VersionID,
		Currency:      ds.currency,
		Cost:          cvt.PtrToVal(ds.spCost),
		HcProductCode: "SavingsPlanNetCost",
		HcProductName: "SavingsPlanNetCost",
		Extension:     cvt.ValToPtr(rawjson.RawMessage("{}")),
	}
	return []bill.BillItemCreateReq[rawjson.RawMessage]{billItemCreate}, nil
}
