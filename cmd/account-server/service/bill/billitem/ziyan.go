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
	"errors"
	"fmt"
	"strings"

	"hcm/cmd/task-server/logics/action/bill/monthtask"
	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	accountset "hcm/pkg/api/core/account-set"
	databill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/finops"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

// fetchAccountProductInfo 根据vendor获取所有关联的数据
func (b *billItemSvc) fetchAccountProductInfo(kt *kit.Kit, vendor enumor.Vendor) (
	rootAccountMap map[string]*accountset.BaseRootAccount, mainAccountMap map[string]*accountset.BaseMainAccount,
	opProductMap map[int64]finops.OperationProduct, err error) {

	opProductMap, err = b.listOpProduct(kt)
	if err != nil {
		logs.Errorf("fail to list op product, err: %v, rid: %s", err, kt.Rid)
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
	return rootAccountMap, mainAccountMap, opProductMap, nil
}

func (b *billItemSvc) listOpProduct(kt *kit.Kit) (map[int64]finops.OperationProduct, error) {

	offset := uint32(0)
	result := make(map[int64]finops.OperationProduct)
	for {
		param := &finops.ListOpProductParam{
			Page: core.BasePage{
				Start: offset,
				Limit: core.DefaultMaxPageLimit,
			},
		}
		productResult, err := b.finops.ListOpProduct(kt, param)
		if err != nil {
			return nil, err
		}
		if len(productResult.Items) == 0 {
			break
		}
		offset += uint32(core.DefaultMaxPageLimit)
		for _, product := range productResult.Items {
			result[product.OpProductId] = product
		}
	}

	return result, nil
}

// PullAIBills 拉取ai账单
func (b *billItemSvc) PullAIBills(cts *rest.Contexts) (any, error) {

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	req := new(bill.PullAIBillsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authMeta := meta.ResourceAttribute{Basic: &meta.Basic{
		// 第三方账单拉取权限, 按厂商鉴权
		Type:   meta.AccountBillThirdParty,
		Action: meta.Find,
		// 把vendor 当成一种资源
		ResourceID: string(vendor),
	}}
	err := b.authorizer.AuthorizeWithPerm(cts.Kit, authMeta)
	if err != nil {
		return nil, err
	}
	flt, err := b.buildAIFilters(cts.Kit, vendor, req)
	if err != nil {
		return nil, err
	}

	billListReq := &databill.BillItemListReq{
		ItemCommonOpt: &databill.ItemCommonOpt{
			Vendor: vendor,
			Year:   int(req.BillYear),
			Month:  int(req.BillMonth),
		},
		ListReq: &core.ListReq{Filter: flt, Page: req.Page},
	}
	itemResp, err := b.client.DataService().Global.Bill.ListBillItemRaw(cts.Kit, billListReq)
	if err != nil {
		logs.Errorf("fail to list bill item for pull ai bill, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if req.Page.Count || len(itemResp.Details) == 0 {
		return itemResp, nil
	}
	// 查询账号
	var mainAccountMap = make(map[string]*accountset.BaseMainAccount)
	var currencyRateMap = make(map[enumor.CurrencyCode]*decimal.Decimal)
	for i := range itemResp.Details {
		item := itemResp.Details[i]
		mainAccountMap[item.MainAccountID] = nil
		currencyRateMap[item.Currency] = nil
	}
	err = b.fillMainAccountInfo(cts.Kit, mainAccountMap)
	if err != nil {
		return nil, fmt.Errorf("fail to list main account for pull ai bill, err: %v, rid: %s", err, cts.Kit.Rid)
	}
	err = b.fillRate(cts.Kit, req.BillYear, req.BillMonth, currencyRateMap)
	if err != nil {
		return nil, fmt.Errorf("fail to list rate for pull ai bill, err: %v, rid: %s", err, cts.Kit.Rid)
	}

	var items = make([]bill.AIBillItem, len(itemResp.Details))
	// 转换bill item
	for i := range itemResp.Details {
		item := itemResp.Details[i]
		var skuDesc string
		switch vendor {
		case enumor.Aws:
			skuDesc = item.HcProductName
		case enumor.Gcp:
			skuDesc = gjson.Get(string(item.Extension), "sku_description").String()
		default:
			skuDesc = "unknown"
		}
		llmType := getLlmType(skuDesc)
		mainAccount := cvt.PtrToVal(mainAccountMap[item.MainAccountID])
		items[i] = bill.AIBillItem{
			ID:                 item.ID,
			Year:               item.BillYear,
			Month:              item.BillMonth,
			Day:                item.BillDay,
			Vendor:             item.Vendor,
			ProductID:          item.ProductID,
			MainAccountEmail:   mainAccount.Email,
			MainAccountName:    mainAccount.Name,
			MainAccountCloudID: mainAccount.CloudID,
			LLMTYpe:            llmType,
			Cost:               item.Cost.String(),
			Currency:           item.Currency,
			RawBill:            item.Extension,
			UpdatedAt:          item.UpdatedAt,
		}
		if rate := currencyRateMap[item.Currency]; rate != nil {
			items[i].Rate = rate.String()
			items[i].CostRMB = rate.Mul(item.Cost).String()
		}
	}
	return core.ListResultT[bill.AIBillItem]{Count: itemResp.Count, Details: items}, err
}

func getLlmType(skuDesc string) string {
	llmType := "unknown"
	if strings.Contains(strings.ToLower(skuDesc), "gemini") {
		llmType = "Gemini"
	} else if strings.Contains(strings.ToLower(skuDesc), "claude") {
		llmType = "Claude"
	}
	return llmType
}

func (b *billItemSvc) fillMainAccountInfo(kt *kit.Kit, mainAccountMap map[string]*accountset.BaseMainAccount) error {

	if mainAccountMap == nil {
		return errors.New("main account map is nil")
	}
	if len(mainAccountMap) == 0 {
		return nil
	}
	mainIDs := cvt.MapKeyToStringSlice(mainAccountMap)
	// 查询关联二级账号id，bill item 表中没有账号云id
	for _, idBatch := range slice.Split(mainIDs, int(core.DefaultMaxPageLimit)) {
		mainAccountReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("id", idBatch)),
			Page: core.NewDefaultBasePage(),
		}
		mainAccountList, err := b.client.DataService().Global.MainAccount.List(kt, mainAccountReq)
		if err != nil {
			logs.Errorf("fail to find main account for list bill item, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		for i := range mainAccountList.Details {
			mainAccountMap[mainAccountList.Details[i].ID] = mainAccountList.Details[i]
		}
	}
	return nil
}

func (b *billItemSvc) fillRate(kt *kit.Kit, year, month uint, rateMap map[enumor.CurrencyCode]*decimal.Decimal) error {
	if rateMap == nil {
		return errors.New("rate map is nil")
	}
	one := decimal.NewFromInt(1)
	var cnyRate = cvt.ValToPtr(one)
	if rate, ok := rateMap[enumor.CurrencyCNY]; ok && rate != nil {
		cnyRate = rate
	} else {
		if ok {
			delete(rateMap, enumor.CurrencyCNY)
		}
	}
	defer func() {
		rateMap[enumor.CurrencyCNY] = cnyRate
	}()
	if len(rateMap) == 0 {
		return nil
	}
	currencyCodes := cvt.MapKeyToSlice(rateMap)
	for _, currencyBatch := range slice.Split(currencyCodes, int(core.DefaultMaxPageLimit)) {
		// 查询汇率
		rateReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("from_currency", currencyBatch),
				tools.RuleEqual("to_currency", enumor.CurrencyCNY),
				tools.RuleEqual("year", year),
				tools.RuleEqual("month", month),
			),
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"id", "from_currency", "exchange_rate"},
		}
		rateList, err := b.client.DataService().Global.Bill.ListExchangeRate(kt, rateReq)
		if err != nil {
			logs.Errorf("fail to find exchange rate for list bill item, err: %v, source:%v, rid: %s",
				currencyBatch, err, kt.Rid)
			return err
		}
		for i := range rateList.Details {
			rateMap[rateList.Details[i].FromCurrency] = rateList.Details[i].ExchangeRate
		}
	}

	return nil
}

