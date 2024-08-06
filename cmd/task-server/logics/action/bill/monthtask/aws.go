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

package monthtask

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/account-set"
	"hcm/pkg/api/data-service/bill"
	hcbill "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/shopspring/decimal"
)

func newAwsRunner() MonthTaskRunner {
	return &AwsMonthTask{}
}

// AwsMonthTask ...
type AwsMonthTask struct {
}

// GetBatchSize for aws is always 999
func (a AwsMonthTask) GetBatchSize(kt *kit.Kit) uint64 {
	return 999
}

// Pull aws root account bill
func (a AwsMonthTask) Pull(kt *kit.Kit, rootAccountID string, billYear, billMonth int,
	index uint64) (itemList []bill.RawBillItem, isFinished bool, err error) {

	// 查询根账号信息
	rootAccount, err := actcli.GetDataService().Aws.RootAccount.Get(kt, rootAccountID)
	if err != nil {
		return nil, false, err
	}

	// 获取指定月份最后一天
	lastDay := getLastDayOfMonth(billYear, billMonth)

	billResp, err := actcli.GetHCService().Aws.Bill.GetRootAccountBillList(kt, &hcbill.AwsRootBillListReq{
		RootAccountID:      rootAccountID,
		MainAccountCloudID: rootAccount.CloudID,
		BeginDate:          fmt.Sprintf("%d-%02d-%02d", billYear, billMonth, 1),
		EndDate:            fmt.Sprintf("%d-%02d-%02d", billYear, billMonth, lastDay),
		Page: &hcbill.AwsBillListPage{
			Offset: index,
			Limit:  a.GetBatchSize(kt),
		},
	})
	if err != nil {
		return nil, false, err
	}
	if len(billResp.Details) == 0 {
		return nil, true, nil
	}
	for _, record := range billResp.Details {
		cost, err := getDecimal(record, "line_item_net_unblended_cost")
		if err != nil {
			return nil, false, err
		}
		amount, err := getDecimal(record, "line_item_usage_amount")
		if err != nil {
			return nil, false, err
		}

		extensionBytes, err := json.Marshal(record)
		if err != nil {
			return nil, false, fmt.Errorf("marshal aws bill item %v failed, err: %w", record, err)
		}
		itemList = append(itemList, bill.RawBillItem{
			Region:        record["product_region"],
			HcProductCode: record["line_item_product_code"],
			HcProductName: record["product_product_name"],
			BillCurrency:  enumor.CurrencyCode(record["line_item_currency_code"]),
			BillCost:      *cost,
			ResAmount:     *amount,
			ResAmountUnit: record["pricing_unit"],
			Extension:     types.JsonField(extensionBytes),
		})
	}
	finished := uint64(len(billResp.Details)) < a.GetBatchSize(kt)
	return itemList, finished, nil
}

// Split root account bill into main accounts
func (a AwsMonthTask) Split(kt *kit.Kit, rootAccountID string, billYear, billMonth int,
	rawItemList []*bill.RawBillItem) ([]bill.BillItemCreateReq[json.RawMessage], error) {

	if len(rawItemList) == 0 {
		return nil, nil
	}
	// 查询根账号信息
	rootAccount, err := actcli.GetDataService().Aws.RootAccount.Get(kt, rootAccountID)
	if err != nil {
		logs.Errorf("failt to get root account info, err: %v, accountID: %s, rid: %s", err, rootAccountID, kt.Rid)
		return nil, err
	}

	mainAccountsResp, err := actcli.GetDataService().Global.MainAccount.List(kt, &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("parent_account_id", rootAccountID)),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("failt to list main account for %s month task, err: %v, rid: %s",
			enumor.Aws, err, kt.Rid)
		return nil, err
	}
	mainAccounts := mainAccountsResp.Details
	mainAccountIDs := make([]string, 0, len(mainAccountsResp.Details))
	// 作为二级账号存在的根账号，用于拉取公共费用
	var rootAsMainAccount *protocore.BaseMainAccount
	for _, account := range mainAccounts {
		if account.CloudID == rootAccount.CloudID {
			rootAsMainAccount = account
			continue
		}
		mainAccountIDs = append(mainAccountIDs, account.ID)
	}

	summaryMainResp, err := actcli.GetDataService().Global.Bill.ListBillSummaryMain(kt, &bill.BillSummaryMainListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("main_account_id", mainAccountIDs),
			tools.RuleEqual("bill_year", billYear),
			tools.RuleEqual("bill_month", billMonth),
		),
		Page: core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("failt to list main account bill summary for %s month task, err: %v, rid: %s",
			enumor.Aws, err, kt.Rid)
		return nil, err
	}

	// 聚合本批次 账单总额，并分摊给每个主账号
	batchCost := decimal.Zero
	for _, item := range rawItemList {
		batchCost = batchCost.Add(item.BillCost)
	}
	// 按比例分摊给各个二级账号
	summaryTotal := decimal.Zero
	for _, summaryMain := range summaryMainResp.Details {
		summaryTotal = summaryTotal.Add(summaryMain.CurrentMonthCost)
	}

	billItems := make([]bill.BillItemCreateReq[json.RawMessage], 0, len(summaryMainResp.Details))

	for i := range summaryMainResp.Details {
		summary := summaryMainResp.Details[i]
		cost := batchCost.Mul(summary.CurrentMonthCost).Div(summaryTotal)
		costBillItem := bill.BillItemCreateReq[json.RawMessage]{
			RootAccountID: rootAccountID,
			MainAccountID: summary.MainAccountID,
			Vendor:        summary.Vendor,
			ProductID:     summary.ProductID,
			BkBizID:       summary.BkBizID,
			BillYear:      summary.BillYear,
			BillMonth:     summary.BillMonth,
			BillDay:       0,
			VersionID:     summary.CurrentVersion,
			Currency:      summary.Currency,
			Cost:          cost,
			HcProductCode: "CommonExpense",
			HcProductName: "CommonExpense",
			Extension:     cvt.ValToPtr(json.RawMessage("{}")),
		}
		billItems = append(billItems, costBillItem)

		if rootAsMainAccount == nil {
			continue
		}
		// 此处冲平根账号支出
		reverseBillItem := bill.BillItemCreateReq[json.RawMessage]{
			RootAccountID: rootAccountID,
			MainAccountID: rootAsMainAccount.CloudID,
			Vendor:        summary.Vendor,
			ProductID:     summary.ProductID,
			BkBizID:       summary.BkBizID,
			BillYear:      summary.BillYear,
			BillMonth:     summary.BillMonth,
			BillDay:       0,
			VersionID:     summary.CurrentVersion,
			Currency:      summary.Currency,
			Cost:          cost.Neg(),
			HcProductCode: "CommonExpenseReverse",
			HcProductName: "CommonExpenseReverse",
			Extension:     cvt.ValToPtr(json.RawMessage("{}")),
		}
		billItems = append(billItems, reverseBillItem)
	}

	return billItems, nil
}

// 获取指定月份的最后一天
func getLastDayOfMonth(year int, month int) int {
	date := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.Local)

	return date.AddDate(0, 0, -1).Day()
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
