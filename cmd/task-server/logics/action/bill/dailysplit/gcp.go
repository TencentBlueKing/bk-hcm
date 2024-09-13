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
	billcore "hcm/pkg/api/core/bill"
	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

// GcpSplitter GCP account splitter
type GcpSplitter struct{}

// DoSplit implements RawBillSplitter for GCP
func (ds *GcpSplitter) DoSplit(kt *kit.Kit, opt *DailyAccountSplitActionOption, billDay int,
	item *bill.RawBillItem, mainAccount *protocore.BaseMainAccount) ([]bill.BillItemCreateReq[rawjson.RawMessage],
	error) {

	var billItems []bill.BillItemCreateReq[rawjson.RawMessage]
	var ext billcore.GcpRawBillItem
	if err := rawjson.Unmarshal([]byte(item.Extension), &ext); err != nil {
		logs.Errorf("fail to unmarshal gcp raw bill item extension for split, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// 聚合credit数组
	groupedCredits := ds.groupCredits(ext.CreditInfos)
	ext.CreditInfos = groupedCredits

	rawExt, err := json.Marshal(ext)
	if err != nil {
		logs.Errorf("fail to marshal gcp raw bill item extension for split, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	usageBillItem := bill.BillItemCreateReq[rawjson.RawMessage]{
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
		Extension:     cvt.ValToPtr[rawjson.RawMessage](rawExt),
	}

	billItems = append(billItems, usageBillItem)
	// 将每个 credit 拆分为一条独立账单明细
	for _, credit := range groupedCredits {
		ext.CreditInfos = []billcore.GcpCredit{credit}
		ext.ReturnCost = credit.Amount
		ext.Cost = credit.Amount
		ext.TotalCost = credit.Amount
		rawExt, err := json.Marshal(ext)
		if err != nil {
			logs.Errorf("fail to marshal gcp raw bill item extension for split, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		creditBillItem := bill.BillItemCreateReq[rawjson.RawMessage]{
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
			Cost:          cvt.PtrToVal(credit.Amount),
			HcProductCode: constant.GcpCreditReturnCost,
			HcProductName: credit.ID,
			Extension:     cvt.ValToPtr[rawjson.RawMessage](rawExt),
		}
		billItems = append(billItems, creditBillItem)

	}
	return billItems, nil
}

func (ds *GcpSplitter) groupCredits(creditList []billcore.GcpCredit) (groupedCredits []billcore.GcpCredit) {

	if len(creditList) == 0 {
		return creditList
	}
	creditMap := make(map[string]billcore.GcpCredit)
	for _, credit := range creditList {
		newAmount := cvt.PtrToVal(credit.Amount)
		if _, ok := creditMap[credit.ID]; ok {
			// 存在则加上现有的金额
			newAmount = newAmount.Add(cvt.PtrToVal(creditMap[credit.ID].Amount))
		}
		credit.Amount = cvt.ValToPtr(newAmount)
		creditMap[credit.ID] = credit
	}
	return cvt.MapValueToSlice(creditMap)
}