func (b *billItemSvc) buildAIFilters(kt *kit.Kit, vendor enumor.Vendor, req *bill.PullAIBillsReq) (
	*filter.Expression, error) {

	// AI Service Filter
	var rules []filter.RuleFactory
	switch vendor {
	case enumor.Aws, enumor.Gcp:
		aiRules, err := monthtask.GenAIFilterRules(kt, vendor)
		if err != nil {
			logs.Errorf("fail to gen ai filter rules, vendor: %s, err: %v, rid: %s", vendor, err, kt.Rid)
			return nil, err
		}
		rules = append(rules, aiRules...)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "invalid vendor: %s", vendor)
	}

	// day filter
	if req.BeginBillDay != nil {
		rules = append(rules, tools.RuleGreaterThanEqual("bill_day", cvt.PtrToVal(req.BeginBillDay)))
	}
	if req.EndBillDay != nil {
		rules = append(rules, tools.RuleLessThanEqual("bill_day", cvt.PtrToVal(req.EndBillDay)))
	}

	if len(req.MainAccountCloudIds) > 0 {
		mainAccountIds, err := b.getMainAccountIDs(kt, vendor, req.MainAccountCloudIds)
		if err != nil {
			return nil, err
		}
		rules = append(rules, tools.RuleIn("main_account_id", mainAccountIds))
	}

	if len(req.RootAccountCloudIds) > 0 {
		rootAccountIds, err := b.getRootAccountIDs(kt, vendor, req.RootAccountCloudIds)
		if err != nil {
			return nil, err
		}
		rules = append(rules, tools.RuleIn("root_account_id", rootAccountIds))
	}
	flt := &filter.Expression{
		Op:    filter.And,
		Rules: rules,
	}
	return flt, nil
}

