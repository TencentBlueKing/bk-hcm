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
	"fmt"
	"time"

	actcli "hcm/cmd/task-server/logics/action/cli"
	typesbill "hcm/pkg/adaptor/types/bill"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/api/hc-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/times"
)

// GcpCreditMonthTask 赠金抽取到对应账号下
type GcpCreditMonthTask struct {
	gcpMonthTaskBaseRunner
}

// GetBatchSize ...
func (g GcpCreditMonthTask) GetBatchSize(kt *kit.Kit) uint64 {
	return 1000
}

// Pull gcp credits list
func (g GcpCreditMonthTask) Pull(kt *kit.Kit, opt *MonthTaskActionOption, index uint64) (itemList []dsbill.RawBillItem,
	isFinished bool, err error) {

	limit := g.GetBatchSize(kt)
	hcCli := actcli.GetHCService()
	req := &bill.GcpRootAccountBillListReq{
		RootAccountID: opt.RootAccountID,
		// 查询所有赠金使用情况
		Month: fmt.Sprintf("%d%02d", opt.BillYear, opt.BillMonth),
		Page: &typesbill.GcpBillPage{
			Offset: index,
			Limit:  limit,
		},
	}
	resp, err := hcCli.Gcp.Bill.RootCreditUsageList(kt, req)
	if err != nil {
		logs.Warnf("list gcp root credit list failed, req: %+v, err: %s, rid: %s", req, err.Error(), kt.Rid)
		return nil, false, fmt.Errorf(
			"list gcp root credit list failed, req: %+v, offset: %d, limit: %d, err: %s",
			req, index, limit, err.Error())
	}
	if len(resp.Details) == 0 {
		return nil, true, nil
	}
	itemLen := len(resp.Details)
	firstDay, err := times.GetFirstDayOfMonth(opt.BillYear, opt.BillMonth)
	if err != nil {
		return nil, false, fmt.Errorf("times.GetFirstDayOfMonth failed, err: %s", err.Error())
	}
	utcBillYearMonth := time.Date(opt.BillYear, time.Month(opt.BillMonth), 1, 0, 0, 0, 0, time.UTC)
	nextMonthYear, nextMonth := times.GetRelativeMonth(utcBillYearMonth, 1)
	beginDate := fmt.Sprintf("%d-%02d-%02dT00:00:00Z", opt.BillYear, opt.BillMonth, firstDay)
	endDate := fmt.Sprintf("%d-%02d-%02dT23:59:59Z", nextMonthYear, nextMonth, 1)
	var recordList []billcore.GcpRawBillItem
	for _, item := range resp.Details {
		record := billcore.GcpRawBillItem{
			BillingAccountID:       item.BillingAccountId,
			Cost:                   item.PromotionCredit,
			TotalCost:              item.PromotionCredit,
			Currency:               cvt.ValToPtr(item.Currency),
			CurrencyConversionRate: item.CurrencyConversionRate,
			Month:                  cvt.ValToPtr(item.Month),
			ProjectID:              cvt.ValToPtr(item.ProjectId),
			ProjectName:            cvt.ValToPtr(item.ProjectName),
			ProjectNumber:          cvt.ValToPtr(item.ProjectNumber),
			CreditInfos:            item.Credits,
			UsageStartTime:         cvt.ValToPtr(beginDate),
			UsageEndTime:           cvt.ValToPtr(endDate),
		}
		recordList = append(recordList, record)
	}
	billItems, err := convertToRawBill(recordList)
	if err != nil {
		return nil, false, err
	}
	return billItems, uint64(itemLen) < limit, nil
}

