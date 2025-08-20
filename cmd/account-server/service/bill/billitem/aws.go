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

package billitem

import (
	"fmt"

	"hcm/cmd/account-server/logics/bill/export"
	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/account-set"
	billapi "hcm/pkg/api/core/bill"
	databill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/TencentBlueKing/gopkg/conv"
	"github.com/shopspring/decimal"
)

func (b *billItemSvc) exportAwsBillItems(kt *kit.Kit, req *bill.ExportBillItemReq,
	rate *decimal.Decimal) (any, error) {

	rootAccountMap, mainAccountMap, bizNameMap, err := b.fetchAccountBizInfo(kt, enumor.Aws)
	if err != nil {
		logs.Errorf("[exportAwsBillItems] fetch account and biz info failed: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	filename, filepath, writer, closeFunc, err := export.CreateWriterByFileName(kt, generateFilename(enumor.Aws))
	defer func() {
		if closeFunc != nil {
			closeFunc()
		}
	}()
	if err != nil {
		logs.Errorf("create writer failed: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for _, header := range export.AwsBillItemHeaders {
		if err = writer.Write(header); err != nil {
			logs.Errorf("csv write header failed: %v, val: %v, rid: %s", err, header, kt.Rid)
			return nil, err
		}
	}

	convFunc := func(items []*billapi.AwsBillItem) error {
		if len(items) == 0 {
			return nil
		}
		table, err := convertAwsBillItems(kt, items, bizNameMap, mainAccountMap, rootAccountMap, rate)
		if err != nil {
			logs.Errorf("[exportAwsBillItems] convert to raw data error: %v, rid: %s", err, kt.Rid)
			return err
		}
		err = writer.WriteAll(table)
		if err != nil {
			logs.Errorf("csv write data failed: %v, rid: %s", err, kt.Rid)
			return err
		}

		return nil
	}
	err = b.fetchAwsBillItems(kt, req, convFunc)
	if err != nil {
		logs.Errorf("fetch aws bill items for export failed, req: %v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	return &bill.FileDownloadResp{
		ContentTypeStr:        "application/octet-stream",
		ContentDispositionStr: fmt.Sprintf(`attachment; filename="%s"`, filename),
		FilePath:              filepath,
	}, nil
}

func convertAwsBillItems(kt *kit.Kit, items []*billapi.AwsBillItem, bizNameMap map[int64]string,
	mainAccountMap map[string]*protocore.BaseMainAccount, rootAccountMap map[string]*protocore.BaseRootAccount,
	rate *decimal.Decimal) ([][]string, error) {

	result := make([][]string, 0, len(items))
	for _, item := range items {
		mainAccount, ok := mainAccountMap[item.MainAccountID]
		if !ok {
			return nil, fmt.Errorf("main account(%s) not found", item.MainAccountID)
		}
		rootAccount, ok := rootAccountMap[item.RootAccountID]
		if !ok {
			return nil, fmt.Errorf("root account(%s) not found", item.RootAccountID)
		}
		bizName, ok := bizNameMap[item.BkBizID]
		if !ok {
			logs.Warnf("biz(%d) not found", item.BkBizID)
		}

		extension := item.Extension.AwsRawBillItem
		if extension == nil {
			extension = &billapi.AwsRawBillItem{}
		}

		table := &export.AwsBillItemTable{
			Site:                string(mainAccount.Site),
			AccountDate:         fmt.Sprintf("%d-%02d", item.BillYear, item.BillMonth),
			BizID:               conv.ToString(item.BkBizID),
			BizName:             bizName,
			RootAccountName:     rootAccount.Name,
			MainAccountName:     mainAccount.Name,
			Region:              extension.ProductToRegionCode,
			LocationName:        extension.ProductFromLocation,
			BillInvoiceIC:       extension.BillInvoiceId,
			BillEntity:          extension.BillBillingEntity,
			ProductCode:         extension.LineItemProductCode,
			ProductFamily:       extension.ProductProductFamily,
			ProductName:         extension.ProductProductName,
			ApiOperation:        extension.LineItemOperation, // line_item_operation
			ProductUsageType:    extension.ProductUsagetype,
			InstanceType:        extension.ProductInsightstype,
			ResourceId:          extension.LineItemResourceId,
			PricingTerm:         extension.PricingTerm,
			LineItemType:        extension.LineItemLineItemType,
			LineItemDescription: extension.LineItemLineItemDescription,
			UsageAmount:         extension.LineItemUsageAmount,
			PricingUnit:         extension.PricingUnit,
			Cost:                item.Cost.String(),
			Currency:            string(item.Currency),
			RMBCost:             item.Cost.Mul(*rate).String(),
			Rate:                rate.String(),
		}
		values, err := table.GetValuesByHeader()
		if err != nil {
			logs.Errorf("get header fields failed, table: %v, error: %v, rid: %s", table, err, kt.Rid)
			return nil, err
		}
		result = append(result, values)
	}
	return result, nil
}

func (b *billItemSvc) fetchAwsBillItems(kt *kit.Kit, req *bill.ExportBillItemReq,
	convertFunc func([]*billapi.AwsBillItem) error) error {

	totalCount, err := b.fetchAwsBillItemCount(kt, req)
	if err != nil {
		logs.Errorf("fetch aws bill item count failed: %v, rid: %s", err, kt.Rid)
		return err
	}
	exportLimit := min(totalCount, req.ExportLimit)

	commonOpt := &databill.ItemCommonOpt{
		Vendor: enumor.Aws,
		Year:   req.BillYear,
		Month:  req.BillMonth,
	}
	lastID := ""
	for offset := uint64(0); offset < exportLimit; offset = offset + uint64(core.DefaultMaxPageLimit) {
		left := exportLimit - offset
		expr := req.Filter
		if len(lastID) > 0 {
			expr, err = tools.And(
				expr,
				tools.RuleIDGreaterThan(lastID),
			)
			if err != nil {
				logs.Errorf("build filter failed: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
		billListReq := &databill.BillItemListReq{
			ItemCommonOpt: commonOpt,
			ListReq: &core.ListReq{
				Filter: expr,
				Page: &core.BasePage{
					Start: 0,
					Limit: min(uint(left), core.DefaultMaxPageLimit),
					Sort:  "id",
					Order: core.Ascending,
				},
			},
		}
		result, err := b.client.DataService().Aws.Bill.ListBillItem(kt, billListReq)
		if err != nil {
			logs.Errorf("list aws bill item failed: %v, rid: %s", err, kt.Rid)
			return err
		}
		if len(result.Details) == 0 {
			continue
		}
		if err = convertFunc(result.Details); err != nil {
			logs.Errorf("convert aws bill item failed: %v, rid: %s", err, kt.Rid)
			return err
		}
		lastID = result.Details[len(result.Details)-1].ID
	}
	return nil
}

func (b *billItemSvc) fetchAwsBillItemCount(kt *kit.Kit, req *bill.ExportBillItemReq) (uint64, error) {
	countReq := &databill.BillItemListReq{
		ItemCommonOpt: &databill.ItemCommonOpt{
			Vendor: enumor.Aws,
			Year:   req.BillYear,
			Month:  req.BillMonth,
		},
		ListReq: &core.ListReq{Filter: req.Filter, Page: core.NewCountPage()},
	}
	details, err := b.client.DataService().Aws.Bill.ListBillItem(kt, countReq)
	if err != nil {
		return 0, err
	}
	return details.Count, nil
}
