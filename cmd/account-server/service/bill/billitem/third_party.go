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
	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	databill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
)

// PullBillItemForThirdParty 查询账单明细外部接口，根据vendor鉴权
func (b *billItemSvc) PullBillItemForThirdParty(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}
	if !isSupportedVendor(vendor) {
		return nil, errf.Newf(errf.InvalidParameter, "vendor %s is not supported", vendor)
	}

	req := new(bill.ListBillItemByVendorReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authMeta := meta.ResourceAttribute{Basic: &meta.Basic{
		Type:   meta.AccountBillThirdParty,
		Action: meta.Find,
		// 把vendor 当成一种资源 TODO: 改为属性鉴权
		ResourceID: string(vendor),
	}}
	err := b.authorizer.AuthorizeWithPerm(cts.Kit, authMeta)
	if err != nil {
		return nil, err
	}

	rules, err := b.buildBillItemFilter(cts.Kit, vendor, req)
	if err != nil {
		logs.Errorf("fail to build bill item filter for list bill item by %s, err: %v, rid: %s",
			vendor, err, cts.Kit.Rid)
		return nil, err
	}

	billListReq := &databill.BillItemListReq{
		ItemCommonOpt: &databill.ItemCommonOpt{
			Vendor: vendor,
			Year:   int(req.BillYear),
			Month:  int(req.BillMonth),
		},
		ListReq: &core.ListReq{Filter: tools.ExpressionAnd(rules...), Page: req.Page},
	}

	return b.client.DataService().Global.Bill.ListBillItemRaw(cts.Kit, billListReq)

}

func isSupportedVendor(vendor enumor.Vendor) bool {
	switch vendor {
	case enumor.Aws, enumor.HuaWei, enumor.Gcp, enumor.Zenlayer:
		return true

	default:
		return false
	}
}

func (b *billItemSvc) buildBillItemFilter(kt *kit.Kit, vendor enumor.Vendor, req *bill.ListBillItemByVendorReq) (
	[]*filter.AtomRule, error) {

	// build filter
	var rules []*filter.AtomRule
	if req.BeginBillDay != nil {
		rules = append(rules, tools.RuleGreaterThanEqual("bill_day", cvt.PtrToVal(req.BeginBillDay)))
	}
	if req.EndBillDay != nil {
		rules = append(rules, tools.RuleLessThanEqual("bill_day", cvt.PtrToVal(req.EndBillDay)))
	}

	var mainAccountIds, rootAccountIds []string
	if len(req.MainAccountCloudIds) > 0 {
		// 查询关联二级账号id，bill item 表中没有账号云id
		mainAccountReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("cloud_id", req.MainAccountCloudIds),
				tools.RuleEqual("vendor", vendor)),
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"id", "cloud_id"},
		}
		mainAccountList, err := b.client.DataService().Global.MainAccount.List(kt, mainAccountReq)
		if err != nil {
			logs.Errorf("fail to find main account for list bill item, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if len(req.MainAccountCloudIds) != len(mainAccountList.Details) {
			return nil, errf.Newf(errf.InvalidParameter,
				"some account can not be found by main account cloud id")
		}
		for i := range mainAccountList.Details {
			mainAccountIds = append(mainAccountIds, mainAccountList.Details[i].ID)
		}
	}

	if len(req.RootAccountCloudIds) > 0 {
		// 查询关联一级账号id，bill item 表中没有账号云id
		rootAccountReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("cloud_id", req.RootAccountCloudIds),
				tools.RuleEqual("vendor", vendor)),
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"id", "cloud_id"},
		}
		rootAccountList, err := b.client.DataService().Global.RootAccount.List(kt, rootAccountReq)
		if err != nil {
			logs.Errorf("fail to find root account by cloud id for list bill item, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if len(req.RootAccountCloudIds) != len(rootAccountList.Details) {
			return nil, errf.Newf(errf.InvalidParameter, "some account can not be found by root account cloud id")
		}
		for i := range rootAccountList.Details {
			rootAccountIds = append(rootAccountIds, rootAccountList.Details[i].ID)
		}
	}
	if len(req.RootAccountIds) > 0 {
		rootAccountIds = append(rootAccountIds, req.RootAccountIds...)
	}
	if len(req.MainAccountIds) > 0 {
		mainAccountIds = append(mainAccountIds, req.MainAccountIds...)
	}

	if len(mainAccountIds) > 0 {
		rules = append(rules, tools.RuleIn("main_account_id", mainAccountIds))
	}
	if len(rootAccountIds) > 0 {
		rules = append(rules, tools.RuleIn("root_account_id", rootAccountIds))
	}

	return rules, nil
}
