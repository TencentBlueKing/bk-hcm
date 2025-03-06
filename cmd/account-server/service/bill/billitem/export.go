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
	"time"

	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	accountset "hcm/pkg/api/core/account-set"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/cmdb"

	"github.com/shopspring/decimal"
)

const (
	defaultExportFilename = "bill_item-%s-%s.csv"
)

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

	rate, err := b.getExchangeRate(cts.Kit, req.BillYear, req.BillMonth)
	if err != nil {
		logs.Errorf("fail get exchange rate for exporting bill items, err: %v, year: %d, month: %d, rid: %s",
			err, req.BillYear, req.BillMonth, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.HuaWei:
		return b.exportHuaweiBillItems(cts.Kit, req, rate)
	case enumor.Gcp:
		return b.exportGcpBillItems(cts.Kit, req, rate)
	case enumor.Aws:
		return b.exportAwsBillItems(cts.Kit, req, rate)
	default:
		return nil, fmt.Errorf("unsupport %s vendor", vendor)
	}
}

func (b *billItemSvc) getExchangeRate(kt *kit.Kit, year, month int) (*decimal.Decimal, error) {
	// 获取汇率
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("from_currency", enumor.CurrencyUSD),
			tools.RuleEqual("to_currency", enumor.CurrencyRMB),
			tools.RuleEqual("year", year),
			tools.RuleEqual("month", month),
		),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}
	result, err := b.client.DataService().Global.Bill.ListExchangeRate(kt, listReq)
	if err != nil {
		return nil, fmt.Errorf("get exchange rate from %s to %s in %d-%d failed, err %s",
			enumor.CurrencyUSD, enumor.CurrencyRMB, year, month, err.Error())
	}
	if len(result.Details) == 0 {
		return nil, fmt.Errorf("get no exchange rate from %s to %s in %d-%d, rid %s",
			enumor.CurrencyUSD, enumor.CurrencyRMB, year, month, kt.Rid)
	}
	if result.Details[0].ExchangeRate == nil {
		return nil, fmt.Errorf("get exchange rate is nil, from %s to %s in %d-%d, rid %s",
			enumor.CurrencyUSD, enumor.CurrencyRMB, year, month, kt.Rid)
	}
	return result.Details[0].ExchangeRate, nil
}

func (b *billItemSvc) listRootAccount(kt *kit.Kit, vendor enumor.Vendor) (
	map[string]*accountset.BaseRootAccount, error) {

	filter := tools.ExpressionAnd(tools.RuleEqual("vendor", vendor))
	countReq := &core.ListReq{
		Filter: filter,
		Page:   core.NewCountPage(),
	}
	countResp, err := b.client.DataService().Global.MainAccount.List(kt, countReq)
	if err != nil {
		logs.Errorf("list main account failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	total := countResp.Count

	result := make(map[string]*accountset.BaseRootAccount)
	for i := uint64(0); i < total; i += uint64(core.DefaultMaxPageLimit) {
		listReq := &core.ListReq{
			Filter: filter,
			Page: &core.BasePage{
				Start: uint32(i),
				Limit: core.DefaultMaxPageLimit,
			},
		}
		listResult, err := b.client.DataService().Global.RootAccount.List(kt, listReq)
		if err != nil {
			logs.Errorf("list main account failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		for _, detail := range listResult.Details {
			result[detail.ID] = detail
		}
	}
	return result, nil
}

func (b *billItemSvc) listBiz(kt *kit.Kit) (map[int64]string, error) {
	params := &cmdb.SearchBizParams{
		Fields: []string{"bk_biz_id", "bk_biz_name"},
	}
	resp, err := b.cmdbCli.SearchBusiness(kt, params)
	if err != nil {
		logs.Errorf("call cmdb search business api failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
	}

	infos := resp.Info
	data := make(map[int64]string, len(infos))
	for _, biz := range infos {
		data[biz.BizID] = biz.BizName
	}

	return data, nil
}

// fetchAccountBizInfo 根据vendor获取所有关联的数据
func (b *billItemSvc) fetchAccountBizInfo(kt *kit.Kit, vendor enumor.Vendor) (
	rootAccountMap map[string]*accountset.BaseRootAccount, mainAccountMap map[string]*accountset.BaseMainAccount,
	bizNameMap map[int64]string, err error) {

	bizNameMap, err = b.listBiz(kt)
	if err != nil {
		logs.Errorf("fail to list biz, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}
	mainAccounts, err := b.listMainAccount(kt, vendor)
	if err != nil {
		logs.Errorf("fail to list main account, vendor: %s, err: %v, rid: %s", vendor, err, kt.Rid)
		return nil, nil, nil, err
	}
	mainAccountMap = make(map[string]*accountset.BaseMainAccount, len(mainAccounts))
	for _, account := range mainAccounts {
		mainAccountMap[account.ID] = account
	}

	rootAccountMap, err = b.listRootAccount(kt, vendor)
	if err != nil {
		logs.Errorf("fail to list root account, vendor: %s, err: %v, rid: %s", vendor, err, kt.Rid)
		return nil, nil, nil, err
	}
	return rootAccountMap, mainAccountMap, bizNameMap, nil
}

func generateFilename(vendor enumor.Vendor) string {
	return fmt.Sprintf(defaultExportFilename, vendor, time.Now().Format("2006-01-02-15_04_05"))
}
