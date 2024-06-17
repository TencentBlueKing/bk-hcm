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
	rawjson "encoding/json"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/bill"
	dataproto "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

// ListBillItemRaw list bill item raw data without parsing extension
func (svc *service) ListBillItemRaw(cts *rest.Contexts) (any, error) {
	req := new(dataproto.BillItemListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	data, err := svc.dao.AccountBillItem().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	details := make([]*bill.BillItemRaw, 0, len(data.Details))
	for _, d := range data.Details {
		details = append(details, &bill.BillItemRaw{
			BaseBillItem: convBillItem(&d),
			Extension:    rawjson.RawMessage(d.Extension),
		})
	}

	return &core.ListResultT[*bill.BillItemRaw]{Details: details, Count: cvt.PtrToVal(data.Count)}, nil
}

// ListBillItem list bill item with options
func (svc *service) ListBillItem(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.BillItemListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	data, err := svc.dao.AccountBillItem().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	details := make([]*bill.BaseBillItem, len(data.Details))
	for idx, d := range data.Details {
		details[idx] = convBillItem(&d)
	}

	return &dataproto.BillItemBaseListResult{Details: details, Count: cvt.PtrToVal(data.Count)}, nil
}

// ListBillItemExt ...
func (svc *service) ListBillItemExt(cts *rest.Contexts) (any, error) {

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.Aws:
		return listBillItemExt[bill.AwsBillItemExtension](cts, svc, vendor)
	case enumor.HuaWei:
		return listBillItemExt[bill.HuaweiBillItemExtension](cts, svc, vendor)
	case enumor.Azure:
		return listBillItemExt[bill.AzureBillItemExtension](cts, svc, vendor)
	case enumor.Gcp:
		return listBillItemExt[bill.GcpBillItemExtension](cts, svc, vendor)
	case enumor.Kaopu:
		return listBillItemExt[bill.KaopuBillItemExtension](cts, svc, vendor)
	case enumor.Zenlayer:
		return listBillItemExt[bill.ZenlayerBillItemExtension](cts, svc, vendor)
	default:
		return nil, fmt.Errorf("unsupport %s vendor", vendor)
	}
}

func listBillItemExt[E bill.BillItemExtension](cts *rest.Contexts, svc *service, vendor enumor.Vendor) (any, error) {
	req := new(dataproto.BillItemListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	data, err := svc.dao.AccountBillItem().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	details := make([]*bill.BillItem[E], len(data.Details))
	for idx, d := range data.Details {
		details[idx], err = convBillItemExt[E](&d)
		if err != nil {
			logs.Errorf("fail to convert bill item extension for %s vendor, err: %v, rid: %s", vendor, err, cts.Kit.Rid)
			return nil, err
		}
	}

	return &core.ListResultT[*bill.BillItem[E]]{Details: details, Count: cvt.PtrToVal(data.Count)}, nil
}

func convBillItemExt[E bill.BillItemExtension](m *tablebill.AccountBillItem) (*bill.BillItem[E], error) {

	extension := new(E)
	if len(m.Extension) != 0 {
		if err := json.UnmarshalFromString(string(m.Extension), &extension); err != nil {
			return nil, fmt.Errorf("UnmarshalFromString bill item extension failed, err: %v", err)
		}
	}
	ext := &bill.BillItem[E]{
		BaseBillItem: convBillItem(m),
		Extension:    extension,
	}
	return ext, nil
}

func convBillItem(m *tablebill.AccountBillItem) *bill.BaseBillItem {
	return &bill.BaseBillItem{
		ID:            m.ID,
		RootAccountID: m.RootAccountID,
		MainAccountID: m.MainAccountID,
		Vendor:        m.Vendor,
		ProductID:     m.ProductID,
		BkBizID:       m.BkBizID,
		BillYear:      m.BillYear,
		BillMonth:     m.BillMonth,
		BillDay:       m.BillDay,
		VersionID:     m.VersionID,
		Currency:      m.Currency,
		Cost:          m.Cost.Decimal,
		HcProductCode: m.HcProductCode,
		HcProductName: m.HcProductName,
		ResAmount:     m.ResAmount.Decimal,
		ResAmountUnit: m.ResAmountUnit,
		Revision: &core.Revision{
			CreatedAt: m.CreatedAt.String(),
			UpdatedAt: m.UpdatedAt.String(),
		},
	}
}
