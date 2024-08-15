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
	"strings"

	protocore "hcm/pkg/api/core/account-set"
	corebill "hcm/pkg/api/core/bill"
	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/shopspring/decimal"
)

const AwsSavingsPlanAccountCloudIDKey = "aws_savings_plan_account_cloud_id"
const AwsSavingsPlanARNPrefixKey = "aws_savings_plan_arn_prefix"
const AwsLineItemTypeSavingPlanCoveredUsage = "SavingsPlanCoveredUsage"
const AwsSavingsPlansCostCode = "SavingsPlanCost"
const AwsSavingsPlansCostCodeReverse = "SavingsPlanCostReverse"

// AwsSplitter default account splitter
type AwsSplitter struct {
}

// DoSplit implements RawBillSplitter
func (ds *AwsSplitter) DoSplit(kt *kit.Kit, opt *DailyAccountSplitActionOption, billDay int,
	item *bill.RawBillItem, mainAccount *protocore.BaseMainAccount) ([]bill.BillItemCreateReq[rawjson.RawMessage],
	error) {

	var billItems []bill.BillItemCreateReq[rawjson.RawMessage]

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
		Extension:     cvt.ValToPtr(rawjson.RawMessage(item.Extension)),
	}
	billItems = append(billItems, usageBillItem)

	spItems, err := ds.extractSpCostItem(kt, opt, billDay, mainAccount, item)
	if err != nil {
		logs.Errorf("fail to extract sp cost item, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	billItems = append(billItems, spItems...)

	return billItems, nil
}

func (ds *AwsSplitter) extractSpCostItem(kt *kit.Kit, opt *DailyAccountSplitActionOption, billDay int,
	mainAccount *protocore.BaseMainAccount, item *bill.RawBillItem) ([]bill.BillItemCreateReq[rawjson.RawMessage],
	error) {

	if opt.Extension == nil {
		return nil, nil
	}
	spPrefix := opt.Extension[AwsSavingsPlanARNPrefixKey]
	if len(spPrefix) == 0 {
		return nil, nil
	}

	var ext corebill.AwsBillItemExtension
	if err := rawjson.Unmarshal([]byte(item.Extension), &ext); err != nil {
		return nil, err
	}

	if ext.LineItemLineItemType != AwsLineItemTypeSavingPlanCoveredUsage {
		return nil, nil
	}
	if !strings.HasPrefix(ext.SavingsPlanSavingsPlanARN, spPrefix) {
		return nil, nil
	}

	spNetCostStr := ext.SavingsPlanNetSavingsPlanEffectiveCost
	spNetCost, err := decimal.NewFromString(spNetCostStr)
	if err != nil {
		logs.Errorf("fail to parse aws sp net cost, err: %v, raw str: %s  rid: %s", err, spNetCostStr, kt.Rid)
		return nil, err
	}
	// 将的saving plan 消耗转化为两笔明细：
	// 1. 消耗资源账号的支出明细，对应cost 为 savings_plan_net_savings_plan_effective_cost
	// 2. 购买SP的账号的收入明细，对应cost 为 -savings_plan_net_savings_plan_effective_cost (该步骤在month task中实现)
	spUsageItem := bill.BillItemCreateReq[rawjson.RawMessage]{
		RootAccountID: opt.RootAccountID,
		// 转化为消耗账户实际支出
		MainAccountID: opt.MainAccountID,
		Vendor:        opt.Vendor,
		ProductID:     mainAccount.OpProductID,
		BkBizID:       mainAccount.BkBizID,
		BillYear:      opt.BillYear,
		BillMonth:     opt.BillMonth,
		BillDay:       billDay,
		VersionID:     opt.VersionID,
		Currency:      item.BillCurrency,
		Cost:          spNetCost,
		HcProductCode: AwsSavingsPlansCostCode,
		HcProductName: AwsSavingsPlansCostCode,
		Extension:     cvt.ValToPtr(rawjson.RawMessage(item.Extension)),
	}

	return []bill.BillItemCreateReq[rawjson.RawMessage]{spUsageItem}, nil
}

// BuildAwsDailySplitOptionExt build aws daily split option extension
func BuildAwsDailySplitOptionExt(spArnPrefix string) map[string]string {
	return map[string]string{
		AwsSavingsPlanARNPrefixKey: spArnPrefix,
	}
}
