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

package billitem

import (
	"fmt"

	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ListBillItems 查询账单明细
func (b *billItemSvc) ListBillItems(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	req := new(bill.ListBillItemReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := b.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Find}})
	if err != nil {
		return nil, err
	}
	expressions := []filter.RuleFactory{
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("bill_year", req.BillYear),
		tools.RuleEqual("bill_month", req.BillMonth),
	}
	if req.Filter != nil {
		expressions = append(expressions, req.Filter)
	}
	mergedFilter, err := tools.And(expressions...)
	if err != nil {
		logs.Errorf("fail merge filter for listing bill items, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}
	billListReq := &core.ListReq{Filter: mergedFilter, Page: req.Page}

	return b.client.DataService().Global.Bill.ListBillItemRaw(cts.Kit, billListReq)

}

// ExportBillItems 导出账单明细
func (b *billItemSvc) ExportBillItems(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	req := new(bill.ExportBillItemReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := b.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	mergedFilter, err := tools.And(
		tools.RuleEqual("vendor", vendor),
		tools.RuleEqual("bill_year", req.BillYear),
		tools.RuleEqual("bill_month", req.BillMonth),
		req.Filter)
	if err != nil {
		logs.Errorf("fail merge filter for exporting bill items, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.HuaWei:
		return exportHuaweiBillItems(cts.Kit, b, mergedFilter, req.ExportLimit)
	default:
		return nil, fmt.Errorf("unsupport %s vendor", vendor)
	}
}

func exportHuaweiBillItems(kt *kit.Kit, b *billItemSvc, filter *filter.Expression, requireCount uint64) (
	any, error) {

	billListReq := &core.ListReq{Filter: filter, Page: core.NewDefaultBasePage()}
	_, err := b.client.DataService().HuaWei.Bill.ListBillItem(kt, billListReq)
	if err != nil {
		logs.Errorf("fail to list bill item for export, err: %v, req: %+v, rid: %s", err, billListReq, kt.Rid)
		return nil, err
	}

	// TODO
	// export as csv

	return nil, nil
}
