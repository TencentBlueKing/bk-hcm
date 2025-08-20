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
	"hcm/pkg/tools/converter"

	"github.com/TencentBlueKing/gopkg/conv"
	"github.com/shopspring/decimal"
)

func (b *billItemSvc) exportGcpBillItems(kt *kit.Kit, req *bill.ExportBillItemReq,
	rate *decimal.Decimal) (any, error) {

	rootAccountMap, mainAccountMap, bizNameMap, err := b.fetchAccountBizInfo(kt, enumor.Gcp)
	if err != nil {
		logs.Errorf("[exportGcpBillItems] prepare related data failed: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	regionMap, err := b.listGcpRegions(kt)
	if err != nil {
		logs.Errorf("list gcp regions failed: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	filename, filepath, writer, closeFunc, err := export.CreateWriterByFileName(kt, generateFilename(enumor.Gcp))
	defer func() {
		if closeFunc != nil {
			closeFunc()
		}
	}()
	if err != nil {
		logs.Errorf("create writer failed: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for _, header := range export.GcpBillItemHeaders {
		if err := writer.Write(header); err != nil {
			logs.Errorf("csv write header failed: %v, val: %v, rid: %s", err, header, kt.Rid)
			return nil, err
		}
	}

	convFunc := func(items []*billapi.GcpBillItem) error {
		if len(items) == 0 {
			return nil
		}
		table, err := convertGcpBillItem(kt, items, bizNameMap, mainAccountMap, rootAccountMap, regionMap, rate)
		if err != nil {
			logs.Errorf("[exportGcpBillItems] convert to raw data error: %v, rid: %s", err, kt.Rid)
			return err
		}
		err = writer.WriteAll(table)
		if err != nil {
			logs.Errorf("csv write data failed: %v, rid: %s", err, kt.Rid)
			return err
		}
		return nil
	}
	err = b.fetchGcpBillItems(kt, req, convFunc)
	if err != nil {
		return nil, err
	}

	return &bill.FileDownloadResp{
		ContentTypeStr:        "application/octet-stream",
		ContentDispositionStr: fmt.Sprintf(`attachment; filename="%s"`, filename),
		FilePath:              filepath,
	}, nil
}

func convertGcpBillItem(kt *kit.Kit, items []*billapi.GcpBillItem, bizNameMap map[int64]string,
	mainAccountMap map[string]*protocore.BaseMainAccount, rootAccountMap map[string]*protocore.BaseRootAccount,
	regionMap map[string]string, rate *decimal.Decimal) ([][]string, error) {

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
		extension := item.Extension.GcpRawBillItem
		if extension == nil {
			extension = &billapi.GcpRawBillItem{}
		}

		table := export.GcpBillItemTable{
			Site:                       string(mainAccount.Site),
			AccountDate:                converter.PtrToVal[string](extension.Month),
			BizID:                      conv.ToString(item.BkBizID),
			BizName:                    bizName,
			RootAccountName:            rootAccount.Name,
			MainAccountName:            mainAccount.Name,
			Region:                     converter.PtrToVal[string](extension.Region),
			RegionName:                 regionMap[converter.PtrToVal[string](extension.Region)],
			ProjectID:                  converter.PtrToVal[string](extension.ProjectID),
			ProjectName:                converter.PtrToVal[string](extension.ProjectName),
			ServiceCategory:            converter.PtrToVal[string](extension.ServiceDescription), // 服务分类
			ServiceCategoryDescription: converter.PtrToVal[string](extension.ServiceDescription), // 服务分类名称
			SkuDescription:             converter.PtrToVal[string](extension.SkuDescription),
			Currency:                   string(item.Currency),
			UsageUnit:                  converter.PtrToVal[string](extension.UsageUnit),
			UsageAmount:                (converter.PtrToVal[decimal.Decimal](extension.UsageAmount)).String(),
			Cost:                       item.Cost.String(),
			ExchangeRate:               rate.String(),
			RMBCost:                    item.Cost.Mul(*rate).String(),
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

func (b *billItemSvc) fetchGcpBillItems(kt *kit.Kit, req *bill.ExportBillItemReq,
	convertFunc func([]*billapi.GcpBillItem) error) error {

	totalCount, err := b.fetchGcpBillItemCount(kt, req)
	if err != nil {
		logs.Errorf("fetch gcp bill item count failed: %v, rid: %s", err, kt.Rid)
		return err
	}
	exportLimit := min(totalCount, req.ExportLimit)

	commonOpt := &databill.ItemCommonOpt{
		Vendor: enumor.Gcp,
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
				logs.Errorf("[fetchGcpBillItems] build filter failed, lastID: %s, filter: %v, error: %v, rid: %s",
					lastID, expr, err, kt.Rid)
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
		result, err := b.client.DataService().Gcp.Bill.ListBillItem(kt, billListReq)
		if err != nil {
			return err
		}
		if len(result.Details) == 0 {
			continue
		}
		if err = convertFunc(result.Details); err != nil {
			logs.Errorf("convert gcp bill item failed: %v, rid: %s", err, kt.Rid)
			return err
		}
		lastID = result.Details[len(result.Details)-1].ID
	}
	return nil
}

func (b *billItemSvc) fetchGcpBillItemCount(kt *kit.Kit, req *bill.ExportBillItemReq) (uint64, error) {
	countReq := &databill.BillItemListReq{
		ItemCommonOpt: &databill.ItemCommonOpt{
			Vendor: enumor.Gcp,
			Year:   req.BillYear,
			Month:  req.BillMonth,
		},
		ListReq: &core.ListReq{Filter: req.Filter, Page: core.NewCountPage()},
	}
	details, err := b.client.DataService().Gcp.Bill.ListBillItem(kt, countReq)
	if err != nil {
		return 0, err
	}
	return details.Count, nil
}

func (b *billItemSvc) listGcpRegions(kt *kit.Kit) (map[string]string, error) {

	offset := uint32(0)
	regionMap := make(map[string]string)
	for {
		listReq := &core.ListReq{
			Filter: tools.AllExpression(),
			Page: &core.BasePage{
				Start: offset,
				Limit: core.DefaultMaxPageLimit,
			},
		}
		regions, err := b.client.DataService().Gcp.Region.ListRegion(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list gcp region failed: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if len(regions.Details) == 0 {
			break
		}
		for _, region := range regions.Details {
			regionMap[region.RegionID] = region.RegionName
		}
		offset = offset + uint32(core.DefaultMaxPageLimit)
	}
	return regionMap, nil
}