func (b *billItemSvc) getRootAccountIDs(kt *kit.Kit, vendor enumor.Vendor, cloudIDs []string) ([]string, error) {
	var rootAccountIds []string
	// 查询关联一级账号id，bill item 表中没有账号云id
	uniqueCloudIDs := slice.Unique(cloudIDs)
	if len(uniqueCloudIDs) != len(cloudIDs) {
		return nil, errf.Newf(errf.InvalidParameter, "some root account cloud id is duplicated")
	}
	for _, cloudIDBatch := range slice.Split(cloudIDs, int(core.DefaultMaxPageLimit)) {
		rootAccountReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("cloud_id", cloudIDBatch),
				tools.RuleEqual("vendor", vendor)),
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"id", "cloud_id"},
		}
		rootAccountList, err := b.client.DataService().Global.RootAccount.List(kt, rootAccountReq)
		if err != nil {
			logs.Errorf("fail to find root account by cloud id for list bill item, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if len(cloudIDBatch) != len(rootAccountList.Details) {
			return nil, errf.Newf(errf.InvalidParameter, "some account can not be found by root account cloud id")
		}
		for i := range rootAccountList.Details {
			rootAccountIds = append(rootAccountIds, rootAccountList.Details[i].ID)
		}
	}
	return rootAccountIds, nil
}

func (b *billItemSvc) getMainAccountIDs(kt *kit.Kit, vendor enumor.Vendor, cloudIDs []string) ([]string, error) {

	var mainAccountIds []string
	// 查询关联二级账号id，bill item 表中没有账号云id
	uniqueCloudIDs := slice.Unique(cloudIDs)
	if len(uniqueCloudIDs) != len(cloudIDs) {
		return nil, errf.Newf(errf.InvalidParameter, "some main account cloud id is duplicated")
	}
	for _, cloudIDBatch := range slice.Split(cloudIDs, int(core.DefaultMaxPageLimit)) {
		mainAccountReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("cloud_id", cloudIDBatch),
				tools.RuleEqual("vendor", vendor)),
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"id", "cloud_id"},
		}
		mainAccountList, err := b.client.DataService().Global.MainAccount.List(kt, mainAccountReq)
		if err != nil {
			logs.Errorf("fail to find main account for list bill item, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if len(cloudIDBatch) != len(mainAccountList.Details) {
			return nil, errf.Newf(errf.InvalidParameter,
				"some account can not be found by main account cloud id")
		}
		for i := range mainAccountList.Details {
			mainAccountIds = append(mainAccountIds, mainAccountList.Details[i].ID)
		}
	}

	return mainAccountIds, nil
}