// Split gcp credits
func (g GcpCreditMonthTask) Split(kt *kit.Kit, opt *MonthTaskActionOption, rawItemList []*dsbill.RawBillItem) (
	result []dsbill.BillItemCreateReq[json.RawMessage], err error) {

	if len(rawItemList) == 0 {
		return nil, nil
	}
	// 	根据配置把指定的赠金id返还到指定的账号下
	if err = g.initExtension(opt); err != nil {
		logs.Errorf("init extension failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	summaryMainReq := &dsbill.BillSummaryMainListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bill_year", opt.BillYear),
			tools.RuleEqual("bill_month", opt.BillMonth),
			tools.RuleEqual("root_account_id", opt.RootAccountID)),
		Page: core.NewDefaultBasePage(),
	}
	summaryMainResp, err := actcli.GetDataService().Global.Bill.ListBillSummaryMain(kt, summaryMainReq)
	if err != nil {
		logs.Errorf("fail to list bill summary main for credit split, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	creditToSummary := make(map[string]*dsbill.BillSummaryMain)
	summaryMap := make(map[string]*dsbill.BillSummaryMain)
	for i, detail := range summaryMainResp.Details {
		summaryMap[detail.MainAccountCloudID] = summaryMainResp.Details[i]
	}
	for creditID, mainCloudID := range g.creditReturnMap {
		if summaryMap[mainCloudID] == nil {
			return nil, fmt.Errorf("summary main for credit %s not found, main account: %s, rid: %s",
				mainCloudID, creditID, kt.Rid)
		}
		creditToSummary[creditID] = summaryMap[mainCloudID]
	}

	for _, item := range rawItemList {
		gcpRaw := billcore.GcpRawBillItem{}
		err := json.Unmarshal([]byte(item.Extension), &gcpRaw)
		if err != nil {
			logs.Errorf("unmarshal gcp raw bill item failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		for _, credit := range gcpRaw.CreditInfos {
			creditOwnerSummary := creditToSummary[credit.ID]
			if creditOwnerSummary == nil {
				continue
			}
			ownerItems, err := g.convCreditToOwnerRaw(gcpRaw, credit, creditOwnerSummary)
			if err != nil {
				logs.Errorf("fail to conv credit to owner raw bill item, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			result = append(result, ownerItems...)
			if cvt.PtrToVal(gcpRaw.ProjectID) == "" {
				// cost do not belong to any project, handled by support month task
				continue
			}
			// if match credit return info return to owner account(negative cost) and usage account(positive cost)
			usageSummary := summaryMap[cvt.PtrToVal(gcpRaw.ProjectID)]
			if usageSummary == nil {
				logs.Errorf("gcp credit usage summary for project %s not found, rid: %s",
					cvt.PtrToVal(gcpRaw.ProjectID), kt.Rid)
				return nil, fmt.Errorf("gcp credit usage summary for project %s not found",
					cvt.PtrToVal(gcpRaw.ProjectID))
			}
			usageItems, err := g.convCreditToUsageRaw(gcpRaw, credit, usageSummary)
			if err != nil {
				logs.Errorf("fail to conv credit to usage raw bill item, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			result = append(result, usageItems...)

		}
	}
	return result, nil
}

func (g GcpCreditMonthTask) convCreditToUsageRaw(gcpRaw billcore.GcpRawBillItem, credit billcore.GcpCredit,
	usageSummary *dsbill.BillSummaryMain) ([]dsbill.BillItemCreateReq[json.RawMessage], error) {

	// 1. 使用方 +123.000
	usageCost := cvt.ValToPtr(credit.Amount.Neg())
	ext := billcore.GcpRawBillItem{
		BillingAccountID:       gcpRaw.BillingAccountID,
		Cost:                   credit.Amount,
		Currency:               gcpRaw.Currency,
		CurrencyConversionRate: gcpRaw.CurrencyConversionRate,
		Month:                  gcpRaw.Month,
		ProjectID:              gcpRaw.ProjectID,
		ProjectName:            gcpRaw.ProjectName,
		ProjectNumber:          gcpRaw.ProjectNumber,
		ServiceID:              cvt.ValToPtr(credit.ID),
		ServiceDescription:     cvt.ValToPtr(credit.Name),
		SkuDescription:         cvt.ValToPtr(credit.FullName),
		SkuID:                  cvt.ValToPtr(credit.ID),
		TotalCost:              usageCost,
		ReturnCost:             usageCost,
		UsageEndTime:           gcpRaw.UsageEndTime,
		UsagePricingUnit:       gcpRaw.UsagePricingUnit,
		UsageStartTime:         gcpRaw.UsageStartTime,
		CreditInfos:            []billcore.GcpCredit{credit},
	}

	extByte, err := json.Marshal(ext)
	if err != nil {
		return nil, fmt.Errorf("fail to marshal credit user extension, err: %v", err)
	}
	itemUsage := dsbill.BillItemCreateReq[json.RawMessage]{
		RootAccountID: usageSummary.RootAccountID,
		MainAccountID: usageSummary.MainAccountID,
		Vendor:        usageSummary.Vendor,
		ProductID:     usageSummary.ProductID,
		BkBizID:       usageSummary.BkBizID,
		BillYear:      usageSummary.BillYear,
		BillMonth:     usageSummary.BillMonth,
		BillDay:       enumor.MonthTaskSpecialBillDay,
		VersionID:     usageSummary.CurrentVersion,
		Currency:      usageSummary.Currency,
		Cost:          cvt.PtrToVal(usageCost),
		HcProductCode: constant.GcpCreditReturnCostReverse,
		HcProductName: cvt.PtrToVal(ext.ServiceID),
		Extension:     cvt.ValToPtr[json.RawMessage](extByte),
	}

	return []dsbill.BillItemCreateReq[json.RawMessage]{itemUsage}, nil
}

func (g GcpCreditMonthTask) convCreditToOwnerRaw(gcpRaw billcore.GcpRawBillItem, credit billcore.GcpCredit,
	ownerSummary *dsbill.BillSummaryMain) ([]dsbill.BillItemCreateReq[json.RawMessage], error) {

	// 2. credit所有者 -123.000
	ownerExt := billcore.GcpRawBillItem{
		BillingAccountID:       gcpRaw.BillingAccountID,
		Cost:                   credit.Amount,
		Currency:               gcpRaw.Currency,
		CurrencyConversionRate: gcpRaw.CurrencyConversionRate,
		Month:                  gcpRaw.Month,
		ProjectID:              cvt.ValToPtr(ownerSummary.MainAccountCloudID),
		ServiceID:              cvt.ValToPtr(credit.ID),
		ServiceDescription:     cvt.ValToPtr(credit.Name),
		SkuDescription:         cvt.ValToPtr(credit.FullName),
		SkuID:                  cvt.ValToPtr(credit.ID),
		TotalCost:              credit.Amount,
		ReturnCost:             credit.Amount,
		UsageEndTime:           gcpRaw.UsageEndTime,
		UsagePricingUnit:       gcpRaw.UsagePricingUnit,
		UsageStartTime:         gcpRaw.UsageStartTime,
		CreditInfos:            []billcore.GcpCredit{credit},
	}
	extByte, err := json.Marshal(ownerExt)
	if err != nil {
		return nil, fmt.Errorf("fail to marshal credit owner extension, err: %v", err)
	}
	itemOwner := dsbill.BillItemCreateReq[json.RawMessage]{
		RootAccountID: ownerSummary.RootAccountID,
		MainAccountID: ownerSummary.MainAccountID,
		Vendor:        ownerSummary.Vendor,
		ProductID:     ownerSummary.ProductID,
		BkBizID:       ownerSummary.BkBizID,
		BillYear:      ownerSummary.BillYear,
		BillMonth:     ownerSummary.BillMonth,
		BillDay:       enumor.MonthTaskSpecialBillDay,
		VersionID:     ownerSummary.CurrentVersion,
		Currency:      ownerSummary.Currency,
		Cost:          cvt.PtrToVal(credit.Amount),
		HcProductCode: constant.GcpCreditReturnCost,
		HcProductName: cvt.PtrToVal(ownerExt.ServiceID),
		Extension:     cvt.ValToPtr[json.RawMessage](extByte),
	}

	return []dsbill.BillItemCreateReq[json.RawMessage]{itemOwner}, nil
}

// GetHcProductCodes type to product codes
func (g GcpCreditMonthTask) GetHcProductCodes() []string {
	return []string{constant.GcpCreditReturnCost, constant.GcpCreditReturnCostReverse}
}
